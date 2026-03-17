#include "accessibility_darwin.h"
#include "bridge_shim_common.h"

#include <dlfcn.h>
#include <stdlib.h>
#include <string.h>

#import <AppKit/AppKit.h>
#import <ApplicationServices/ApplicationServices.h>
#import <Foundation/Foundation.h>

#include <unistd.h>

namespace {

const int gut_ax_focus_retry_count = 5;
const useconds_t gut_ax_focus_retry_delay_usec = 20000;

int gut_ax_get_system_value(CFStringRef attribute, AXUIElementRef *element);
int gut_ax_copy_focused_application(AXUIElementRef *application);
int gut_ax_copy_focused_window_from_application(AXUIElementRef application, AXUIElementRef *window);
int gut_ax_copy_attribute_element_with_retry(AXUIElementRef source, CFStringRef attribute, AXUIElementRef *element);

bool gut_ax_error_is_transient(AXError ax_error, CFTypeRef value) {
	return ax_error == kAXErrorNoValue || ax_error == kAXErrorCannotComplete || (ax_error == kAXErrorSuccess && value == NULL);
}

int gut_ax_copy_attribute_element_once(AXUIElementRef source, CFStringRef attribute, AXUIElementRef *element) {
	if (source == NULL || element == NULL) {
		return 2;
	}
	*element = NULL;
	CFTypeRef value = NULL;
	AXError ax_error = AXUIElementCopyAttributeValue(source, attribute, &value);
	if (gut_ax_error_is_transient(ax_error, value)) {
		if (value != NULL) {
			CFRelease(value);
		}
		return 4;
	}
	if (ax_error != kAXErrorSuccess || value == NULL || CFGetTypeID(value) != AXUIElementGetTypeID()) {
		if (value != NULL) {
			CFRelease(value);
		}
		return 2;
	}
	*element = (AXUIElementRef)value;
	return 0;
}

int gut_ax_copy_attribute_element_with_retry(AXUIElementRef source, CFStringRef attribute, AXUIElementRef *element) {
	int status = 4;
	for (int attempt = 0; attempt < gut_ax_focus_retry_count; attempt++) {
		status = gut_ax_copy_attribute_element_once(source, attribute, element);
		if (status != 4 || attempt + 1 == gut_ax_focus_retry_count) {
			return status;
		}
		usleep(gut_ax_focus_retry_delay_usec);
	}
	return status;
}

int gut_ax_copy_frontmost_application(AXUIElementRef *application) {
	if (application == NULL) {
		return 2;
	}
	*application = NULL;
	NSRunningApplication *frontmost = [[NSWorkspace sharedWorkspace] frontmostApplication];
	if (frontmost == nil) {
		return 4;
	}
	pid_t pid = frontmost.processIdentifier;
	if (pid <= 0) {
		return 4;
	}
	*application = AXUIElementCreateApplication(pid);
	if (*application == NULL) {
		return 2;
	}
	return 0;
}

int gut_ax_copy_focused_application(AXUIElementRef *application) {
	if (application == NULL) {
		return 2;
	}
	*application = NULL;
	AXUIElementRef system = AXUIElementCreateSystemWide();
	if (system != NULL) {
		int status = gut_ax_copy_attribute_element_with_retry(system, kAXFocusedApplicationAttribute, application);
		CFRelease(system);
		if (status == 0 && *application != NULL) {
			return 0;
		}
		if (status != 4) {
			return status;
		}
	}
	return gut_ax_copy_frontmost_application(application);
}

int gut_ax_copy_focused_window_from_application(AXUIElementRef application, AXUIElementRef *window) {
	if (application == NULL || window == NULL) {
		return 2;
	}
	*window = NULL;
	int status = gut_ax_copy_attribute_element_with_retry(application, kAXFocusedWindowAttribute, window);
	if (status == 0 || status != 4) {
		return status;
	}
	return gut_ax_copy_attribute_element_with_retry(application, kAXMainWindowAttribute, window);
}

char *gut_ax_copy_ns_string(NSString *value) {
	if (value == nil) {
		return NULL;
	}
	const char *utf8 = [value UTF8String];
	if (utf8 == NULL) {
		return NULL;
	}
	size_t length = strlen(utf8);
	char *copy = (char *)malloc(length + 1);
	if (copy == NULL) {
		return NULL;
	}
	memcpy(copy, utf8, length);
	copy[length] = '\0';
	return copy;
}

NSString *gut_ax_copy_description(CFTypeRef value) {
	if (value == NULL) {
		return nil;
	}
	if (CFGetTypeID(value) == CFStringGetTypeID()) {
		return [(__bridge NSString *)value copy];
	}
	CFStringRef description = CFCopyDescription(value);
	if (description == NULL) {
		return nil;
	}
	NSString *copy = [(__bridge NSString *)description copy];
	CFRelease(description);
	return copy;
}

char *gut_ax_copy_attribute_string(AXUIElementRef element, CFStringRef attribute) {
	CFTypeRef value = NULL;
	if (AXUIElementCopyAttributeValue(element, attribute, &value) != kAXErrorSuccess || value == NULL) {
		return NULL;
	}
	NSString *description = gut_ax_copy_description(value);
	CFRelease(value);
	char *copy = gut_ax_copy_ns_string(description);
	[description release];
	return copy;
}

bool gut_ax_copy_bool_attribute(AXUIElementRef element, CFStringRef attribute, int *result) {
	if (result == NULL) {
		return false;
	}
	CFTypeRef value = NULL;
	if (AXUIElementCopyAttributeValue(element, attribute, &value) != kAXErrorSuccess || value == NULL) {
		return false;
	}
	bool ok = CFGetTypeID(value) == CFBooleanGetTypeID();
	if (ok) {
		*result = CFBooleanGetValue((CFBooleanRef)value) ? 1 : 0;
	}
	CFRelease(value);
	return ok;
}

bool gut_ax_copy_rect(AXUIElementRef element, gut_rect *rect) {
	if (rect == NULL) {
		return false;
	}
	CFTypeRef position_value = NULL;
	CFTypeRef size_value = NULL;
	if (AXUIElementCopyAttributeValue(element, kAXPositionAttribute, &position_value) != kAXErrorSuccess || position_value == NULL) {
		return false;
	}
	if (AXUIElementCopyAttributeValue(element, kAXSizeAttribute, &size_value) != kAXErrorSuccess || size_value == NULL) {
		CFRelease(position_value);
		return false;
	}
	CGPoint position = CGPointZero;
	CGSize size = CGSizeZero;
	bool ok = CFGetTypeID(position_value) == AXValueGetTypeID() && AXValueGetType((AXValueRef)position_value) == kAXValueCGPointType && AXValueGetValue((AXValueRef)position_value, (AXValueType)kAXValueCGPointType, &position);
	ok = ok && CFGetTypeID(size_value) == AXValueGetTypeID() && AXValueGetType((AXValueRef)size_value) == kAXValueCGSizeType && AXValueGetValue((AXValueRef)size_value, (AXValueType)kAXValueCGSizeType, &size);
	if (ok) {
		rect->x = (int64_t)position.x;
		rect->y = (int64_t)position.y;
		rect->width = (int64_t)size.width;
		rect->height = (int64_t)size.height;
	}
	CFRelease(position_value);
	CFRelease(size_value);
	return ok;
}

int64_t gut_ax_window_number(AXUIElementRef element) {
	CFTypeRef value = NULL;
	if (AXUIElementCopyAttributeValue(element, CFSTR("AXWindowNumber"), &value) != kAXErrorSuccess || value == NULL) {
		return 0;
	}
	int64_t window_number = 0;
	if (CFGetTypeID(value) == CFNumberGetTypeID()) {
		CFNumberGetValue((CFNumberRef)value, kCFNumberSInt64Type, &window_number);
	}
	CFRelease(value);
	return window_number;
}

bool gut_ax_rect_matches_dictionary(gut_rect rect, CGRect bounds) {
	return rect.x == (int64_t)bounds.origin.x &&
		rect.y == (int64_t)bounds.origin.y &&
		rect.width == (int64_t)bounds.size.width &&
		rect.height == (int64_t)bounds.size.height;
}

bool gut_ax_point_is_on_screen(int64_t x, int64_t y) {
	uint32_t display_count = 0;
	if (CGGetActiveDisplayList(0, NULL, &display_count) != kCGErrorSuccess || display_count == 0) {
		return true;
	}
	CGDirectDisplayID *display_ids = (CGDirectDisplayID *)calloc(display_count, sizeof(CGDirectDisplayID));
	if (display_ids == NULL) {
		return true;
	}
	bool on_screen = false;
	if (CGGetActiveDisplayList(display_count, display_ids, &display_count) == kCGErrorSuccess) {
		CGPoint point = CGPointMake((CGFloat)x, (CGFloat)y);
		for (uint32_t index = 0; index < display_count; index++) {
			if (CGRectContainsPoint(CGDisplayBounds(display_ids[index]), point)) {
				on_screen = true;
				break;
			}
		}
	}
	free(display_ids);
	return on_screen;
}

int64_t gut_ax_find_window_number(pid_t pid, gut_rect rect, bool has_rect) {
	CFArrayRef windows = CGWindowListCopyWindowInfo(kCGWindowListOptionOnScreenOnly | kCGWindowListExcludeDesktopElements, kCGNullWindowID);
	if (windows == NULL) {
		return 0;
	}
	int64_t handle = 0;
	for (CFIndex index = 0; index < CFArrayGetCount(windows); index++) {
		CFDictionaryRef dictionary = (CFDictionaryRef)CFArrayGetValueAtIndex(windows, index);
		if (dictionary == NULL) {
			continue;
		}
		CFNumberRef owner_pid = (CFNumberRef)CFDictionaryGetValue(dictionary, kCGWindowOwnerPID);
		int64_t candidate_pid = 0;
		if (owner_pid == NULL || !CFNumberGetValue(owner_pid, kCFNumberSInt64Type, &candidate_pid) || candidate_pid != pid) {
			continue;
		}
		if (has_rect) {
			CFDictionaryRef bounds_dict = (CFDictionaryRef)CFDictionaryGetValue(dictionary, kCGWindowBounds);
			CGRect bounds = CGRectZero;
			if (bounds_dict == NULL || !CGRectMakeWithDictionaryRepresentation(bounds_dict, &bounds) || !gut_ax_rect_matches_dictionary(rect, bounds)) {
				continue;
			}
		}
		CFNumberRef window_number = (CFNumberRef)CFDictionaryGetValue(dictionary, kCGWindowNumber);
		if (window_number != NULL && CFNumberGetValue(window_number, kCFNumberSInt64Type, &handle) && handle != 0) {
			break;
		}
	}
	CFRelease(windows);
	return handle;
}

void gut_ax_fill_running_application(pid_t pid, gut_window_metadata *metadata) {
	if (metadata == NULL || pid <= 0) {
		return;
	}
	metadata->owner_pid = pid;
	NSRunningApplication *application = [NSRunningApplication runningApplicationWithProcessIdentifier:pid];
	if (application == nil) {
		return;
	}
	metadata->owner_name = gut_ax_copy_ns_string(application.localizedName);
	metadata->bundle_id = gut_ax_copy_ns_string(application.bundleIdentifier);
}

void gut_ax_fill_actions(AXUIElementRef element, gut_element_metadata *metadata) {
	if (metadata == NULL) {
		return;
	}
	CFArrayRef actions = NULL;
	if (AXUIElementCopyActionNames(element, &actions) != kAXErrorSuccess || actions == NULL) {
		return;
	}
	CFIndex count = CFArrayGetCount(actions);
	if (count <= 0) {
		CFRelease(actions);
		return;
	}
	metadata->actions.items = (char **)calloc((size_t)count, sizeof(char *));
	if (metadata->actions.items == NULL) {
		CFRelease(actions);
		return;
	}
	metadata->actions.length = count;
	for (CFIndex index = 0; index < count; index++) {
		CFStringRef action = (CFStringRef)CFArrayGetValueAtIndex(actions, index);
		if (action != NULL) {
			metadata->actions.items[index] = gut_ax_copy_ns_string((__bridge NSString *)action);
		}
	}
	CFRelease(actions);
}

void gut_ax_fill_element_metadata(AXUIElementRef element, gut_element_metadata *metadata) {
	metadata->role = gut_ax_copy_attribute_string(element, kAXRoleAttribute);
	metadata->subrole = gut_ax_copy_attribute_string(element, kAXSubroleAttribute);
	metadata->title = gut_ax_copy_attribute_string(element, kAXTitleAttribute);
	metadata->description = gut_ax_copy_attribute_string(element, kAXDescriptionAttribute);
	metadata->value = gut_ax_copy_attribute_string(element, kAXValueAttribute);
	gut_ax_copy_bool_attribute(element, kAXEnabledAttribute, &metadata->enabled);
	gut_ax_copy_bool_attribute(element, kAXFocusedAttribute, &metadata->focused);
	metadata->has_frame = gut_ax_copy_rect(element, &metadata->frame) ? 1 : 0;
	gut_ax_fill_actions(element, metadata);
}

NSString *gut_ax_action_token_to_ns_string(const char *action_token) {
	if (action_token == NULL || action_token[0] == '\0') {
		return nil;
	}
	return [[[NSString alloc] initWithUTF8String:action_token] autorelease];
}

bool gut_ax_action_supported(AXUIElementRef element, NSString *action_name) {
	if (element == NULL || action_name == nil) {
		return false;
	}
	CFArrayRef actions = NULL;
	if (AXUIElementCopyActionNames(element, &actions) != kAXErrorSuccess || actions == NULL) {
		return false;
	}
	bool supported = false;
	for (CFIndex index = 0; index < CFArrayGetCount(actions); index++) {
		CFStringRef candidate = (CFStringRef)CFArrayGetValueAtIndex(actions, index);
		if (candidate != NULL && CFStringCompare(candidate, (__bridge CFStringRef)action_name, 0) == kCFCompareEqualTo) {
			supported = true;
			break;
		}
	}
	CFRelease(actions);
	return supported;
}

int gut_ax_perform_action(AXUIElementRef element, const char *action_token) {
	NSString *action_name = gut_ax_action_token_to_ns_string(action_token);
	if (element == NULL || action_name == nil) {
		return 1;
	}
	if (!gut_ax_action_supported(element, action_name)) {
		return 4;
	}
	AXError ax_error = AXUIElementPerformAction(element, (__bridge CFStringRef)action_name);
	if (ax_error == kAXErrorSuccess) {
		return 0;
	}
	if (ax_error == kAXErrorActionUnsupported || ax_error == kAXErrorAttributeUnsupported || ax_error == kAXErrorNoValue || ax_error == kAXErrorCannotComplete) {
		return 4;
	}
	return 2;
}

int gut_ax_copy_focused_window(AXUIElementRef *window) {
	if (window == NULL) {
		return 2;
	}
	*window = NULL;
	AXUIElementRef application = NULL;
	int status = gut_ax_copy_focused_application(&application);
	if (status != 0 || application == NULL) {
		return status == 0 ? 4 : status;
	}
	status = gut_ax_copy_focused_window_from_application(application, window);
	CFRelease(application);
	return status;
}

int gut_ax_copy_focused_element(AXUIElementRef *element) {
	if (element == NULL) {
		return 2;
	}
	*element = NULL;
	int status = gut_ax_get_system_value(kAXFocusedUIElementAttribute, element);
	if (status != 0) {
		return status;
	}
	if (*element != NULL) {
		return 0;
	}

	AXUIElementRef application = NULL;
	status = gut_ax_copy_focused_application(&application);
	if (status != 0 || application == NULL) {
		return status == 0 ? 4 : status;
	}

	AXUIElementRef window = NULL;
	status = gut_ax_copy_focused_window_from_application(application, &window);
	CFRelease(application);
	if (status != 0 || window == NULL) {
		return status == 0 ? 4 : status;
	}

	status = gut_ax_copy_attribute_element_with_retry(window, kAXFocusedUIElementAttribute, element);
	CFRelease(window);
	return status;
}

int gut_ax_copy_element_at_point(int64_t x, int64_t y, AXUIElementRef *element) {
	if (element == NULL) {
		return 2;
	}
	*element = NULL;
	if (!gut_ax_point_is_on_screen(x, y)) {
		return gut_status_no_element_at_point;
	}
	AXUIElementRef system = AXUIElementCreateSystemWide();
	if (system == NULL) {
		return 2;
	}
	AXError ax_error = AXUIElementCopyElementAtPosition(system, (float)x, (float)y, element);
	CFRelease(system);
	if (ax_error == kAXErrorIllegalArgument) {
		if (*element != NULL) {
			CFRelease(*element);
			*element = NULL;
		}
		return gut_status_no_element_at_point;
	}
	if (ax_error == kAXErrorNoValue || (ax_error == kAXErrorSuccess && *element == NULL)) {
		if (*element != NULL) {
			CFRelease(*element);
			*element = NULL;
		}
		return 4;
	}
	if (ax_error != kAXErrorSuccess || *element == NULL) {
		if (*element != NULL) {
			CFRelease(*element);
			*element = NULL;
		}
		return 2;
	}
	return 0;
}

int gut_ax_get_system_value(CFStringRef attribute, AXUIElementRef *element) {
	if (element == NULL) {
		return 2;
	}
	*element = NULL;
	AXUIElementRef system = AXUIElementCreateSystemWide();
	if (system == NULL) {
		return 2;
	}
	int status = gut_ax_copy_attribute_element_with_retry(system, attribute, element);
	CFRelease(system);
	if (status == 4) {
		return 0;
	}
	return status;
}

void gut_ax_wait_for_focused_element(AXUIElementRef target) {
	if (target == NULL) {
		return;
	}
	for (int attempt = 0; attempt < gut_ax_focus_retry_count; attempt++) {
		AXUIElementRef focused = NULL;
		int status = gut_ax_copy_focused_element(&focused);
		if (status == 0 && focused != NULL) {
			bool matched = CFEqual(focused, target);
			CFRelease(focused);
			if (matched) {
				return;
			}
		}
		usleep(gut_ax_focus_retry_delay_usec);
	}
}

bool gut_ax_screen_recording_supported(void) {
	return dlsym(RTLD_DEFAULT, "CGPreflightScreenCaptureAccess") != NULL;
}

bool gut_ax_screen_recording_granted(void) {
	typedef bool (*gut_screen_capture_preflight_fn)(void);
	gut_screen_capture_preflight_fn preflight = (gut_screen_capture_preflight_fn)dlsym(RTLD_DEFAULT, "CGPreflightScreenCaptureAccess");
	if (preflight == NULL) {
		return false;
	}
	return preflight();
}

} 

