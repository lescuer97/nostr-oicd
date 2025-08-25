# nostr-oicd

Go Chi server with Nostr NIP-07 authentication, Templ templates, and Tailwind CSS Play CDN.

Development notes

- Templ components are stored in templates/ and compiled using `templ generate` locally. Generated Go files should not be committed (see .gitignore).
- Tailwind CSS is loaded from the Play CDN in templates/layouts/base.templ. No npm or build step is required for prototyping.
- SQLite driver: github.com/mattn/go-sqlite3 (requires CGO). Use modernc.org/sqlite if you prefer pure-Go builds.

Environment

- Copy `.env.example` to `.env` and update secrets (JWT_SECRET, DATABASE_PATH, COOKIE_*).
- The server will load a `.env` file automatically when present. For production, prefer platform-managed environment variables or secrets stores.

Dev run (local development)

1. Copy the example env file and edit values:

```sh
cp .env.example .env
# edit .env to set JWT_SECRET and other vars
```

2. (Optional) Install templ generator if you edit templ files:

```sh
go install github.com/a-h/templ/cmd/templ@latest
```

3. Generate templ Go code (if needed):

```sh
templ generate
```

4. Run the server (CGO required for mattn/go-sqlite3):

```sh
# Ensure CGO is enabled if using mattn/go-sqlite3
CGO_ENABLED=1 go run ./cmd/server
```

Notes & tips

- If you don't want to use a `.env` file, set environment variables directly (e.g., in your shell, systemd unit, or container runtime).
- The app will run migrations from `./database/migrations` at startup. Back up your DB before running in production.
- For CI, ensure `templ generate` is run or that the templ CLI is available.

