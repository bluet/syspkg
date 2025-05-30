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

# Install Go 1.23.4
RUN curl -L https://go.dev/dl/go1.23.4.linux-amd64.tar.gz | tar -C /usr/local -xz
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
