# syntax=docker/dockerfile:1.6

ARG GO_VERSION=1.23.5

# Fast builder (uses BuildKit mounts if available)
FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /src
ENV GOPROXY=https://proxy.golang.org,direct

# Leverage layer caching for deps
COPY go.mod .
COPY go.sum* .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# Build a static binary
ENV CGO_ENABLED=0
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN mkdir -p /out
RUN --mount=type=bind,source=.,target=/src,ro \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -mod=readonly -buildvcs=false -ldflags "-s -w" -o /out/itadakimasu-dl .

# Fallback builder (no BuildKit; uses plain COPY)
FROM golang:${GO_VERSION}-alpine AS build_nobk
WORKDIR /src
ENV GOPROXY=https://proxy.golang.org,direct
COPY go.mod .
COPY go.sum* .
RUN go mod download
COPY . .
ENV CGO_ENABLED=0
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN mkdir -p /out
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -mod=readonly -buildvcs=false -ldflags "-s -w" -o /out/itadakimasu-dl .

# Runtime
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app

# Copy binary from selected builder
COPY --from=build /out/itadakimasu-dl /app/itadakimasu-dl

# Ship default config; can be overridden with a bind mount
COPY config.json /app/config.json

# Location for downloads (mounted via compose)
VOLUME ["/downloads"]

ENTRYPOINT ["/app/itadakimasu-dl"]