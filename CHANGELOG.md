# Change Log

## 2026-09-12

### Administrator balance allocation

- Added an administrator-facing balance allocation action to the user management page.
- Added add, subtract, and set operations with amount validation, post-adjustment preview, optional notes, and success/error notifications.
- Reused `POST /api/v1/admin/users/:id/balance`, preserving atomic balance updates, non-negative balance protection, idempotency, cache invalidation, and adjustment history records.
- Added preview-mode behavior so balance changes can be exercised without modifying the database.

### Deployment

- Rebuilt `termrelay-commercial-dark:local` from the `release/smirel-commercial-dark` source baseline.
- Verified the frontend type check and production build.
- Recreated only the application container; PostgreSQL, Redis, and persistent volumes were preserved.

### Branch alignment and cleanup

- Kept the implementation aligned with `release/smirel-commercial-dark` and removed inactive CQU browser-bridge routing, configuration, tests, and provider assets.
- Added editing for existing upstream accounts in the administrator console.
- Kept the model plaza and group views focused on the active commercial provider set.
- Preserved the local build adjustments required for Apple Container deployment, including the npm mirror argument and frontend memory limit.
