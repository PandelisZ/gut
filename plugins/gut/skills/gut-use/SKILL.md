---
name: gut-use
description: Use the gut MCP plugin to inspect and control the local desktop through capability-aware tools. Use when the user wants desktop automation, UI inspection, screenshots, window targeting, accessibility search, local clipboard access, mouse actions, or keyboard actions via gut.
---

# gut-use

## Workflow

1. Start with `status`.
2. Prefer read-only tools to localize the target before any mutation.
3. Use `window_active`, `window_list`, `window_elements`, `window_find_elements`, `accessibility_snapshot`, and `accessibility_search` before guessing screen coordinates.
4. Use `screen_capture`, `screen_color_at`, or `screen_find_color` when accessibility metadata is insufficient.
5. Use mutating tools only after the user intent is clear.

## Guardrails

- Treat `status.mutationAllowed=false` as a hard stop for mutating tools.
- Do not assume OCR or image-template search exists in v1; they are intentionally not exposed.
- Prefer accessibility refs, window handles, and exact element matches over brittle coordinate clicks.

## References

- [Safety](../../references/safety.md)
- [Cookbook](../../references/tool-cookbook.md)
- [Key Map](../../references/key-map.md)
