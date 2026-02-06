#!/bin/bash
set -e

echo "====================================="
echo "Building dbank CLI"
echo "====================================="

# Tidy dependencies
echo "Downloading dependencies..."
go mod tidy

# Build for current platform
echo "Building for current platform..."
go build -o dbank .

echo ""
echo "Build complete! Binary: ./dbank"
echo ""
echo "Usage:"
echo "  ./dbank login -u <username> -p <password>"
echo "  ./dbank whoami"
echo "  ./dbank accounts list"
echo "  ./dbank transfers list"
echo "  ./dbank --help"
echo ""
echo "To install globally:"
echo "  sudo mv dbank /usr/local/bin/"
