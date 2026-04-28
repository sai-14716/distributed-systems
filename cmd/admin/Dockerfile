FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY internal ./internal
COPY cmd/admin ./cmd/admin
RUN go build -o /bin/raft-admin ./cmd/admin

FROM alpine:3.19
RUN adduser -D -H app
WORKDIR /app
USER app
COPY --from=builder /bin/raft-admin /bin/raft-admin
COPY cluster_config.yaml /app/cluster_config.yaml
ENTRYPOINT ["/bin/raft-admin"]
