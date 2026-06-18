# Design Spec — Family Board

## Who we're designing for

Family Board is used by **real households**, not productivity power users. The people using this app are:

- A parent adding groceries between meetings, one-handed, on a phone
- A teenager checking if they're supposed to bring something to school tomorrow
- A partner confirming whether that dentist appointment is this week or next
- A grandparent (less tech-savvy) who was invited to join the family board

**Key insight:** the app is glanced at, not studied. Users open it for 15 seconds, do one thing (check an event, add an item, mark a task done), and close it. It must deliver value in that 15-second window every single time.

---

## Design principles

### 1. Calm over dense
Families have enough stress. The app should feel like a clean kitchen whiteboard, not a project management dashboard. Whitespace is load-bearing. Never pack the screen to use every pixel.

### 2. Glanceable over powerful
The most important information must be readable without interaction. Today's events and tasks should be obvious at a glance. Filters and sort controls exist but are never the default.

### 3. Forgiving over strict
Deleting a grocery item is fine without confirmation. Deleting a list with 20 items needs one. Match the weight of the confirmation to the weight of the action. Never punish small mistakes.

### 4. Shared but personal
The family shares data, but each member is an individual. Assignees matter. "What's mine?" should be answerable in one tap.

### 5. Mobile-first, tablet-comfortable, desktop-usable
Design for the phone first. The kitchen tablet gets the same design at a larger size. Desktop is a bonus, not the primary form factor.

---

## Visual language

### Warmth
- Warm neutral backgrounds (not pure white, not pure black)
- Avoid clinical blues and greys; prefer cream, warm stone, soft amber accents
- The app should feel like a home, not an office

### Clarity through restraint
- One primary action per screen
- Typography hierarchy: large → medium → small; never more than 3 levels visible at once
- Icons only when they add meaning; text labels alongside icons for all nav items

### Color as signal, not decoration
- Red: overdue, destructive actions, errors
- Blue: events (consistent across all views) — solid fill
- Amber/orange: tasks (consistent across all views) — outlined (light tint bg, amber border); also used for important task star indicator
- Green: success, done states
- Label colors: a curated palette of ~10 warm, distinct colors (not neon)

---

## Typography

- **Font:** Inter (variable)
- **Headings:** `font-semibold` or `font-bold`
- **Body:** `font-normal`, line-height `1.5`
- **Metadata (dates, assignees, labels):** `text-xs text-muted-foreground`
- Never use more than one typeface. Never use decorative fonts.

### Scale (Tailwind defaults)
| Role | Class |
|---|---|
| Page title | `text-xl font-semibold` |
| Section header | `text-xs font-semibold uppercase tracking-wide` |
| Card title | `text-sm font-medium` |
| Card metadata | `text-xs text-muted-foreground` |
| Input / button | `text-sm` |

---

## Color tokens

Defined as CSS variables via shadcn-svelte's theming. Must support light and dark mode.

| Token | Light | Dark | Usage |
|---|---|---|---|
| `--background` | warm off-white | warm near-black | Page background |
| `--foreground` | slate-900 | slate-50 | Primary text |
| `--muted` | stone-100 | stone-800 | Input backgrounds, subtle fills |
| `--muted-foreground` | stone-500 | stone-400 | Secondary text, metadata |
| `--border` | stone-200 | stone-700 | Dividers, card borders |
| `--primary` | warm amber or teal | same | Buttons, active nav, checkboxes |
| `--destructive` | red-600 | red-500 | Delete buttons, error text |
| `--sidebar` | slightly darker than background | slightly lighter | Sidebar fill |
| `--sidebar-border` | matches border | matches border | Sidebar right edge |

Exact hex values live in `tailwind.config.ts` and `app.css`. This table describes intent, not implementation.

---

## Component conventions

