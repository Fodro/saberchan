# Deploy & local development

## Topology

| Environment | App | Postgres | Redis | Object storage | Auth |
|-------------|-----|----------|-------|----------------|------|
| **Production** | `docker-compose.yaml` (+ optional Caddy) | external | external | external S3 | external Keycloak / OIDC |
| **Local** | `docker-compose.local.yaml` | in compose | in compose | MinIO in compose | Keycloak in compose |

Production keeps the data plane (DB, S3, Redis) and IdP outside the app compose file. Point env vars at those services via `.env.prod`.

## Prerequisites

- Go (backend toolchain matching `backend/go.mod`)
- Node 22+
- Docker Engine + Compose plugin (`docker compose version`)
  - On Homebrew: `brew install docker-compose`, then add to `~/.docker/config.json`:
    ```json
    "cliPluginsExtraDirs": ["/opt/homebrew/lib/docker/cli-plugins"]
    ```
  - Colima (or Docker Desktop) must be running for `make local-up`
  - **Colima RAM:** give the VM **≥4 GiB** (2 GiB often SIGKILLs the frontend image at `@sveltejs/adapter-node`). Example: `colima stop && colima start --cpu 2 --memory 4`
  - `make local-up` builds images **one at a time** (`COMPOSE_PARALLEL_LIMIT=1`) to reduce peak memory
- Global Go tools: `make backend-install-deps` (installs `mockgen` into `$(go env GOPATH)/bin` — keep that dir on `PATH`)

## Local CI gate

```bash
cp frontend/.env.dist frontend/.env   # once; make ci also auto-copies if missing
make ci
```

This runs:

- `go test ./...` and `go build` in `backend/`
- `npm ci`, `npm run check`, `npm test`, `npm run build` in `frontend/`

Regenerate gomock mocks after interface changes:

```bash
make backend-mocks   # runs backend-install-deps, then go generate
```

Other targets: `make help`.

## Full local stack

```bash
cp .env.local.dist .env.local
# edit secrets / OIDC client secret after Keycloak setup if needed
make local-up
```

Services (defaults from `.env.local.dist`):

| Service | Host port |
|---------|-----------|
| Frontend | http://localhost:3000 |
| Backend API | http://localhost:8888 (local only — debugging) |
| Keycloak | http://localhost:9090 |
| MinIO API | http://localhost:9000 |
| MinIO console | http://localhost:9001 |
| Postgres | localhost:5432 |
| Redis | localhost:6379 |

Useful:

```bash
make local-ps
make local-logs
make local-down
```

### First-time Keycloak

1. Open http://localhost:9090 — login `admin` / `admin` (or values from `.env.local`).
2. Create realm `saberchan` (or match `OIDC_REALM`).
3. Create a confidential client `saberchan` with:
   - **Valid redirect URIs:** `http://localhost:3000/admin/auth/signIn`
   - **Valid post logout redirect URIs:** `http://localhost:3000/admin/auth/signOut`
   - **Web origins:** `http://localhost:3000`
4. Put the client secret into `.env.local` as `OIDC_CLIENT_SECRET` and rebuild frontend (`make local-up` again).

OIDC issuer URLs (local Docker):

| Var | Example | Used for |
|-----|---------|----------|
| `OIDC_REALM` | `http://localhost:9090/realms/saberchan` | Browser authorize + logout links + JWT `iss` |
| `OIDC_REALM_INTERNAL` | `http://keycloak:8080/realms/saberchan` | Token/refresh + JWKS from the **frontend container** |

If `OIDC_REALM_INTERNAL` is missing, the callback hits `localhost:9090` *inside* the frontend container → `ECONNREFUSED` / HTTP 500.

Admin UI actions require a **cryptographically verified** Keycloak access token (JWKS) before the BFF attaches `X-Admin-Token`.

### Request body size (posts with images)

`adapter-node` defaults to **512KB**. Local/prod frontend should set:

