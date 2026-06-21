# Roles Spec

## Roles

Every real member of a family has one of two roles:

| Role | Description |
|------|-------------|
| `admin` | Full control over family settings and content |
| `member` | Can create and manage content, cannot change family structure |

Virtual members have no role — they cannot act.

---

## Permission table

| Action | admin | member |
|--------|:-----:|:------:|
| View all content (tasks, events, lists, calendar) | ✓ | ✓ |
| Create / edit / complete tasks | ✓ | ✓ |
| Create / edit / delete events | ✓ | ✓ |
| Create / edit / delete lists and list items | ✓ | ✓ |
| **Manage categories** (create / edit / delete) | ✓ | ✗ |
| **Create virtual members** | ✓ | ✗ |
| **Remove virtual members** | ✓ | ✗ |
| **Kick real members** | ✓ | ✗ |
| **Generate / revoke invite links** | ✓ | ✗ |
| **Change member roles** | ✓ | ✗ |

---

## Rules

- **Family creator** is always the first admin.
- **New members** joining via invite land as `member` by default.
- **Role change** is admin-only and applies to any other real member.
- **Last-admin protection:** an admin cannot be demoted if they are the only admin in the family. At least one admin must always exist.
- **Self-kick is forbidden.** An admin cannot remove themselves (use "Leave family" instead, not yet implemented).
- Admins can promote a member to admin or demote another admin to member (subject to last-admin protection).

---

## API

| Method | Path | Guard | Description |
|--------|------|-------|-------------|
| `PUT` | `/api/v1/families/{familyID}/members/{memberID}/role` | admin | Change a member's role. Body: `{"role": "admin"\|"member"}` |
| `DELETE` | `/api/v1/families/{familyID}/members/{memberID}` | admin | Kick a real member |
| `POST` | `/api/v1/families/{familyID}/members/virtual` | admin | Create virtual member |
| `DELETE` | `/api/v1/families/{familyID}/members/virtual/{memberID}` | admin | Remove virtual member |
| `POST` | `/api/v1/families/{familyID}/invites` | admin | Generate invite link |
| `DELETE` | `/api/v1/families/{familyID}/invites/{token}` | admin | Revoke invite link |
| `POST` | `/api/v1/families/{familyID}/categories` | admin | Create category |
| `PUT` | `/api/v1/families/{familyID}/categories/{categoryID}` | admin | Edit category |
| `DELETE` | `/api/v1/families/{familyID}/categories/{categoryID}` | admin | Delete category |

---

## UI

### Settings > Members

- Every member row shows a role badge (`Admin` / `Member`).
- If the viewer is an admin:
  - Non-self real members show a role toggle button (promote ↔ demote).
  - Non-self real members show a kick button.
  - Last-admin protection: the toggle is disabled when demoting would leave zero admins.
- If the viewer is a member: badges are read-only; no kick or toggle buttons shown.

---

## Delta vs current implementation

| Item | Status |
|------|--------|
| `admin` / `member` roles stored in DB | ✅ implemented |
| Family creator assigned `admin` | ✅ implemented |
| Admin-only kick endpoint | ✅ implemented |
| Admin-only role change endpoint (`PUT /members/{id}/role`) | ✅ implemented |
| Last-admin demotion protection | ✅ implemented |
| Admin-only gate on categories | ✅ implemented |
| Admin-only gate on virtual members | ✅ implemented |
| Admin-only gate on invites | ✅ implemented |
| Role badge in settings UI | ✅ implemented |
| Role toggle button (admin only) in settings UI | ✅ implemented |
| Conditional kick button (admin only) in settings UI | ✅ implemented |
| New members join as `member` | ✅ implemented (invite accept sets role = "member") |
| Leave family flow | ☐ not yet implemented |
