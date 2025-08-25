- [x] Wire DB into RegisterRoutes and pass *sql.DB to LoginHandler
  - [x] Update cmd/server/main.go to open DB and run migrations
  - [x] Pass cfg and db into RegisterRoutes
- [x] Run migrations at startup using internal/database.RunMigrations
- [x] Load JWT_SECRET from environment in login handler (centralized via config)
- [x] Replace placeholder session token hash with secure opaque token + HMAC (SESSION_SIGNING_KEY)
- [x] Implement Auth middleware to validate cookie tokens and attach user to context
- [x] Create protected Dashboard page and templ fragment (with HTMX logout form)
- [x] Implement logout endpoint (POST /api/auth/logout) that deactivates session and clears cookie

- [ ] Move JWT secret & session signing key handling to a more explicit startup config (done: LoadFromEnv; consider validation)
- [ ] Replace inline login success/error HTML with templ fragments (templates/fragments/login_success.templ, login_error.templ) and render them from LoginHandler
  - [ ] Create fragments
  - [ ] Update LoginHandler to render fragments
  - [ ] Run `templ generate` and remove any shim files
- [ ] Add session cleanup goroutine to remove expired sessions periodically (e.g., every 10m)
  - [ ] Implement models.DeleteExpiredSessions(ctx, db)
  - [ ] Start a background goroutine in main.go to run periodically
- [ ] Add logout confirmation/change UI (HTMX fragment) instead of redirect

- [ ] Add CSRF protection for mutating routes (including /api/auth/login and /api/auth/logout)
  - [ ] Consider SameSite enforcement, CSRF token in forms (templ), and HTMX considerations
- [ ] Add rate limiting middleware on auth endpoints (per-IP or per-user)
- [ ] Add auth middleware tests and model tests
  - [ ] Unit tests for ValidateAndDeleteChallenge
  - [ ] Unit tests for models.EnsureUser, CreateSession, GetSessionByHash, DeactivateSessionByHash (use in-memory sqlite)
  - [ ] Integration test for login → session → dashboard → logout flow

- [ ] Developer UX / CI
  - [ ] Add Makefile targets: gen, run, build, test
  - [ ] Add CI workflow: run templ generate (or install templ), go vet, go test, build (with CGO if desired)
  - [ ] Document CGO requirement and templ CLI in README (done; consider adding more examples)

Notes / blockers
- Run `templ generate` locally to produce templ-generated Go files; some shims exist in templates/pages to make the repo build without generated code.
- SESSION_SIGNING_KEY should be set in .env for production; rotating keys or key IDs are not yet supported.

If you want, I can pick the next item and implement it now. Suggested next actions (in order):
1. Create templ fragments for login success/error and render them from LoginHandler (small, user-visible UX improvement).
2. Add session cleanup goroutine.
3. Add CSRF protection for login/logout endpoints.

Which would you like me to work on next?