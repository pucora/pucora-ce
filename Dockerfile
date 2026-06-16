ARG GOLANG_VERSION=1.25.11
ARG ALPINE_VERSION=3.23
ARG VERSION=2.0.0

FROM golang:${GOLANG_VERSION}-bookworm AS builder

ARG VERSION

WORKDIR /app

COPY . .

# Offline build when vendor/ is present (local dev); otherwise fetch modules (CI).
RUN set -eux; \
    test -f cmd/velonetics-ce/schema/schema.json; \
    if [ -d vendor ]; then \
      CGO_ENABLED=0 GOPROXY=off go build -mod=vendor \
        -ldflags="-s -w \
          -X github.com/velonetics/velonetics-ce/v2/pkg.Version=${VERSION} \
          -X github.com/velonetics/lura/v2/core.VeloneticsVersion=${VERSION}" \
        -o velonetics ./cmd/velonetics-ce; \
    else \
      CGO_ENABLED=0 go build \
        -ldflags="-s -w \
          -X github.com/velonetics/velonetics-ce/v2/pkg.Version=${VERSION} \
          -X github.com/velonetics/lura/v2/core.VeloneticsVersion=${VERSION}" \
        -o velonetics ./cmd/velonetics-ce; \
    fi

FROM alpine:${ALPINE_VERSION}

LABEL maintainer="community@velonetics.io"

# CA bundle from builder — avoids apk in runtime (fixes TLS issues on some Docker hosts).
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/velonetics /usr/bin/velonetics

RUN mkdir -p /etc/velonetics \
    && echo '{"version":3}' > /etc/velonetics/velonetics.json

USER 1000:1000

WORKDIR /etc/velonetics

ENTRYPOINT ["/usr/bin/velonetics"]
CMD ["run", "-c", "/etc/velonetics/velonetics.json"]

EXPOSE 8080 8090
