# TODO checklist for wiring DB and JWT environment

- [x] Wire DB into RegisterRoutes and pass *sql.DB to LoginHandler
- [x] Run migrations at startup using internal/database.RunMigrations
- [x] Load JWT_SECRET from environment in login handler
- [ ] Move JWT secret into startup config (centralized) and propagate to handlers
- [ ] Replace placeholder session token hash with secure HMAC or random token + hash
- [ ] Create templ fragments for login success/error and render them
- [ ] Add session cleanup goroutine
- [ ] Add rate limiting middleware for auth endpoints
- [ ] Add tests for auth flow
