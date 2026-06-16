FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY main.go .
RUN go build -o sentry main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/sentry .
CMD ["./sentry"]
