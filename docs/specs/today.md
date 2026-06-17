# Today — Functional Spec

## Core concept

Today is the **home screen** of the app. It answers a single question: "what do I need to deal with right now?" The scope is strictly the current day — nothing tomorrow, nothing hypothetical. It is the first thing a family member sees when they open the app.

The goal is a fast, calming overview. Open the app, understand your day in 10 seconds, close it.

---

## What appears

| Item | Condition |
|---|---|
| Overdue tasks | `status = todo` AND `end_date < today` |
| Events | `start_at` or `end_at` falls within today (00:00–23:59 local) |
| Due-today tasks | `status = todo` AND `end_date = today` |

No filters, no sorting controls — Today is opinionated. The view is always "everything that matters right now."

### What does NOT appear
- Tasks without a due date (they belong on the Board's "Later" group)
- Done tasks
- Future events or tasks

---

## Layout

```
Today
Wednesday, 18 June

[Add a task for today…]            [+ Event]

OVERDUE ──────────────────────────────── (red)
  [task] Pay electricity bill — 3 days ago

EVENTS ────────────────────────────────
  [event] Doctor 14:00–15:00
  [event] School pickup 16:30

DUE TODAY ─────────────────────────────
  [task] ☐ Buy groceries
  [task] ☐ Call insurance
```

### Section order
1. **Overdue** (red section divider) — always shown first; these are urgent
2. **Events** — time-ordered, all-day events last
3. **Due today** — sorted by priority (high → medium → low)

Sections with no items are hidden. When all three are empty: full-screen empty state.

---

## Quick-add task

- Single text input at the top, placeholder "Add a task for today…"
- Enter submits: creates a task with `end_date = today`, `priority = medium`, no assignee
- Input clears after submission; focus stays for rapid entry
- Quick-added tasks appear in "Due today" immediately (SSE or optimistic)

---

## Quick-add event

- "+ Event" button (top-right of header)
- Opens the Create dialog pre-selected on the Event tab, with today's date pre-filled
- No quick-text shortcut for events — they always need at least a time or confirmation

---

## Task interactions

- Tap checkbox → toggle done (optimistic update); done tasks disappear from view
- Tap card body → open Edit dialog
- Checking an overdue task: it disappears immediately — no confirmation needed

---

## Event interactions

- Tap card → open Edit dialog
- Live dot on currently-ongoing events (animated orange dot)

---

## Empty state

```
     ✓
  All clear for today
  No events, no tasks due.
```

Shown when overdue = 0, events = 0, due-today = 0. No suggestions, no upsells — just a clean confirmation.

---

## Real-time sync

- SSE connection for the family; any `refresh` event reloads data
- Useful when a family member adds a task or event while another has Today open

---

## What's out of scope

| Feature | Notes |
|---|---|
| Filters or sort controls | Today is intentionally opinionated — no customisation |
| "Tomorrow" preview | Belongs in the Calendar or Board |
| Weather / external data | Third-party dependency; YAGNI |
| Streaks / gamification | Could be added as a family engagement feature later |
| Agenda for the full week | That's Calendar |

---

## Delta vs current implementation

| Current | Target |
|---|---|
| Implemented: overdue / events / due-today sections | — |
| Implemented: quick-add task (creates with today's due date) | — |
| Implemented: "+ Event" button opens create dialog | — |
| Implemented: SSE refresh, optimistic checkbox toggle | — |
| Implemented: empty state | — |
| Not implemented: all-day events sorted to bottom within Events section | Sort all-day after timed events |
| Not implemented: live dot on ongoing events in Today view | Port from EventCard (already in Board) |
