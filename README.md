# Saberchan

Anonymous imageboard written in Go + SvelteKit. Post with or without images in
customizable boards, moderated via a Keycloak-authenticated admin panel.

## Features

- Threaded anonymous posting with image/video upload
- Board creation and management (admin)
- Captcha-based post throttling (one-time Redis token)
- BBCode markup (`[b]`, `[i]`, `[spoiler]`, etc.)
- OIDC authentication via Keycloak
- Post deletion / thread archiving with configurable grace periods
- Follow/unfollow threads (client-side cookie)
- i18n (locale switching)
- Dark mode
- S3-compatible object storage (MinIO local, Garage production, any S3 API)
- Rate limiting (captcha generation, post creation)
- Prometheus metrics endpoint

## Architecture

| Component | Role |
|-----------|------|
| **Frontend** | SvelteKit BFF — server-side rendering, API proxy, OIDC login flow |
| **Backend** | Go REST API — post/thread CRUD, captcha, admin actions, S3 upload |
| **Keycloak** | OIDC provider — admin authentication, JWT issuance + verification |
| **Postgres** | App database + Keycloak database |
| **Redis** | Captcha token store, rate limiter backend |
| **S3** | Image/video storage (MinIO, Garage, AWS S3, etc.) |
| **Nginx** | Reverse proxy (user-provided config) |

## Production deploy

See [DEPLOY.md](DEPLOY.md) for full details.

### Prerequisites

- VPS with Docker + Compose plugin
- External Postgres (two databases: `saberchan` + `keycloak`)
- External Redis
- External S3-compatible storage (Garage, MinIO, AWS S3, DigitalOcean Spaces, etc.)
- Nginx on the host (you write the config — see service map below)

### Service map

```
Internet → example.com → 127.0.0.1:3000 (frontend)
         → example.com/cloak/ → 127.0.0.1:8080 (Keycloak)
         → example.com/media/ → S3 bucket
```

- Frontend binds `127.0.0.1:3000`
- Keycloak binds `127.0.0.1:8080`
- Backend is **not published** — reachable only on the compose network as `http://backend:8888`

### Quick start

```bash
cp .env.prod.dist .env.prod
# edit every placeholder (hostnames, secrets, DB/S3 credentials)
make prod-up          # pulls images from ghcr.io and starts
make prod-ps          # wait for healthy
```

Images are pulled from `ghcr.io/fodro/saberchan/{frontend,backend}`. Set
`FRONTEND_TAG` / `BACKEND_TAG` in `.env.prod` to pin a specific release.

### Nginx

You will need to write your own nginx config to proxy:

| Path | Upstream |
|------|----------|
| `/` | `127.0.0.1:3000` (frontend) |
| `/cloak/` | `127.0.0.1:8080` (Keycloak) |
| `/media/` | your S3 bucket (optional) |

## Local development

```bash
cp .env.local.dist .env.local
make local-up
```

See [DEPLOY.md](DEPLOY.md) for Keycloak setup, OIDC issuer split, and S3 config.

## CI

Pull requests run lint, type-check, tests, and build via GitHub Actions.
Merges to `main` create a version tag + GitHub release and push Docker images
to ghcr.io. Tag is bumped based on PR labels (`major`, `patch`, or default
minor).

## Planned features

- Auto-archiving thread as plain HTML to S3 (with files intact) + built-in archive browser with search
- Auto-moderation
- More sane mobile version
- NGINX pre-built configuration