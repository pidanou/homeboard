# Profile Management — Functional Spec

## Core concept

A **profile** is the user's personal identity within the app. It controls how they appear to other family members (name, avatar) and manages account-level settings (email, password). Profile data is scoped to the **user account**, not the family — a user who belongs to multiple families has one profile shown consistently across all of them.

---

## Profile fields

| Field | Notes |
|---|---|
| display name | Required. Shown everywhere the user appears (task assignees, event attendees, member lists). |
| avatar | Optional. A photo or image. Falls back to initials if absent. |
| email | Required. Used for login and invites. Changing it requires confirmation. |
| password | Not displayed — change via a dedicated flow. |

---

## Avatar

### Upload

- User taps their avatar (or the placeholder) → file picker opens, filtered to images (`image/*`)
- After selection: **crop UI** appears — square crop with circular preview overlay
  - User can pan and pinch/zoom to frame their face
  - Confirm or cancel
- On confirm: image is uploaded to the backend, stored as a file, and the `avatar_url` on the member record is updated
- **Accepted formats:** JPEG, PNG, WebP
- **Max file size:** 5 MB (enforced client- and server-side)
- **Stored size:** backend rescales to 256×256 before saving — never store raw upload

### Removal

- "Remove photo" option in the avatar action menu
- Immediately reverts to initials fallback
- No confirmation required — low stakes, re-upload is trivial

### Fallback

- Initials derived from display name (first letter of first and last word)
- Fallback background color: deterministic from user ID so it's stable across sessions
- Single initial if display name is one word

### Storage (self-hosted)

- Files stored on the server filesystem under a configurable path (`UPLOAD_DIR` env var)
- Served via a dedicated route: `GET /api/v1/uploads/avatars/:filename`
- **SaaS path:** swap filesystem storage for an object store (S3/R2) behind the same upload interface — no frontend changes required

---

## Editing profile

### Display name

- Inline edit: tap name → editable input, confirm on Enter or blur, cancel on Escape
- Validation: non-empty, max 64 chars
- Change reflects immediately everywhere (SSE broadcast to family members)

### Email

- Change email flow:
  1. User enters new email + current password to confirm intent
  2. Verification email sent to **new** address
  3. Email updates only after verification link is clicked
  4. Until then, login and notifications still use old email
- Rationale: prevents account lockout from typos

### Password

- Separate "Change password" section: current password + new password + confirm
- Minimum 8 characters (same as signup)
- On success: current session remains valid, other sessions are not invalidated (defer session management to a later milestone)

---

## Layout

### Entry point

The user's avatar + display name appears in the **top-right corner of the app shell** (persistent across all views). Tapping it opens the profile page. This is the standard pattern used by Google, Notion, Linear, and most SaaS apps — users already know to look there.

No dedicated "Profile" tab in the bottom nav — that real estate is for content views, not account management.

### Mobile (< 768px) — full-screen page

```
┌─────────────────────────────┐
│  ←  My Profile              │  ← back button, page title
├─────────────────────────────┤
│                             │
│         [Avatar]            │  ← large (80px), centered, tap to change
│       Pierre Dupont         │  ← display name below avatar
│    pierre@example.com       │  ← email, muted
│                             │
├─────────────────────────────┤
│  ACCOUNT                    │  ← section header (muted caps, small)
│  Display name        ›      │
│  Email               ›      │
│  Change password     ›      │
├─────────────────────────────┤
│  DANGER ZONE                │
│  Delete account      ›      │  ← destructive color
└─────────────────────────────┘
```

- Each row taps through to a dedicated sub-page (name edit, email change, password change)
- Sub-pages follow the same full-screen + back-button pattern
- iOS/Android Settings app pattern — universally understood on mobile

### Desktop (≥ 768px) — centered card, max-width 480px

Same content as mobile, but rendered as a card centered in the viewport rather than full-screen. No sidebar or two-column layout — profile is not complex enough to warrant it, and a narrow centered card reads more like a focused form (better for editing).

```
┌──────────────────────────────────────┐
│              My Profile              │
│                                      │
│             [Avatar 96px]            │
│            Pierre Dupont             │
│         pierre@example.com           │
│                                      │
│  ┌────────────────────────────────┐  │
│  │ ACCOUNT                        │  │
│  │ Display name              ›    │  │
│  │ Email                     ›    │  │
│  │ Change password           ›    │  │
│  └────────────────────────────────┘  │
│  ┌────────────────────────────────┐  │
│  │ DANGER ZONE                    │  │
│  │ Delete account            ›    │  │
│  └────────────────────────────────┘  │
└──────────────────────────────────────┘
```

### Avatar tap → action sheet / popover

Tapping the avatar opens a small action menu (bottom sheet on mobile, popover on desktop):

- **Upload new photo**
- **Remove photo** (only shown if avatar exists)
- Cancel

No inline crop in the action sheet — crop UI opens as a full-screen overlay after the file is selected.

---

## What's out of scope

- Multiple profile photos / gallery
- Profile visibility controls (everything is visible within the family)
- Social features (bio, status, presence indicators)
- OAuth avatar import (Google/Apple profile photo auto-sync)
- Session management (viewing / revoking active sessions)
- Two-factor authentication

---

## Delta vs current implementation

| Feature | Status |
|---|---|
| Display name editing | Not implemented |
| Avatar upload + crop | Not implemented |
| Avatar fallback (initials) | Partially implemented (virtual members) |
| Email change flow | Not implemented |
| Password change | Not implemented |
| Profile settings page | Not implemented |
