# Unified Polyglot Build
FROM golang:1.22-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o agent main.go

FROM python:3.11-slim
RUN apt-get update && apt-get install -y ca-certificates curl jq git && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=go-builder /app/agent .
COPY web/ ./web/
COPY *.html ./
COPY data/ ./data/
RUN mkdir -p /data/applications && chmod -R 777 /data
EXPOSE 8080
CMD ["./agent"]
