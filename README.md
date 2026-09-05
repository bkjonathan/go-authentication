# go-auth

A JWT authentication API in Go, built as a learning project. Gin for HTTP,
zerolog for logging, and a hand-written composition root instead of a DI
framework.

**Status: v1.0.0, the foundation.** The runtime skeleton is complete —
configuration, logging, the HTTP server with graceful shutdown, CORS, a health
check, and a Postgres connection opened at startup and closed on the way out.
The database schema — users, refresh tokens, password reset tokens — is
described by GORM models in [`cmd/schema/models/`](cmd/schema/models/) and
applied by versioned Atlas migrations. The authentication itself (registration,
login, token issuing and refresh, the protected routes) is not written yet.
[`ARCHITECTURE.md`](ARCHITECTURE.md) describes the layered shape the code is
growing into; the repository and service layers do not exist in the tree today.

## What works today

| Method | Path      | Response              |
| ------ | --------- | --------------------- |
| `GET`  | `/health` | `{"status":"ok"}`     |

## Requirements

- Go 1.26.4 or newer
- Docker, for the Postgres container — or your own Postgres reachable at the
  `DB_*` settings below
- [Atlas](https://atlasgo.io/getting-started#installation) on your `PATH`, for
  the migration targets. `brew install ariga/tap/atlas`

## Getting started

```sh
git clone https://github.com/bkjonathan/go-authentication.git
cd go-authentication
cp .env.example .env
make docker-up
make migrate-up
make run
```

`make docker-up` starts Postgres from
[`docker/docker-compose.yml`](docker/docker-compose.yml), reading its user,
password, database and port from the same `.env`. `make migrate-up` creates the
tables. The API refuses to start if the database is unreachable.

The server logs the address it binds to:

```
2026-09-05T01:12:00+07:00 INF http server listening address=:3000
```

Check it:

```sh
curl localhost:3000/health
# {"status":"ok"}
```

`Ctrl-C` shuts it down: the server stops accepting connections, in-flight
requests get up to 15 seconds to finish, then the process exits.

## Database schema

The GORM structs in [`cmd/schema/models/`](cmd/schema/models/) are the only
definition of the schema. Nothing else describes a table — the migration files
are generated from the models, never written by hand.

They live next to the tool that reads them rather than under `internal/`
because nothing in the API imports them yet: today the models exist to generate
SQL. When the repository layer arrives and the API starts loading rows, they
move to `internal/models` and `cmd/schema` imports them from there.

| Table | Model | Holds |
| --- | --- | --- |
| `users` | [`user.go`](cmd/schema/models/user.go) | Account, credentials, role, status, lockout counters. |
| `refresh_tokens` | [`refresh_token.go`](cmd/schema/models/refresh_token.go) | One row per issued refresh token, in rotation families, with a revocation reason. |
| `password_reset_tokens` | [`password_reset_token.go`](cmd/schema/models/password_reset_token.go) | One row per reset request; single use. |

Every table embeds [`Base`](cmd/schema/models/base.go): a UUID primary key
minted before insert, `created_at` / `updated_at`, a soft-delete `deleted_at`,
and a `version` counter incremented in SQL on every update.

### Roles

A user holds exactly one role, stored as a string in `users.role` and checked
by a database constraint. There is no permission table and no join table — the
role *is* the authorisation, and a route guard asks for the roles it accepts.

| Role | |
| --- | --- |
| `owner` | Full control, including the things an admin cannot undo. |
| `admin` | Administration of the whole workspace. |
| `manager` | Runs a team; more than staff, less than admin. |
| `staff` | The default for a new account. |

The constants and the validity check live in
[`cmd/schema/models/role.go`](cmd/schema/models/role.go); the check constraint
`chk_users_role` keeps the same four values true in the database. `Status` in
[`user_status.go`](cmd/schema/models/user_status.go) is a separate axis: the
role says what an account may do, the status says whether it may sign in at all.

## Migrations

Atlas compares three states, all built from `.env` by the Makefile and named in
[`atlas.hcl`](atlas.hcl):

- **the shadow schema** — a scratch schema in your database that
  `go run ./cmd/schema` drops and rebuilds from the models. This is the desired
  state, and rebuilding it before every diff is what keeps it from going stale.
- **the migration directory** — [`cmd/schema/migrations/`](cmd/schema/migrations/),
  replayed in a throwaway container to work out what the files add up to. This
  is the current state.
- **your database** — where the generated files get applied.

A migration is the difference between the first two.

### Setting up a new database, step by step

Starting from a fresh clone, an empty Postgres, or a database you have just
wiped with `make docker-reset`:

1. **Configure.** `cp .env.example .env`, then edit `DB_NAME`, `DB_USER`,
   `DB_PASSWORD` if you want something other than the defaults. The Makefile,
   `docker-compose.yml` and the application all read this one file.
2. **Start Postgres.** `make docker-up`, then `make docker-ps` until the
   container reports `healthy`. Point at your own server instead by leaving the
   container down and setting `DB_HOST` / `DB_PORT`.
3. **Check what is applied.** `make migrate-status`. On an empty database it
   reports `PENDING`, no current version, and lists every file as pending.
4. **Apply the migrations.** `make migrate-up`. Atlas runs each pending file in
   order and records it in an `atlas_schema_revisions` table it creates for the
   purpose.
5. **Confirm.** `make migrate-status` should now say `OK` and
   `Already at latest version`.
6. **Confirm the models and the database agree.** `make db-diff name=check` —
   it should print `The migration directory is synced with the desired state`
   and write nothing. If it does generate a file, the models and the migrations
   have drifted; delete it and read the next section.
7. **Run the API.** `make run`.

### Changing the schema, step by step

Never edit an applied migration — add a new one:

1. **Edit the model.** Change a struct in `cmd/schema/models/`, or add a file
   for a new table.
2. **Register a new table.** Add it to `All()` in
   [`models.go`](cmd/schema/models/models.go), *after* every table it
   references — the list is the order the tables are created in.
3. **Generate the migration.** `make db-diff name=add_wishlist`. The name
   becomes the filename, so make it describe the change. This rebuilds the
   shadow schema first, so it always diffs against the models as they are on
   disk right now.
4. **Read the SQL it wrote.** Two files land in `cmd/schema/migrations/`, `.up`
   and `.down`. Atlas is good but not clairvoyant: a rename reads to it as a
   drop plus an add, and dropping a column throws the data away. Fix anything
   destructive by hand, then run `make db-hash` so `atlas.sum` matches again.
5. **Apply it.** `make migrate-up`.
6. **Verify.** `make db-diff name=check` should report nothing to do.

Useful while iterating:

- `make db-inspect` prints the DDL the models currently describe, without
  writing a migration.
- `make migrate-down` rolls back the last migration, `make migrate-reset` rolls
  back all of them.
- `make docker-reset` deletes the database volume entirely — the fastest way
  back to step 1 when a half-applied experiment gets tangled.

## Configuration

Every setting is read from the environment once at startup, into a struct
([`internal/config/cofig.go`](internal/config/cofig.go)). A `.env` file in the
project root is loaded if present, and real environment variables win over it.

| Variable                  | Default                 | Meaning                                            |
| ------------------------- | ----------------------- | -------------------------------------------------- |
| `PORT`                    | `8090`                  | Port the HTTP server listens on.                    |
| `GIN_MODE`                | `debug`                 | Gin's mode: `debug`, `release`, or `test`. Anything other than `release` also turns on human-readable console logging. |
| `JWT_SECRET`              | `you_jwt_secret_key`    | Signing key for access tokens. Replace it before this runs anywhere real. |
| `JWT_EXPIRES_IN`          | `24h`                   | Access token lifetime, as a Go duration.            |
| `REFRESH_TOKEN_EXPIRES_IN`| `720h`                  | Refresh token lifetime, as a Go duration.           |
| `DB_HOST`                 | `localhost`             | Postgres host.                                      |
| `DB_PORT`                 | `5432`                  | Postgres port. Also the port the container publishes.|
| `DB_USER`                 | `postgres`              | Postgres user.                                      |
| `DB_PASSWORD`             | `password`              | Postgres password. Replace it before this runs anywhere real. |
| `DB_NAME`                 | `go_auth`               | Database name.                                      |
| `DB_SSLMODE`              | `disable`               | `sslmode` in the connection string.                 |

`.env` is gitignored; `.env.example` is the checked-in template.

## Make targets

| Command       | Does                                                |
| ------------- | --------------------------------------------------- |
| `make help`   | Lists the targets.                                   |
| `make run`    | Runs the API from source.                            |
| `make dev`    | Same as `run`.                                       |
| `make build`  | Builds a binary to `bin/app`.                        |
| `make format` | `gofmt -s -w .` over the tree.                       |
| `make lint`   | Formats, then runs `golangci-lint run ./...`.        |
| `make docker-up`    | Starts Postgres in the background.              |
| `make docker-down`  | Stops it.                                       |
| `make docker-logs`  | Follows the container logs.                     |
| `make docker-ps`    | Shows the container status.                     |
| `make docker-reset` | Stops it and deletes the data volume.           |
| `make db-shadow`         | Rebuilds the shadow schema from the models.|
| `make db-diff name=xxx`  | Generates a migration from the models.     |
| `make db-inspect`        | Prints the DDL the models describe.        |
| `make db-hash`           | Re-hashes `atlas.sum` after a hand edit.   |
| `make migrate-up`        | Applies pending migrations.                |
| `make migrate-down`      | Rolls back the last migration.             |
| `make migrate-reset`     | Rolls back every migration.                |
| `make migrate-status`    | Shows the applied version.                 |

`make lint` needs [golangci-lint](https://golangci-lint.run/) on your `PATH`;
the Makefile adds `$(go env GOPATH)/bin` for you.

## Layout

```
cmd/
  api/                   entry point: load config, build the app, run it
  schema/
    main.go              rebuilds the shadow schema Atlas diffs against
    models/              the GORM structs - the schema is generated from these
    migrations/          generated SQL, applied in order
internal/
  config/                environment variables parsed into a struct
    app/                 the composition root
      app.go             wiring, signal handling, graceful shutdown
      container.go       builds the dependency graph, owns the DB handle
  database/              opens the GORM/Postgres connection
  server/
    server.go            the http.Server and its timeouts
    router.go            the route table
  handlers/              HTTP handlers (currently an empty Registry)
  middleware/            CORS today; auth guards to come
  logger/                zerolog setup
docker/                  the Postgres service used in development
atlas.hcl                the three states Atlas compares
```

The rule the packages follow is that a layer only knows about the one below it —
no `gin.Context` below `handlers`, no database access above `repositories`.
[`ARCHITECTURE.md`](ARCHITECTURE.md) has the reasoning, including why there is a
hand-written container instead of `wire` or `fx`.

## Next up

- Repository layer over the existing models
- Register and login handlers, password hashing
- Access and refresh token issuing, and the middleware that verifies them
- A role guard on the protected routes
- Tests
