# Roadmap

Status legend: ✅ Done · 🚧 In progress · ⬜ Not started

---

## M1 — Foundation ✅
Auth, families, invite flow.

- ✅ Email + password auth (register, login, JWT)
- ✅ Family creation
- ✅ Member invite via link
- ✅ Protected routes, auth redirect

---

## M2 — Tasks ✅
Core task management.

- ✅ Create / edit / delete tasks
- ✅ Priority (high / medium / low) with color borders
- ✅ Due date
- ✅ Assignee (family member)
- ✅ Labels (create, color, apply to task)
- ✅ Status toggle (todo / done)

---

## M3 — Calendar events ✅
Events on the calendar and board.

- ✅ Create / edit / delete events
- ✅ Date range + time + all-day
- ✅ Location
- ✅ Attendees
- ✅ Labels on events
- ✅ Month view calendar

---

## M4 — Real-time sync ✅
SSE push to all family members.

- ✅ SSE hub in Go backend
- ✅ Broadcast on every mutation
- ✅ EventSource on frontend, reconnect on error
- ✅ Optimistic checkbox toggle

---

## M5 — Board, Today, Lists, Navigation ✅
Full UI structure.

- ✅ Today page (overdue / events / due-today sections)
- ✅ Board page (time groups, filter, sort, quick-add)
- ✅ Lists page (multi-list, check/uncheck, clear checked)
- ✅ Sidebar (desktop) + bottom tab bar (mobile)
- ✅ Settings page (members, invite)
- ✅ Labels management

---

## M6 — Lists polish ✅

- ✅ Ordering: unchecked newest-first (`created_at DESC`), checked by `checked_at DESC` (migration 000010)
- ✅ Section counts ("To buy (N)" / "In cart (N)")
- ✅ Hide "In cart" section when empty
- ✅ Delete list: inline confirmation banner showing item count
- ✅ Active list ID persisted to `localStorage` (per family)
- ✅ "All done! Tap Clear all when you're home." nudge when to-buy is empty but cart has items

---

## M7 — Virtual members ✅
Kids and non-app family members.

- ✅ DB: `virtual_members` table (migration 000011); FK constraints on `tasks.assigned_to` and `event_attendees.user_id` dropped to allow virtual IDs
- ✅ API: `POST /families/{id}/members/virtual`, `DELETE /families/{id}/members/virtual/{id}`, `POST /families/{id}/members/virtual/{id}/link`, `GET /families/{id}/members/virtual/unlinked`
- ✅ `GET /families/{id}/members` returns real + unlinked virtual members merged (virtual flag set)
- ✅ Settings > Members: "+ Without account" button + inline name form; virtual members shown with person-off icon + "No account" label + delete button
- ✅ Invite acceptance: `POST /invites/{token}/accept` now returns `{ family_id, unlinked_virtual_members }` — frontend shows "Are you one of these?" prompt before redirect
- ✅ Link action: migrates task assignments + event attendees from virtual ID to real user ID in a single transaction
- ⬜ Settings > Members: manual link/unlink for admins (admin can link a virtual member to an existing real account) — deferred

---

## M8 — Calendar improvements ✅
Close the gap between current month-only calendar and the Calendar spec.

- ✅ Tasks with due dates shown as chips on their due date (month + agenda)
- ✅ Agenda view — every day listed including empty; auto-scrolls to today; mobile default
- ✅ Day panel — tap any day → dialog with that day's items, quick-add task, + Event button
- ✅ Tapping a day cell pre-fills date in create dialog
- ✅ "Today" jump button (scrolls to today in agenda, resets month in month view)
- ✅ Filter panel (by member, by label) with active filter count badge
- ✅ "+N more" overflow on month view day cells (cap at 2, opens day panel)

---

## M9 — Calendar week / day views ✅
Secondary calendar views; lower priority than M8.

- ✅ Week view (7-column grid, hourly slots, events as sized blocks)
- ✅ Day view (single day, hourly slots, reached from month/week tap or day panel "Day view" button)
- ✅ Swipe navigation for month, week, and day views
- ✅ Current-time red indicator line (updates every minute)
- ✅ Click time grid to pre-fill event start/end time
- ✅ All-day events and tasks in pinned row above time grid

---

## M10 — Auth improvements ⬜
OAuth and invite polish.

- ⬜ Google OAuth
- ⬜ Apple OAuth
- ⬜ Forgot password / reset flow
- ⬜ Resend invite email

---

## M11 — Offline & PWA hardening ⬜
- ⬜ "Offline" banner when network is unavailable
- ⬜ Auto-reload data on reconnect
- ⬜ App shell cached via service worker (install prompt)
- ⬜ iOS home screen icon + splash screen

---

