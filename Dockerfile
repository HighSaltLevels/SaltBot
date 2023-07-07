FROM golang:1.20.5-buster AS builder

WORKDIR /build

COPY go.mod go.sum saltbot.go /build/
COPY cache /build/cache
COPY poll /build/poll
COPY expirychecker /build/expirychecker
COPY giphy /build/giphy
COPY handler /build/handler
COPY jeopardy /build/jeopardy
COPY poll /build/poll
COPY reminder /build/reminder
COPY util /build/util
COPY youtube /build/youtube

RUN ls -lah /build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -a -o saltbot saltbot.go

FROM scratch

COPY --from=builder /usr/share/zoneinfo/US/Eastern /usr/share/zoneinfo/US/Eastern
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /build/saltbot /saltbot

USER 69:420

ENTRYPOINT ["/saltbot"]
