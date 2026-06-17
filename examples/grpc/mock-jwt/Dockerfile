FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
COPY vendor ./vendor
COPY main.go ./
RUN CGO_ENABLED=0 GOPROXY=off go build -mod=vendor -o /mock-jwt .

FROM alpine:3.21
COPY --from=builder /mock-jwt /usr/local/bin/mock-jwt
EXPOSE 8081
ENTRYPOINT ["/usr/local/bin/mock-jwt"]
