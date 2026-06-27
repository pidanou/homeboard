# Homeboard

A self-hostable family wall — shared calendar, tasks, and shopping lists for your household.

**Live, real-time sync** across all devices. Install as a PWA or run as a native iOS/Android app via Capacitor.

---

<img src="docs/screenshots/demo.gif" alt="Homeboard demo" width="100%"/>

<table>
  <tr>
    <td align="center"><img src="docs/screenshots/today.png" alt="Today view" width="220"/><br/><sub>Today</sub></td>
    <td align="center"><img src="docs/screenshots/calendar.png" alt="Calendar" width="220"/><br/><sub>Calendar</sub></td>
    <td align="center"><img src="docs/screenshots/board.png" alt="Board" width="220"/><br/><sub>Board</sub></td>
    <td align="center"><img src="docs/screenshots/lists.png" alt="Shopping lists" width="220"/><br/><sub>Lists</sub></td>
  </tr>
</table>

---

## Features

- **Calendar** — month, week, day, and agenda views; recurring events; birthday events; drag-and-drop rescheduling
- **Tasks** — due dates, assignees, priority, labels, drag-to-reorder
- **Shopping lists** — multiple lists, checked/unchecked sections, bulk clear
- **Today view** — overdue items, today's events and tasks in one place
- **Real-time sync** — all changes pushed to every family member via SSE
- **Virtual members** — add kids or non-app members; link to a real account later
- **Roles** — admin and member roles with access control
- **PWA** — installable on iOS, Android, and desktop from the browser
- **Native apps** — iOS and Android via Capacitor

---

## Self-hosting

**Requirements:** Docker and Docker Compose.

```bash
git clone https://github.com/your-username/homeboard.git
cd homeboard
cp .env.example .env
# Edit .env — set POSTGRES_PASSWORD and JWT_SECRET at minimum
docker compose up -d
```

The app is now running at `http://localhost:3000`. The API is at `http://localhost:8080`.

### Environment variables

| Variable | Required | Description |
|---|---|---|
| `POSTGRES_PASSWORD` | ✅ | PostgreSQL password |
| `JWT_SECRET` | ✅ | Secret for signing JWTs — use a long random string |
| `API_BASE_URL` | ✅ | Public URL of the backend (e.g. `https://api.yourdomain.com`) |
| `PUBLIC_API_URL` | ✅ | Same value — used by the frontend at runtime |
| `APP_BASE_URL` | | Frontend URL — added to CORS allowed origins |
| `CORS_ALLOWED_ORIGINS` | | Extra CORS origins, comma-separated, or `*` |
| `UPLOAD_DIR` | | Directory for avatar uploads (defaults to `./uploads`) |

### Reverse proxy

To expose the app on a domain, proxy `yourdomain.com` → port `3000` and `api.yourdomain.com` → port `8080`, then set:

```env
API_BASE_URL=https://api.yourdomain.com
PUBLIC_API_URL=https://api.yourdomain.com
APP_BASE_URL=https://yourdomain.com
```

---

## Installing on your phone

Homeboard works as a PWA — no app store needed. Open the app URL in your browser and install it to your home screen.

**iOS (Safari)**
1. Open the app in **Safari** (other browsers don't support PWA install on iOS)
2. Tap the **Share** button (box with arrow at the bottom)
3. Scroll down and tap **Add to Home Screen**
4. Confirm — the app icon appears on your home screen

**Android (Chrome)**
1. Open the app in **Chrome**
2. Tap the **⋮ menu** → **Add to Home Screen** (or look for the install prompt in the address bar)
3. Tap **Install**

The app will open full-screen with no browser chrome, just like a native app.

---

## Architecture

```
homeboard/
├── backend/    # Go REST API (Postgres, custom JWT auth)
├── web/        # SvelteKit PWA + Capacitor (iOS/Android)
└── docker-compose.yml
```

- **Backend:** Go, PostgreSQL, `golang-migrate` for schema migrations, SSE for real-time push
- **Frontend:** SvelteKit, Tailwind CSS, shadcn-svelte, Capacitor

---

## Development

```bash
# Backend (requires a running Postgres)
cd backend && go run ./cmd/server

# Web
cd web && npm install && npm run dev
```

Copy `.env.example` to `.env` and fill in `DATABASE_URL` and `JWT_SECRET` before starting the backend.

---

## License

MIT
