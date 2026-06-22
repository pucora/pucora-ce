ARG GOLANG_VERSION=1.26.4
ARG ALPINE_VERSION=3.23
ARG VERSION=2.0.0

FROM golang:${GOLANG_VERSION}-bookworm AS builder

ARG VERSION

WORKDIR /app

COPY . .

# Offline build when vendor/ is present (local dev); otherwise fetch modules (CI).
RUN set -eux; \
    test -f cmd/pucora-ce/schema/schema.json; \
    if [ -d vendor ]; then \
      CGO_ENABLED=0 GOPROXY=off go build -mod=vendor \
        -ldflags="-s -w \
          -X github.com/pucora/pucora-ce/v2/pkg.Version=${VERSION} \
          -X github.com/pucora/lura/v2/core.PucoraVersion=${VERSION}" \
        -o pucora ./cmd/pucora-ce; \
    else \
      CGO_ENABLED=0 go build \
        -ldflags="-s -w \
          -X github.com/pucora/pucora-ce/v2/pkg.Version=${VERSION} \
          -X github.com/pucora/lura/v2/core.PucoraVersion=${VERSION}" \
        -o pucora ./cmd/pucora-ce; \
    fi

FROM alpine:${ALPINE_VERSION}

LABEL maintainer="community@pucora.io"

# CA bundle from builder — avoids apk in runtime (fixes TLS issues on some Docker hosts).
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/pucora /usr/bin/pucora

RUN mkdir -p /etc/pucora \
    && echo '{"version":3}' > /etc/pucora/pucora.json

USER 1000:1000

WORKDIR /etc/pucora

ENTRYPOINT ["/usr/bin/pucora"]
CMD ["run", "-c", "/etc/pucora/pucora.json"]

EXPOSE 8080 8090
