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
RUN apk add --no-cache ca-certificates wget unzip && \
    wget https://github.com/projectdiscovery/nuclei/releases/download/v3.0.0/nuclei_3.0.0_linux_amd64.zip && \
    unzip nuclei_3.0.0_linux_amd64.zip && \
    mv nuclei /usr/local/bin/ && \
    rm nuclei_3.0.0_linux_amd64.zip

WORKDIR /app
COPY --from=builder /app/agent .

# Create data directory for MetaClaw persistence
RUN mkdir -p /data/applications && chown -R 1000:1000 /data

USER 1000:1000

EXPOSE 8080

CMD ["./agent"]
