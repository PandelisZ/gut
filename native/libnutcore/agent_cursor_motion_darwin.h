#import <Cocoa/Cocoa.h>
#include <math.h>

typedef struct gut_agent_cursor_motion_profile {
	CGFloat distance;
	CGFloat heading_radians;
	CGFloat hop_height;
	CGFloat sway_amplitude;
	CGFloat trail_strength;
} gut_agent_cursor_motion_profile;

typedef struct gut_agent_cursor_motion_sample {
	NSPoint point;
	CGFloat rotation_radians;
	CGFloat hover_lift;
	CGFloat scale_boost;
	CGFloat trail_strength;
} gut_agent_cursor_motion_sample;

typedef struct gut_agent_cursor_settle_sample {
	CGFloat rotation_radians;
	CGFloat hover_lift;
	CGFloat scale_boost;
	CGFloat trail_strength;
} gut_agent_cursor_settle_sample;

static inline CGFloat gut_agent_cursor_clamp(CGFloat value, CGFloat min_value, CGFloat max_value) {
	if (value < min_value) {
		return min_value;
	}
	if (value > max_value) {
		return max_value;
	}
	return value;
}

static inline CGFloat gut_agent_cursor_ease_out_cubic(CGFloat progress) {
	CGFloat clamped = gut_agent_cursor_clamp(progress, 0.0, 1.0);
	CGFloat inverse = 1.0 - clamped;
	return 1.0 - inverse * inverse * inverse;
}

static inline CGFloat gut_agent_cursor_ease_in_out_sine(CGFloat progress) {
	CGFloat clamped = gut_agent_cursor_clamp(progress, 0.0, 1.0);
	return 0.5 - 0.5 * cos(clamped * M_PI);
}

static inline gut_agent_cursor_motion_profile gut_agent_cursor_make_motion_profile(NSPoint from, NSPoint to) {
	CGFloat delta_x = to.x - from.x;
	CGFloat delta_y = to.y - from.y;
	CGFloat distance = hypot(delta_x, delta_y);

	gut_agent_cursor_motion_profile profile;
	profile.distance = distance;
	profile.heading_radians = distance > 0.0 ? atan2(delta_y, delta_x) : 0.0;
	profile.hop_height = gut_agent_cursor_clamp(8.0 + distance * 0.03, 8.0, 34.0);
	profile.sway_amplitude = gut_agent_cursor_clamp(3.0 + distance * 0.018, 3.0, 16.0);
	profile.trail_strength = gut_agent_cursor_clamp(distance / 180.0, 0.0, 1.0);
	return profile;
}

static inline gut_agent_cursor_motion_sample gut_agent_cursor_sample_motion(
	gut_agent_cursor_motion_profile profile,
	NSPoint from,
	NSPoint to,
	CGFloat progress
) {
	CGFloat eased = gut_agent_cursor_ease_in_out_sine(progress);
	CGFloat direction_x = profile.distance > 0.0 ? cos(profile.heading_radians) : 1.0;
	CGFloat direction_y = profile.distance > 0.0 ? sin(profile.heading_radians) : 0.0;
	CGFloat normal_x = -direction_y;
	CGFloat normal_y = direction_x;
	CGFloat sway_envelope = sin(eased * M_PI);
	CGFloat sway = sin(eased * M_PI * 1.45) * sway_envelope * profile.sway_amplitude;
	CGFloat hop = -sin(eased * M_PI) * profile.hop_height;
	CGFloat settle = pow(1.0 - eased, 1.4);

	gut_agent_cursor_motion_sample sample;
	sample.point = NSMakePoint(
		from.x + (to.x - from.x) * eased + normal_x * sway,
		from.y + (to.y - from.y) * eased + normal_y * sway + hop
	);
	sample.rotation_radians = gut_agent_cursor_clamp(
		direction_x * 0.09 + sin(eased * M_PI * 2.0) * 0.04 * settle,
		-0.18,
		0.18
	);
	sample.hover_lift = -1.2 * sway_envelope;
	sample.scale_boost = 0.012 * sway_envelope;
	sample.trail_strength = profile.trail_strength * (0.25 + 0.75 * sway_envelope);
	return sample;
}

static inline gut_agent_cursor_settle_sample gut_agent_cursor_sample_settle(CGFloat progress) {
	CGFloat clamped = gut_agent_cursor_clamp(progress, 0.0, 1.0);
	CGFloat envelope = pow(1.0 - clamped, 1.7);
	CGFloat wobble = sin(clamped * M_PI * 2.6) * envelope;

	gut_agent_cursor_settle_sample sample;
	sample.rotation_radians = wobble * 0.09;
	sample.hover_lift = -fabs(wobble) * 1.8;
	sample.scale_boost = fabs(wobble) * 0.03;
	sample.trail_strength = fabs(wobble) * 0.18;
	return sample;
}
