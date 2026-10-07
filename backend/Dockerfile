# syntax=docker/dockerfile:1
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o /api ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates \
    && addgroup -S bc \
    && adduser -S bc -G bc
COPY --from=build --chown=bc:bc /api /api
USER bc
EXPOSE 8080
ENTRYPOINT ["/api"]
