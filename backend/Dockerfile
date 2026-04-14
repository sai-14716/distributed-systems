# Stage 1: Build the binary
FROM golang:1.22-alpine AS builder
WORKDIR /app
# Initialize a module (required for modern Go builds)
RUN go mod init backend
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go

# Stage 2: Final lightweight image
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]