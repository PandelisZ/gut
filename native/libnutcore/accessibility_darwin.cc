#include "accessibility_darwin.h"
#include "bridge_shim_common.h"

#include <algorithm>
#include <deque>
#include <vector>
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
int gut_ax_copy_focused_window(AXUIElementRef *window);
int gut_ax_copy_focused_window_from_application(AXUIElementRef application, AXUIElementRef *window);
int gut_ax_copy_window_for_handle(int64_t window_handle, AXUIElementRef *window, pid_t *owner_pid);
int gut_ax_copy_attribute_element_with_retry(AXUIElementRef source, CFStringRef attribute, AXUIElementRef *element);
char *gut_ax_copy_attribute_string(AXUIElementRef element, CFStringRef attribute);
NSString *gut_ax_action_token_to_ns_string(const char *action_token);
bool gut_ax_action_supported(AXUIElementRef element, NSString *action_name);
bool gut_ax_copy_rect(AXUIElementRef element, gut_rect *rect);
int64_t gut_ax_window_number(AXUIElementRef element);
int64_t gut_ax_find_window_number(pid_t pid, gut_rect rect, bool has_rect);

struct gut_ax_root_resolution {
	AXUIElementRef root;
	pid_t owner_pid;
	int64_t window_handle;
};

struct gut_ax_search_visit {
	AXUIElementRef element;
	std::vector<int64_t> path;
	int depth;
};

bool gut_ax_scope_equals(const char *scope, const char *expected) {
	return scope != NULL && strcmp(scope, expected) == 0;
}

bool gut_ax_matches_case_insensitive_contains(char *candidate, char *needle) {
	if (needle == NULL || needle[0] == '\0') {
		return true;
	}
	if (candidate == NULL || candidate[0] == '\0') {
		return false;
	}
		NSString *candidate_string = [NSString stringWithUTF8String:candidate];
		NSString *needle_string = [NSString stringWithUTF8String:needle];
		if (candidate_string == nil || needle_string == nil) {
			return false;
		}
		return [candidate_string rangeOfString:needle_string options:(NSCaseInsensitiveSearch | NSDiacriticInsensitiveSearch)].location != NSNotFound;
}

bool gut_ax_element_has_action(AXUIElementRef element, const char *action_token) {
	if (action_token == NULL || action_token[0] == '\0') {
		return true;
	}
	NSString *action_name = gut_ax_action_token_to_ns_string(action_token);
	return gut_ax_action_supported(element, action_name);
}

bool gut_ax_matches_bool_filter(int actual, int filter_state) {
	if (filter_state < 0) {
		return true;
	}
	return actual == filter_state;
}

bool gut_ax_matches_query(AXUIElementRef element, gut_element_metadata *metadata, const gut_ax_element_search_query *query) {
	if (element == NULL || metadata == NULL || query == NULL) {
		return false;
	}
	if (query->role != NULL && query->role[0] != '\0' && (metadata->role == NULL || strcmp(metadata->role, query->role) != 0)) {
		return false;
	}
	if (query->subrole != NULL && query->subrole[0] != '\0' && (metadata->subrole == NULL || strcmp(metadata->subrole, query->subrole) != 0)) {
		return false;
	}
	if (!gut_ax_matches_case_insensitive_contains(metadata->title, query->title_contains)) {
		return false;
	}
	if (!gut_ax_matches_case_insensitive_contains(metadata->value, query->value_contains)) {
		return false;
	}
	if (!gut_ax_matches_case_insensitive_contains(metadata->description, query->description_contains)) {
		return false;
	}
	if (!gut_ax_matches_bool_filter(metadata->enabled, query->enabled_state)) {
		return false;
	}
	if (!gut_ax_matches_bool_filter(metadata->focused, query->focused_state)) {
		return false;
	}
	if (!gut_ax_element_has_action(element, query->action)) {
		return false;
	}
	return true;
}

