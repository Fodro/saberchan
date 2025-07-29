FROM node:22-alpine AS builder
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

ENV PORT=5127
EXPOSE ${PORT}
CMD ["node","build"]