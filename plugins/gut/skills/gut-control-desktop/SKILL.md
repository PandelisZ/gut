---
name: gut-control-desktop
description: Control the local desktop with gut after a target has already been identified. Use when the user wants to focus windows, click or drag the mouse, type keys, update the clipboard, or perform accessibility actions through gut.
---

# gut-control-desktop

## Preconditions

1. Call `status` and confirm `mutationAllowed=true`.
2. Resolve the target first with `window_*` or `accessibility_*` read-only tools.
3. If the action is destructive or ambiguous, restate the target in your own words before invoking the tool.

## Preferred Mutating Tools

- `window_action` for focus, move, resize, minimize, and restore
- `mouse_action` for move, click, drag, press, release, and scroll
- `keyboard_action` for text entry and key combos
- `clipboard_write` for clipboard updates
- `accessibility_action` for focused-element, point-based, or ref-based actions

## Guardrails

- Prefer `window_action` or `accessibility_action` over raw mouse coordinates when both are viable.
- Keep multi-step UI flows observable: inspect, act, then inspect again.
- If a mutating call is blocked, tell the user mutation is disabled rather than inventing a fallback.

## References

- [Safety](../../references/safety.md)
- [Cookbook](../../references/tool-cookbook.md)
- [Key Map](../../references/key-map.md)
