# Architecture

> System design, component relationships, and data flow.

---

## Overview

{{PROJECT_NAME}} follows a standard three-tier architecture:

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Web UI    │────▶│   Go API    │────▶│  PostgreSQL │
│   (React)   │◀────│   (Chi)     │◀────│     16      │
└─────────────┘     └─────────────┘     └─────────────┘
     :3000              :8080               :5432
```

---

## Components

### Web (React)

**Responsibility:** User interface, client-side state, API communication.

**Key Patterns:**
- Pages handle routing and layout
- Components are reusable UI elements
- React Query manages server state (fetching, caching, sync)
- Zustand manages client state (UI state, ephemeral data)
- Tailwind CSS for styling, design tokens in `tailwind.config.js`

**Boundaries:**
- Never accesses database directly
- All data flows through the API
- No business logic; validation only for UX

### API (Go)

**Responsibility:** Business logic, data validation, database access, authentication.

**Key Patterns:**
- Chi router for HTTP routing and middleware
- Handlers receive requests, call services, return responses
- Models define domain types
- Database package contains sqlc-generated code
- Middleware handles cross-cutting concerns (auth, logging, CORS)

**Boundaries:**
- Single source of truth for business rules
- All inputs validated and sanitized
- All database access through sqlc (parameterized queries)

### Database (PostgreSQL)

**Responsibility:** Data persistence, referential integrity, transactions.

**Key Patterns:**
- Schema defined in `api/sql/schema/`
- Queries defined in `api/sql/queries/`
- sqlc generates type-safe Go code
- Migrations applied in order

**Boundaries:**
- No business logic in stored procedures
- Constraints enforce data integrity
- Accessed only through the API

---

## Data Flow

### Read Path (Example: Fetch Items)

```
1. User navigates to /items
2. React Query triggers GET /api/items
3. Handler calls database.GetItems()
4. sqlc executes SELECT query
5. Results returned as typed structs
6. Handler serializes to JSON
7. React Query caches response
8. Component renders data
```

### Write Path (Example: Create Item)

```
1. User submits form
2. React sends POST /api/items with JSON body
3. Handler validates input
4. Handler calls database.CreateItem()
5. sqlc executes INSERT query
6. New ID returned
7. Handler returns 201 with created item
8. React Query invalidates items cache
9. UI updates
```

---

## External Integrations

<!-- FILL IN: List any external APIs, services, or systems this project integrates with -->

| Integration | Purpose | Authentication |
|-------------|---------|----------------|
| None yet | — | — |

---

## Security Boundaries

### Trust Boundaries

```
┌─────────────────────────────────────────────┐
│                 UNTRUSTED                   │
│  Browser ─────────────────────────────────  │
└─────────────────────┬───────────────────────┘
                      │ HTTPS
┌─────────────────────▼───────────────────────┐
│                  TRUSTED                    │
│  API (validates all input) ───────────────  │
│           │                                 │
│           ▼                                 │
│  Database (parameterized queries only)      │
└─────────────────────────────────────────────┘
```

### Principles

1. **Never trust client input.** Validate everything in handlers.
2. **Parameterized queries only.** sqlc enforces this.
3. **Principle of least privilege.** Database user has minimal permissions.
4. **Secrets in environment.** Never in code or version control.

---

## Design Principles

### Privacy by Default

All design and implementation decisions assume privacy and GDPR compliance as baseline requirements, not afterthoughts.

- Data collection is minimal and purposeful
- Personal data has explicit retention policies
- User consent is obtained before processing
- Data subject rights (access, deletion, portability) are built into the architecture
- Cross-border data transfer requirements are considered

### Security by Default

Security is not a feature — it is a property of the system. The architecture assumes hostile input and untrusted environments.

- OWASP Top 10 vulnerabilities are architectural failures
- MITRE CWE Top 25 weaknesses are treated as defects
- Known security anti-patterns are rejected at design time
- Defense in depth: no single point of security failure

### SOLID Principles

The codebase adheres strictly to SOLID principles:

- **S**ingle Responsibility: Each module has one reason to change
- **O**pen/Closed: Open for extension, closed for modification
- **L**iskov Substitution: Subtypes are substitutable for their base types
- **I**nterface Segregation: Many specific interfaces over one general interface
- **D**ependency Inversion: Depend on abstractions, not concretions

Exceptions to SOLID require explicit justification and documentation.

---

## Scaling Considerations

<!-- FILL IN: Notes on scaling when relevant -->

Current design assumes single instance. Future considerations:

- Stateless API allows horizontal scaling
- Database connection pooling via pgx
- Consider read replicas if read-heavy
- Consider caching layer if needed

---

## Decision Log

<!-- FILL IN: Record significant architectural decisions here -->

| Date | Decision | Rationale |
|------|----------|-----------|
| — | Template initialized | Starting point |
