#!/bin/bash
set -e

echo "=== go-sqlcmd Development Container Setup ==="

# Dynamic workspace detection with fallback
if [ -n "${WORKSPACE_FOLDER}" ] && [ -d "${WORKSPACE_FOLDER}" ]; then
    cd "${WORKSPACE_FOLDER}"
elif [ -d "/workspaces/go-sqlcmd" ]; then
    cd /workspaces/go-sqlcmd
else
    # Find workspace by looking for go.mod
    workspace_go_mod="$(find /workspaces -maxdepth 2 -name 'go.mod' -type f -print -quit 2>/dev/null)"
    if [ -n "$workspace_go_mod" ]; then
        cd "$(dirname "$workspace_go_mod")"
    else
        echo "Error: Could not determine workspace directory" >&2
        exit 1
    fi
fi
echo "📁 Working in: $(pwd)"

# Download Go dependencies
echo "📦 Downloading Go dependencies..."
go mod download

# Build sqlcmd and add to PATH
echo "🔨 Building sqlcmd..."
go build -o ~/bin/sqlcmd ./cmd/modern
echo "✅ sqlcmd built and added to PATH at ~/bin/sqlcmd"

# Verify build works
echo "🔨 Verifying full build..."
go build ./...

# Wait for SQL Server to be ready (health check should have done this, but let's verify)
echo "🔄 Verifying SQL Server connection..."
max_attempts=30
attempt=1
sql_ready=false
while [ $attempt -le $max_attempts ]; do
    if ~/bin/sqlcmd -S localhost -U sa -P "${SQLCMDPASSWORD}" -C -Q "SELECT 1" > /dev/null 2>&1; then
        echo "✅ SQL Server is ready!"
        sql_ready=true
        break
    fi
    echo "  Waiting for SQL Server... (attempt $attempt/$max_attempts)"
    sleep 2
    attempt=$((attempt + 1))
done

if [ "$sql_ready" = false ]; then
    echo "⚠️ Warning: Could not verify SQL Server connection. Tests may fail."
fi

# Run initial setup SQL if it exists and SQL Server is ready
if [ -f ".devcontainer/mssql/setup.sql" ]; then
    if [ "$sql_ready" = true ]; then
        echo "📋 Running setup.sql..."
        ~/bin/sqlcmd -S localhost -U sa -P "${SQLCMDPASSWORD}" -C -i .devcontainer/mssql/setup.sql
    else
        echo "⚠️ Skipping setup.sql because SQL Server connection could not be verified."
    fi
fi

# Create useful aliases in a dedicated directory (safe and idempotent)
echo "🔧 Setting up helpful aliases..."
mkdir -p ~/.bash_aliases.d
cat > ~/.bash_aliases.d/go-sqlcmd << 'EOF'
# go-sqlcmd development aliases
alias gtest='go test ./...'
alias gtest-short='go test -short ./...'
alias gtest-v='go test -v ./...'
alias gbuild='go build ./cmd/modern && echo "Built: ./modern"'
alias ginstall='go build -o ~/bin/sqlcmd ./cmd/modern && echo "Installed to ~/bin/sqlcmd"'
alias gfmt='go fmt ./...'
alias gvet='go vet ./...'
alias glint='golangci-lint run'
alias ggen='go generate ./...'

# sqlcmd shortcut - uses the locally built version
alias sql='~/bin/sqlcmd -S localhost -U sa -P "$SQLCMDPASSWORD" -C'

# Legacy ODBC sqlcmd for compatibility testing
alias sql-legacy='/opt/mssql-tools18/bin/sqlcmd -S localhost -U sa -P "$SQLCMDPASSWORD" -C'

# Quick test connection
alias test-db='~/bin/sqlcmd -S localhost -U sa -P "$SQLCMDPASSWORD" -C -Q "SELECT @@VERSION"'

# Rebuild and test
alias rebuild='go build -o ~/bin/sqlcmd ./cmd/modern && echo "Rebuilt sqlcmd"'
EOF

# Ensure aliases are sourced from .bashrc
if ! grep -q 'go-sqlcmd aliases' ~/.bashrc 2>/dev/null; then
    {
        echo ''
        echo '# go-sqlcmd aliases'
        echo 'if [ -f ~/.bash_aliases ]; then'
        echo '    # Source traditional aliases file if present'
        echo '    . ~/.bash_aliases'
        echo 'fi'
        echo ''
        echo 'if [ -d ~/.bash_aliases.d ]; then'
        echo '    # Source all alias snippets from ~/.bash_aliases.d'
        echo '    for f in ~/.bash_aliases.d/*; do'
        echo '        [ -r "$f" ] && . "$f"'
        echo '    done'
        echo 'fi'
    } >> ~/.bashrc
fi

echo ""
echo "=== Setup Complete! ==="
echo ""
echo "📖 Quick Reference:"
echo "  gtest       - Run all tests"
echo "  gtest-short - Run short tests"
echo "  gtest-v     - Run tests with verbose output"
echo "  gbuild      - Build sqlcmd locally"
echo "  ginstall    - Build and install sqlcmd to ~/bin"
echo "  gfmt        - Format code"
echo "  gvet        - Run go vet"
echo "  glint       - Run golangci-lint"
echo "  ggen        - Run go generate (for translations)"
echo "  test-db     - Test database connection"
echo "  sql         - Connect to SQL Server (go-sqlcmd)"
echo "  sql-legacy  - Connect using legacy ODBC sqlcmd"
echo "  rebuild     - Rebuild sqlcmd"
echo ""
echo "🔧 Your locally built sqlcmd is at ~/bin/sqlcmd and in PATH"
echo ""
echo "🔗 SQL Server Connection:"
echo "  Server:   localhost,1433"
echo "  User:     sa"
echo "  Password: (from SQLCMDPASSWORD environment variable)"
echo "  Database: master (or SqlCmdTest)"
echo ""
echo "🧪 Environment variables are pre-configured for tests."
echo "  Run 'go test ./...' to execute the full test suite."
echo ""
