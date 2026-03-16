#include "appkit_main_thread_darwin.h"

#import <Cocoa/Cocoa.h>
#import <dispatch/dispatch.h>

void gut_run_on_main_thread_sync(gut_main_thread_callback callback, void *context) {
	if (callback == NULL) {
		return;
	}
	if ([NSThread isMainThread]) {
		callback(context);
		return;
	}
	dispatch_sync(dispatch_get_main_queue(), ^{
		callback(context);
	});
}

void gut_run_on_main_thread_async(gut_main_thread_callback callback, void *context) {
	if (callback == NULL) {
		return;
	}
	dispatch_async(dispatch_get_main_queue(), ^{
		callback(context);
	});
}
