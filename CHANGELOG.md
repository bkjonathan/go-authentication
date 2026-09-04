# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project uses
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-09-05

The foundation release: everything an API needs before the first endpoint is
written. HTTP, configuration, database and the layering pattern are settled, so
the next work is domain code rather than plumbing.

### Added

- **Composition root.** `internal/config/app` builds the dependency graph by
  hand (`container.go`) and owns the process lifecycle (`app.go`): start,
  signal handling, shutdown, resource cleanup. No DI framework —
  [`ARCHITECTURE.md`](ARCHITECTURE.md) explains the choice.
- **HTTP server.** Gin router behind an `http.Server` with read, write and idle
  timeouts, plus graceful shutdown — `SIGINT`/`SIGTERM` stops new connections
  and gives in-flight requests up to 15 seconds to finish.
- **Configuration.** Every setting read from the environment once at startup
  into a typed struct (`Server`, `Database`, `JWT`), with defaults for each.
  A `.env` file is loaded when present; real environment variables win over it.
- **Database.** GORM over Postgres, connected during container construction and
  closed on shutdown. Settings come from the `DB_*` environment variables.
- **Docker.** `docker/docker-compose.yml` runs Postgres 17 with a health check
  and a named volume, driven by the same `.env` as the application, and wrapped
  in `make docker-up` / `docker-down` / `docker-logs` / `docker-ps` /
  `docker-reset`.
- **Middleware.** CORS, constructed with the JWT config so the auth guard can
  be added alongside it later.
- **Logging.** zerolog, with human-readable console output outside
  `GIN_MODE=release` and structured JSON in it.
- **Routes.** `GET /health`, returning `{"status":"ok"}`.
- **Tooling.** `Makefile` targets for build, run, format and lint; `README.md`
  and `ARCHITECTURE.md`.

### Not included

Authentication itself. Users, registration, login, password hashing, access and
refresh token issuing, the middleware that verifies them, the repository and
service layers, and tests are all still to come.

[1.0.0]: https://github.com/bkjonathan/go-authentication/releases/tag/v1.0.0