## M11b — Capacitor native wrapper ⬜
Wrap the existing SvelteKit app in a Capacitor native shell for iOS and Android distribution. **No rewrite** — same codebase, native WebView + plugin access.

**Why Capacitor over Flutter/React Native:**
- Zero code rewrite — SvelteKit stays as-is
- Three targets from one codebase: self-hosted PWA + iOS App Store + Google Play
- Access to native plugins (push, haptics, safe area, status bar, camera)
- Capacitor detects native context via `Capacitor.isNativePlatform()` — conditional behavior where needed

**Steps when ready:**
1. `npm install @capacitor/core @capacitor/cli && npx cap init`
2. `npx cap add ios && npx cap add android`
3. SvelteKit build → `npx cap sync` copies output into native projects
4. Add `@capacitor/push-notifications` — use instead of Web Push when native
5. Add `@capacitor/status-bar`, `@capacitor/splash-screen`, `@capacitor/haptics`
6. CI: add Xcode + Android Studio build steps

**Distribution targets after this milestone:**
- Self-hosted PWA: `docker compose up` (unchanged)
- iOS: Capacitor → Xcode → App Store
- Android: Capacitor → Android Studio → Play Store

> **If traction grows beyond Capacitor:** Flutter (one codebase, true native UI, good if team stays JS/Dart) or fully native Swift/Kotlin (maximum platform integration — widgets, Watch, CarPlay). The Go backend requires zero changes for either path — it's just another REST client.

---

## M12 — Calendar DnD & advanced rendering ⬜
Drag-and-drop interaction and richer visual rendering on the calendar. This milestone likely requires replacing the hand-rolled month grid with a calendar rendering library that supports DnD natively.

**Library candidates to evaluate at implementation time:**
- [`svelte-fullcalendar`](https://github.com/YogliB/svelte-fullcalendar) — wrapper around FullCalendar, mature, heavy (~200kb)
- [`@event-calendar/svelte`](https://github.com/vkurko/calendar) — lightweight, Svelte-native, supports DnD, month/week/day/agenda — **leading candidate**
- Hand-rolled + `svelte-dnd-action` — full control, significant effort

- ⬜ Drag an event on the calendar to a new day → updates `start_at` / `end_at` via PATCH
- ⬜ Multi-day events rendered as horizontal pills spanning across day cells in month view
- ⬜ Drag to resize event duration (week/day view only)
- ⬜ Drag tasks on the Board to reorder within a group (manual ordering)
- ⬜ Drag tasks between time groups (e.g. move "Later" task to "Today" → sets due date to today)

> **Note:** if a library is adopted here, M8 may need partial rework to hand off the grid to the library. Factor that in before starting M12.

---

## M13 — Board drag-and-drop ⬜
If M12 doesn't cover Board DnD (it may not, since Board is a list not a calendar):

- ⬜ Drag tasks within a group to set manual ordering
- ⬜ Drag a task to a different time group → adjusts due date accordingly (e.g. drag to "Today" → sets due date = today)
- ⬜ Requires: `manual_order` column on tasks, or accept "due date changes on drop" as the ordering mechanism

---

## M15 — Design pass ⬜
One focused sweep to reconcile the full UI against `docs/specs/design.md`. Deferred until M13 is done to avoid conflicts with feature work.

- ⬜ Audit color tokens — ensure all semantic tokens (`--primary`, `--muted`, etc.) match the design spec in both light and dark mode
- ⬜ Typography consistency — page titles, section headers, card metadata, all on spec
- ⬜ Spacing and rounding audit — `rounded-lg` everywhere, comfortable padding, no dense screens
- ⬜ Touch target audit — all interactive elements ≥ 44×44px
- ⬜ Empty states — consistent style across all pages
- ⬜ Motion — cap all transitions at 200ms, remove any JS animation

> **Note on new code (M6–M11):** do not add new design debt. Follow the design spec for anything written now. This milestone only exists to fix existing inconsistencies — not to catch up on drift from ongoing work.

---

## Deferred / no milestone yet

These are captured in spec out-of-scope tables. Promote to a milestone when prioritised.

| Feature | Spec reference |
|---|---|
| Recurring events | `calendar.md` |
| External calendar sync (iCal, Google, CalDAV) | `calendar.md` |
| Push notifications | `calendar.md`, `app.md` |
| Search / full-text filter | `board.md` |
| Bulk actions (select multiple, bulk delete/assign) | `board.md` |
| Inline item rename (lists) | `lists.md` |
| Inline list rename | `lists.md` |
| Subtasks | `board.md` |
| List item categories / aisle grouping | `lists.md` |
| Activity feed / audit log | `app.md` |
| File attachments | `app.md` |
| Native mobile app (Flutter) | `app.md` |
