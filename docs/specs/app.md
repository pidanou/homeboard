# Family Board — Global App Spec

## Purpose

Family Board is a **shared wall for household life** — a single place where a family keeps track of tasks, events, and lists without fighting over a paper calendar or scattering things across WhatsApp threads, notes apps, and Google Calendar.

The app is designed for real families: two parents, kids of various ages, varying tech literacy, used across phones and a shared kitchen tablet. It must be fast to open, fast to add something, and impossible to break by accident.

---

## The five core questions the app must answer

1. **What do I need to deal with today?** → Today view
2. **What's coming up and what needs doing?** → Board
3. **What does the week / month look like?** → Calendar
4. **What do we need to buy / pack / prepare?** → Lists
5. **Who else is in the family and how do we share access?** → Settings

Every feature must serve at least one of these questions. If it doesn't, it's out of scope.

---

## Core entities

### Family
The tenant unit. All data is scoped to a family. A user can belong to multiple families (e.g. their own + an aging parent's household).

### Member
A person who belongs to a family. Members come in two kinds:

**Real member** — linked to a user account. Can log in, see the board, and act on their own behalf.

**Virtual member** — a name (+ optional avatar) created by an admin with no login account. Used for kids, or anyone who won't use the app themselves. Can be assigned tasks and added to event attendees; someone else checks things off on their behalf.

| Field | Notes |
|---|---|
| display name | required |
| avatar | optional |
| role | `admin` or `member`; virtual members have no role (they can't act) |
| user_id | nullable — null = virtual, non-null = linked to a real user account |
| email | only for real members (used for invites) |

#### Virtual → real upgrade path

When a new user accepts a family invite, if the family has **unlinked virtual members**, prompt:

> "Are you one of these people already in the family?"
>
> ○ Lucas  ○ Emma  ○ No, I'm someone new

- Name match is a hint to pre-select, not a requirement — the user picks manually
- Linking merges the new account into the existing virtual member: all task assignments, event attendees, and history are preserved, no re-assignment needed
- "No, I'm someone new" creates a fresh member record
- Admins can also manually link or unlink from Settings > Members at any time

### Task
A discrete action item.

| Field | Type | Notes |
|---|---|---|
| title | string | required |
| description | string | optional |
| status | `todo` \| `done` | default `todo` |
| priority | `high` \| `medium` \| `low` | default `medium` |
| end_date | date (UTC) | due date; optional |
| assigned_to | member ID | optional; "unassigned" by default |
| label_ids | []label ID | optional |

### Event
A time-bound happening.

| Field | Type | Notes |
|---|---|---|
| title | string | required |
| description | string | optional |
| start_at | datetime (UTC) | required |
| end_at | datetime (UTC) | required |
| all_day | bool | hides time display when true |
| location | string | optional |
| attendees | []member ID | optional |
| label_ids | []label ID | optional |

### Label
A colored tag. Scoped to the family. Can be applied to both tasks and events.

| Field | Notes |
|---|---|
| name | short text |
| color | one of a fixed palette (~10 colors) |

### List
A named checklist (shopping, packing, chores, etc.). Scoped to the family. Multiple lists per family.

### ListItem
An entry in a List. Has: name, checked bool, created_at.

---

## Navigation model

```
App root
├── /login               Auth screens (login, register, forgot password)
├── /families            Family picker (if user belongs to 0 or 1 families, skip or auto-redirect)
└── /families/[id]
    ├── /                Today
    ├── /board           Board
    ├── /calendar        Calendar
    ├── /lists           Lists
    └── /settings        Settings
        ├── /members     Member management + invite
        └── /labels      Label management
```

### Desktop: persistent left sidebar  
### Mobile: bottom tab bar (Today / Board / Calendar / Lists) + Settings icon in top bar header

---

## Auth

- Email + password login
- OAuth: Google, Apple (planned)
- Session: JWT stored in localStorage, refreshed silently
- Protected routes: redirect to `/login` if not authenticated
- Invite flow: member receives email link → clicks → registers or logs in → **virtual member linking prompt** (if applicable) → joins family

---

## Real-time sync

- Every data-mutating API call triggers an SSE broadcast to all family members
- Frontend subscribes via EventSource on mount; reconnects on error
- Optimistic updates for high-frequency interactions (checkbox toggle, list item check)
- SSE reconciles on `refresh` event

---

## Multi-tenancy rules

- Every resource is scoped to a family ID
- Every API call validates family membership server-side before reading or writing
- No cross-family data leakage is ever acceptable — the family is the security boundary

---

## API conventions

- REST under `/api/v1/`
- All dates UTC ISO 8601 strings
- Standard status codes: 200, 201, 204, 400, 401, 403, 404, 500
- Error envelope: `{ "error": "human-readable message" }`
- Pagination: cursor-based (not yet implemented; current endpoints return all records within the display window)

---

## Offline behavior

- PWA with service worker caches the app shell (HTML, CSS, JS)
- Data is not cached offline — stale data is worse than "no connection" for a family app
- When offline: show a subtle "Offline — changes won't save" banner
- On reconnect: reload data automatically

## Push notifications

- Web Push via VAPID — no Firebase, no third-party accounts
- Backend sends notifications on event create and task create to all family members
- Subscriptions stored per user per family in `push_subscriptions`; expired endpoints (410/404) auto-removed
- Frontend: service worker handles `push` events and calls `showNotification`; toggle in household settings
- Browser support: Chrome, Firefox, Edge, Android WebView; iOS 16.4+ when PWA is added to home screen
- VAPID keys configured via `VAPID_PUBLIC_KEY` / `VAPID_PRIVATE_KEY` / `VAPID_SUBJECT` env vars

---

## Performance targets

- Time to interactive on 4G mobile: < 3s
- API response time (p95): < 300ms for list/read endpoints
- SSE reconnect: within 5s of network recovery
- App shell cached: instant repeat loads

---

## Accessibility

- WCAG 2.1 AA minimum
- All interactive elements keyboard-navigable
- Touch targets: minimum 44×44px
- Color is never the sole signal — always paired with text, icon, or shape
- Screen reader: semantic HTML, ARIA labels on icon-only buttons

---

## Self-hosting requirements

- Single `docker compose up` to run the full stack
- No external dependencies beyond PostgreSQL
- Configuration entirely via environment variables (`.env`)
- No telemetry, no call-home, no external API required for core functionality

---

## SaaS readiness (future)

These constraints must not be violated during any implementation:

- The frontend never talks to the database or auth provider directly
- The repository layer is the only place that knows about the database
- Multi-family multi-user model is built in from day one
- Auth can be swapped (custom JWT → Supabase GoTrue) without changing the API contract
- Feature flags live in a `settings` table, not hardcoded

---

## What's out of scope for v1

| Feature | Reason |
|---|---|
| Native mobile app (Capacitor) | PWA covers it now; Capacitor wrapper planned for M11b |
| Recurring events | Data model is non-trivial; deferred |
| External calendar sync (iCal, Google) | High effort; deferred |
| Search / full-text | High value; needs backend index; deferred |
| Subtasks | Different data model; deferred |
| Activity feed / audit log | Useful for larger households; deferred |
| File attachments | Storage complexity; YAGNI |
| Chat / messaging | Out of scope — use WhatsApp |
