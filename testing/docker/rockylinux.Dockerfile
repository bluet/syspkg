# Rocky Linux test container for go-syspkg (YUM testing)
FROM rockylinux:8

# Install build dependencies and YUM
RUN yum update -y && yum install -y \
    yum-utils \
    curl \
    git \
    make \
    which \
    && yum clean all

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
ENV GOROOT="/usr/local/go"

# Set working directory
WORKDIR /workspace

# Copy go mod files for dependency caching
COPY go.mod go.sum ./
RUN go mod download

# Set test environment variables
ENV IN_CONTAINER=true
ENV CGO_ENABLED=0
ENV TEST_OS=rockylinux
ENV TEST_OS_VERSION=8
ENV TEST_PACKAGE_MANAGER=yum

# Default command runs YUM-specific tests
CMD ["go", "test", "-v", "-tags=unit,integration", "./manager/yum", "./osinfo"]