```bash
BODY_SIZE_LIMIT=16M
```

(already in `.env.local.dist` and `frontend.Dockerfile`). Client + BFF also pre-validate title/text/captcha/files and toast a clear error instead of a generic Internal Error.

Captcha is enforced on the **Go API** for create post/thread (one-time Redis token). The BFF only checks that captcha fields are present and forwards them — it does not consume the token first.

### `ORIGIN` (required for multipart uploads)

Create post/thread uses `multipart/form-data`. SvelteKit CSRF compares the browser `Origin` header to the request URL origin. If `ORIGIN` is unset, adapter-node often treats the app as `https://…` even when you browse over `http://`, and you get:

`Cross-site POST form submissions are forbidden` (HTTP 403).

Set it to the exact public URL users type in the browser (scheme + host + port):

```bash
ORIGIN=http://localhost:3000   # local Docker
# ORIGIN=https://board.example.com   # production
```

Local compose already passes `ORIGIN` (defaults to `AUTH_HOST`).

Logout uses Keycloak’s end-session endpoint with  
`post_logout_redirect_uri={AUTH_HOST}/admin/auth/signOut`  
(local default matches `OIDC_LOGOUT_REDIRECT_URI` in `.env.local.dist`). That URI must be listed under **Valid post logout redirect URIs**, or Keycloak will reject the redirect after logout.

### Local MinIO / S3

Local defaults in `.env.local.dist`:

| Var | Value | Purpose |
|-----|-------|---------|
| `S3_URL` | `minio:9000` | SDK endpoint host (in-network) |
| `S3_USE_SSL` | `false` | HTTP to MinIO |
| `S3_FORCE_PATH_STYLE` | `true` | path-style API (`/bucket/key`) |
| `S3_PUBLIC_URL` | `http://localhost:9000` | browser-facing link prefix |

Object links look like `http://localhost:9000/saberchan/<key>`. Production typically keeps `S3_USE_SSL=true`, `S3_FORCE_PATH_STYLE=false`, and leaves `S3_PUBLIC_URL` empty (or set to a CloudFront URL).

## Production (AWS VPS + compose)

```bash
cp .env.prod.dist .env.prod
# fill RDS / Redis / S3 / OIDC / ADMIN_API_TOKEN / AUTH_HOST=https://…
make prod-up
```

What this starts:

| Service | Published | Notes |
|---------|-----------|--------|
| `frontend` | `${FRONTEND_PORT:-3000}` | Public app / BFF |
| `backend` | **none** | Only on compose network at `http://backend:8888` |
| `caddy` | 80/443 | Optional: `docker compose --env-file .env.prod --profile proxy up -d` |

Edit [`deploy/caddy/Caddyfile`](deploy/caddy/Caddyfile) / set `CADDY_SITE` to your hostname. Put TLS on Caddy (or a host nginx) and set `ORIGIN` + `AUTH_HOST` to that HTTPS URL.

**Security notes for prod:**

- Never publish `:8888` — captcha is enforced on Go, but the BFF is still the intended public entrypoint.
- `ADMIN_API_TOKEN` must match on Go + frontend; rotate requires rebuilding the frontend image (build-args bake `$env/static/private`).
- Frontend image verifies TLS by default. Local-only: pass build-arg `ALLOW_INSECURE_TLS=0` if you must talk to a self-signed IdP (not used by default local HTTP Keycloak).

`PURGE_INTERVAL` (default `10m`) sweeps soft-deleted rows after the 24h grace window (S3 media included) and soft-deletes threads not bumped for 30 days.

`TRUSTED_PROXIES` (comma-separated CIDRs) controls which peers may set `X-Forwarded-For`. The BFF forwards the browser address on create/captcha calls. Empty = never trust XFF.

Rate limits (per client IP, in-process): captcha generate ~30/min, create post/thread ~12/min.

Health endpoints on the Go API (compose network):

- `GET /liveness` — process up
- `GET /readiness` — DB + Redis ping
