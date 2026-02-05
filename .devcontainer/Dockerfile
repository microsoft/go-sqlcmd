# go-sqlcmd Development Container
FROM mcr.microsoft.com/devcontainers/go:1.24-bookworm

# Install additional OS packages
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get install -y curl libkrb5-dev gnupg2 \
    && apt-get clean -y && rm -rf /var/lib/apt/lists/*

# Install Microsoft ODBC driver and mssql-tools18 (legacy ODBC sqlcmd/bcp for compatibility testing)
RUN curl -fsSL https://packages.microsoft.com/keys/microsoft.asc | gpg --dearmor -o /usr/share/keyrings/microsoft-prod.gpg \
    && curl -fsSL https://packages.microsoft.com/config/debian/12/prod.list | tee /etc/apt/sources.list.d/mssql-release.list \
    && apt-get update \
    && ACCEPT_EULA=Y apt-get install -y msodbcsql18 mssql-tools18 unixodbc-dev \
    && apt-get clean -y && rm -rf /var/lib/apt/lists/*
ENV PATH="/opt/mssql-tools18/bin:${PATH}"

# Install golangci-lint for code quality
# Download pre-built binary with SHA256 checksum verification (supply chain security)
# Supports both amd64 and arm64 architectures
ARG GOLANGCI_LINT_VERSION=1.64.8
ARG GOLANGCI_LINT_SHA256_AMD64=b6270687afb143d019f387c791cd2a6f1cb383be9b3124d241ca11bd3ce2e54e
ARG GOLANGCI_LINT_SHA256_ARM64=a6ab58ebcb1c48572622146cdaec2956f56871038a54ed1149f1386e287789a5
RUN ARCH=$(dpkg --print-architecture) \
    && if [ "$ARCH" = "amd64" ]; then \
         CHECKSUM="${GOLANGCI_LINT_SHA256_AMD64}"; \
       elif [ "$ARCH" = "arm64" ]; then \
         CHECKSUM="${GOLANGCI_LINT_SHA256_ARM64}"; \
       else \
         echo "Unsupported architecture: $ARCH" && exit 1; \
       fi \
    && curl -fsSLO "https://github.com/golangci/golangci-lint/releases/download/v${GOLANGCI_LINT_VERSION}/golangci-lint-${GOLANGCI_LINT_VERSION}-linux-${ARCH}.tar.gz" \
    && echo "${CHECKSUM}  golangci-lint-${GOLANGCI_LINT_VERSION}-linux-${ARCH}.tar.gz" | sha256sum -c - \
    && tar -xzf "golangci-lint-${GOLANGCI_LINT_VERSION}-linux-${ARCH}.tar.gz" \
    && mv "golangci-lint-${GOLANGCI_LINT_VERSION}-linux-${ARCH}/golangci-lint" /usr/local/bin/ \
    && rm -rf "golangci-lint-${GOLANGCI_LINT_VERSION}-linux-${ARCH}" "golangci-lint-${GOLANGCI_LINT_VERSION}-linux-${ARCH}.tar.gz" \
    && golangci-lint --version

# Install additional Go tools (pinned versions for reproducibility)
RUN go install golang.org/x/tools/gopls@v0.18.1 \
    && go install github.com/go-delve/delve/cmd/dlv@v1.24.1 \
    && go install honnef.co/go/tools/cmd/staticcheck@v0.6.1 \
    && go install golang.org/x/text/cmd/gotext@v0.22.0

# Create bin directory for local sqlcmd builds
RUN mkdir -p /home/vscode/bin && chown vscode:vscode /home/vscode/bin
ENV PATH="/home/vscode/bin:${PATH}"