### Cards (tasks, events)
- Rounded: `rounded-lg`
- Border: `border border-border`
- Background: `bg-card` (same as background or slightly elevated)
- Hover: subtle `bg-muted/40` lift
- Left accent border: 4px solid amber for tasks in calendar view; events use solid fill (label color or default blue)
- Important tasks: amber star icon (filled) inline with title; non-important tasks have no indicator
- Tap target: the full card is tappable; min height ensures 44px

### Buttons
- Primary: filled, `bg-primary text-primary-foreground`
- Secondary: outlined or ghost
- Destructive: `bg-destructive text-white`
- Icon buttons: 36–40px square minimum; always `aria-label`
- `cursor-pointer` on all interactive elements

### Inputs
- Consistent height: `h-9` or `h-10`
- Border: `border-border`, focus: `ring-2 ring-primary/20`
- Quick-add inputs: dashed border, solid on focus — signals "lightweight entry"

### Dialogs
- Max width: `max-w-md` on mobile, `max-w-lg` on desktop
- Always closeable by clicking the overlay or pressing Escape
- Destructive action (delete) in bottom-left; primary action (save) in bottom-right
- Title field auto-focused on open

### Labels (chips)
- Small pill: `rounded-full px-2 py-0.5 text-xs font-medium`
- Color matches the label's assigned color from the palette
- Unselected state: `opacity-40` (selected = full opacity)

### Section dividers
- Pattern: `text-xs font-semibold uppercase tracking-wide` label + `flex-1 h-px bg-border` line
- Overdue section: `text-destructive` + `bg-destructive/20` line

### Empty states
- Centered vertically in the content area
- Large icon or emoji + short heading + one-line explanation
- No marketing copy, no calls to action beyond the relevant "create" action if applicable

---

## Motion and interaction

- **Transitions:** subtle, fast (150–200ms). CSS transitions only — no JS animation libraries.
- **Checkboxes:** immediate visual feedback; optimistic update
- **Dialogs:** fade-in (100ms) + slight scale up from 95%
- **Navigation:** instant; no page transition animations
- **Error banners:** slide in from top or fade in; auto-dismiss after 4s
- **No loading spinners** for operations that take < 300ms; prefer skeleton placeholders for initial page load

---

## Responsive breakpoints

| Breakpoint | Width | Layout |
|---|---|---|
| Mobile (base) | < 768px | Single column, bottom tab bar, stacked sections |
| Tablet (md) | 768px–1279px | Left sidebar (collapsed or mini), larger cards |
| Desktop (lg) | 1280px+ | Full sidebar (w-56), content area with max-width |

Content max-width: `max-w-2xl` for list-style pages (Today, Board, Lists); `max-w-4xl` for Calendar.

---

## Dark mode

- Follows system preference via `prefers-color-scheme`
- Tailwind `dark:` variant on all color classes
- No manual toggle in v1 — system preference is sufficient for a household app
- Test: all color token pairs must pass WCAG AA contrast (4.5:1 for text, 3:1 for UI components)

---

## Iconography

- **Library:** Lucide Svelte (consistent with shadcn-svelte defaults)
- **Size:** `w-4 h-4` in text/cards, `w-5 h-5` in nav and buttons, `w-6 h-6` for page headers
- Never use icons without labels in navigation (accessibility + clarity)
- Icon color: inherits from text color unless it's a semantic color (red = destructive, etc.)

---

## What to avoid

| Anti-pattern | Why |
|---|---|
| Dense information layout | Families glance, they don't parse dashboards |
| Confirmation dialogs for low-stakes actions | Friction kills habit; match to action weight |
| Jargon (workspace, sprint, ticket) | This is a family app, not Jira |
| Tiny tap targets | Parents use this with one hand while holding a child |
| Modals for navigation | Disorienting; use pages for content, modals for focused actions |
| Animations > 300ms | Feels sluggish on mid-range phones |
| Pure white or pure black backgrounds | Harsh; warm neutrals feel more domestic |
| Color as the only signal | Colorblindness + dark mode both break color-only cues |
