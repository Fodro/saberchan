FROM node:22-alpine AS builder
WORKDIR /app
COPY ./frontend ./
RUN ls
RUN npm ci
COPY . .
RUN ls
RUN npm run build
RUN ls

FROM node:22-alpine
ENV NODE_ENV=production
USER node
WORKDIR /app
RUN ls
COPY --from=builder --chown=node:node /app/build ./build
RUN ls
COPY --from=builder --chown=node:node /app/node_modules ./node_modules
RUN ls
COPY --chown=node:node package.json .
RUN ls
ENV PORT=5127
EXPOSE ${PORT}
CMD ["node","build"]