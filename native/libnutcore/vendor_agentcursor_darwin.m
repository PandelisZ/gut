#include "agent_cursor_darwin.h"
#include "appkit_main_thread_darwin.h"

#import <Cocoa/Cocoa.h>
#include <stdlib.h>
#include <string.h>

static const CGFloat gut_agent_cursor_window_size = 156.0;
static const CGFloat gut_agent_cursor_hotspot_x = 28.0;
static const CGFloat gut_agent_cursor_hotspot_top = 18.0;
static const NSTimeInterval gut_agent_cursor_hide_delay = 1.2;
static const NSTimeInterval gut_agent_cursor_fade_duration = 0.18;
static const NSTimeInterval gut_agent_cursor_click_pulse_duration = 0.24;
static const NSTimeInterval gut_agent_cursor_scroll_pulse_duration = 0.34;

typedef struct gut_agent_cursor_context {
	int64_t position_x;
	int64_t position_y;
	int has_target;
	int64_t target_x;
	int64_t target_y;
	int pressed;
	int64_t duration_ms;
	char kind[32];
	char button[16];
	char direction[16];
} gut_agent_cursor_context;

static NSRect gut_agent_cursor_desktop_frame(void) {
	NSArray<NSScreen *> *screens = [NSScreen screens];
	if (screens.count == 0) {
		NSScreen *mainScreen = [NSScreen mainScreen];
		return mainScreen == nil ? NSMakeRect(0, 0, 0, 0) : mainScreen.frame;
	}

	NSRect unionFrame = ((NSScreen *)screens[0]).frame;
	for (NSUInteger index = 1; index < screens.count; index++) {
		unionFrame = NSUnionRect(unionFrame, ((NSScreen *)screens[index]).frame);
	}
	return unionFrame;
}

static NSString *gut_agent_cursor_string(const char *value) {
	if (value == NULL || value[0] == '\0') {
		return @"";
	}
	return [NSString stringWithUTF8String:value] ?: @"";
}

static void gut_agent_cursor_copy_token(char *destination, size_t destination_size, const char *source) {
	if (destination == NULL || destination_size == 0) {
		return;
	}
	destination[0] = '\0';
	if (source == NULL) {
		return;
	}
	strncpy(destination, source, destination_size - 1);
	destination[destination_size - 1] = '\0';
}

static CGFloat gut_agent_cursor_clamp(CGFloat value, CGFloat min_value, CGFloat max_value) {
	if (value < min_value) {
		return min_value;
	}
	if (value > max_value) {
		return max_value;
	}
	return value;
}

static CGFloat gut_agent_cursor_ease_out_cubic(CGFloat progress) {
	CGFloat clamped = gut_agent_cursor_clamp(progress, 0.0, 1.0);
	CGFloat inverse = 1.0 - clamped;
	return 1.0 - inverse * inverse * inverse;
}

@class GutAgentCursorController;

@interface GutAgentCursorView : NSView
@property(nonatomic, assign) GutAgentCursorController *controller;
@end

@interface GutAgentCursorController : NSObject
@property(nonatomic, retain) NSWindow *window;
@property(nonatomic, retain) GutAgentCursorView *view;
@property(nonatomic, retain) NSTimer *timer;
@property(nonatomic, copy) NSString *scrollDirection;
@property(nonatomic, assign) NSRect desktopFrame;
@property(nonatomic, assign) NSPoint currentPoint;
@property(nonatomic, assign) NSPoint animationFrom;
@property(nonatomic, assign) NSPoint animationTo;
@property(nonatomic, assign) NSTimeInterval animationStart;
@property(nonatomic, assign) NSTimeInterval animationDuration;
@property(nonatomic, assign) NSTimeInterval hideDeadline;
@property(nonatomic, assign) NSTimeInterval fadeStart;
@property(nonatomic, assign) NSTimeInterval clickPulseStartA;
@property(nonatomic, assign) NSTimeInterval clickPulseStartB;
@property(nonatomic, assign) NSTimeInterval scrollPulseStart;
@property(nonatomic, assign) BOOL hasPoint;
@property(nonatomic, assign) BOOL animating;
@property(nonatomic, assign) BOOL fading;
@property(nonatomic, assign) BOOL pressed;

