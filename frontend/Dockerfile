FROM node:22-alpine AS builder
WORKDIR /app
COPY . ./
RUN rm -rf node_modules package-lock.json && npm i
COPY . .
RUN npm run build

FROM node:22-alpine
ENV NODE_ENV=production
USER node
WORKDIR /app
COPY --from=builder --chown=node:node /app/build ./build
COPY --from=builder --chown=node:node /app/node_modules ./node_modules
COPY --chown=node:node package.json .
ENV PORT=443
EXPOSE ${PORT}
CMD ["node","build"]