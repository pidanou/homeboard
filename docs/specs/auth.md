# Auth Spec

## Model

Two independent login methods, both optional to enable, at least one must be usable:

| Method | Enabled by | Notes |
|--------|-----------|-------|
| Email + password | default on | can be disabled via `ALLOW_PASSWORD_LOGIN=false` |
| Generic OIDC (SSO) | `OIDC_ISSUER_URL` + `OIDC_CLIENT_ID` + `OIDC_CLIENT_SECRET` all set, and discovery succeeds at boot | works with any OIDC-compliant self-hosted IdP (Authentik, Keycloak, Authelia, Zitadel, etc.) — not hardcoded per-provider integrations |

A `users` row now has a nullable `password_hash` — an OIDC-only account has none. The link between a user and an external identity lives in a separate `oidc_identities` table, keyed by `(issuer, subject)`, so a user can in principle be linked to more than one IdP identity later without another migration.

Both login methods issue the same app JWT via `AuthService.IssueToken` — there is one source of truth for token shape/expiry regardless of how the user authenticated.

---

## Config (env vars)

| Var | Required | Default | Notes |
|---|---|---|---|
| `ALLOW_PASSWORD_LOGIN` | no | `true` | set `false` to go OIDC-only |
| `OIDC_ISSUER_URL` | for OIDC | — | e.g. `https://auth.example.com/application/o/homeboard/` |
| `OIDC_CLIENT_ID` | for OIDC | — | |
| `OIDC_CLIENT_SECRET` | for OIDC | — | |
| `OIDC_REDIRECT_URL` | no | `${API_BASE_URL}/api/v1/auth/oidc/callback` | override for reverse-proxy path rewrites |
| `OIDC_PROVIDER_NAME` | no | `SSO` | shown on the frontend button, e.g. "Sign in with Authentik" |

`GET /api/v1/config` exposes `oidcEnabled`, `oidcProviderName`, and `allowPasswordLogin` read-only to the frontend — availability is never hardcoded client-side.

If OIDC env vars are set but discovery fails at startup (unreachable issuer, malformed discovery document), the backend logs a warning and starts with OIDC disabled rather than failing to boot. Password login (if enabled) keeps working.

---

## Account linking rule

On a successful OIDC login:

1. **Known identity** — `(issuer, subject)` already has a linked user → log in as that user.
2. **Unknown identity, verified email match** — the IdP asserts `email_verified: true` and the email matches an existing account (password-based or otherwise) → **auto-link**, no confirmation step, log in as that user.
3. **Unknown identity, verified email, no match** — provision a new user (`password_hash = NULL`), subject to the same `ALLOW_REGISTRATION` / first-user bootstrap gate that email/password registration uses.
4. **Unverified email** (`email_verified` false or absent) → reject outright. No account is created or linked. This is a deliberately safe default — an IdP that doesn't assert email verification can't be trusted to prove identity ownership.

---

## Backend flow

State, nonce, and PKCE verifier are carried in a short-lived signed `HttpOnly` cookie (`oidc_flow`, `SameSite=Lax`, 10 min TTL) — there's no server-side session store in this stack. The cookie is HMAC-signed with a key derived from `JWT_SECRET` via HKDF (a distinct key from the one that signs the app JWT).

1. `GET /api/v1/auth/oidc/login` — generates state/nonce/PKCE verifier, signs them into the `oidc_flow` cookie, redirects to the IdP's authorization endpoint.
2. `GET /api/v1/auth/oidc/callback` — verifies the cookie (HMAC + expiry), checks `state` against the callback param (CSRF), exchanges the code (PKCE), verifies the ID token (JWKS/signature/issuer/audience/expiry via `go-oidc`), checks `nonce` against the cookie (replay defense — `go-oidc` does not do this automatically), clears the cookie, then runs the account-linking rule above.
3. On success, the backend mints the app JWT and stores `{code → jwt}` in an in-memory, single-use, 60s-TTL map, then redirects to a **fixed** `${APP_BASE_URL}/callback?code=...` — never a caller-supplied URL, which rules out backend-side open-redirect entirely.
4. On failure, redirects to `${APP_BASE_URL}/callback?error=<reason>` where `reason` is one of `email_not_verified`, `registration_closed`, `oidc_failed` — internal error detail is never leaked to the redirect target.
5. `POST /api/v1/auth/oidc/exchange {code}` — consumes the handoff code, returns `{token}` as JSON. The JWT is never placed in a URL at any hop.

