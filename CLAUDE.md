# gamenglam-server — CLAUDE.md

Go modular monolith backing the Aldine / GamenGlam e-commerce platform.
Stack: **Gin · pgx/v5 · PostgreSQL · JWT (RS256)**

---

## Quick start

```bash
go run cmd/server/main.go        # dev server (reads .env)
go build -o bin/server ./cmd/server
go test ./...
```

Copy `.env.example` → `.env` and set `DATABASE_URL`, `JWT_PRIVATE_KEY_PATH`, `JWT_PUBLIC_KEY_PATH`.

---

## Repository layout

```
cmd/server/          # main.go entry point
database/
  migrate.go         # 5-step migration runner
  seeds/             # seed SQL files
internal/
  module/
    auth/            # JWT issue/verify, admin login, forgot/reset/change password
    catalog/         # products, variants, categories, attributes, attribute values, variant images
      order.adapter.go   # CatalogProvider impl for order module
    customer/        # customers, addresses, customer auth (register/login/logout/refresh)
      order.adapter.go   # CustomerProvider impl for order module
    employee/        # sections, pages, roles (JSONB permissions), users (RBAC staff management)
    order/           # cart, orders, order status history
      models/        # request/response structs
      services/      # business logic (cart.service.go, order.service.go)
      repository/postgres/  # DB layer (cart, order repos)
  middleware/        # JWT auth, CORS, etc.
  shared/            # shared types, errors, pagination
```

---

## Database conventions

### Schemas
All product/catalog tables live in the **`catalog`** schema — never `product`.
Other schemas: `auth`, `customer`, `orders`.

### Dual-ID pattern (critical)
Every entity has two identifiers:

| Field | Type | Purpose |
|-------|------|---------|
| `*_no` | `SERIAL` / `uint32` | Internal PK, used in JOINs and FK columns |
| `*_id` | `UUID` | External identifier exposed in **all APIs** |

**Always use `_id` (UUID) in API request/response. Never expose `_no` serials.**

Repositories resolve `_id → _no` internally when the DB schema stores `_no` in FK columns.

Example in cart repository:
```go
// resolve variant_id → variant_no inside the transaction
var variantNo uint32
err = tx.QueryRow(ctx, `
    SELECT variant_no FROM catalog.product_variants
    WHERE variant_id = $1 AND status != 'D'
`, req.VariantId).Scan(&variantNo)
```

### Soft deletes
All tables use `status` column. Active records have `status != 'D'`.
**Always add `WHERE status != 'D'` to every SELECT.**

### Price priority
```sql
COALESCE(selling_price, mrp, cost_price, 0) AS price
```
Always apply this order when resolving a variant's price.

---

## Coding patterns

### ACID write pattern
```go
tx, err := r.db.Begin(ctx)
defer func() { if err != nil { _ = tx.Rollback(ctx) } }()

// SELECT FOR UPDATE to lock the row
err = tx.QueryRow(ctx, `SELECT ... FOR UPDATE`, ...).Scan(...)

// mutate
err = tx.QueryRow(ctx, `UPDATE ... RETURNING ...`, ...).Scan(...)

err = tx.Commit(ctx)
```

### Error wrapping
```go
return nil, fmt.Errorf("repoName.MethodName: %w", err)
// e.g.:
return nil, fmt.Errorf("cartRepository.AddItem: %w", err)
```

### Not-found convention
Repositories return `nil, nil` when a row is not found (never return a 404 error from the repo layer). Services convert `nil, nil` → 404 response.

### Repository scan discipline
SELECT column count **must** exactly match Scan destination count.
When adding/removing a SELECT column, update the matching Scan call in the same edit.

---

## Cross-module adapter pattern

Modules don't import each other's packages directly. Instead:

1. The **consuming** module defines a provider interface in `services/interfaces.go`
2. The **providing** module implements that interface in its own `order.adapter.go`
3. The adapter is wired up at app startup via dependency injection

```go
// order/services/interfaces.go
type CatalogProvider interface {
    GetVariantById(ctx context.Context, id uuid.UUID) (*catalogModels.ProductVariant, error)
}

// catalog/order.adapter.go
type catalogAdapter struct { svc ProductVariantService }
func (a *catalogAdapter) GetVariantById(ctx context.Context, id uuid.UUID) (*models.ProductVariant, error) {
    return a.svc.FindOneById(ctx, id)
}
```

### Current provider interfaces

**`CatalogProvider`** (in order module):
- `GetVariantById(ctx, uuid.UUID) (*catalogModels.ProductVariant, error)`

**`CustomerProvider`** (in order module):
- `GetAddressByNo(ctx, uint32) (*customerModels.Address, error)`
- `GetAddressByID(ctx, uuid.UUID) (*customerModels.Address, error)`

---

## Key model fields

### `AddCartItemRequest`
```go
VariantId uuid.UUID
Quantity   int
```

### `PlaceOrderItemRequest`
```go
VariantId  uuid.UUID
Quantity   int
UnitPrice  float64
```

---

## Migration runner (`database/migrate.go`)

Five steps run in order:
1. `createSchema` — create schemas if not exists
2. `createEnums` — create PG enum types (checked via `pg_namespace` join for schema qualification)
3. `createTables` — DDL for all tables
4. `addForeignKeys` — ALTER TABLE … ADD CONSTRAINT
5. `seedData` — insert default/reference data

Run automatically on server start (idempotent).

---

## Common mistakes to avoid

- **Never use `_no` in API payloads** — always `_id` (UUID).
- **Never query the `product` schema** — all catalog tables are in `catalog`.
- **Don't forget `WHERE status != 'D'`** in every SELECT.
- **Don't call `GetVariantByNo` from order module** — that method was removed. Use `GetVariantById`.
- **Column count must match Scan count** — mismatch causes a runtime panic (`number of field descriptions must equal number of destinations`).
- **Don't import catalog/customer packages from order** — use the adapter/provider pattern.
