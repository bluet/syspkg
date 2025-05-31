# Alpine test container for go-syspkg
FROM alpine:3.18

# Install build dependencies and apk package manager
RUN apk add --no-cache \
    curl \
    tar \
    git \
    make \
    alpine-sdk \
    bash

# Install Go 1.23.4 directly (Alpine package manager has older version)
RUN curl -L https://go.dev/dl/go1.23.4.linux-amd64.tar.gz | tar -C /usr/local -xz
ENV PATH="/usr/local/go/bin:${PATH}"

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
