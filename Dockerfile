# Build stage for Go backend
FROM golang:1.22-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o agent main.go

# Runtime stage
FROM python:3.11-slim-bookworm

# Install system dependencies for Playwright and Xvfb
RUN apt-get update && apt-get install -y \
    libglib2.0-0 \
    libnss3 \
    libnspr4 \
    libatk1.0-0 \
    libatk-bridge2.0-0 \
    libcups2 \
    libdrm2 \
    libdbus-1-3 \
    libxcb1 \
    libxkbcommon0 \
    libx11-6 \
    libxcomposite1 \
    libxdamage1 \
    libxext6 \
    libxfixes3 \
    libxrandr2 \
    libgbm1 \
    libasound2 \
    libpangocairo-1.0-0 \
    libpango-1.0-0 \
    libcairo2 \
    curl \
    xvfb \
    xauth \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Set Playwright browser path to a shared location
ENV PLAYWRIGHT_BROWSERS_PATH=/ms-playwright
RUN mkdir -p $PLAYWRIGHT_BROWSERS_PATH && chown -R 1000:1000 $PLAYWRIGHT_BROWSERS_PATH

# Install Python dependencies
COPY browser-agent/requirements.txt ./browser-agent/
RUN pip install --no-cache-dir -r browser-agent/requirements.txt

# Install Playwright browsers to the shared location
RUN playwright install chromium
RUN playwright install-deps chromium

# Copy application files
COPY --from=builder /app/agent .
COPY browser-agent/ ./browser-agent/
COPY start.sh .
RUN chmod +x start.sh

# Setup data directory and home directory for non-root user
RUN mkdir -p /data && chown -R 1000:1000 /data
RUN mkdir -p /home/node && chown -R 1000:1000 /home/node

# Run as non-root user
USER 1000:1000

# Set HOME to the directory we just created
ENV HOME=/home/node

EXPOSE 8080 8081

CMD ["./start.sh"]
