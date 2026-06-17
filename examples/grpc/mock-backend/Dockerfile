FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
COPY vendor ./vendor
COPY main.go ./
COPY flight_finder ./flight_finder
RUN CGO_ENABLED=0 GOPROXY=off go build -mod=vendor -o /mock-grpc-backend .

FROM alpine:3.23
COPY --from=builder /mock-grpc-backend /usr/local/bin/mock-grpc-backend
COPY flights.pb /etc/grpc/flights.pb
EXPOSE 4242
ENTRYPOINT ["/usr/local/bin/mock-grpc-backend"]
