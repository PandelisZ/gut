---
name: gut-inspect-desktop
description: Inspect the local desktop with gut without mutating it. Use when the user wants to understand windows, accessibility state, screen colors, screenshots, or clipboard text before taking action.
---

# gut-inspect-desktop

## Preferred Tools

- `status` for capability and permission readiness
- `window_active` and `window_list` for top-level targeting
- `window_elements` and `window_find_elements` for exact UI structure
- `accessibility_snapshot` and `accessibility_search` for focused UI state and searchable element refs
- `screen_capture`, `screen_color_at`, and `screen_find_color` for visual inspection
- `clipboard_read` for clipboard text

## Heuristics

- Use the active window first unless the task clearly references another window.
- Capture a screenshot when the UI hierarchy and visible state may disagree.
- If an inspection call returns permission or capability errors, surface them directly instead of retrying blindly.

## References

- [Safety](../../references/safety.md)
- [Cookbook](../../references/tool-cookbook.md)
