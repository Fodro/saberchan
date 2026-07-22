# Deploy & local development

## Topology

| Environment | App | Postgres | Redis | Object storage | Auth | TLS |
|-------------|-----|----------|-------|----------------|------|-----|
| **Production** | `docker-compose.yaml` (frontend + backend + Keycloak — images from ghcr.io) | external | external | external S3 / Garage | Keycloak in compose | **host nginx** |
| **Local** | `docker-compose.local.yaml` | in compose | in compose | MinIO in compose | Keycloak in compose | none (HTTP) |

Production keeps Postgres, Redis, and object storage outside compose. Keycloak runs in compose and is reached via nginx at `/cloak/`. Point env vars at those services via `.env.prod`.

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
# ORIGIN=https://example.com   # production
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
| `S3_PUBLIC_URL` | `http://localhost:9000` | browser-facing origin (bucket appended) |

Object links look like `http://localhost:9000/saberchan/<key>`.

### Production media: any S3-compatible storage

Any S3-compatible storage works (Garage, MinIO, AWS S3, DigitalOcean Spaces, etc.).

| Role | Config |
|------|--------|
| Upload/delete (API → Garage) | `S3_URL=garage.internal:3900`, `S3_FORCE_PATH_STYLE=true`, `S3_USE_SSL=false` |
| Browser URLs | `S3_PUBLIC_URL=https://example.com/media` → `https://example.com/media/<key>` |
`S3_PUBLIC_URL` rules:

- Origin only (`https://example.com`) → append `/{bucket}` (same shape as local MinIO).
- Origin + path (`https://example.com/media`) → use as the full link prefix (no bucket in the public URL); nginx must add the bucket when proxying.
- Already ends with `/{bucket}` → used as-is.

The API rewrites attachment links from the object `key` + current public prefix on read (no DB migration when you change the public origin).

```bash
S3_URL=garage.internal:3900
S3_BUCKET=saberchan
S3_PUBLIC_URL=https://example.com/media
S3_USE_SSL=false
S3_FORCE_PATH_STYLE=true
```

## Production deploy (VPS + compose + host nginx)

Assumes one VPS with Docker Compose, host nginx for TLS, and external Postgres / Redis / S3-compatible storage.

### 0. DNS

| Hostname | Points to | Role |
|----------|-----------|------|
| `example.com` | VPS public IP | App + Keycloak at `/cloak/` |

### 1. External data plane (before compose)

On Postgres (same host as the app DB is fine):

```sql
CREATE DATABASE saberchan;
CREATE DATABASE keycloak;
-- grant your DB user access to both
```

Also have reachable Redis and Garage (S3 API). Create bucket `saberchan` (or match `S3_BUCKET`).

### 2. Env file

```bash
cp .env.prod.dist .env.prod
```

Edit **every** placeholder. Critical mappings:

| Var | Example | Notes |
|-----|---------|--------|
| `AUTH_HOST` / `ORIGIN` | `https://example.com` | Exact browser URL (scheme + host) |
| `KC_HOSTNAME` | `https://example.com` | Keycloak public URL (nginx proxies `/cloak/`) |
| `OIDC_REALM` | `https://example.com/cloak/realms/saberchan` | Browser authorize / logout / JWT `iss` |
| `OIDC_REALM_INTERNAL` | `http://keycloak:8080/realms/saberchan` | Token/JWKS from **frontend container** |
| `KC_DB_URL` | `jdbc:postgresql://…:5432/keycloak` | Keycloak’s own DB |
| `DB_*` | app DB | Backend migrations + data |
| `S3_*` | S3-compatible storage | See media section above |
| `ADMIN_API_TOKEN` | long random | Same value for Go + frontend |
| `OIDC_CLIENT_SECRET` | from Keycloak | Fill after step 5, then restart |

`FRONTEND_PORT` / `KEYCLOAK_PORT` default to `3000` / `8080` and are bound to **`127.0.0.1` only**.

### 3. Host nginx

Write your own nginx config. Targets:

| Public | Upstream |
|--------|----------|
| `https://example.com/` | `127.0.0.1:3000` (frontend) |
| `https://example.com/cloak/` | `127.0.0.1:8080` (Keycloak) |
| `https://example.com/media/` | your S3 bucket (optional) |

Do **not** open Docker-published ports on `0.0.0.0`. Backend stays unpublished.

### 4. Start compose

```bash
make prod-up          # pulls images from ghcr.io and starts
make prod-ps          # wait until keycloak / backend / frontend are healthy
```

