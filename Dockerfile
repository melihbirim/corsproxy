# Build stage
FROM golang:1.27-alpine AS builder

ARG VERSION=dev
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath \
    -ldflags="-s -w -X main.version=${VERSION}" -o cors-proxy .

# Final stage: alpine for CA certs + wget (used by the compose healthcheck),
# but the proxy itself runs as an unprivileged user, not root.
FROM alpine:3.20

RUN apk --no-cache add ca-certificates \
    && adduser -D -H -u 10001 corsproxy
COPY --from=builder /app/cors-proxy /usr/local/bin/cors-proxy

USER corsproxy
EXPOSE 8080
CMD ["cors-proxy"]
