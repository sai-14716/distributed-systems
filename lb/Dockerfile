# Stage 1: Build the binary
FROM golang:1.22-alpine AS builder
WORKDIR /app
RUN go mod init lb
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o lb .

# Stage 2: Final lightweight image
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/lb .
EXPOSE 8080
CMD ["./lb"]