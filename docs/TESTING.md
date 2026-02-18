# Testing Strategy

> Philosophy, patterns, and practices for useful tests.

---

## Philosophy

**Tests exist to provide value, not coverage numbers.**

The goal is confidence that the system works correctly and continues to work through refactoring and prototyping. Every test should earn its place by either:

1. Protecting an interface contract (API, UI, module boundary)
2. Catching regressions in functionality we want to keep
3. Documenting expected behavior

Coverage metrics are a secondary signal, not a target. A codebase with 60% meaningful coverage is better than 90% coverage padded with trivial tests.

---

## Core Principles

### Test Interfaces, Not Implementations

Focus testing effort on boundaries:

- **External APIs** — HTTP endpoints, request/response contracts
- **Human-Machine Interfaces** — UI components, user interactions
- **Module boundaries** — Public functions of internal packages

Implementation details should be free to change without breaking tests. If a refactor breaks tests but not functionality, the tests were testing the wrong thing.

### Protect Against Regressions

This codebase will undergo significant refactoring and prototyping. Tests are the safety net that lets us move fast without breaking things we want to keep.

When something works and should keep working:
1. Write a test that verifies the behavior
2. That test now guards against regression
3. Refactor freely

### Test-Driven Development

Write tests first:

1. **Red** — Write a failing test that describes desired behavior
2. **Green** — Write minimal code to make it pass
3. **Refactor** — Clean up while keeping tests green

TDD forces clarity about what we're building before we build it. The test becomes both specification and verification.

---

## Test Types

### Unit Tests (During Development)

**When:** Write these as you develop, following TDD.

**What to test:**
- Pure functions and business logic
- Edge cases in data transformation
- Validation rules
- Error handling paths

**What not to test:**
- Simple getters/setters
- Framework boilerplate
- Implementation details that may change

**Agent responsibility:** Write unit tests for new functionality during implementation.

### Integration Tests (Post-Validation)

**When:** After the product/feature is validated and worth keeping.

**What to test:**
- API endpoint contracts (request → response)
- Database queries return expected data
- Service interactions work correctly
- Authentication/authorization flows

**Scope:** These are a separate effort from feature development. Prompt explicitly when ready to add integration test coverage.

### End-to-End Tests (Post-Validation)

**When:** After the product is validated and core flows are stable.

**What to test:**
- Critical user journeys
- Flows involving multiple components
- Scenarios that have broken in production

**Scope:** E2E tests are expensive to write and maintain. Reserve for high-value paths. This is a separate prompt/phase from development.

---

## Mocking Strategy

**Default: Minimal mocking.** Use real dependencies where practical.

Real dependencies catch integration issues that mocks hide. A test with real database calls verifies more than one with mocked queries.

**When to mock:**

| Situation | Approach |
|-----------|----------|
| External APIs (third-party services) | Mock or use sandbox/test instances |
| Expensive operations (payment processing) | Mock |
| Non-deterministic behavior (time, randomness) | Inject controllable implementations |
| Slow dependencies impacting dev cycle | Consider test instances |
| Flaky external services | Mock for reliability |

**When not to mock:**

- Database queries (use test database)
- Internal module interactions
- Anything where the integration itself is what we're verifying

---

## Edge Cases

Focus edge case testing on scenarios that:

1. **Have happened** — Bugs that occurred in practice
2. **Would be costly** — Data corruption, security issues, financial errors
3. **Are likely** — Common user mistakes, boundary conditions
4. **Are subtle** — Off-by-one, empty inputs, Unicode, timezones

Use judgment. Not every edge case needs a test. Prioritize:

- Boundaries of valid input ranges
- Empty/null/missing data
- Concurrent access (if applicable)
- Error recovery paths
- Permission boundaries

---

## Mandatory Enforcement

Certain violations are not negotiable. Tests and builds must fail when these constraints are broken.

### Privacy and GDPR Compliance

