# Kyaraben test container
# Build: podman build -t kyaraben-test -f Containerfile .
# Run:   podman run -it --rm kyaraben-test

FROM docker.io/library/golang:1.24-bookworm AS builder

WORKDIR /app

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN go build -o /kyaraben ./cmd/kyaraben

# Run tests
RUN go test ./... -v

# Runtime image
FROM docker.io/library/debian:bookworm-slim

# Install minimal dependencies for testing
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Create a non-root user for testing
RUN useradd -m -s /bin/bash testuser

# Copy the binary
COPY --from=builder /kyaraben /usr/local/bin/kyaraben

# Switch to non-root user
USER testuser
WORKDIR /home/testuser

# Create default config
RUN mkdir -p ~/.config/kyaraben && \
    kyaraben init -u ~/Emulation -s e2e-test

# Default command shows status
CMD ["kyaraben", "status"]
