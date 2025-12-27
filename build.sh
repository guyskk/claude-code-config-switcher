#!/bin/bash

set -e

# Create dist directory if it doesn't exist
mkdir -p dist

echo "Building ccc command line tool..."

# Build for current platform
echo "Building for current platform..."
go build -o dist/ccc main.go

# Make it executable
chmod +x dist/ccc

echo "Build completed successfully!"
echo "Binary location: dist/ccc"
echo ""
echo "To install system-wide, run:"
echo "  sudo cp dist/ccc /usr/local/bin/ccc"