bool gut_ax_copy_children(AXUIElementRef element, std::vector<AXUIElementRef> *children) {
	if (element == NULL || children == NULL) {
		return false;
	}
	children->clear();
	CFTypeRef value = NULL;
	AXError ax_error = AXUIElementCopyAttributeValue(element, kAXChildrenAttribute, &value);
	if (ax_error != kAXErrorSuccess || value == NULL) {
		if (value != NULL) {
			CFRelease(value);
		}
		return false;
	}
	if (CFGetTypeID(value) != CFArrayGetTypeID()) {
		CFRelease(value);
		return false;
	}
	CFArrayRef array = (CFArrayRef)value;
	CFIndex count = CFArrayGetCount(array);
	children->reserve((size_t)count);
	for (CFIndex index = 0; index < count; index++) {
		CFTypeRef item = CFArrayGetValueAtIndex(array, index);
		if (item == NULL || CFGetTypeID(item) != AXUIElementGetTypeID()) {
			continue;
		}
		AXUIElementRef child = (AXUIElementRef)item;
		CFRetain(child);
		children->push_back(child);
	}
	CFRelease(value);
	return true;
}

void gut_ax_release_children(std::vector<AXUIElementRef> *children) {
	if (children == NULL) {
		return;
	}
	for (AXUIElementRef child : *children) {
		if (child != NULL) {
			CFRelease(child);
		}
	}
	children->clear();
}

void gut_ax_fill_action_point(gut_element_metadata *metadata, gut_point *point, int *has_action_point) {
	if (point == NULL || has_action_point == NULL) {
		return;
	}
	*has_action_point = 0;
	if (metadata == NULL || metadata->has_frame == 0 || metadata->frame.width <= 0 || metadata->frame.height <= 0) {
		return;
	}
	point->x = metadata->frame.x + metadata->frame.width / 2;
	point->y = metadata->frame.y + metadata->frame.height / 2;
	*has_action_point = 1;
}

void gut_ax_fill_element_ref(gut_ax_element_ref *ref, const char *scope, pid_t owner_pid, int64_t window_handle, const std::vector<int64_t> &path) {
	if (ref == NULL) {
		return;
	}
	memset(ref, 0, sizeof(*ref));
	ref->scope = gut_copy_c_string(scope == NULL ? "" : scope);
	ref->owner_pid = owner_pid;
	ref->window_handle = window_handle;
	if (!path.empty()) {
		ref->path.items = (int64_t *)calloc(path.size(), sizeof(int64_t));
		if (ref->path.items != NULL) {
			ref->path.length = (int64_t)path.size();
			for (size_t index = 0; index < path.size(); index++) {
				ref->path.items[index] = path[index];
			}
		}
	}
}

int gut_ax_resolve_root(const char *scope, int64_t requested_window_handle, gut_ax_root_resolution *resolution) {
	if (scope == NULL || scope[0] == '\0' || resolution == NULL) {
		return 1;
	}
	memset(resolution, 0, sizeof(*resolution));
	if (gut_ax_scope_equals(scope, "focused_window")) {
		int status = gut_ax_copy_focused_window(&resolution->root);
		if (status == 4) {
			return 0;
		}
		if (status != 0) {
			return status;
		}
		AXUIElementGetPid(resolution->root, &resolution->owner_pid);
		resolution->window_handle = gut_ax_window_number(resolution->root);
		if (resolution->window_handle == 0 && resolution->owner_pid > 0) {
			gut_rect rect = {};
			bool has_rect = gut_ax_copy_rect(resolution->root, &rect);
			resolution->window_handle = gut_ax_find_window_number(resolution->owner_pid, rect, has_rect);
		}
		return 0;
	}
	if (gut_ax_scope_equals(scope, "frontmost_application")) {
		int status = gut_ax_copy_focused_application(&resolution->root);
		if (status == 4) {
			return 0;
		}
		if (status != 0) {
			return status;
		}
		AXUIElementGetPid(resolution->root, &resolution->owner_pid);
		resolution->window_handle = 0;
		return 0;
	}
	if (gut_ax_scope_equals(scope, "window_handle")) {
		if (requested_window_handle <= 0) {
			return 1;
		}
		int status = gut_ax_copy_window_for_handle(requested_window_handle, &resolution->root, &resolution->owner_pid);
		if (status == 4) {
			return 0;
		}
		if (status != 0) {
			return status;
		}
		if (resolution->owner_pid <= 0) {
			AXUIElementGetPid(resolution->root, &resolution->owner_pid);
		}
		resolution->window_handle = requested_window_handle;
		return 0;
	}
	return 1;
}

