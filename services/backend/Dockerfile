# Stage 1: Build the binary
FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY services/backend ./services/backend
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./services/backend

# Stage 2: Final lightweight image
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /out/server .
EXPOSE 8080
CMD ["./server"]
