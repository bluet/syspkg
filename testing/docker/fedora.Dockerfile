# Fedora test container for go-syspkg (DNF testing)
FROM fedora:39

# Install build dependencies and DNF
RUN dnf update -y && dnf install -y \
    dnf-utils \
    curl \
    git \
    make \
    which \
    golang \
    && dnf clean all

# Set working directory
WORKDIR /workspace

# Copy go mod files for dependency caching
COPY go.mod go.sum ./
RUN go mod download

# Set test environment variables
ENV IN_CONTAINER=true
ENV CGO_ENABLED=0
ENV TEST_OS=fedora
ENV TEST_OS_VERSION=39
ENV TEST_PACKAGE_MANAGER=dnf

# Default command runs DNF-specific tests
CMD ["go", "test", "-v", "-tags=unit,integration", "./manager/dnf", "./osinfo"]
