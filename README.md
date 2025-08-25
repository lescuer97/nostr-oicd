# nostr-oicd

Go Chi server with Nostr NIP-07 authentication, Templ templates, and Tailwind CSS Play CDN.

Development notes

- Templ components are stored in templates/ and compiled using `templ generate` locally. Generated Go files should not be committed (see .gitignore).
- Tailwind CSS is loaded from the Play CDN in templates/layouts/base.templ. No npm or build step is required for prototyping.
- SQLite driver: github.com/mattn/go-sqlite3 (requires CGO). Use modernc.org/sqlite if you prefer pure-Go builds.

Environment

- Copy `.env.example` to `.env` and update secrets (JWT_SECRET, DATABASE_PATH, COOKIE_*).

Dev run

1. (Optional) Install templ: `go install github.com/a-h/templ/cmd/templ@latest`
2. Generate templ code: `templ generate`
3. Run the server: `go run ./cmd/server`

