# gut Cookbook

## Inspect the current app

1. `status`
2. `window_active`
3. `accessibility_snapshot`
4. `window_elements` or `accessibility_search`

## Find a button and click it

1. `window_active`
2. `accessibility_search` with a role/title filter or `window_find_elements`
3. `accessibility_action` with `perform_ref_action` when you have a ref
4. Fall back to `mouse_action` only if you cannot get a stable ref

## Visual verification

1. `screen_capture`
2. `screen_color_at` or `screen_find_color`
3. Re-run `window_active` or `accessibility_snapshot` after the UI changes

## Type into the focused control

1. `accessibility_snapshot` to confirm focus
2. `keyboard_action` with `type` or `tap`
3. `clipboard_write` if a paste-based flow is more reliable
