FROM golang:1.22-bookworm AS builder

WORKDIR /src
COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-X main.version=demo" -o /out/envguard .

FROM ghcr.io/charmbracelet/vhs:latest

RUN apt-get update \
    && apt-get install -y --no-install-recommends git ca-certificates \
    && git config --global --add safe.directory /vhs/.demo/envguard-demo \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /out/envguard /usr/local/bin/envguard
