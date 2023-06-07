FROM golang:1.20.5-buster AS builder

WORKDIR /build

COPY go.mod go.sum main.go /build/
COPY cache /build/cache
COPY poll /build/poll

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -a -o messenger main.go

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /build/messenger /messenger

USER 69:420

ENTRYPOINT ["/messenger"]
