# Muxway Creative Canvas · foundation (Preview 01)

This is the first incremental implementation of the image/video creative workbench. It **does not replace** the API relay or any existing billing route.

## First milestone delivered

- Authenticated `/create` route and new sidebar entry, with both user/admin workspace navigation.
- Responsive canvas with pan, zoom, draggable nodes, prompt/image/video/note types, connection lines, DAG checks, node inspector, duplicate/remove, undo/redo, light/dark theme support.
- One starter workflow (prompt → image → video). All outputs show **not run**, never fake generated assets.
- Multi-project editing, debounced drafts isolated by signed-in user ID, safe JSON export/import, and per-user IndexedDB for local image/video uploads.
- Manual generation of image/video nodes through **existing** `/v1/images/generations`, `/v1/videos` and owner-scoped `GET` status routes. The chosen API key lives in component memory only, not project JSON or storage.
- Generated outputs appear on nodes only after a successful, settled backend task. Existing price/entitlement/escrow checks continue to reject unapproved models/options.
- During Preview 01, only administrators see an enabled paid-run button; all other users can edit the canvas but cannot submit paid media from this route. Every paid POST requires a direct click and browser confirmation. A job ID is stored immediately; on timeout/unknown state, the node blocks repeat submissions. Only GET status calls are automatically repeated.

## Explicit MVP limits

- Canvas drafts and uploaded media are **browser-local only**. There is no cloud sync, cross-device persistence, binary export, team collaboration, or server file workspace. Local browser data can be cleared; users should export their project JSON for structural backups. JSON does not contain uploaded binary assets.
- Connections model the creative graph but **do not automatically execute** other nodes or charge users. Prompt text connects to generation nodes; image-to-video reference media parameter mapping is intentionally deferred until provider-specific capabilities and pricing are validated.
- Current model IDs are editable starting examples, **not** a guarantee of key authorization or an approved price. Video options must match the server's exact verified pricing tier; the default `{}` is a draft.
- The client does not give an authoritative upfront quote. An administrator-approved USD price is enforced server-side, but a price quote preview before confirmation is a next milestone. Until that is complete, this route is presented as **Preview 01**, with paid execution restricted to signed-in administrators using authorized keys and explicit confirmations.
- Task status polling is bounded; after page reload the user needs to select the original API Key to resume status reads. Another user's key cannot read the task.

## Next milestones

1. Server-owned canvas projects and tenant-scoped asset storage, with optimistic updates, quotas and lifecycle cleanup.
2. Signed read-only price quote endpoint for an authenticated key + exact node model/options; show USD totals before paid confirmation.
3. Provider capability schema for image-to-video, video duration/aspect resolution and uploads; then sequential DAG execution with queue admission.
4. Private gallery/history, task recovery, branch variants, pricing per node and clear success/failed/refunded states.
5. Accessibility and full-browser interaction tests on iPad Safari and desktop; release from Preview into the main product after billing review.

## Test and release gates

- `pnpm run build` (Vue TSC + Vite).
- `pnpm run test:run -- src/smirel/__tests__/creative-canvas.spec.ts src/smirel/__tests__/creative-canvas-page.spec.ts`.
- Run full CI, security scan and production-control checks before merge.
- No production deployment or enabled billing/account changes just to test this UI. Preserve existing Cangyuan QA safety settings.
