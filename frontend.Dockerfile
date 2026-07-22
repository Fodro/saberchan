# syntax=docker/dockerfile:1

# Build context: frontend/ (see docker-compose*.yaml).
# Secrets are provided at runtime via docker-compose environment — no build args needed.
# Keep NODE_OPTIONS modest — adapter-node tracing OOMs low-RAM Colima VMs (see DEPLOY.md).
#
# Production: do NOT set ALLOW_INSECURE_TLS (default empty = normal TLS verify).
# Local compose may pass ALLOW_INSECURE_TLS=0 for self-signed OIDC experiments;
# local Keycloak over HTTP does not need it.

FROM node:24-alpine AS builder

ENV NODE_OPTIONS="--max-old-space-size=1536" \
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

EXPOSE 3000
CMD ["node", "build"]
