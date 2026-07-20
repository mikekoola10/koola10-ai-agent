# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o echelon main.go

# Run stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/echelon .

RUN mkdir -p /data && chown -R 1000:1000 /data

USER 1000:1000

EXPOSE 8080

CMD ["./echelon"]
