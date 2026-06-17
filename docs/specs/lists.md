# Lists — Functional Spec

## Core concept

A **List** is a named checklist scoped to the family. The canonical use case is shopping, but the model is generic (packing list, chores, school supplies, etc.). Items persist until explicitly deleted — checking an item doesn't delete it, it moves it to "In cart."

---

## List management

### Creating a list
- User types a name and confirms
- No template or type selection — just a name
- First list creation: placeholder text in the name input should suggest "Shopping"
- Soft limit: ~10 lists per family (no hard enforcement needed now)

### Switching lists
- Horizontal pill tabs at top — visible at a glance, one tap to switch
- On mobile: scrollable horizontally if many lists
- Active list ID persists in `localStorage` — don't reset to the first list on every page load

### Deleting a list
- Only available when the list is active AND there are at least 2 lists (can't delete the last one)
- Requires confirmation: "Delete 'Shopping'? This will remove all X items." with a destructive confirm button
- Deleting a list deletes all its items permanently

### Renaming a list *(not yet implemented)*
- Tap the active list pill → inline rename
- Confirm on Enter or blur, cancel on Escape

---

## Items

### Adding an item
- Primary action: text input at top of list, Enter to submit
- New items appear at the **top** of the unchecked section (newest first)
  - Rationale: when adding multiple items rapidly, each new one is immediately visible without scrolling

### Checking an item
- Tap checkbox → item moves instantly to "In cart" section, struck through and dimmed
- Tap checked item → moves back to unchecked (put it back on the shelf)
- **Optimistic update**: apply state change immediately, reconcile on SSE refresh

### Deleting a single item
- Tap × icon → immediate delete, no confirmation
- Rationale: grocery items are low-stakes, undo is not needed

### Renaming an item *(not yet implemented)*
- Tap the item's text (not the checkbox) → inline edit
- Confirm on Enter or blur, cancel on Escape

### Clearing checked items
- "Clear all" button in the "In cart" section header
- No confirmation — checked items are already done, intent is unambiguous
- Primary end-of-shopping flow: clear everything and the list resets for next time

---

## Layout

```
─── To buy (4) ──────────────────────────
  ○ Butter          (newest at top)
  ○ Eggs
  ○ Bread
  ○ Milk

─── In cart (2) ─────────────  [Clear all]
  ✓ ~~Coffee~~      (most recently checked at top)
  ✓ ~~Orange juice~~
```

- **To buy**: unchecked items, ordered by created_at DESC
- **In cart**: checked items, ordered by checked_at DESC (most recently checked first)
- Section headers show item counts
- "In cart" section is hidden entirely when empty
- When "To buy" is empty but "In cart" has items: show a subtle "All done! Tap 'Clear all' when you're home." nudge

---

## Empty states

| Situation | Message |
|---|---|
| No lists exist | "Your family has no lists yet. Create a Shopping list or any other shared list." + [+ Create a list] button |
| List exists, no items | "List is empty. Add your first item above." |

---

## Real-time sync

- Every mutation (add, check, delete, clear) broadcasts an SSE event to all family members
- Critical use case: one person adds items from home while another is already at the store
- Optimistic updates for check/uncheck so the UI feels instant; SSE reconciles any conflicts

---

## What's out of scope (flag for later)

| Feature | Reason deferred |
|---|---|
| Item categories / aisle grouping | Useful but complex; solves a problem only on long lists |
| Quantity + unit (e.g. "2× Milk") | Nice to have, adds friction to quick-add flow |
| Drag to reorder | Complex on mobile; creation-order is good enough |
| Shared link / guest access | SaaS concern, not v1 |
| Recurring items ("always buy milk") | Requires a different data model |
| List templates | YAGNI until users request it |

---

## Delta vs current implementation

| Current | Target |
|---|---|
| ✅ Items ordered: unchecked `created_at DESC`, checked `checked_at DESC` | — |
| ✅ List delete: inline confirmation banner with item count | — |
| ✅ Active list persisted to `localStorage` (per family) | — |
| ✅ "In cart" hidden when empty | — |
| ✅ Section headers show count | — |
| ✅ "All done!" nudge when to-buy empty but cart has items | — |
| Not implemented: inline item rename | Tap item text → inline edit |
