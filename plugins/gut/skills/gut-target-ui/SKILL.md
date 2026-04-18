---
name: gut-target-ui
description: Find a stable desktop UI target with gut before automation. Use when an agent needs to identify the right window, accessibility element, ref, color anchor, or coordinate for a local UI task.
---

# gut-target-ui

## Targeting Order

1. Use `window_active` or `window_list` to choose the right window.
2. Use `accessibility_snapshot` to inspect the current focused window and element.
3. Use `accessibility_search` when the UI can be described by role, title/value substrings, or a window scope.
4. Use `window_elements` or `window_find_elements` when you need the local element tree for a known window.
5. Use `screen_capture` and `screen_find_color` only when accessibility metadata is missing or unreliable.

## Output Preference

- Return the most stable selector available: window handle, accessibility ref, or exact element match.
- Return coordinates only when the stable selector path is unavailable.
- Include the tool evidence you used so later steps can validate the same target quickly.

## References

- [Safety](../../references/safety.md)
- [Cookbook](../../references/tool-cookbook.md)
