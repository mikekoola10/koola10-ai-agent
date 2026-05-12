# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod ./
# RUN go mod download

COPY . .
RUN go build -o agent main.go

# Run stage
FROM alpine:latest

# Install necessary runtime dependencies
RUN apk add --no-cache ca-certificates curl bash

# Install flyctl
RUN curl -L https://fly.io/install.sh | sh
ENV PATH="/root/.fly/bin:${PATH}"

WORKDIR /app
COPY --from=builder /app/agent .

# Create data directory for MetaClaw persistence
RUN mkdir -p /data/applications && chown -R 1000:1000 /data

USER 1000:1000

EXPOSE 8080

CMD ["./agent"]