extern "C" {

int gut_darwin_accessibility_permission_granted(void) {
	@autoreleasepool {
		NSDictionary *options = @{(__bridge id)kAXTrustedCheckOptionPrompt: @NO};
		return AXIsProcessTrustedWithOptions((__bridge CFDictionaryRef)options) ? 1 : 0;
	}
}

int gut_darwin_get_permission_snapshot(gut_permission_snapshot *snapshot) {
	if (snapshot == NULL) {
		return 2;
	}
	memset(snapshot, 0, sizeof(*snapshot));
	@autoreleasepool {
		snapshot->accessibility_supported = 1;
		snapshot->accessibility_granted = gut_darwin_accessibility_permission_granted();
		snapshot->screen_recording_supported = gut_ax_screen_recording_supported() ? 1 : 0;
		if (snapshot->screen_recording_supported) {
			snapshot->screen_recording_granted = gut_ax_screen_recording_granted() ? 1 : 0;
		}
	}
	return 0;
}

int gut_darwin_get_focused_window_metadata(gut_window_metadata *metadata) {
	if (metadata == NULL) {
		return 2;
	}
	memset(metadata, 0, sizeof(*metadata));
	if (!gut_darwin_accessibility_permission_granted()) {
		return 5;
	}
	@autoreleasepool {
		AXUIElementRef window = NULL;
		int status = gut_ax_copy_focused_window(&window);
		if (status != 0) {
			return status == 4 ? 0 : status;
		}
		metadata->title = gut_ax_copy_attribute_string(window, kAXTitleAttribute);
		metadata->role = gut_ax_copy_attribute_string(window, kAXRoleAttribute);
		metadata->subrole = gut_ax_copy_attribute_string(window, kAXSubroleAttribute);
		metadata->has_rect = gut_ax_copy_rect(window, &metadata->rect) ? 1 : 0;
		// The focused-window lookup is authoritative here: if kAXFocusedWindowAttribute
		// resolved to this AXWindow, report it as focused instead of depending on the
		// window object's kAXFocusedAttribute, which is not reliably populated.
		metadata->focused = 1;
		gut_ax_copy_bool_attribute(window, kAXMainAttribute, &metadata->main);
		gut_ax_copy_bool_attribute(window, kAXMinimizedAttribute, &metadata->minimized);
		pid_t pid = 0;
		AXUIElementGetPid(window, &pid);
		gut_ax_fill_running_application(pid, metadata);
		metadata->handle = gut_ax_window_number(window);
		if (metadata->handle == 0 && pid > 0) {
			metadata->handle = gut_ax_find_window_number(pid, metadata->rect, metadata->has_rect != 0);
		}
		CFRelease(window);
	}
	return 0;
}

int gut_darwin_raise_focused_window(void) {
	if (!gut_darwin_accessibility_permission_granted()) {
		return 5;
	}
	@autoreleasepool {
		AXUIElementRef window = NULL;
		int status = gut_ax_copy_focused_window(&window);
		if (status != 0) {
			return status;
		}
		status = gut_ax_perform_action(window, "AXRaise");
		CFRelease(window);
		return status;
	}
}

int gut_darwin_get_focused_element_metadata(gut_element_metadata *metadata) {
	if (metadata == NULL) {
		return 2;
	}
	memset(metadata, 0, sizeof(*metadata));
	if (!gut_darwin_accessibility_permission_granted()) {
		return 5;
	}
	@autoreleasepool {
		AXUIElementRef element = NULL;
		int status = gut_ax_copy_focused_element(&element);
		if (status != 0) {
			return status == 4 ? 0 : status;
		}
		gut_ax_fill_element_metadata(element, metadata);
		CFRelease(element);
	}
	return 0;
}

int gut_darwin_perform_focused_element_action(const char *action_token) {
	if (!gut_darwin_accessibility_permission_granted()) {
		return 5;
	}
	@autoreleasepool {
		AXUIElementRef element = NULL;
		int status = gut_ax_copy_focused_element(&element);
		if (status != 0) {
			return status;
		}
		status = gut_ax_perform_action(element, action_token);
		CFRelease(element);
		return status;
	}
}

int gut_darwin_get_element_metadata_at_point(int64_t x, int64_t y, gut_element_metadata *metadata) {
	if (metadata == NULL) {
		return 2;
	}
	memset(metadata, 0, sizeof(*metadata));
	if (!gut_darwin_accessibility_permission_granted()) {
		return 5;
	}
	@autoreleasepool {
		AXUIElementRef element = NULL;
		int status = gut_ax_copy_element_at_point(x, y, &element);
		if (status != 0) {
			return status == gut_status_capability_unavailable ? 0 : status;
		}
		gut_ax_fill_element_metadata(element, metadata);
		CFRelease(element);
	}
	return 0;
}

int gut_darwin_perform_element_action_at_point(int64_t x, int64_t y, const char *action_token) {
	if (!gut_darwin_accessibility_permission_granted()) {
		return 5;
	}
	@autoreleasepool {
		AXUIElementRef element = NULL;
		int status = gut_ax_copy_element_at_point(x, y, &element);
		if (status != 0) {
			return status;
		}
		status = gut_ax_perform_action(element, action_token);
		CFRelease(element);
		return status;
	}
}

int gut_darwin_focus_element_at_point(int64_t x, int64_t y) {
	if (!gut_darwin_accessibility_permission_granted()) {
		return 5;
	}
	@autoreleasepool {
		AXUIElementRef element = NULL;
		int status = gut_ax_copy_element_at_point(x, y, &element);
		if (status != 0) {
			return status;
		}
		Boolean settable = false;
		AXError ax_error = AXUIElementIsAttributeSettable(element, kAXFocusedAttribute, &settable);
		if (ax_error != kAXErrorSuccess || !settable) {
			CFRelease(element);
			return 4;
		}
		ax_error = AXUIElementSetAttributeValue(element, kAXFocusedAttribute, kCFBooleanTrue);
		if (ax_error == kAXErrorSuccess) {
			gut_ax_wait_for_focused_element(element);
			CFRelease(element);
			return 0;
		}
		CFRelease(element);
		if (ax_error == kAXErrorAttributeUnsupported || ax_error == kAXErrorNoValue || ax_error == kAXErrorCannotComplete) {
			return 4;
		}
		return 2;
	}
}

void gut_darwin_free_window_metadata(gut_window_metadata *metadata) {
	if (metadata == NULL) {
		return;
	}
	free(metadata->title);
	free(metadata->role);
	free(metadata->subrole);
	free(metadata->owner_name);
	free(metadata->bundle_id);
	memset(metadata, 0, sizeof(*metadata));
}

void gut_darwin_free_element_metadata(gut_element_metadata *metadata) {
	if (metadata == NULL) {
		return;
	}
	free(metadata->role);
	free(metadata->subrole);
	free(metadata->title);
	free(metadata->description);
	free(metadata->value);
	for (int64_t index = 0; index < metadata->actions.length; index++) {
		free(metadata->actions.items[index]);
	}
	free(metadata->actions.items);
	memset(metadata, 0, sizeof(*metadata));
}

}
