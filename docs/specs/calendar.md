# Calendar — Functional Spec

## Core concept

The calendar is the **time-oriented view** of the family's life. It shows events (time-specific) and tasks with due dates side by side on a shared timeline. The primary value is answering "what's happening this week / this month?" at a glance.

### How it differs from the Board

The Board and the Calendar both show tasks and events but serve opposite mental models:

| | Board | Calendar |
|---|---|---|
| Primary axis | Priority + urgency | Chronological time |
| Tasks without due date | Shown (in "Later") | Never shown — no date = no place on a timeline |
| Done tasks | Accessible via filter | Never shown |
| Past events | Never shown | Visible (scrollable in agenda view) |
| Past overdue tasks | Shown in "Overdue" group | Shown on their original due date |
| Empty days | Not a concept | Always visible — seeing "nothing on Thursday" is useful |
| Purpose | Act on things, triage | Understand the shape of time |

The agenda view is **not** a duplicate of the Board. Its distinct value is: every day is represented (including empty ones), past days are scrollable, and the primary question is "what does this week look like?" not "what should I do next?"

---

## Views

### Month view (default)
- Standard calendar grid: 5–6 weeks, Mon–Sun columns (locale-aware)
- Each day cell shows a compact list of that day's events and due tasks
- "Today" is visually highlighted
- Tap a day → expand to Day panel (see below)
- Navigate months: swipe left/right or arrow buttons

### Agenda view *(not yet implemented — highest priority)*
- A chronological list of events and due tasks, grouped by day — **every day is shown**, including days with no items
- Starts from today, scrolls forward; scrolls back into the past
- Empty days are shown with a subtle "Nothing scheduled" line — this is intentional, not a bug
- Best for mobile: no grid, just a readable vertical list
- Should be the **default on mobile** (month grid is hard to read on small screens)
- Does NOT replace the Board: tasks without due dates are absent, done tasks are absent, priority is not shown

### Week view *(not yet implemented — secondary priority)*
- 7-column grid with hourly time slots
- Events shown as blocks sized by duration
- Tasks shown as all-day chips at the top
- Useful for scheduling and seeing conflicts
- Swipe to navigate weeks

### Day view *(not yet implemented)*
- Single day, hourly time slots
- Reached by tapping a day in month or week view
- Shows all events for that day with precise times
- Tasks due that day shown at top
- Back button returns to the view that opened it

---

## What appears on the calendar

### Events
- All events with `start_at` / `end_at` in the visible range
- Multi-day events span across day cells in month view
- Color indicator: blue left border (consistent with Board cards)
- Show: title + start time (if not all-day)

### Tasks with due dates
- Any task with `end_date` set appears on its due date
- Shown as a chip/dot distinct from events (e.g. checkbox icon, no time)
- Overdue tasks: shown on their original due date, not today (to preserve history)
- Completed tasks: hidden by default; toggle "Show done" to reveal

### What does NOT appear
- Tasks without a due date
- Tasks marked done (unless toggled on)

---

## Day panel (tapping a day in month view)

Slides up from the bottom (sheet / drawer) or expands below the grid:

```
Wednesday, 18 June
─────────────────────────────
  [event] Doctor 14:00–15:00
  [event] School pickup 16:30
  [task]  ☐ Pay electricity bill
─────────────────────────────
  [+ Event]  [+ Task]
```

- Lists all events and due tasks for that day
- Tap an item → open edit dialog
- Tap [+ Event] → open create dialog pre-filled with that date
- Tap [+ Task] → open create dialog pre-filled with that date as due date
- Dismiss by tapping outside or swiping down

---

## Event display rules

### Month view cell
- Max 2–3 items visible per cell; overflow shown as "+2 more"
- Tapping "+2 more" opens the Day panel
- Events shown first, then tasks
- Multi-day events shown as a spanning pill across cells

### Conflict / overlap indicator
- If a day has 2+ events with overlapping times: subtle warning dot on the day cell
- No automatic conflict resolution — just visibility

---

## Navigation

| Action | Behavior |
|---|---|
| Swipe left/right (month) | Go to next/previous month |
| Tap day | Open Day panel |
| Tap event anywhere | Open edit dialog |
| "Today" button | Jump back to current month/week/day |
| Arrow buttons | Same as swipe, for desktop |

---

## Creating events from the calendar

- Tap [+ Event] button (top-right of calendar header)
- In month view: tapping an empty day cell → create dialog pre-filled with that date
- In week/day view: tap-and-drag on a time slot → create dialog pre-filled with start/end time

---

## Filtering

Consistent with the Board page filter panel:
- Filter by attendee (show only events where selected family member is attending)
- Filter by label
- "Show done tasks" toggle (hidden by default)
- Filters persist per session, not across sessions

---

## Tasks on calendar — interaction rules

- Checking off a task from the calendar marks it done (same API call as Board)
- Overdue tasks appear greyed with a red indicator on their original due date
- Tapping a task opens the edit dialog (same EditDialog component as Board)

---

## Recurring events *(not yet implemented — deferred)*

Full recurrence is out of scope for v1 but the data model should not preclude it. When implemented:
- Rules: daily, weekly (select days), monthly, yearly
- End: never / after N occurrences / on date
- Edit: "this event" / "this and following" / "all events"
- Stored as a `recurrence_rule` TEXT field on the events table (RFC 5545 RRULE)

---

## External calendar sync *(deferred)*

| Feature | Notes |
|---|---|
| iCal export (.ics) | Per-family export URL, token-authenticated |
| Google Calendar import | One-way pull of external events as read-only |
| CalDAV server | Full two-way sync; requires significant backend work |

---

## Reminders / notifications *(deferred)*

- Push notifications via PWA Web Push API
- Trigger: X minutes before an event start (configurable: 15 / 30 / 60 min)
- Requires: service worker with push subscription, server-side notification sender
- Per-user setting, not per-family (each member configures their own)

---

## Delta vs current implementation

| Current | Target |
|---|---|
| Month view only | Add Agenda view (mobile default), Week view, Day panel |
| Events only on calendar | Tasks with due dates shown as chips |
| No overflow handling | "+N more" on day cells |
| No day tap interaction | Day panel slides up on tap |
| Create button opens dialog with today's date | Tapping a day pre-fills that date |
| No filter UI on calendar | Port filter panel from Board (by member, by label) |
| No "Today" button | Add jump-to-today |
| Swipe not implemented | Add swipe navigation for month/week |
