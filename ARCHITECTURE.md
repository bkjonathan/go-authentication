# Architecture

The API is layered. A request always travels in one direction, and each layer
only knows about the one below it:

```
                        HTTP request
                             |
                             v
   +-- the HTTP layer: knows gin, knows no SQL ----------------------+
   |  server       the route table, the Handle wrappers, the server  |
   |  middleware   CORS, authentication, the role guard              |
   |  handlers     bind, delegate, return                            |
   +-----------------------------------------------------------------+
                             |
                             v
     services      business rules: credentials, lockouts, tokens
                   knows neither gin nor gorm
                             |
                             v
     repositories  every query in the application
                   knows gorm, knows nothing about HTTP
                             |
                             v
                          PostgreSQL
```

Two rules keep it honest:

- **A layer never skips the one below it.** No SQL in a handler, no
  `gin.Context` in a service.
- **The arrows point one way.** `repositories` does not import `services`,
  `services` does not import `server`.

## The packages

Marked *(planned)* means the layer is described here but is not in the tree yet.

| Package | Holds | Depends on |
| --- | --- | --- |
| `cmd/api` | The entry point. Loads config, hands over to `app`. | `app`, `config`, `logger` |
| `cmd/schema` | The schema tool: rebuilds the shadow schema Atlas diffs against. | `models`, `config`, `database` |
| `cmd/schema/models` | The GORM structs. The database schema is generated from these. | — |
| `cmd/schema/migrations` | Generated SQL, applied in order. Not code. | — |
| `internal/config/app` | The composition root: builds the graph, runs it, shuts it down. | everything |
| `internal/server` | The router - the whole route table - the `Handle` wrappers, and the HTTP server. | `handlers`, `middleware`, `config` |
| `internal/handlers` | One handler type per area of the API. Thin: bind, delegate, return. | `services`, `middleware`, `dto` |
| `internal/middleware` | CORS, authentication, the role guard, and who the caller is. | `config`, `models`, `utils` |
| `internal/database` | Opens the GORM connection, and renders the DSN both binaries share. | `config` |
| `internal/logger` | zerolog setup. | — |
| `internal/config` | Environment variables, parsed once into a struct. | — |
| `internal/services` *(planned)* | Business logic. One service per area of the domain. | `repositories`, `dto`, `apperror` |
| `internal/repositories` *(planned)* | Every database query, behind an interface per aggregate. | `models`, `gorm` |
| `internal/dto` *(planned)* | Request and response bodies. The API contract. | — |
| `internal/apperror` *(planned)* | Failures described in HTTP terms. | — |
| `internal/providers` *(planned)* | Adapters for the outside world (local disk, S3). | `config` |
| `internal/utils` *(planned)* | Response envelope, validation messages, JWT, hashing, pagination. | — |

## The schema lives in one place

The GORM structs are the only definition of the database. There is no HCL
schema file, no hand-written `CREATE TABLE`: `cmd/schema` materialises the
models into a scratch schema, Atlas diffs that against the migration directory,
and the difference is the migration. Describing a table twice is how the two
descriptions start disagreeing, so the project describes it once.

`cmd/schema/models` sits under the binary that reads it because that binary is
its only importer today — nothing in the API loads a row yet. When the
repository layer lands the models move to `internal/models`, `cmd/schema`
imports them from there, and nothing else about the arrangement changes. The
table in the previous section already names `models` as the dependency the
`repositories` and `middleware` layers will take.

