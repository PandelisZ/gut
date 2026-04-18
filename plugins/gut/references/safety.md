# gut Safety

Use `status` first.

- `mutationAllowed=false` means `window_action`, `mouse_action`, `keyboard_action`, `clipboard_write`, and `accessibility_action` will be blocked.
- macOS hosts may need Accessibility and Screen Recording permissions before screen or accessibility tools succeed.
- Linux hosts depend on the active display session. Status will show capability-level failures when the environment is incomplete.
- v1 does not expose OCR or image-template matching. Do not build plans around those capabilities.
- Prefer exact window handles and accessibility refs over coordinate-only plans.
