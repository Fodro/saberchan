# syntax=docker/dockerfile:1

# Build context: backend/ (see docker-compose*.yaml).
# Runtime config comes from compose/k8s env — do not bake secrets into the image.

FROM golang:1.26-alpine AS builder

RUN apk add --no-cache ca-certificates git

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
	go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
	--mount=type=cache,target=/root/.cache/go-build \
	CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/saberchan ./main.go

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata wget \
	&& adduser -D -H -u 65532 nonroot

WORKDIR /app

COPY --from=builder /out/saberchan /app/main
COPY --from=builder /src/migrations /app/migrations

USER nonroot:nonroot
EXPOSE 8888
CMD ["/app/main"]