+ (instancetype)sharedController;
- (void)handleEventWithKind:(NSString *)kind
                   position:(NSPoint)position
                  hasTarget:(BOOL)hasTarget
                     target:(NSPoint)target
                     button:(NSString *)button
                  direction:(NSString *)direction
                    pressed:(BOOL)pressed
                   duration:(NSTimeInterval)duration;
- (void)hideCursor;
- (CGFloat)clickPulseAtTime:(NSTimeInterval)now;
- (CGFloat)scrollPulseAtTime:(NSTimeInterval)now;
@end

@implementation GutAgentCursorView

- (BOOL)isOpaque {
	return NO;
}

- (void)drawRect:(NSRect)dirtyRect {
	(void)dirtyRect;
	[[NSColor clearColor] setFill];
	NSRectFill(self.bounds);

	GutAgentCursorController *controller = self.controller;
	if (controller == nil || !controller.hasPoint) {
		return;
	}

	NSTimeInterval now = [NSDate timeIntervalSinceReferenceDate];
	CGFloat clickPulse = [controller clickPulseAtTime:now];
	CGFloat scrollPulse = [controller scrollPulseAtTime:now];
	CGFloat tipY = gut_agent_cursor_window_size - gut_agent_cursor_hotspot_top;
	NSPoint tip = NSMakePoint(gut_agent_cursor_hotspot_x, tipY);
	CGFloat scale = controller.pressed ? 0.94 : 1.0;
	scale += clickPulse * 0.08;
	CGFloat pressOffset = controller.pressed ? -3.0 : 0.0;

	NSRect glowRect = NSMakeRect(tip.x - 34.0 - clickPulse * 10.0, tip.y - 36.0 - clickPulse * 10.0, 68.0 + clickPulse * 20.0, 68.0 + clickPulse * 20.0);
	NSBezierPath *glow = [NSBezierPath bezierPathWithOvalInRect:glowRect];
	[[NSColor colorWithCalibratedRed:1.0 green:0.45 blue:0.79 alpha:0.18 + clickPulse * 0.08] setFill];
	[glow fill];

	if (clickPulse > 0.01) {
		CGFloat radius = 26.0 + clickPulse * 22.0;
		NSRect pulseRect = NSMakeRect(tip.x - radius, tip.y - radius, radius * 2.0, radius * 2.0);
		NSBezierPath *pulse = [NSBezierPath bezierPathWithOvalInRect:pulseRect];
		pulse.lineWidth = 2.0;
		[[NSColor colorWithCalibratedRed:1.0 green:0.76 blue:0.92 alpha:(1.0 - clickPulse) * 0.5] setStroke];
		[pulse stroke];
	}

	if (scrollPulse > 0.01) {
		CGFloat radius = 20.0 + scrollPulse * 26.0;
		NSRect rippleRect = NSMakeRect(tip.x - radius, tip.y - radius, radius * 2.0, radius * 2.0);
		NSBezierPath *ripple = [NSBezierPath bezierPathWithOvalInRect:rippleRect];
		ripple.lineWidth = 1.6;
		[[NSColor colorWithCalibratedRed:1.0 green:0.58 blue:0.85 alpha:(1.0 - scrollPulse) * 0.45] setStroke];
		[ripple stroke];
	}

	[NSGraphicsContext saveGraphicsState];
	NSAffineTransform *transform = [NSAffineTransform transform];
	[transform translateXBy:tip.x yBy:tip.y + pressOffset];
	[transform scaleBy:scale];
	[transform concat];

	NSShadow *shadow = [[NSShadow alloc] init];
	shadow.shadowOffset = NSMakeSize(0.0, -4.0);
	shadow.shadowBlurRadius = 18.0;
	shadow.shadowColor = [NSColor colorWithCalibratedRed:0.54 green:0.0 blue:0.3 alpha:0.3];
	[shadow set];
	[shadow release];

	NSBezierPath *cursor = [NSBezierPath bezierPath];
	[cursor moveToPoint:NSMakePoint(0.0, 0.0)];
	[cursor lineToPoint:NSMakePoint(0.0, 84.0)];
	[cursor lineToPoint:NSMakePoint(22.0, 62.0)];
	[cursor lineToPoint:NSMakePoint(38.0, 102.0)];
	[cursor lineToPoint:NSMakePoint(53.0, 95.0)];
	[cursor lineToPoint:NSMakePoint(37.0, 58.0)];
	[cursor lineToPoint:NSMakePoint(69.0, 54.0)];
	[cursor closePath];
	cursor.lineJoinStyle = NSRoundLineJoinStyle;

	[NSGraphicsContext saveGraphicsState];
	[cursor addClip];
	NSGradient *fill = [[NSGradient alloc] initWithColorsAndLocations:
		[NSColor colorWithCalibratedRed:1.0 green:0.62 blue:0.89 alpha:1.0], 0.0,
		[NSColor colorWithCalibratedRed:1.0 green:0.32 blue:0.72 alpha:1.0], 0.55,
		[NSColor colorWithCalibratedRed:0.92 green:0.1 blue:0.55 alpha:1.0], 1.0,
		nil];
	[fill drawFromPoint:NSMakePoint(10.0, 102.0) toPoint:NSMakePoint(56.0, 0.0) options:0];
	[fill release];
	[NSGraphicsContext restoreGraphicsState];

	cursor.lineWidth = 2.4;
	[[NSColor colorWithCalibratedRed:1.0 green:0.94 blue:0.98 alpha:0.85] setStroke];
	[cursor stroke];

	NSBezierPath *highlight = [NSBezierPath bezierPath];
	[highlight moveToPoint:NSMakePoint(6.0, 15.0)];
	[highlight lineToPoint:NSMakePoint(6.0, 71.0)];
	[highlight lineToPoint:NSMakePoint(20.0, 58.0)];
	highlight.lineWidth = 3.0;
	highlight.lineCapStyle = NSRoundLineCapStyle;
	[[NSColor colorWithCalibratedRed:1.0 green:0.98 blue:1.0 alpha:0.46] setStroke];
	[highlight stroke];

	[NSGraphicsContext restoreGraphicsState];
}

