# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project uses
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2026-09-05

Authentication, end to end: an account can be created, signed in on as many
devices as it likes, and each of those sessions listed and ended on its own. The
layers `ARCHITECTURE.md` had been describing — repositories, services, DTOs,
errors — exist now.

### Added

- **Repository layer.** `internal/repositories`: `UserRepository` and
  `RefreshTokenRepository` behind interfaces, a `Store` that holds them, and
  `Store.Atomic` for anything that writes to more than one table.
- **Service layer.** `internal/services`: `AuthService` (register, login,
  refresh, logout, change password) and `SessionService` (list, revoke one,
  revoke the others). Neither knows gin or gorm.
- **Sessions.** A session is a refresh token family. Signing in opens one
  without touching any other, so an account can be signed in on several devices
  at once, and each session records the user agent and IP it was last used from.
  `GET /api/v1/sessions` lists them with the caller's own marked `current`;
  `DELETE /api/v1/sessions/:id` ends one; `DELETE /api/v1/sessions` ends every
  other one.
- **Refresh token rotation.** Each refresh spends the presented token and issues
  a replacement in the same family, linked by `replaced_by_token_id`. The revoke
  is guarded by `revoked_at IS NULL`, so two requests racing on one token cannot
  both win. Only the SHA-256 digest is stored.
- **Reuse detection.** A revoked token presented again means a copy is in
  circulation: the whole family is revoked with `reuse_detected` and that device
  has to sign in again.
- **Lockout.** Consecutive failures lock an account for `LOCKOUT_DURATION`. A
  failed attempt always answers `invalid_credentials`, whether or not the
  account exists and whether or not it just locked — the lock is only disclosed
  to someone who then gets the password right.
- **Access tokens.** HS256 JWTs carrying the user, the role and the session
  (`sid`). `Middleware.Authenticate` verifies them, `Middleware.RequireRole`
  guards on the role, and neither needs a database round trip.
- **Password hashing.** bcrypt at a configurable cost, plus a decoy comparison
  on an unknown address so response time does not reveal which addresses are
  registered.
- **Request and response shapes.** `internal/dto` for the contract,
  `internal/apperror` for failures in HTTP terms, and one envelope for every
  response: `{"success":…,"data":…}` or `{"success":…,"error":…}`. Handlers
  return their error; `internal/server/handle.go` renders it in one place and
  logs the ones that mean the server is broken.
- **Configuration.** `BCRYPT_COST`, `MAX_FAILED_LOGIN_ATTEMPTS` and
  `LOCKOUT_DURATION`, and `.env.example` now carries the `JWT_*` settings the
  README documents.

### Changed

- **Models moved to `internal/models`.** They sat under `cmd/schema` while that
  binary was their only importer; the repository layer means it no longer is.
  The migration workflow is unchanged.
- **Unique index on `refresh_tokens.token_digest`**, which every refresh looks
  up by. Migration `20260905045435_add_refresh_token_digest_index`.
- **`Middleware` is constructed with the token issuer** rather than the raw JWT
  secret, so verification lives in one place.
- **`database.New` sets `TranslateError`,** so a unique-index violation arrives
  as `gorm.ErrDuplicatedKey` and a repository can recognise "already taken"
  without knowing it is talking to Postgres.

### Not included

Password reset — the table and the model are there, but the flow needs
somewhere to send mail. Tests.

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

[1.1.0]: https://github.com/bkjonathan/go-authentication/releases/tag/v1.1.0
[1.0.0]: https://github.com/bkjonathan/go-authentication/releases/tag/v1.0.0
