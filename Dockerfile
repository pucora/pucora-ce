ARG GOLANG_VERSION
ARG ALPINE_VERSION
FROM golang:${GOLANG_VERSION}-alpine${ALPINE_VERSION} as builder

RUN apk --no-cache --virtual .build-deps add make gcc musl-dev binutils-gold

COPY . /app
WORKDIR /app

RUN make build


FROM alpine:${ALPINE_VERSION}

LABEL maintainer="community@velonetics.io"

RUN apk upgrade --no-cache --no-interactive && apk add --no-cache ca-certificates tzdata && \
    adduser -u 1000 -S -D -H velonetics && \
    mkdir /etc/velonetics && \
    echo '{ "version": 3 }' > /etc/velonetics/velonetics.json

COPY --from=builder /app/velonetics /usr/bin/velonetics

USER 1000

WORKDIR /etc/velonetics

ENTRYPOINT [ "/usr/bin/velonetics" ]
CMD [ "run", "-c", "/etc/velonetics/velonetics.json" ]

EXPOSE 8000 8090
