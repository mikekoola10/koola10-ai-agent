FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o aether main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/aether .
EXPOSE 8081
CMD ["./aether"]
