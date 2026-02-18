# Feature Documentation

> Pattern for documenting individual features. Create one file per feature.

---

## How to Use This Directory

For each significant feature, create a file named:

```
docs/features/FEATURE_NAME.md
```

Use lowercase with hyphens for the filename (e.g., `user-authentication.md`, `item-management.md`).

---

## Feature Document Template

Copy this template when creating a new feature doc:

```markdown
# Feature: [Feature Name]

> One-line description of what this feature does.

---

## Overview

**Status:** 🔴 Not Started / 🟡 In Progress / 🟢 Complete

**Priority:** High / Medium / Low

**Estimated Size:** S / M / L / XL

**Recommended Model:** Opus / Sonnet / Haiku

---

## User Story

As a [type of user], I want to [action] so that [benefit].

---

## Requirements

### Functional

- [ ] Requirement 1
- [ ] Requirement 2
- [ ] Requirement 3

### Non-Functional

- [ ] Performance: [specific requirement]
- [ ] Security: [specific requirement]
- [ ] Accessibility: [specific requirement]

---

## Design

### User Flow

1. User does X
2. System responds with Y
3. User sees Z

### UI Components

- Component A: [description]
- Component B: [description]

### API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/...` | ... |
| POST | `/api/...` | ... |

### Data Model

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | UUID | PK | ... |
| ... | ... | ... | ... |

---

## Implementation Notes

<!-- Technical details, gotchas, decisions -->

---

## Testing

### Test Cases

- [ ] Test case 1
- [ ] Test case 2

### Edge Cases

- Edge case 1: [how to handle]
- Edge case 2: [how to handle]

---

## Open Questions

- [ ] Question 1
- [ ] Question 2

---

## Related

- Links to related features
- Links to external documentation
```

---

## Guidelines

1. **Create feature docs before implementation.** This forces thinking through requirements.

2. **Keep them updated.** If implementation diverges, update the doc.

3. **Link from TASKS.md.** Reference feature docs when listing tasks.

4. **Include enough detail for handoff.** Another agent (or future you) should understand the feature from this doc alone.

5. **Don't over-document.** Small features may not need a full doc. Use judgment.

---

## Example Features That Should Have Docs

- User authentication
- Core CRUD operations
- File upload handling
- Notification system
- Search functionality
- Reporting/analytics
- Third-party integrations

---

## Example Features That Probably Don't Need Docs

- Basic layout components
- Simple utility functions
- Configuration changes
- Minor UI tweaks