int gut_ax_resolve_ref_element(const gut_ax_element_ref *ref, AXUIElementRef *element) {
	if (ref == NULL || element == NULL) {
		return 1;
	}
	*element = NULL;
	gut_ax_root_resolution resolution = {};
	int status = gut_ax_resolve_root(ref->scope, ref->window_handle, &resolution);
	if (status != 0) {
		return status;
	}
	if (resolution.root == NULL) {
		return 4;
	}
	if (ref->owner_pid > 0 && resolution.owner_pid > 0 && ref->owner_pid != resolution.owner_pid) {
		CFRelease(resolution.root);
		return 4;
	}
	if (ref->window_handle != 0 && resolution.window_handle != 0 && ref->window_handle != resolution.window_handle) {
		CFRelease(resolution.root);
		return 4;
	}
	AXUIElementRef current = resolution.root;
	for (int64_t depth = 0; depth < ref->path.length; depth++) {
		int64_t child_index = ref->path.items[depth];
		if (child_index < 0) {
			CFRelease(current);
			return 1;
		}
		std::vector<AXUIElementRef> children;
		gut_ax_copy_children(current, &children);
		if (child_index >= (int64_t)children.size()) {
			gut_ax_release_children(&children);
			CFRelease(current);
			return 4;
		}
		AXUIElementRef next = children[(size_t)child_index];
		CFRetain(next);
		gut_ax_release_children(&children);
		CFRelease(current);
		current = next;
	}
	*element = current;
	return 0;
}

int gut_ax_validate_search_query(const gut_ax_element_search_query *query) {
	if (query == NULL) {
		return 1;
	}
	if (query->scope == NULL || query->scope[0] == '\0') {
		return 1;
	}
	if (!gut_ax_scope_equals(query->scope, "focused_window") &&
		!gut_ax_scope_equals(query->scope, "frontmost_application") &&
		!gut_ax_scope_equals(query->scope, "window_handle")) {
		return 1;
	}
	if (gut_ax_scope_equals(query->scope, "window_handle") && query->window_handle <= 0) {
		return 1;
	}
	if (query->limit <= 0) {
		return 1;
	}
	if (query->max_depth < 0) {
		return 1;
	}
	if (query->enabled_state < -1 || query->enabled_state > 1) {
		return 1;
	}
	if (query->focused_state < -1 || query->focused_state > 1) {
		return 1;
	}
	return 0;
}

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

