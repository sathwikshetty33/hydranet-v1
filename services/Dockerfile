# syntax=docker/dockerfile:1.7

# ── Build stage ──
FROM golang:1.26-alpine AS build

WORKDIR /src

# Cache modules separately from source so source-only changes don't
# re-pull the dependency graph.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

# CGO off ⇒ fully static binary that runs on distroless/scratch.
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/core ./cmd/core

# ── Runtime stage ──
FROM gcr.io/distroless/static-debian12:nonroot AS runtime

WORKDIR /
COPY --from=build /out/core /core

# 8080 = HTTP API (dashboard, agent registration)
# 50051 = gRPC (agent → core incident stream)
EXPOSE 8080 50051

USER nonroot:nonroot
ENTRYPOINT ["/core"]
