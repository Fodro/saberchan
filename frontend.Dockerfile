# syntax=docker/dockerfile:1

# Build context: frontend/ (see docker-compose*.yaml).
# Private SvelteKit env ($env/static/private) is baked at build time via ARG/ENV below.
# Keep NODE_OPTIONS modest — adapter-node tracing OOMs low-RAM Colima VMs (see DEPLOY.md).
#
# Production: do NOT set ALLOW_INSECURE_TLS (default empty = normal TLS verify).
# Local compose may pass ALLOW_INSECURE_TLS=0 for self-signed OIDC experiments;
# local Keycloak over HTTP does not need it.

FROM node:24-alpine AS builder

ARG OIDC_REALM
ARG OIDC_REALM_INTERNAL
ARG OIDC_CLIENT_ID
ARG OIDC_CLIENT_SECRET
ARG AUTH_HOST
ARG MAIN_BACKEND_URL
ARG PORT=3000
ARG AUTH_SECRET
ARG ADMIN_API_TOKEN
ARG OIDC_REDIRECT_URI
ARG OIDC_LOGOUT_REDIRECT_URI

ENV OIDC_REALM=${OIDC_REALM} \
	OIDC_REALM_INTERNAL=${OIDC_REALM_INTERNAL} \
	OIDC_CLIENT_ID=${OIDC_CLIENT_ID} \
	OIDC_CLIENT_SECRET=${OIDC_CLIENT_SECRET} \
	AUTH_HOST=${AUTH_HOST} \
	MAIN_BACKEND_URL=${MAIN_BACKEND_URL} \
	PORT=${PORT} \
	AUTH_SECRET=${AUTH_SECRET} \
	ADMIN_API_TOKEN=${ADMIN_API_TOKEN} \
	OIDC_REDIRECT_URI=${OIDC_REDIRECT_URI} \
	OIDC_LOGOUT_REDIRECT_URI=${OIDC_LOGOUT_REDIRECT_URI} \
	NODE_OPTIONS="--max-old-space-size=1536" \
	UV_THREADPOOL_SIZE=2

WORKDIR /app

COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm \
	npm ci

COPY . .
RUN npm run build \
	&& npm prune --omit=dev

FROM node:24-alpine AS runner

# Empty = verify TLS (prod). Set build-arg ALLOW_INSECURE_TLS=0 only for local HTTPS experiments.
ARG ALLOW_INSECURE_TLS=
ENV NODE_ENV=production \
	BODY_SIZE_LIMIT=16M \
	NODE_TLS_REJECT_UNAUTHORIZED=${ALLOW_INSECURE_TLS}

USER node
WORKDIR /app

COPY --from=builder --chown=node:node /app/build ./build
COPY --from=builder --chown=node:node /app/node_modules ./node_modules
COPY --from=builder --chown=node:node /app/package.json ./

ARG OIDC_REALM
ARG OIDC_REALM_INTERNAL
ARG OIDC_CLIENT_ID
ARG OIDC_CLIENT_SECRET
ARG AUTH_HOST
ARG MAIN_BACKEND_URL
ARG PORT=3000
ARG AUTH_SECRET
ARG ADMIN_API_TOKEN
ARG OIDC_REDIRECT_URI
ARG OIDC_LOGOUT_REDIRECT_URI

ENV OIDC_REALM=${OIDC_REALM} \
	OIDC_REALM_INTERNAL=${OIDC_REALM_INTERNAL} \
	OIDC_CLIENT_ID=${OIDC_CLIENT_ID} \
	OIDC_CLIENT_SECRET=${OIDC_CLIENT_SECRET} \
	AUTH_HOST=${AUTH_HOST} \
	MAIN_BACKEND_URL=${MAIN_BACKEND_URL} \
	PORT=${PORT} \
	AUTH_SECRET=${AUTH_SECRET} \
	ADMIN_API_TOKEN=${ADMIN_API_TOKEN} \
	OIDC_REDIRECT_URI=${OIDC_REDIRECT_URI} \
	OIDC_LOGOUT_REDIRECT_URI=${OIDC_LOGOUT_REDIRECT_URI}

EXPOSE 3000
CMD ["node", "build"]