[`README.md`](README.md#migrations) has the workflow: the three states Atlas
compares, setting up a new database, and generating a migration from a model
change.

## Authorisation: one role per user

A user holds a single role — `owner`, `admin`, `manager` or `staff` — in a
string column on `users`, with a check constraint that keeps the four values
true in the database as well as in Go.

The alternative, a `roles` table with a permission array and a `user_roles`
join, was tried and removed. It buys runtime-editable permissions, and it costs
a join on every authenticated request, a preload every caller must remember (or
fail open), and a set of permission strings that only Go actually defines —
the database could not check them, so the constants were the real referent
anyway. For an application with four fixed roles, the column *is* the model.
`User.HasRole(roles ...Role)` answers the question a guard asks: "owner or
admin may do this."

Roles are orthogonal to `UserStatus`. The role says what an account may do; the
status (`active`, `pending_verification`, `suspended`, `deactivated`) says
whether it may sign in at all. A suspended owner is still an owner.

Reconsider the join-table shape if permissions ever have to be edited by an
administrator at runtime, or if a role's meaning has to vary per tenant.

## Dependency injection

Everything is constructor-injected, and the whole graph is built in one place:
[`internal/config/app/container.go`](internal/config/app/container.go).

```go
store := repositories.NewStore(db)

userService := services.NewUserService(store.Users, store.RefreshTokens)

registry := &handlers.Registry{
    Auth: handlers.NewAuthHandler(userService),
}
```

Nothing below the container constructs its own dependencies, and nothing reads
a global. That is what makes a service testable: hand it a fake repository and
it never touches Postgres.

### Why no DI framework?

Because at this size the framework costs more than it saves. `container.go` is
one function you can read top to bottom, step through in a debugger, and get a
compile error from the moment a dependency is missing. `google/wire` and
`uber-go/fx` solve a problem this project does not have yet.

Reconsider when the wiring genuinely hurts — roughly:

- the graph is large enough that reading the constructor order is a chore
  (say, 30+ providers), or
- components need lifecycle hooks (start/stop ordering, health checks), or
- the same graph has to be built several different ways.

`google/wire` is the smaller step: it generates exactly the code that is in
`container.go` today, from the same constructors, so switching is mechanical
and nothing else in the codebase changes.

## Interfaces: where and why

Interfaces sit at the edges — the places the application talks to something
outside itself, and therefore the places worth faking in a test:

- **`repositories.UserRepository`, `RefreshTokenRepository`, ...** — the
  database.
- **`providers.UploadProvider`** — file storage. `UPLOAD_PROVIDER=s3` swaps the
  implementation without a service noticing.

Services are deliberately **concrete structs**. A service is the code under
test, not the thing you stub out, and an interface per service would be
boilerplate the compiler cannot check against anything. If handler tests ever
need to fake one, add the interface then — it is a two-line change.

The Go convention "accept interfaces, return structs" is what the constructors
follow: `NewUserService` takes repository interfaces and returns a
`*UserService`.

## Transactions

Anything that writes to more than one table goes through `Store.Atomic`. The
store handed to the callback is bound to the transaction, so every repository
call inside it commits or rolls back as one:

```go
err := s.store.Atomic(func(tx *repositories.Store) error {
    if err := tx.Users.Create(&user); err != nil {
        return err
    }
    return tx.RefreshTokens.RevokeFamily(familyID, models.RevocationReasonPasswordChanged)
})
```

Token rotation is the real example: it revokes the presented token, writes its
replacement and links the two. A failure at any point leaves the database
exactly as it was — which matters here, because a half-written rotation is
indistinguishable from token theft.

## Adding a feature

A wishlist, end to end:

1. **Model** — `cmd/schema/models/wishlist.go`, registered in `All()` in
   [`models.go`](cmd/schema/models/models.go), then
   `make db-diff name=add_wishlist` and `make migrate-up`.
2. **Repository** — `internal/repositories/wishlist_repository.go`: the
   `WishlistRepository` interface plus its gorm implementation. Add it to
   `Store` and `NewStore`.
3. **DTOs** — `internal/dto/wishlist.go`: request and response shapes.
4. **Service** — `internal/services/wishlist_service.go`, taking
   `repositories.WishlistRepository`.
5. **Handler** — `internal/handlers/wishlist.go`: a struct holding the service,
   one thin method per route. Add it to the `Registry`.
6. **Routes** — add `registerWishlistRoutes` in
   [`router.go`](internal/server/router.go) and call it from `NewRouter`,
   behind the role guard if it is not public.
7. **Wire it up** — three lines in
   [`container.go`](internal/config/app/container.go).
