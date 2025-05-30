# Ubuntu test container for go-syspkg
FROM ubuntu:22.04

# Avoid interactive prompts
ENV DEBIAN_FRONTEND=noninteractive

# Install package managers and dependencies
RUN apt-get update && apt-get install -y \
    apt-utils \
    software-properties-common \
    flatpak \
    curl \
    git \
    make \
    && rm -rf /var/lib/apt/lists/*

# Install Go
RUN curl -L https://go.dev/dl/go1.21.0.linux-amd64.tar.gz | tar -C /usr/local -xz
ENV PATH="/usr/local/go/bin:${PATH}"

# Note: snap requires systemd which doesn't work in standard Docker containers
# Options for snap testing:
# 1. Use mock data for snap command outputs
# 2. Use GitHub Actions with native Ubuntu runners
# 3. Use systemd-enabled containers (complex setup)

# Set working directory
WORKDIR /workspace

# Install test dependencies
COPY go.mod go.sum ./
RUN go mod download

# Set test environment
ENV IN_CONTAINER=true
ENV CGO_ENABLED=0

# Default command runs tests
CMD ["go", "test", "-v", "./..."]
