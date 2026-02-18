# Template README

**Author:** Einar Sundgren  
**Contact:** mail@einarsundgren.se

---

## What Is This?

A documentation-first project template optimized for agentic coding with Claude Code. Designed to minimize token usage while maintaining clarity, security, and maintainability.

The structure allows Claude to read only what's relevant to the current task rather than loading entire codebases into context.

---

## Tech Stack (Pre-configured)

| Layer | Technology | Rationale |
|-------|------------|-----------|
| Backend | Go + Chi | Minimal abstraction, stdlib-compatible, token-efficient |
| Database | PostgreSQL 16 + sqlc | SQL as source of truth, compile-time type safety |
| Frontend | React + Vite | Fast dev server, standard tooling |
| State | React Query + Zustand | Server state and client state separated cleanly |
| Styling | Tailwind CSS | Colocated styles, design tokens in config |
| Containers | Podman Compose | Go + Postgres + Vite dev server |

---

## Setup Checklist

When starting a new project from this template:

### 1. Replace Placeholders

Find and replace `{{PROJECT_NAME}}` throughout all files:

```bash
grep -r "{{PROJECT_NAME}}" .
```

Replace with your actual project name (lowercase, no spaces recommended).

### 2. Fill In Project-Specific Documentation

These files have placeholder sections marked with `<!-- FILL IN -->`:

| File | What to fill in |
|------|-----------------|
| `CLAUDE.md` | Current focus, project-specific quick commands |
| `docs/ARCHITECTURE.md` | Your specific components, data flows, external integrations |
| `docs/DESIGN_SYSTEM.md` | Colors, typography, spacing values, component patterns |
| `docs/STATE.md` | Initial project state, first milestones |
| `docs/TASKS.md` | Your backlog, prioritized tasks |

### 3. Define the Project

In `CLAUDE.md`, fill in the `## Project Definition` section:

- What is this project?
- Who is it for?
- What problem does it solve?
- What are the success criteria?

### 4. Create Initial Plan

In `docs/TASKS.md`, create your initial backlog. Use the model assessment pattern:

- Break work into phases
- For each phase, propose which model (Opus/Sonnet/Haiku) and why
- Estimate complexity (S/M/L/XL)

### 5. Initialize the Codebase

```bash
# Initialize Go module
cd api && go mod init {{PROJECT_NAME}} && cd ..

# Initialize React app (or let agent do this)
cd web && npm create vite@latest . -- --template react && cd ..

# Start containers
podman-compose up -d db
```

---

## Agent Interaction Pattern

This template enforces a specific workflow:

1. **Assess** — Agent determines which model is appropriate for the task
2. **Execute** — Agent completes the task
3. **Stop** — Agent stops and summarizes what was done
4. **Prompt** — Agent provides a continuation prompt for the next task/model

This prevents runaway token usage and allows you to switch to cheaper models for routine work.

---

## Directory Structure

```
{{PROJECT_NAME}}/
├── CLAUDE.md                # Agent entry point (start here)
├── TEMPLATE_README.md       # This file (human reference)
├── docs/
│   ├── ARCHITECTURE.md      # System design, component relationships
│   ├── TECH_STACK.md        # Technology decisions (pre-filled)
│   ├── DESIGN_SYSTEM.md     # Visual language (fill in per project)
│   ├── TESTING.md           # Test strategy and patterns (pre-filled)
│   ├── CICD.md              # Pipeline and security gates (TBD placeholder)
│   ├── STATE.md             # Current progress, blockers
│   ├── TASKS.md             # Backlog with estimates
│   └── features/            # Feature specifications
│       └── README.md        # Pattern for feature docs
├── api/                     # Go backend
│   ├── cmd/server/          # Entry point
│   ├── internal/            # Private packages
│   │   ├── handlers/        # HTTP handlers
│   │   ├── models/          # Domain types
│   │   ├── database/        # DB access (sqlc generated)
│   │   └── middleware/      # HTTP middleware
│   ├── sql/
│   │   ├── queries/         # sqlc query files
│   │   └── schema/          # Migration files
│   └── sqlc.yaml            # sqlc configuration
├── web/                     # React frontend
│   └── src/
│       ├── components/      # Reusable UI components
│       ├── pages/           # Route-level components
│       ├── hooks/           # Custom React hooks
│       ├── stores/          # Zustand stores
│       └── api/             # API client (React Query)
├── podman-compose.yaml      # Local development setup
└── .gitignore
```

---

## Architectural Qualities (Priority Order)

All decisions should optimize for these qualities in order:

1. **Security** — Secure by default, explicit trust boundaries
2. **Clarity** — Obvious over clever, explicit over implicit
3. **Extendability** — Easy to add features without rewriting
4. **Maintainability** — Easy to understand in 6 months

---

## Notes

- Keep this file for reference; it documents the template's intent
- The `docs/` folder is the primary context for agents — keep it updated
- When in doubt, add documentation rather than comments in code
