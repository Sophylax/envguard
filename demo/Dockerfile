FROM golang:1.22-bookworm AS builder

WORKDIR /src
COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-X main.version=demo" -o /out/envguard .

FROM ghcr.io/charmbracelet/vhs@sha256:9d5fc3dc0c160b0fb1d2212baff07e6bdf3fa9438c504a3237484567302fcf93

RUN apt-get update \
    && apt-get install -y --no-install-recommends git ca-certificates \
    && git config --global --add safe.directory /vhs/.demo/envguard-demo \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /out/envguard /usr/local/bin/envguard
