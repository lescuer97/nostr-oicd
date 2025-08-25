- [x] Wire DB into RegisterRoutes and pass *sql.DB to LoginHandler
  - [x] Update cmd/server/main.go to open DB and run migrations
  - [x] Pass cfg and db into RegisterRoutes
- [x] Run migrations at startup using internal/database.RunMigrations
- [x] Load JWT_SECRET from environment in login handler (centralized via config)
- [x] Replace placeholder session token hash with secure opaque token + HMAC (SESSION_SIGNING_KEY)
- [x] Implement Auth middleware to validate cookie tokens and attach user to context
- [x] Create protected Dashboard page and templ fragment (with HTMX logout form)
- [x] Implement logout endpoint (POST /api/auth/logout) that deactivates session and clears cookie

- [x] Replace inline login success/error HTML with templ fragments and render them from LoginHandler
  - [x] Created templates/fragments/login.templ
  - [x] LoginHandler renders fragments.LoginSuccess / fragments.LoginError
- [ ] Run `templ generate` and remove any shim files (still pending; run locally)

- [ ] Add session cleanup goroutine to remove expired sessions periodically (e.g., every 10m)
  - [ ] Implement models.DeleteExpiredSessions(ctx, db) (function exists and marks expired sessions inactive)
  - [ ] Start a background goroutine in main.go to run periodically (not implemented)
- [ ] Add CSRF protection for mutating routes (including /api/auth/login and /api/auth/logout)
  - [ ] Consider SameSite enforcement, CSRF token in forms (templ), and HTMX considerations (HTMX sends X-Requested-With / HX-Request)
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
- Run `templ generate` locally to produce templ-generated Go files; some shims exist to allow building without generated code. The templ CLI is required to generate component Go files (not checked into git).
- SESSION_SIGNING_KEY should be set in .env for production; rotating keys or key IDs are not yet supported.
- models.DeleteExpiredSessions already exists and marks expired sessions inactive; starting a background cleaner in main.go is the remaining bit.

Suggested next actions (pick one)
1. Run `templ generate` locally and remove generated-code shims. (You must run this locally; I can provide exact command and verify expected files.)
2. Add CSRF protection for HTMX login/logout endpoints. (I can implement middleware + templ changes to include a CSRF token compatible with HTMX.)
3. Add a session cleanup goroutine in cmd/server/main.go to call models.DeleteExpiredSessions every 10 minutes.
4. Add rate limiting middleware for auth endpoints (simple in-memory for dev or redis-backed for production).
5. Add unit/integration tests for auth/session flows.

If you want me to implement one of the above, tell me which and whether you'd like code changes applied here (I'll edit files) or a patch/diff you will apply locally.
