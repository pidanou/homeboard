# CLAUDE.md

This file provides guidance to Claude Code when working in this repository.

---

## Collaboration Style

Act as an expert consultant — not an order-taker. This means:
- Give honest opinions, even when they contradict what the user just did or asked for
- If a design decision, architecture choice, or implementation approach is wrong or inconsistent, say so directly before (or instead of) executing it
- When validating a user's idea, only confirm it if you genuinely agree — don't rubber-stamp
- Recommend the better path, explain why briefly, then let the user decide
- This applies across all domains in this project: UX/UI, backend architecture, API design, database modeling, frontend patterns

---

## Project Overview

**Family Board** — a self-hostable family wall app. A family can share a calendar, tasks, and (eventually) more. The UI is a SvelteKit PWA (web + mobile/tablet via browser). Designed to be self-hostable first; SaaS compatibility must be kept in mind during every architectural decision.

> **Native mobile note:** The preferred path for native distribution is **Capacitor** — it wraps the existing SvelteKit app in a native shell (iOS + Android) without a rewrite. This gives access to native push notifications, haptics, and App Store distribution while keeping one codebase. Flutter was considered but deferred; the backend API is fully decoupled so either path remains viable.

---

## Architecture

```
family-board/
├── backend/        # Go API server (business logic + Postgres)
├── web/            # SvelteKit PWA
└── docker-compose.yml
```

### Backend (`backend/` — Go)
- REST API under `/api/v1/` — the frontend never touches the database or auth provider directly
- The Go layer owns all business logic, tenant scoping, and validation
- Structured as: `cmd/`, `internal/handler/`, `internal/service/`, `internal/repository/`, `internal/model/`
- Repository interfaces are defined against plain Go types — the database is an implementation detail behind those interfaces
- **Current implementation:** raw PostgreSQL via `pgx`, custom JWT auth
- **Future SaaS path:** swap repository implementations to use Supabase (Postgres + GoTrue + Realtime) — no frontend or API contract changes required

### Web (`web/` — SvelteKit PWA)
- Talks exclusively to the Go backend via REST
- Real-time updates via Server-Sent Events (SSE) from the Go backend
- PWA: installable on iOS/Android/desktop, offline-capable via service worker
- Uses `@vite-pwa/sveltekit` for PWA support
- Mobile-responsive; serves as both the web and mobile experience

### Database
- PostgreSQL (plain, no Supabase)
- Migrations managed with `golang-migrate` (SQL files in `backend/migrations/`)
- All dates stored as UTC, ISO 8601

### Auth
- Custom JWT auth in Go (email/password + OAuth: Google, Apple)
- JWTs verified server-side on every request
- **Future SaaS path:** replace with Supabase GoTrue behind the same auth interface

### Deployment
- `docker-compose.yml` spins up the full stack (Go backend + PostgreSQL + web)
- Must remain self-hostable with a single `docker compose up`
- SaaS readiness: multi-tenant model (family = tenant) built in from day one

---

## Commands

### Development

```bash
# Backend
cd backend && go run ./cmd/server

# Web
cd web && npm run dev
```

### Build & Deploy

```bash
# Full stack (self-hosted)
docker compose up --build

# Backend only
cd backend && go build -o bin/server ./cmd/server

# Web
cd web && npm run build
```

---

## Code Style

### Go (backend)
- Follow standard Go conventions (`gofmt`, `golangci-lint`)
- Errors returned, not panicked — handle every error explicitly
- Use `context.Context` for cancellation and deadlines
- Repository pattern: interfaces in `internal/repository/`, implementations in `internal/repository/postgres/`
- No `any` unless unavoidable; use typed structs

### TypeScript / Svelte (web)
- Strict TypeScript; avoid `any`
- Prefer `const` over `let`; never use `var`
- Named exports over default exports
- One primary export per file, named after the file
- Use SvelteKit's `load` functions for data fetching; avoid client-side fetching on page load

### General
- Keep functions small and single-purpose
- Write self-documenting code; comments only for non-obvious logic

---

## Testing

- Write tests alongside new features, not after
- **Go**: unit tests with `testing` package; integration tests hit a real Postgres instance (no mocks for DB)
- **Svelte**: component tests with Vitest + Testing Library
- Test file naming: `*_test.go`, `*.test.ts`
- Aim for meaningful coverage, not 100%

---

## Git