Tests must fail when patterns that break GDPR or privacy principles are detected:

- Personal data stored without documented purpose
- Missing or inadequate consent mechanisms
- Data retained beyond defined retention periods
- Personal data exposed in logs or error messages
- Missing data subject rights implementation (where required)

### Security Standards

Builds must fail when OWASP Top 10 or MITRE CWE Top 25 violations are detected:

- Injection vulnerabilities (SQL, command, etc.)
- Broken authentication or session management
- Sensitive data exposure
- Security misconfiguration
- Known vulnerable dependencies (enforced via SCA)

Known security anti-patterns fail tests regardless of context.

### SOLID Violations

Tests must fail when SOLID principles are violated:

- Classes/modules with multiple responsibilities
- Modifications required to extend functionality
- Interface implementations that throw "not supported"
- Fat interfaces forcing unused dependencies
- Concrete dependencies where abstractions should exist

Static analysis and code review should catch these. When violations are detected, they block merge.

---

## Recommended Frameworks

Chosen for simplicity, AI-friendliness, and token efficiency.

### Go

**Testing:** Built-in `testing` package

- No external dependencies
- Agents know it extremely well
- Table-driven tests are idiomatic and token-efficient

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid", "user@example.com", false},
        {"missing @", "userexample.com", true},
        {"empty", "", true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateEmail(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
            }
        })
    }
}
```

**Assertions (optional):** `testify/assert` if you prefer assertion style

- Widely used, agents handle it well
- Cleaner failure messages

### React

**Testing:** Vitest + React Testing Library

- Vitest: Fast, Vite-native, Jest-compatible API
- React Testing Library: Tests behavior, not implementation
- Both are well-represented in training data

```typescript
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ItemList } from './ItemList'

test('displays items', () => {
  render(<ItemList items={[{ id: 1, name: 'Test' }]} />)
  expect(screen.getByText('Test')).toBeInTheDocument()
})

test('calls onDelete when delete button clicked', async () => {
  const onDelete = vi.fn()
  render(<ItemList items={[{ id: 1, name: 'Test' }]} onDelete={onDelete} />)
  
  await userEvent.click(screen.getByRole('button', { name: /delete/i }))
  
  expect(onDelete).toHaveBeenCalledWith(1)
})
```

### Integration / E2E

**API testing:** Go's `net/http/httptest` for in-process testing

**E2E (when needed):** Playwright

- Cross-browser, reliable
- Good TypeScript support
- Reasonable agent familiarity

---

## Test Organization

### Go

```
api/
  internal/
    handlers/
      items.go
      items_test.go      # Unit tests alongside code
    models/
      item.go
      item_test.go
  integration_test/       # Integration tests (separate package)
    api_test.go
```

### React

```
web/
  src/
    components/
      ItemList.tsx
      ItemList.test.tsx   # Unit tests alongside components
    hooks/
      useItems.ts
      useItems.test.ts
  e2e/                    # E2E tests (separate directory)
    items.spec.ts
```

---

## Agent Instructions

### During Feature Development

1. Follow TDD: write failing test → implement → refactor
2. Test public interfaces and edge cases
3. Use real dependencies unless there's a specific reason to mock
4. Keep tests focused — one behavior per test

### When Asked for Integration/E2E Tests

This is a separate phase. When prompted:

1. Identify critical paths and contracts to verify
2. Set up test infrastructure if needed
3. Write tests for API contracts and user flows
4. Focus on regression protection, not coverage metrics

### Test Quality Checklist

Before considering a test complete:

- [ ] Test fails when the behavior breaks
- [ ] Test name describes the expected behavior
- [ ] Test is independent (no order dependencies)
- [ ] Test is deterministic (same result every run)
- [ ] Test is fast enough to run frequently

---

## Notes

- Tests are documentation. Someone reading them should understand what the code does.
- Delete tests that no longer provide value (testing dead code, obsolete features).
- Flaky tests are worse than no tests. Fix or remove them.
