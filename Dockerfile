# syntax=docker/dockerfile:1.7

FROM node:24-alpine AS node-base
RUN corepack enable && corepack prepare pnpm@10.33.0 --activate
WORKDIR /workspace

FROM golang:1.25-alpine AS go-base
RUN apk add --no-cache git ca-certificates
WORKDIR /workspace

FROM node-base AS deps
COPY pnpm-lock.yaml pnpm-workspace.yaml package.json ./
COPY client/package.json client/package.json
RUN pnpm install --frozen-lockfile

FROM deps AS client-builder
COPY client client
COPY openapi openapi
RUN pnpm --dir client run codegen && pnpm --dir client build

FROM go-base AS server-builder
COPY server server
COPY openapi openapi
WORKDIR /workspace/server
RUN go generate ./... && go mod tidy && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -o /out/server ./cmd/api

FROM go-base AS server-dev
RUN go install github.com/air-verse/air@latest
WORKDIR /workspace/server
ENV PATH="/go/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
CMD ["air", "-c", ".air.toml"]

FROM node-base AS client-dev
WORKDIR /workspace/client
CMD ["sh", "-c", "pnpm install --frozen-lockfile && pnpm dev --host 0.0.0.0"]

FROM alpine:3.22 AS app
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=server-builder /out/server /app/server
COPY --from=client-builder /workspace/client/dist /app/assets
ENV API_ADDR=:8080
ENV ASSETS_DIR=/app/assets
ENTRYPOINT ["/app/server"]
