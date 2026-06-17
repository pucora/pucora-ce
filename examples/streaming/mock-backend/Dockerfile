ARG GOLANG_VERSION=1.25.11
ARG ALPINE_VERSION=3.23

FROM golang:${GOLANG_VERSION}-bookworm AS builder
WORKDIR /src
COPY . .
RUN set -eux; \
    if [ -d vendor ]; then \
      CGO_ENABLED=0 GOPROXY=off go build -mod=vendor -trimpath -ldflags="-s -w" -o /mock-backend .; \
    else \
      CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /mock-backend .; \
    fi

FROM alpine:${ALPINE_VERSION}
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /mock-backend /usr/local/bin/mock-backend
EXPOSE 8081
HEALTHCHECK --interval=5s --timeout=3s --retries=5 CMD ["/usr/local/bin/mock-backend", "healthcheck"]
ENTRYPOINT ["/usr/local/bin/mock-backend"]
