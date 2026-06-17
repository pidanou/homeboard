# Board — Functional Spec

## Core concept

The Board is the **planning view** of the family's life. It answers "what needs doing and what's coming up?" across any time horizon. Unlike Today (scoped to now) or Calendar (time-grid), the Board is a flat list optimised for scanning, triaging, and managing work.

Tasks and events are shown together because families don't think in separate silos — a school pickup and a dentist appointment to book live in the same mental space.

---

## Content model

### What appears

| Item | Condition |
|---|---|
| Tasks | Status is `todo` (not `done`) unless Done filter is active |
| Events | `end_at` is in the future (within 90-day window) |
| Done tasks | Only when "Done" filter is active |

Events older than 90 days from today are not fetched — this is a display window, not an archive.

### What does NOT appear
- Tasks marked done (unless Done filter active)
- Past events (end_at < now)
- Tasks without a match against active member/label filters

---

## Layout

### Time groups

When sorting by date (default), items are grouped into four sections with a labelled divider:

```
OVERDUE ───────────────────────
  [task] Pay electricity bill — 3 days ago

TODAY ──────────────────────────
  [event] Doctor 14:00–15:00
  [task]  ☐ Buy groceries

THIS WEEK ──────────────────────
  [event] School trip Thu
  [task]  ☐ Call insurance

LATER ──────────────────────────
  [task]  ☐ Plan summer holiday
  [task]  ☐ Renew passport (no due date)
```

- **Overdue**: tasks with `end_date` < today. Events never appear here (already filtered out by `end_at >= now`).
- **Today**: tasks due today + events starting today or currently ongoing.
- **This week**: items due/starting within the next 7 days (excluding today).
- **Later**: items beyond 7 days + tasks with no due date.
- Groups with no items are hidden entirely.
- When sorting by priority or title, grouping is disabled — flat list, no dividers.

### Done view

When "Done" filter is active: flat list of completed tasks, no grouping, sorted by completion recency (most recently done first). No events appear in the Done view.

---

## Quick-add

