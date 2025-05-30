# Alpine test container for go-syspkg
FROM alpine:3.18

# Install build dependencies and apk package manager
RUN apk add --no-cache \
    go \
    git \
    make \
    alpine-sdk \
    bash

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