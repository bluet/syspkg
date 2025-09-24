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

# Install Go using Docker's TARGETARCH ARG for platform detection
ARG GO_VERSION=1.23.4
ARG TARGETARCH
SHELL ["/bin/bash", "-euxo", "pipefail", "-c"]
RUN if [ -z "${TARGETARCH}" ]; then \
        echo "Error: TARGETARCH is not set. Use a BuildKit-enabled builder." >&2; \
        exit 1; \
    fi && \
    curl -L "https://go.dev/dl/go${GO_VERSION}.linux-${TARGETARCH}.tar.gz" | tar -C /usr/local -xz
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
