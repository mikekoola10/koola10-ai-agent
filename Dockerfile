# Build stage using RHEL UBI Go Toolset
FROM registry.access.redhat.com/ubi8/go-toolset:1.22 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o agent main.go

# Run stage using RHEL UBI Minimal
FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

# Install necessary runtime dependencies if any
# RUN microdnf install -y ... && microdnf clean all

WORKDIR /app
COPY --from=builder /app/agent .

# Create data directory for persistence
RUN mkdir -p /data/applications && chmod -R 777 /data

EXPOSE 8080

CMD ["./agent"]
