# go-auth

A JWT authentication API in Go, built as a learning project. Gin for HTTP,
zerolog for logging, and a hand-written composition root instead of a DI
framework.

**Status: v1.0.0, the foundation.** The runtime skeleton is complete —
configuration, logging, the HTTP server with graceful shutdown, CORS, a health
check, and a Postgres connection opened at startup and closed on the way out.
The authentication itself (users, registration, login, token issuing and
refresh, the protected routes) is not written yet.
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

## Getting started

```sh
git clone https://github.com/bkjonathan/go-authentication.git
cd go-authentication
cp .env.example .env
make docker-up
make run
```

`make docker-up` starts Postgres from
[`docker/docker-compose.yml`](docker/docker-compose.yml), reading its user,
password, database and port from the same `.env`. The API refuses to start if
the database is unreachable.

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

`make lint` needs [golangci-lint](https://golangci-lint.run/) on your `PATH`;
the Makefile adds `$(go env GOPATH)/bin` for you.

## Layout

```
cmd/api/                 entry point: load config, build the app, run it
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
```

The rule the packages follow is that a layer only knows about the one below it —
no `gin.Context` below `handlers`, no database access above `repositories`.
[`ARCHITECTURE.md`](ARCHITECTURE.md) has the reasoning, including why there is a
hand-written container instead of `wire` or `fx`.

## Next up

- User model and storage
- Register and login handlers, password hashing
- Access and refresh token issuing, and the middleware that verifies them
- Tests
