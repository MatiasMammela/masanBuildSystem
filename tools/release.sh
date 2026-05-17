#!/bin/bash
set -e

OUTPUT_DIR="../bin"
BINARY_NAME="mbs"

mkdir -p "$OUTPUT_DIR"

echo "Building release..."

echo "Building Linux x86_64..."
GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o "$OUTPUT_DIR/${BINARY_NAME}" \
    ../

echo "Building Windows x86_64..."
GOOS=windows GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o "$OUTPUT_DIR/${BINARY_NAME}.exe" \
    ../

echo ""
echo "Build complete. Binaries in $OUTPUT_DIR/:"
ls -lh "$OUTPUT_DIR/"