@end

@implementation GutAgentCursorController

+ (instancetype)sharedController {
	static GutAgentCursorController *sharedController = nil;
	if (sharedController == nil) {
		sharedController = [[GutAgentCursorController alloc] init];
		sharedController.scrollDirection = @"";
	}
	return sharedController;
}

- (void)dealloc {
	[_timer invalidate];
	[_timer release];
	[_view release];
	[_window release];
	[_scrollDirection release];
	[super dealloc];
}

- (void)ensureWindow {
	NSRect desktopFrame = gut_agent_cursor_desktop_frame();
	if (NSEqualRects(desktopFrame, NSZeroRect)) {
		return;
	}

	if (_window != nil && NSEqualRects(_desktopFrame, desktopFrame)) {
		return;
	}

	_desktopFrame = desktopFrame;
	if (_window != nil) {
		[_window orderOut:nil];
		[_window release];
		_window = nil;
		[_view release];
		_view = nil;
	}

	NSWindow *window = [[NSWindow alloc] initWithContentRect:NSMakeRect(desktopFrame.origin.x, desktopFrame.origin.y, gut_agent_cursor_window_size, gut_agent_cursor_window_size)
		styleMask:NSWindowStyleMaskBorderless
		backing:NSBackingStoreBuffered
		defer:NO];
	if (window == nil) {
		return;
	}

	window.releasedWhenClosed = NO;
	window.opaque = NO;
	window.backgroundColor = [NSColor clearColor];
	window.hasShadow = NO;
	window.ignoresMouseEvents = YES;
	window.level = NSStatusWindowLevel;
	window.collectionBehavior = NSWindowCollectionBehaviorCanJoinAllSpaces | NSWindowCollectionBehaviorFullScreenAuxiliary | NSWindowCollectionBehaviorStationary;
	window.alphaValue = 0.0;

	GutAgentCursorView *view = [[GutAgentCursorView alloc] initWithFrame:NSMakeRect(0.0, 0.0, gut_agent_cursor_window_size, gut_agent_cursor_window_size)];
	view.controller = self;
	window.contentView = view;

	_window = window;
	_view = view;
}

- (void)ensureTimer {
	if (_timer != nil) {
		return;
	}
	_timer = [[NSTimer timerWithTimeInterval:(1.0 / 60.0) target:self selector:@selector(tick:) userInfo:nil repeats:YES] retain];
	[[NSRunLoop mainRunLoop] addTimer:_timer forMode:NSRunLoopCommonModes];
}

- (void)invalidateTimer {
	[_timer invalidate];
	[_timer release];
	_timer = nil;
}

