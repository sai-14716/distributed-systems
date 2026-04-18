FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY internal ./internal
COPY pkg ./pkg
COPY cmd/node ./cmd/node
COPY cmd/lb ./cmd/lb
RUN go build -o /bin/raft-node ./cmd/node
RUN go build -o /bin/lb ./cmd/lb

FROM alpine:3.19
RUN adduser -D -H app
RUN mkdir -p /var/lib/raft && chown -R app:app /var/lib/raft
COPY cmd/node/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
USER app
COPY --from=builder /bin/raft-node /bin/raft-node
COPY --from=builder /bin/lb /bin/lb
EXPOSE 8000
EXPOSE 19090
ENTRYPOINT ["/entrypoint.sh"]
