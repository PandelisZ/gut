#include "appkit_main_thread_darwin.h"
#include "../../../libnut-core/src/highlightwindow.h"

#import <Cocoa/Cocoa.h>
#import <dispatch/dispatch.h>

typedef struct gut_highlight_context {
	int32_t x;
	int32_t y;
	int32_t width;
	int32_t height;
	long duration;
	float opacity;
} gut_highlight_context;

static void gut_close_highlight_window(void *context) {
	NSWindow *window = (NSWindow *)context;
	if (window == nil) {
		return;
	}
	[window orderOut:nil];
	[window close];
	[window release];
}

static void gut_show_highlight_window(void *context) {
	gut_highlight_context *highlight = (gut_highlight_context *)context;
	if (highlight == NULL) {
		return;
	}

	@autoreleasepool {
		[NSApplication sharedApplication];

		NSScreen *screen = [NSScreen mainScreen];
		if (screen == nil) {
			return;
		}

		NSRect screenFrame = [screen frame];
		NSRect frame = NSMakeRect(
			highlight->x,
			screenFrame.size.height - highlight->y - highlight->height,
			highlight->width,
			highlight->height
		);
		NSWindow *window = [[NSWindow alloc] initWithContentRect:frame
			styleMask:NSWindowStyleMaskBorderless
			backing:NSBackingStoreBuffered
			defer:NO];
		if (window == nil) {
			return;
		}

		[window setReleasedWhenClosed:NO];
		[window setOpaque:NO];
		[window setBackgroundColor:[NSColor redColor]];
		[window setAlphaValue:highlight->opacity];
		[window setIgnoresMouseEvents:YES];
		[window setLevel:NSStatusWindowLevel];
		[window setCollectionBehavior:NSWindowCollectionBehaviorCanJoinAllSpaces | NSWindowCollectionBehaviorFullScreenAuxiliary];
		[window orderFrontRegardless];

		dispatch_after(
			dispatch_time(DISPATCH_TIME_NOW, (int64_t)(highlight->duration * NSEC_PER_MSEC)),
			dispatch_get_main_queue(),
			^{
				gut_close_highlight_window(window);
			}
		);
	}
}

static void gut_show_highlight_window_without_app_loop(gut_highlight_context *highlight) {
	if (highlight == NULL) {
		return;
	}

	@autoreleasepool {
		[NSApplication sharedApplication];

		NSScreen *screen = [NSScreen mainScreen];
		if (screen == nil) {
			return;
		}

		NSRect screenFrame = [screen frame];
		NSRect frame = NSMakeRect(
			highlight->x,
			screenFrame.size.height - highlight->y - highlight->height,
			highlight->width,
			highlight->height
		);
		NSWindow *window = [[NSWindow alloc] initWithContentRect:frame
			styleMask:NSWindowStyleMaskBorderless
			backing:NSBackingStoreBuffered
			defer:NO];
		if (window == nil) {
			return;
		}

		[window setReleasedWhenClosed:NO];
		[window setOpaque:NO];
		[window setBackgroundColor:[NSColor redColor]];
		[window setAlphaValue:highlight->opacity];
		[window setIgnoresMouseEvents:YES];
		[window setLevel:NSStatusWindowLevel];
		[window setCollectionBehavior:NSWindowCollectionBehaviorCanJoinAllSpaces | NSWindowCollectionBehaviorFullScreenAuxiliary];
		[window orderFrontRegardless];

		if (highlight->duration > 0) {
			[NSThread sleepForTimeInterval:(NSTimeInterval)highlight->duration / 1000.0];
		}

		gut_close_highlight_window(window);
	}
}

static void gut_show_highlight_window_without_app_loop_async(void *context) {
	gut_highlight_context *highlight = (gut_highlight_context *)context;
	if (highlight == NULL) {
		return;
	}
	gut_show_highlight_window_without_app_loop(highlight);
	free(highlight);
}

void showHighlightWindow(int32_t x, int32_t y, int32_t width, int32_t height, long duration, float opacity) {
	gut_highlight_context context = {
		.x = x,
		.y = y,
		.width = width,
		.height = height,
		.duration = duration,
		.opacity = opacity,
	};

	if (NSApp == nil || ![NSApp isRunning]) {
		gut_highlight_context *async_context = (gut_highlight_context *)malloc(sizeof(gut_highlight_context));
		if (async_context == NULL) {
			return;
		}
		*async_context = context;
		gut_run_on_main_thread_async(gut_show_highlight_window_without_app_loop_async, async_context);
		return;
	}

	gut_run_on_main_thread_sync(gut_show_highlight_window, &context);
}
