# Architecture

The API is layered. A request always travels in one direction, and each layer
only knows about the one below it:

```
                        HTTP request
                             |
                             v
   +-- the HTTP layer: knows gin, knows no SQL ----------------------+
   |  server       the route table, the Handle wrappers, the server  |
   |  middleware   CORS, authentication, the admin guard             |
   |  handlers     bind, delegate, return                            |
   +-----------------------------------------------------------------+
                             |
                             v
     services      business rules: prices, stock, ownership, tokens
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

| Package | Holds | Depends on |
| --- | --- | --- |
| `cmd/api` | The entry point. Loads config, hands over to `app`. | `app`, `config`, `logger` |
| `internal/app` | The composition root: builds the graph, runs it, shuts it down. | everything |
| `internal/server` | The router - the whole route table - the `Handle` wrappers, and the HTTP server. | `handlers`, `middleware`, `config` |
| `internal/handlers` | One handler type per area of the API. Thin: bind, delegate, return. | `services`, `middleware`, `dto` |
| `internal/middleware` | CORS, authentication, the admin guard, and who the caller is. | `config`, `models`, `utils` |
| `internal/services` | Business logic. One service per area of the domain. | `repositories`, `dto`, `apperror` |
| `internal/repositories` | Every database query, behind an interface per aggregate. | `models`, `gorm` |
| `internal/models` | The GORM structs. The database schema is generated from these. | — |
| `internal/dto` | Request and response bodies. The API contract. | — |
| `internal/providers` | Adapters for the outside world (local disk, S3). | `config` |
| `internal/config` | Environment variables, parsed once into a struct. | — |
| `internal/apperror` | Failures described in HTTP terms. | — |
| `internal/utils` | Response envelope, validation messages, JWT, hashing, pagination. | — |

## Dependency injection

Everything is constructor-injected, and the whole graph is built in one place:
[`internal/app/container.go`](internal/app/container.go).

```go
store := repositories.NewStore(db)

cartService := services.NewCartService(store.Carts, store.Products)

registry := &handlers.Registry{
    Cart: handlers.NewCartHandler(cartService),
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

- **`repositories.UserRepository`, `ProductRepository`, ...** — the database.
- **`providers.UploadProvider`** — file storage. `UPLOAD_PROVIDER=s3` swaps the
  implementation without a service noticing.

Services are deliberately **concrete structs**. A service is the code under
test, not the thing you stub out, and an interface per service would be
boilerplate the compiler cannot check against anything. If handler tests ever
need to fake one, add the interface then — it is a two-line change.

The Go convention "accept interfaces, return structs" is what the constructors
follow: `NewCartService` takes repository interfaces and returns a
`*CartService`.

## Transactions

Anything that writes to more than one table goes through `Store.Atomic`. The
store handed to the callback is bound to the transaction, so every repository
call inside it commits or rolls back as one:

```go
err := s.store.Atomic(func(tx *repositories.Store) error {
    if err := tx.Users.Create(&user); err != nil {
        return err
    }
    return tx.Carts.Create(&models.Cart{UserID: user.ID})
})
```

Checkout ([`order_service.go`](internal/services/order_service.go)) is the real
example: it prices the cart, takes stock, writes the order and empties the cart.
A failure at any point leaves the database exactly as it was.

## Adding a feature

A wishlist, end to end:

1. **Model** — `internal/models/wishlist.go`, then register it in
   [`db/loader/main.go`](db/loader/main.go) and run `make db-diff name=add_wishlist`.
2. **Repository** — `internal/repositories/wishlist_repository.go`: the
   `WishlistRepository` interface plus its gorm implementation. Add it to
   `Store` and `NewStore`.
3. **DTOs** — `internal/dto/wishlist.go`: request and response shapes.
4. **Service** — `internal/services/wishlist_service.go`, taking
   `repositories.WishlistRepository`. Map to DTOs in
   [`mapper.go`](internal/services/mapper.go).
5. **Handler** — `internal/handlers/wishlist.go`: a struct holding the service,
   one thin method per route. Add it to the `Registry`.
6. **Routes** — add `registerWishlistRoutes` in
   [`router.go`](internal/server/router.go) and call it from `NewRouter`.
7. **Wire it up** — three lines in
   [`container.go`](internal/app/container.go).

[`internal/handlers/README.md`](internal/handlers/README.md) covers steps 3-6 in
detail: validation tags, error mapping, and the response envelope.
