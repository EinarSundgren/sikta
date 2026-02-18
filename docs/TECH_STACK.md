# Tech Stack

> Technology choices and rationale. Optimized for AI-assisted development.

---

## Selection Criteria

All technology choices optimize for:

1. **Token efficiency** — Minimal API surface, less to explain in context
2. **Explicitness** — No magic, behavior is visible in code
3. **Agent reliability** — Well-represented in training data, predictable patterns
4. **Maintainability** — Easy to understand, debug, and extend

---

## Backend

### Language: Go

**Version:** 1.21+

**Rationale:**
- Strong typing catches errors at compile time
- Simple language, small spec — agents handle it reliably
- Excellent standard library reduces dependencies
- Single binary deployment

### Router: Chi

**Rationale:**
- Thin layer over `net/http`, almost no abstraction
- 100% compatible with stdlib middleware
- Route grouping and URL parameters without framework lock-in
- Agents understand it as well as they understand stdlib

**Patterns:**
```go
r := chi.NewRouter()
r.Use(middleware.Logger)
r.Route("/api", func(r chi.Router) {
    r.Get("/items", handlers.ListItems)
    r.Post("/items", handlers.CreateItem)
})
```

### Database Access: sqlc

**Rationale:**
- SQL is the source of truth (no ORM translation)
- Compile-time type safety
- Generated code is plain Go, easy to read
- Agents see actual queries, not abstracted methods
- Parameterized queries by design (SQL injection safe)

**Workflow:**
1. Define schema in `api/sql/schema/*.sql`
2. Define queries in `api/sql/queries/*.sql`
3. Run `sqlc generate`
4. Use generated Go code in handlers

### Database Driver: pgx

**Rationale:**
- Most performant PostgreSQL driver for Go
- Used by sqlc under the hood
- Connection pooling built in

---

## Frontend

### Framework: React 18+

**Rationale:**
- Dominant market share, agents know it extremely well
- Functional components with hooks are predictable
- Large ecosystem if needed

### Build Tool: Vite

**Rationale:**
- Fast development server with HMR
- Simple configuration
- ES modules native

### Server State: React Query (TanStack Query)

**Rationale:**
- Separates server state from UI state
- Handles caching, refetching, synchronization
- Small, predictable API
- Reduces boilerplate for data fetching

**Patterns:**
```typescript
const { data, isLoading } = useQuery({
  queryKey: ['items'],
  queryFn: () => api.getItems()
})
```

### Client State: Zustand

**Rationale:**
- Tiny API (~5 functions to learn)
- No boilerplate, no reducers, no actions
- Works outside React components if needed
- Extremely token-efficient to describe

**Patterns:**
```typescript
const useStore = create((set) => ({
  count: 0,
  increment: () => set((s) => ({ count: s.count + 1 }))
}))
```

### Styling: Tailwind CSS

**Rationale:**
- Styles colocated with markup (agent reads one file)
- Design tokens in `tailwind.config.js`
- Utility classes are dense and explicit
- No context switching between files

**Patterns:**
```jsx
<button className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600">
  Click me
</button>
```

---

## Database

### PostgreSQL 16

**Rationale:**
- Industry standard, extremely well documented
- Agents have extensive training data
- Robust, battle-tested
- Good JSON support if needed

---

## Infrastructure

### Containers: Podman

**Rationale:**
- Docker-compatible, no daemon
- Rootless by default (better security)
- Works with docker-compose syntax

### Local Development: Podman Compose

Services:
- `api` — Go backend (manual rebuild)
- `web` — React/Vite dev server (hot reload via volume mount)
- `db` — PostgreSQL 16

---

## Not Using (And Why)

| Technology | Why Not |
|------------|---------|
| GORM | ORM magic, agents hallucinate wrong patterns, hard to debug |
| Redux | Large API surface, boilerplate, token-expensive |
| styled-components | Runtime overhead, CSS-in-JS complexity |
| Gin/Echo | More opinionated than Chi, larger API surface |
| MongoDB | Less structured, PostgreSQL covers our needs |

---

## Version Pinning

Pin major versions in:
- `go.mod` for Go dependencies
- `package.json` for Node dependencies
- `podman-compose.yaml` for container images

Avoid `latest` tags in production.