- (void)hideCursor {
	[self invalidateTimer];
	_hasPoint = NO;
	_animating = NO;
	_fading = NO;
	_pressed = NO;
	_hideDeadline = 0.0;
	_fadeStart = 0.0;
	_clickPulseStartA = 0.0;
	_clickPulseStartB = 0.0;
	_scrollPulseStart = 0.0;
	if (_window != nil) {
		[_window orderOut:nil];
		[_window setAlphaValue:0.0];
	}
}

- (void)tick:(NSTimer *)timer {
	(void)timer;
	NSTimeInterval now = [NSDate timeIntervalSinceReferenceDate];
	[self advanceAnimationsToTime:now];

	if (!_fading && _hideDeadline > 0.0 && now >= _hideDeadline) {
		_fading = YES;
		_fadeStart = now;
	}

	if (_fading) {
		CGFloat fadeProgress = gut_agent_cursor_clamp((now - _fadeStart) / gut_agent_cursor_fade_duration, 0.0, 1.0);
		if (fadeProgress >= 1.0) {
			[self hideCursor];
			return;
		}
		[_window setAlphaValue:(1.0 - fadeProgress)];
	} else if (_window != nil) {
		[_window setAlphaValue:1.0];
	}

	[self updateWindowFrame];
	[_view setNeedsDisplay:YES];
}

- (void)handleEventWithKind:(NSString *)kind
                   position:(NSPoint)position
                  hasTarget:(BOOL)hasTarget
                     target:(NSPoint)target
                     button:(NSString *)button
                  direction:(NSString *)direction
                    pressed:(BOOL)pressed
                   duration:(NSTimeInterval)duration {
	(void)button;
	(void)pressed;

	[NSApplication sharedApplication];
	[self ensureWindow];
	if (_window == nil) {
		return;
	}

	if ([kind isEqualToString:@"hide"]) {
		[self hideCursor];
		return;
	}

	NSTimeInterval now = [NSDate timeIntervalSinceReferenceDate];
	[self advanceAnimationsToTime:now];

	if (!_hasPoint) {
		_currentPoint = position;
		_hasPoint = YES;
	}

	if ([kind isEqualToString:@"move"]) {
		NSPoint from = position;
		_currentPoint = from;
		if (hasTarget && duration > 0.0) {
			_animationFrom = from;
			_animationTo = target;
			_animationStart = now;
			_animationDuration = duration;
			_animating = YES;
		} else {
			_currentPoint = hasTarget ? target : position;
			_animating = NO;
		}
		_pressed = NO;
	} else if ([kind isEqualToString:@"drag_start"]) {
		NSPoint from = position;
		_currentPoint = from;
		_pressed = YES;
		if (hasTarget && duration > 0.0) {
			_animationFrom = from;
			_animationTo = target;
			_animationStart = now;
			_animationDuration = duration;
			_animating = YES;
		} else {
			_currentPoint = hasTarget ? target : from;
			_animating = NO;
		}
	} else {
		_animating = NO;
		_currentPoint = hasTarget ? target : position;

		if ([kind isEqualToString:@"drag_end"] || [kind isEqualToString:@"mouse_up"]) {
			_pressed = NO;
		} else if ([kind isEqualToString:@"mouse_down"]) {
			_pressed = YES;
		}

		if ([kind isEqualToString:@"click"]) {
			_clickPulseStartA = now;
		} else if ([kind isEqualToString:@"double_click"]) {
			_clickPulseStartA = now;
			_clickPulseStartB = now + 0.12;
		} else if ([kind isEqualToString:@"scroll"]) {
			_scrollPulseStart = now;
			self.scrollDirection = direction ?: @"";
		} else if ([kind isEqualToString:@"drag_end"]) {
			_clickPulseStartA = now;
		}
	}

	_hideDeadline = now + gut_agent_cursor_hide_delay;
	_fading = NO;
	[_window orderFrontRegardless];
	[_window setAlphaValue:1.0];
	[self updateWindowFrame];
	[_view setNeedsDisplay:YES];
	[self ensureTimer];
}

- (void)advanceAnimationsToTime:(NSTimeInterval)now {
	if (!_animating) {
		return;
	}

	if (_animationDuration <= 0.0) {
		_currentPoint = _animationTo;
		_animating = NO;
		return;
	}

	CGFloat progress = gut_agent_cursor_clamp((now - _animationStart) / _animationDuration, 0.0, 1.0);
	CGFloat eased = gut_agent_cursor_ease_out_cubic(progress);
	_currentPoint = NSMakePoint(
		_animationFrom.x + (_animationTo.x - _animationFrom.x) * eased,
		_animationFrom.y + (_animationTo.y - _animationFrom.y) * eased
	);
	if (progress >= 1.0) {
		_animating = NO;
		_currentPoint = _animationTo;
	}
}

