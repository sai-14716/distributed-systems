FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY internal ./internal
COPY cmd/admin ./cmd/admin
RUN go build -o /bin/raft-admin ./cmd/admin

FROM alpine:3.19
RUN adduser -D -H app
USER app
COPY --from=builder /bin/raft-admin /bin/raft-admin
ENTRYPOINT ["/bin/raft-admin"]