bool gut_ax_rect_equals(gut_rect left, gut_rect right) {
	return left.x == right.x &&
		left.y == right.y &&
		left.width == right.width &&
		left.height == right.height;
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

bool gut_ax_copy_window_lookup_info(int64_t window_handle, pid_t *owner_pid, gut_rect *rect, bool *has_rect, NSString **title) {
	if (owner_pid != NULL) {
		*owner_pid = 0;
	}
	if (has_rect != NULL) {
		*has_rect = false;
	}
	if (title != NULL) {
		*title = nil;
	}
	if (window_handle <= 0) {
		return false;
	}

	CFArrayRef windows = CGWindowListCopyWindowInfo(kCGWindowListOptionOnScreenOnly | kCGWindowListExcludeDesktopElements, kCGNullWindowID);
	if (windows == NULL) {
		return false;
	}

	bool found = false;
	for (CFIndex index = 0; index < CFArrayGetCount(windows); index++) {
		CFDictionaryRef dictionary = (CFDictionaryRef)CFArrayGetValueAtIndex(windows, index);
		if (dictionary == NULL) {
			continue;
		}

		CFNumberRef window_number = (CFNumberRef)CFDictionaryGetValue(dictionary, kCGWindowNumber);
		int64_t candidate_window_handle = 0;
		if (window_number == NULL || !CFNumberGetValue(window_number, kCFNumberSInt64Type, &candidate_window_handle) || candidate_window_handle != window_handle) {
			continue;
		}

		found = true;
		if (owner_pid != NULL) {
			CFNumberRef candidate_owner_pid = (CFNumberRef)CFDictionaryGetValue(dictionary, kCGWindowOwnerPID);
			int64_t value = 0;
			if (candidate_owner_pid != NULL && CFNumberGetValue(candidate_owner_pid, kCFNumberSInt64Type, &value)) {
				*owner_pid = (pid_t)value;
			}
		}
		if (rect != NULL && has_rect != NULL) {
			CFDictionaryRef bounds_dictionary = (CFDictionaryRef)CFDictionaryGetValue(dictionary, kCGWindowBounds);
			CGRect bounds = CGRectZero;
			if (bounds_dictionary != NULL && CGRectMakeWithDictionaryRepresentation(bounds_dictionary, &bounds)) {
				rect->x = (int64_t)bounds.origin.x;
				rect->y = (int64_t)bounds.origin.y;
				rect->width = (int64_t)bounds.size.width;
				rect->height = (int64_t)bounds.size.height;
				*has_rect = true;
			}
		}
		if (title != NULL) {
			CFStringRef window_title = (CFStringRef)CFDictionaryGetValue(dictionary, kCGWindowName);
			if (window_title != NULL && CFGetTypeID(window_title) == CFStringGetTypeID()) {
				*title = [(__bridge NSString *)window_title copy];
			}
		}
		break;
	}

	CFRelease(windows);
	return found;
}

bool gut_ax_title_equals(NSString *expected, const char *candidate_utf8) {
	if (expected == nil || candidate_utf8 == NULL || candidate_utf8[0] == '\0') {
		return false;
	}
	NSString *candidate = [NSString stringWithUTF8String:candidate_utf8];
	if (candidate == nil) {
		return false;
	}
	return [expected isEqualToString:candidate];
}

bool gut_ax_window_matches_lookup(AXUIElementRef window, int64_t target_window_handle, gut_rect target_rect, bool has_target_rect, NSString *target_title) {
	if (window == NULL) {
		return false;
	}

	int64_t candidate_window_handle = gut_ax_window_number(window);
	if (candidate_window_handle != 0 && candidate_window_handle == target_window_handle) {
		return true;
	}

	bool title_match = false;
	if (target_title != nil) {
		char *candidate_title = gut_ax_copy_attribute_string(window, kAXTitleAttribute);
		title_match = gut_ax_title_equals(target_title, candidate_title);
		free(candidate_title);
	}

	bool rect_match = false;
	if (has_target_rect) {
		gut_rect candidate_rect = {};
		rect_match = gut_ax_copy_rect(window, &candidate_rect) && gut_ax_rect_equals(target_rect, candidate_rect);
	}

	if (rect_match) {
		return target_title == nil || title_match;
	}
	return target_title != nil && title_match && !has_target_rect;
}

int gut_ax_copy_window_for_handle(int64_t window_handle, AXUIElementRef *window, pid_t *owner_pid) {
	if (window == NULL) {
		return 2;
	}
	*window = NULL;
	if (owner_pid != NULL) {
		*owner_pid = 0;
	}

	pid_t pid = 0;
	gut_rect rect = {};
	bool has_rect = false;
	NSString *title = nil;
	if (!gut_ax_copy_window_lookup_info(window_handle, &pid, &rect, &has_rect, &title) || pid <= 0) {
		[title release];
		return 4;
	}

	AXUIElementRef application = AXUIElementCreateApplication(pid);
	if (application == NULL) {
		[title release];
		return 2;
	}

	CFTypeRef window_value = NULL;
	AXError ax_error = AXUIElementCopyAttributeValue(application, kAXWindowsAttribute, &window_value);
	if (ax_error != kAXErrorSuccess || window_value == NULL || CFGetTypeID(window_value) != CFArrayGetTypeID()) {
		if (window_value != NULL) {
			CFRelease(window_value);
		}
		CFRelease(application);
		[title release];
		return ax_error == kAXErrorNoValue ? 4 : 2;
	}

	CFArrayRef windows = (CFArrayRef)window_value;
	for (CFIndex index = 0; index < CFArrayGetCount(windows); index++) {
		CFTypeRef item = CFArrayGetValueAtIndex(windows, index);
		if (item == NULL || CFGetTypeID(item) != AXUIElementGetTypeID()) {
			continue;
		}
		AXUIElementRef candidate = (AXUIElementRef)item;
		if (!gut_ax_window_matches_lookup(candidate, window_handle, rect, has_rect, title)) {
			continue;
		}
		CFRetain(candidate);
		*window = candidate;
		if (owner_pid != NULL) {
			*owner_pid = pid;
		}
		break;
	}

	CFRelease(windows);
	CFRelease(application);
	[title release];

	return *window == NULL ? 4 : 0;
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

int gut_darwin_search_ax_elements(const gut_ax_element_search_query *query, gut_ax_element_match_list *matches) {
	if (matches == NULL) {
		return 2;
	}
	memset(matches, 0, sizeof(*matches));
	if (!gut_darwin_accessibility_permission_granted()) {
		return 5;
	}
	if (gut_ax_validate_search_query(query) != 0) {
		return 1;
	}
	@autoreleasepool {
		gut_ax_root_resolution resolution = {};
		int status = gut_ax_resolve_root(query->scope, query->window_handle, &resolution);
		if (status != 0) {
			return status;
		}
		if (resolution.root == NULL) {
			return 0;
		}

		std::deque<gut_ax_search_visit> queue;
		queue.push_back(gut_ax_search_visit{resolution.root, std::vector<int64_t>(), 0});
		std::vector<gut_ax_element_match> collected;
		collected.reserve((size_t)std::min<int64_t>(query->limit, 32));

		while (!queue.empty() && (int64_t)collected.size() < query->limit) {
			gut_ax_search_visit visit = queue.front();
			queue.pop_front();

			gut_element_metadata metadata = {};
			gut_ax_fill_element_metadata(visit.element, &metadata);
			if (gut_ax_matches_query(visit.element, &metadata, query)) {
				gut_ax_element_match match = {};
				gut_ax_fill_element_ref(&match.ref, query->scope, resolution.owner_pid, resolution.window_handle, visit.path);
				match.metadata = metadata;
				match.depth = visit.depth;
				gut_ax_fill_action_point(&match.metadata, &match.action_point, &match.has_action_point);
				collected.push_back(match);
			} else {
				gut_darwin_free_element_metadata(&metadata);
			}

			if (visit.depth < query->max_depth) {
				std::vector<AXUIElementRef> children;
				gut_ax_copy_children(visit.element, &children);
				for (size_t index = 0; index < children.size(); index++) {
					AXUIElementRef child = children[index];
					std::vector<int64_t> child_path = visit.path;
					child_path.push_back((int64_t)index);
					queue.push_back(gut_ax_search_visit{child, child_path, visit.depth + 1});
				}
				children.clear();
			}

			CFRelease(visit.element);
		}

		while (!queue.empty()) {
			gut_ax_search_visit visit = queue.front();
			queue.pop_front();
			CFRelease(visit.element);
		}

		if (!collected.empty()) {
			matches->items = (gut_ax_element_match *)calloc(collected.size(), sizeof(gut_ax_element_match));
			if (matches->items == NULL) {
				for (size_t index = 0; index < collected.size(); index++) {
					gut_darwin_free_ax_element_match(&collected[index]);
				}
				return 2;
			}
			matches->length = (int64_t)collected.size();
			for (size_t index = 0; index < collected.size(); index++) {
				matches->items[index] = collected[index];
			}
		}
		return 0;
	}
}

int gut_darwin_focus_ax_element(const gut_ax_element_ref *ref) {
	if (!gut_darwin_accessibility_permission_granted()) {
		return 5;
	}
	if (ref == NULL || ref->scope == NULL || ref->scope[0] == '\0') {
		return 1;
	}
	@autoreleasepool {
		AXUIElementRef element = NULL;
		int status = gut_ax_resolve_ref_element(ref, &element);
		if (status != 0) {
			return status == 4 ? 4 : status;
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

int gut_darwin_perform_ax_element_action(const gut_ax_element_ref *ref, const char *action_token) {
	if (!gut_darwin_accessibility_permission_granted()) {
		return 5;
	}
	if (ref == NULL || ref->scope == NULL || ref->scope[0] == '\0' || action_token == NULL || action_token[0] == '\0') {
		return 1;
	}
	@autoreleasepool {
		AXUIElementRef element = NULL;
		int status = gut_ax_resolve_ref_element(ref, &element);
		if (status != 0) {
			return status == 4 ? 4 : status;
		}
		status = gut_ax_perform_action(element, action_token);
		CFRelease(element);
		return status;
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

void gut_darwin_free_ax_element_ref(gut_ax_element_ref *ref) {
	if (ref == NULL) {
		return;
	}
	free(ref->scope);
	free(ref->path.items);
	memset(ref, 0, sizeof(*ref));
}

void gut_darwin_free_ax_element_match(gut_ax_element_match *match) {
	if (match == NULL) {
		return;
	}
	gut_darwin_free_ax_element_ref(&match->ref);
	gut_darwin_free_element_metadata(&match->metadata);
	memset(match, 0, sizeof(*match));
}

void gut_darwin_free_ax_element_match_list(gut_ax_element_match_list *matches) {
	if (matches == NULL) {
		return;
	}
	for (int64_t index = 0; index < matches->length; index++) {
		gut_darwin_free_ax_element_match(&matches->items[index]);
	}
	free(matches->items);
	memset(matches, 0, sizeof(*matches));
}

}