- Full-width dashed input at the very top of the page
- Press Enter → creates a task with `priority: medium`, no due date, no assignee
- Clears the input after submission; stays focused for rapid entry
- If the current filter is "Events only", quick-add still creates a task (it's always a task shortcut)
- The created task appears in the list immediately (optimistic or via SSE refresh)

---

## Task card

### Visual anatomy

```
│ [checkbox] Title of the task                        │  ← red/yellow left border for priority
│            Due tomorrow · ⊙ Laura · ● urgent        │
│            Optional description truncated…           │
```

- **Priority border**: `high` = red left border, `medium` = yellow, `low` = none
- **Checkbox**: tapping toggles done/undone immediately (optimistic update)
- **Due date**: displayed as relative text (`today`, `tomorrow`, `in 3 days`, `2 days ago`)
- **Overdue date**: red + bold
- **Assignee**: shown as name with person icon
- **Labels**: colored chips after assignee
- **Description**: one line, truncated, only shown when status is not done
- **Done state**: checkbox checked, title struck through, muted, reduced opacity, 4-px slate left border

### Interactions
- Tap card body → open Edit dialog
- Tap checkbox → toggle done/todo (stops propagation, does not open edit)

---

## Event card

### Visual anatomy

```
│ 17 Jun → 18 Jun · 14:00–18:00 · ⊙ Laura ⊙ Eang   │  ← blue left border
│ Team dinner                                          │
│ wavestone · ● loisir                                │
│                                              ● live │  ← orange pulsing dot if ongoing
```

- **Blue left border**: permanent, visually distinguishes events from tasks at a glance
- **Live dot**: animated orange dot when `start_at <= now <= end_at`
- **Date range**: short format (`17 Jun → 18 Jun`); single-day events show one date
- **Times**: shown when not all-day; hidden when `all_day = true`
- **Location**: MapPin icon + text, only shown when set
- **Attendees**: person icons with names
- **Labels**: colored chips

### Interactions
- Tap anywhere → open Edit dialog

---

## Create dialog

Opened via the "Task" or "Event" buttons in the toolbar row.

### Task fields
| Field | Required | Default |
|---|---|---|
| Title | Yes | — |
| Description | No | — |
| Priority | No | medium |
| Due date | No | — |
| Assign to | No | Unassigned |
| Labels | No | none |

### Event fields
| Field | Required | Default |
|---|---|---|
| Title | Yes | — |
| Description | No | — |
| Date range | Yes | — |
| Start time | No (if all-day) | 09:00 |
| End time | No (if all-day) | 10:00 |
| All day | No | false |
| Location | No | — |
| Attendees | No | none |
| Labels | No | none |

- Dialog opens pre-selected to Task or Event based on the button tapped
- Title field is auto-focused on open
- Enter in the title field submits (task only; events require date)
- Cancel discards all state

---

## Edit dialog

Opened by tapping any card on the board.

- Same fields as Create, pre-populated with current values
- **Delete** button (destructive, bottom-left): no secondary confirmation — the action is explicit
- **Save** button: disabled when title is empty or (for events) no start date
- Closing via Cancel or overlay click discards unsaved changes

---

## Filter panel

Opened by the "Filter" button. Shows a badge with the count of active filters.

### Show (mutually exclusive)
- **Active** (default): all non-done tasks + future events
- **Tasks**: tasks only
- **Events**: events only
- **Done**: completed tasks only (count shown in button label)

### Who (multi-select)
- "Everyone" clears the selection
- Selecting a member shows only items assigned to / attended by that person
- Hidden if the family has only 1 member

### Labels (multi-select)
- Shows all labels for the family
- Selecting a label shows only items with that label
- Selecting multiple labels shows items matching ANY selected label (OR logic)

### Clear all
- Shown only when at least one filter is active
- Resets Show to "Active", clears Who and Labels

### Manage labels link
- Navigates to Settings > Labels

---

## Sort panel

Opened by the "Sort" button. Shows a visual indicator when non-default sort is active.

### Order by (mutually exclusive)
- **Date** (default): by due date for tasks, by start_at for events; no-date tasks sort to the end
- **Priority**: high → medium → low → no priority; ties broken by date
- **Title**: alphabetical A→Z

### Direction (mutually exclusive)
- **Ascending** (default)
- **Descending**

When sort is not "Date", time grouping (Overdue / Today / This week / Later) is disabled.

---

## Real-time sync

- SSE connection opened on mount, closed on destroy
- Any `refresh` event triggers a full data reload (members, tasks, events, labels)
- Optimistic updates for checkbox toggles: state changes immediately, SSE reconciles

---

## Error handling

- Errors shown as a dismissible banner at the top of the page
- Auto-dismiss after 4 seconds
- Manual dismiss via ✕ button

---

## What's out of scope (flag for later)

| Feature | Notes |
|---|---|
| Search / full-text filter | High value; needs backend index or client-side scan |
| Bulk actions (select multiple, bulk delete/assign) | Needed once lists grow long |
| Drag to reorder / manual ordering | Complex on mobile; sort covers it for now |
| Task status beyond todo/done (e.g. in progress) | Requires UI change in card + filter + board logic |
| Subtasks | Different data model; high complexity |
| Pagination / infinite scroll | Current 90-day window is sufficient for now |
| Bulk clear done tasks | Nice to have once "Done" view fills up |
| Pinned / starred tasks | YAGNI |

---

## Delta vs current implementation

| Current | Target |
|---|---|
| Implemented: quick-add, filter, sort, time grouping, task/event cards | — |
| Implemented: relative due dates, priority borders, label chips | — |
| Implemented: create/edit dialogs with all fields | — |
| Implemented: optimistic checkbox toggle | — |
| Implemented: SSE real-time refresh | — |
| Not implemented: search | Flag for later |
| Not implemented: bulk actions | Flag for later |
| Not implemented: "in progress" task status | Flag for later |
| Not implemented: bulk clear done tasks | Flag for later |
