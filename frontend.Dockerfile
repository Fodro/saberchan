FROM node:22-alpine AS builder

ARG OIDC_REALM
ARG OIDC_CLIENT_ID
ARG OIDC_CLIENT_SECRET
ARG AUTH_HOST
ARG MAIN_BACKEND_URL
ARG PORT

ENV OIDC_REALM=${OIDC_REALM}
ENV OIDC_CLIENT_ID=${OIDC_CLIENT_ID}
ENV OIDC_CLIENT_SECRET=${OIDC_CLIENT_SECRET}
ENV AUTH_HOST=${AUTH_HOST}
ENV MAIN_BACKEND_URL=${MAIN_BACKEND_URL}
ENV PORT=${PORT}

WORKDIR /app
COPY ./frontend/ ./

RUN npm ci

RUN npm run build


FROM node:22-alpine
ENV NODE_ENV=production
USER node
WORKDIR /app

COPY --from=builder --chown=node:node /app/build ./build

COPY --from=builder --chown=node:node /app/node_modules ./node_modules

COPY --from=builder --chown=node:node /app/package.json .

EXPOSE 3000
CMD ["node","build"]