- (void)updateWindowFrame {
	if (_window == nil || !_hasPoint) {
		return;
	}
	CGFloat hotspotBottom = gut_agent_cursor_window_size - gut_agent_cursor_hotspot_top;
	CGFloat frameX = _currentPoint.x - gut_agent_cursor_hotspot_x;
	CGFloat frameY = NSMaxY(_desktopFrame) - _currentPoint.y - hotspotBottom;
	[_window setFrame:NSMakeRect(frameX, frameY, gut_agent_cursor_window_size, gut_agent_cursor_window_size) display:NO];
}

- (CGFloat)clickPulseAtTime:(NSTimeInterval)now {
	CGFloat first = [self pulseProgressAtTime:now start:_clickPulseStartA duration:gut_agent_cursor_click_pulse_duration];
	CGFloat second = [self pulseProgressAtTime:now start:_clickPulseStartB duration:gut_agent_cursor_click_pulse_duration];
	return MAX(first, second);
}

- (CGFloat)scrollPulseAtTime:(NSTimeInterval)now {
	return [self pulseProgressAtTime:now start:_scrollPulseStart duration:gut_agent_cursor_scroll_pulse_duration];
}

- (CGFloat)pulseProgressAtTime:(NSTimeInterval)now start:(NSTimeInterval)start duration:(NSTimeInterval)duration {
	if (start <= 0.0 || now < start || duration <= 0.0) {
		return 0.0;
	}
	CGFloat progress = gut_agent_cursor_clamp((now - start) / duration, 0.0, 1.0);
	if (progress >= 1.0) {
		return 0.0;
	}
	return 1.0 - progress;
}

@end

static void gut_agent_cursor_show_async(void *context) {
	gut_agent_cursor_context *payload = (gut_agent_cursor_context *)context;
	if (payload == NULL) {
		return;
	}

	@autoreleasepool {
		NSString *kind = gut_agent_cursor_string(payload->kind);
		NSString *button = gut_agent_cursor_string(payload->button);
		NSString *direction = gut_agent_cursor_string(payload->direction);
		NSPoint position = NSMakePoint((CGFloat)payload->position_x, (CGFloat)payload->position_y);
		NSPoint target = NSMakePoint((CGFloat)payload->target_x, (CGFloat)payload->target_y);
		NSTimeInterval duration = payload->duration_ms > 0 ? (NSTimeInterval)payload->duration_ms / 1000.0 : 0.0;
		[[GutAgentCursorController sharedController] handleEventWithKind:kind position:position hasTarget:payload->has_target != 0 target:target button:button direction:direction pressed:payload->pressed != 0 duration:duration];
	}

	free(payload);
}

static void gut_agent_cursor_hide_async(void *context) {
	(void)context;
	@autoreleasepool {
		[[GutAgentCursorController sharedController] hideCursor];
	}
}

int gut_darwin_show_agent_cursor(
	const char *kind,
	int64_t position_x,
	int64_t position_y,
	int has_target,
	int64_t target_x,
	int64_t target_y,
	const char *button_token,
	const char *direction_token,
	int pressed,
	int64_t duration_ms
) {
	if (kind == NULL || kind[0] == '\0') {
		return 1;
	}

	gut_agent_cursor_context *context = (gut_agent_cursor_context *)calloc(1, sizeof(gut_agent_cursor_context));
	if (context == NULL) {
		return 2;
	}

	context->position_x = position_x;
	context->position_y = position_y;
	context->has_target = has_target;
	context->target_x = target_x;
	context->target_y = target_y;
	context->pressed = pressed;
	context->duration_ms = duration_ms;
	gut_agent_cursor_copy_token(context->kind, sizeof(context->kind), kind);
	gut_agent_cursor_copy_token(context->button, sizeof(context->button), button_token);
	gut_agent_cursor_copy_token(context->direction, sizeof(context->direction), direction_token);

	gut_run_on_main_thread_async(gut_agent_cursor_show_async, context);
	return 0;
}

int gut_darwin_hide_agent_cursor(void) {
	gut_run_on_main_thread_async(gut_agent_cursor_hide_async, NULL);
	return 0;
}