The in-memory handoff store is safe because the self-hosted deployment runs a single backend replica (`docker-compose.yml` has no scaling config); this would need to move to a shared store if that ever changes.

IdP access/refresh tokens are discarded immediately after the code exchange — the app never calls back into the IdP's API, so there's nothing to refresh and nothing extra at risk if the DB is later compromised.

---

## Frontend flow

Backend and frontend are different origins, so the callback can't hand off a JWT via `localStorage` or a JS-readable cookie directly.

- The login page does a full page navigation (not `fetch`) to `/api/v1/auth/oidc/login` so the browser follows the IdP redirect chain. The intended post-login destination (`redirect` query param) is stashed in `sessionStorage` first, since the backend's fixed redirect target can't carry it.
- `/auth/callback` reads `?code=` (or `?error=`), exchanges the code via `POST /api/v1/auth/oidc/exchange`, stores the returned JWT the same way the password flow does, restores the stashed `redirect`, and navigates there.
- The password form is only rendered when `allowPasswordLogin` is true; the "Continue with {provider}" button is only rendered when `oidcEnabled` is true. The register page redirects to `/login` when password login is disabled (OIDC accounts are provisioned on first login, not through a separate registration step).

---

## API

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/api/v1/auth/oidc/login` | public | Starts the OIDC authorization code flow (redirect) |
| `GET` | `/api/v1/auth/oidc/callback` | public | IdP redirects here; redirects to the frontend with a handoff code or error reason |
| `POST` | `/api/v1/auth/oidc/exchange` | public | `{code}` → `{token}`, single-use, 60s TTL |
| `GET` | `/api/v1/config` | public | Exposes `oidcEnabled`, `oidcProviderName`, `allowPasswordLogin` |

---

## Delta vs current implementation

| Item | Status |
|------|--------|
| `password_hash` nullable, `oidc_identities` table (migration `000030`) | ✅ implemented |
| Generic OIDC discovery + PKCE + nonce via `go-oidc` / `oauth2` | ✅ implemented |
| Auto-link by verified email | ✅ implemented |
| Reject on unverified email | ✅ implemented |
| Graceful startup degrade on discovery failure | ✅ implemented |
| `ALLOW_PASSWORD_LOGIN` gate (backend + handler 403) | ✅ implemented |
| One-time handoff code + POST exchange (no JWT in URL) | ✅ implemented |
| `/api/v1/config` exposes OIDC/password-login flags | ✅ implemented |
| Frontend conditional password form / SSO button | ✅ implemented |
| Frontend `/auth/callback` page | ✅ implemented |
| Google OAuth (hardcoded, per-provider) | ☐ not planned — superseded by generic OIDC |
| Apple OAuth (hardcoded, per-provider) | ☐ not planned — superseded by generic OIDC |
| Unlink / view-linked-identity UI | ☐ not yet implemented |
| Forgot password / reset flow | ☐ not yet implemented (tracked separately in roadmap M10) |

---

## Manual verification checklist

No browser-test infra exists in `web/` (no Vitest/Playwright configured) — run this checklist by hand against a real or mock OIDC IdP (e.g. `ghcr.io/navikt/mock-oauth2-server`) before release:

- [ ] New-user OIDC signup: unknown identity, verified email, no matching account → new user created, logged in.
- [ ] Auto-link: unknown identity, verified email matching an existing password account → logged into the existing account, `oidc_identities` row created.
- [ ] Returning user: known `(issuer, subject)` → logged in without re-checking email.
- [ ] Unverified email rejected: `email_verified: false` → login rejected, no account created or linked, frontend shows a clear error.
- [ ] `ALLOW_REGISTRATION=false` blocks new-user OIDC signup the same way it blocks password registration.
- [ ] `ALLOW_PASSWORD_LOGIN=false` hides the password form and register page redirects to `/login`.
- [ ] Expired or already-used handoff code is rejected by `/exchange` with a clear frontend error.
- [ ] Tampered or missing `state`/`nonce` is rejected by the callback.
- [ ] The JWT never appears in a browser-visible URL at any redirect hop (check devtools network tab / address bar through the whole flow).
- [ ] OIDC issuer unreachable at boot → backend still starts, `oidcEnabled` is `false`, password login still works if enabled.
