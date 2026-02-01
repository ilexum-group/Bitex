#!/bin/bash

# Bitex CI/CD Local Validation Script
# Run this script before committing to ensure your code passes all CI checks

set -e

echo "🚀 Running Bitex CI/CD Local Validation..."
echo

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo "❌ Error: Must be run from the project root directory"
    exit 1
fi

echo "📦 Checking Go modules..."
go mod tidy
go mod verify
echo "✅ Go modules OK"
echo

echo "🧪 Running tests..."
go test -v ./tests/...
echo "✅ Tests passed"
echo

echo "🔨 Building application..."
go build -v ./cmd/bitex
echo "✅ Build successful"
echo

echo "📝 Checking code formatting..."
if [ "$(gofmt -s -l $(find . -name "*.go" -not -path "./vendor/*") | wc -l)" -gt 0 ]; then
    echo "❌ The following files are not formatted properly:"
    gofmt -s -l $(find . -name "*.go" -not -path "./vendor/*")
    echo "Run 'gofmt -w \$(find . -name \"*.go\" -not -path \"./vendor/*\")' to fix formatting issues"
    exit 1
fi
echo "✅ Code formatting OK"
echo

echo "🔍 Running go vet..."
go vet ./...
echo "✅ Go vet passed"
echo

echo "🎉 All checks passed! Your code is ready for commit."
echo
echo "Next steps:"
echo "1. git add ."
echo "2. git commit -m 'Your commit message'"
echo "3. git push origin your-branch"
echo "4. Create a Pull Request"
