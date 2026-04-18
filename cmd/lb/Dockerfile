FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY internal ./internal
COPY pkg ./pkg
COPY cmd/lb ./cmd/lb
RUN go build -o /bin/lb ./cmd/lb

FROM alpine:3.19
RUN adduser -D -H app
USER app
COPY --from=builder /bin/lb /bin/lb
EXPOSE 8080
ENTRYPOINT ["/bin/lb"]