| Service | Published | Notes |
|---------|-----------|--------|
| `keycloak` | `127.0.0.1:${KEYCLOAK_PORT:-8080}` | Prod mode (`start`); nginx proxies `/cloak/` |
| `frontend` | `127.0.0.1:${FRONTEND_PORT:-3000}` | BFF; nginx proxies here |
| `backend` | **none** | Compose network only: `http://backend:8888` |

Useful: `make prod-logs`, `make prod-down`.

### 5. First-time Keycloak (prod)

1. Open `https://example.com/cloak/` — login with `KEYCLOAK_ADMIN` / `KEYCLOAK_ADMIN_PASSWORD`.
2. Create realm `saberchan` (must match the path in `OIDC_REALM`).
3. Create confidential client `saberchan`:
   - **Valid redirect URIs:** `https://example.com/admin/auth/signIn`
   - **Valid post logout redirect URIs:** `https://example.com/admin/auth/signOut`
   - **Web origins:** `https://example.com`
4. Copy the client secret into `.env.prod` as `OIDC_CLIENT_SECRET`.
5. Restart frontend (secrets are read at runtime):

```bash
make prod-up
```

OIDC issuer split (same idea as local):

| Var | Example | Used for |
|-----|---------|----------|
| `OIDC_REALM` | `https://example.com/cloak/realms/saberchan` | Browser authorize + logout + JWT `iss` |
| `OIDC_REALM_INTERNAL` | `http://keycloak:8080/realms/saberchan` | Token/refresh + JWKS from the frontend container |

If `OIDC_REALM_INTERNAL` points at the public hostname only, the frontend container must be able to resolve and reach it; the in-compose service name is the reliable default.

### 6. Smoke checks

```bash
curl -fsS http://127.0.0.1:3000/ >/dev/null          # frontend on loopback
curl -fsS https://example.com/ >/dev/null      # via nginx + TLS
curl -fsS https://example.com/cloak/ >/dev/null        # Keycloak via nginx
# from compose network (optional):
docker compose -f docker-compose.yaml --env-file .env.prod exec backend \
  wget -qO- http://127.0.0.1:8888/readiness
```

Then: open the board, generate captcha, create a post with an image, confirm `/media/…` loads, sign in at `/admin`.

### Security / ops notes

- Never publish `:8888` — captcha is enforced on Go; the BFF is the public entrypoint.
- `ADMIN_API_TOKEN` must match on Go + frontend. Secrets are read from the runtime environment (`$env/dynamic/private`); update `.env.prod` and restart to rotate.
- Frontend verifies TLS by default (no `ALLOW_INSECURE_TLS` in prod).
- `TRUSTED_PROXIES` (comma-separated CIDRs) controls who may set `X-Forwarded-For`. Include Docker bridge + `127.0.0.0/8` when nginx is on the host. Empty = never trust XFF.
- `PURGE_INTERVAL` (default `10m`) sweeps soft-deleted rows after the 24h grace window (S3 media included) and soft-deletes threads not bumped for 30 days.
- Rate limits (per client IP, in-process): captcha generate ~30/min, create post/thread ~12/min.

Health endpoints on the Go API (compose network only):

- `GET /liveness` — process up
- `GET /readiness` — DB + Redis ping

## P2 backlog (after 1.0.0)

Tracked leftovers from the pre-1.0 review — not blockers for the VPS demo.

- [ ] **One-shot migrate job** — run goose outside the long-lived API process so multi-replica starts don’t race
- [ ] **Structured logging + request IDs** — replace stdlib `log` / noisy boot prints
- [ ] **CI** — `govulncheck` / `npm audit` in GitHub Actions
- [ ] **S3 prod checklist** — bucket policy / CloudFront via `S3_PUBLIC_URL`, IAM least-privilege (expand this doc)
- [ ] **Dead `SECRET` env** — wire it for something real or remove from `config/env.go` + dist files
- [x] **Runtime secrets for frontend** — switched from `$env/static/private` to `$env/dynamic/private`
- [x] **S3 / media** — `S3_PUBLIC_URL` path-style public links; links rewritten from `key` on read
- [ ] **S3 ops checklist** — bucket public-read / key rotation
- [ ] **Upload magic-byte sniff** — don't trust `Content-Type` / extension alone
- [ ] **Backup runbook** — managed PG snapshots/PITR + S3 lifecycle/versioning; Redis is ephemeral by design
- [ ] **Shared rate-limit store** — move in-process limiters to Redis if backend replicas > 1