- Commits: imperative mood, present tense — `Add feature` not `Added feature`
- Keep commits atomic and focused
- Branch naming: `feat/`, `fix/`, `chore/`, `refactor/`
- No direct commits to `main`; open a PR

---

## Environment

- Secrets and config go in `.env` (never committed)
- See `.env.example` for required variables
- Configs: `development`, `test`, `production`
- Key env vars: `DATABASE_URL`, `JWT_SECRET`, `OAUTH_GOOGLE_CLIENT_ID`, `OAUTH_APPLE_CLIENT_ID`, `APP_BASE_URL`

---

## Design System

### Stack
- **CSS:** Tailwind CSS
- **Components:** shadcn-svelte (copy-owned components, not a runtime dependency)
- **Icons:** Lucide Svelte (consistent with shadcn-svelte defaults)

### Visual Style
- Warm and friendly — soft colors, approachable feel, suited for a family audience
- Balanced spacing and rounding — `md`/`lg` border radii, comfortable padding, not too dense or too bubbly
- Light + dark mode, following system preference via Tailwind's `dark:` variant

### Typography
- **Font:** Inter (sans-serif) for all UI text
- Scale follows Tailwind defaults; headings use `font-semibold` or `font-bold`

### Color Palette
- Define a custom Tailwind palette in `tailwind.config.ts` with semantic tokens: `primary`, `secondary`, `surface`, `muted`, `destructive`
- Colors must work in both light and dark mode — use CSS variables via shadcn-svelte's theming convention
- Warm neutrals for backgrounds (avoid pure white/black); a soft accent color for primary actions

### Component Conventions
- **Always use a shadcn-svelte component if one exists for the UI element.** Check https://www.shadcn-svelte.com/docs/components before writing custom markup for buttons, inputs, labels, cards, dialogs, selects, etc.
- Add new shadcn components via the CLI: `npx shadcn-svelte@latest add <component>` — never hand-edit files in `ui/`
- All components live in `web/src/lib/components/`
- shadcn-svelte components go in `web/src/lib/components/ui/`
- Custom app components go in `web/src/lib/components/` alongside ui/
- Props should be typed with TypeScript interfaces
- Prefer composition over wrapping — extend shadcn primitives, don't re-wrap them

### Responsive Design
- Mobile-first — base styles target small screens, `md:` and `lg:` for larger
- Minimum tap target: 44×44px for interactive elements
- Test layouts at 375px (phone), 768px (tablet), 1280px (desktop)

---

## Key Conventions

- API routes: REST under `/api/v1/`, e.g. `/api/v1/families`, `/api/v1/calendar/events`
- Every resource is scoped to a **family** (the tenant unit) — always validate family membership in handlers
- WebSocket events follow a `type` + `payload` envelope: `{ "type": "event.created", "payload": { ... } }`
- All dates stored and transmitted as UTC ISO 8601 strings
- Database models in `internal/model/`, HTTP request/response DTOs in `internal/handler/dto/`
- Feature flags or tenant-level config live in a `settings` table — don't hardcode feature availability
- **Tasks with a due date must appear on the calendar view** on their due date alongside calendar events

---

## Specs

Feature specs live in `docs/specs/`. When implementing or modifying a feature that has a spec:

- Update the spec's **Delta vs current implementation** table to reflect what has been built
- Move implemented items out of **What's out of scope** if they've been built
- Add new out-of-scope items discovered during implementation
- If a spec does not exist yet for the feature being worked on, create one before or alongside the implementation

---

## Roadmap

The implementation roadmap lives in `docs/roadmap.md`. Keep it current:

- Mark items ✅ as they are completed
- Mark items 🚧 when actively being worked on
- When a full milestone is done, mark the milestone header ✅
- When promoting a deferred item to a milestone, add it to the relevant milestone (or create a new one)
- Do not remove items — completed items stay in the roadmap as history

---

## Do Not

- Do not commit secrets, API keys, or credentials
- Do not modify generated files (mark them with `// @generated`)
- Do not add dependencies without discussion
- Do not break existing public APIs without a deprecation path
- Do not skip error handling
- Do not store tenant (family) data in a way that would make multi-tenancy hard to retrofit
- Do not couple the web or mobile apps to self-hosting assumptions (e.g. hardcoded `localhost` URLs)
- Do not access the database or auth provider from the frontend — everything goes through the Go API
- Do not put database-specific code outside of `internal/repository/` — keep it behind the repository interface boundary
