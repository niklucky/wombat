#!/bin/bash
set -e

BINARY="wombat"
VERSION="0.1.0"
LDFLAGS="-s -w"

mkdir -p dist

echo "Building standard binaries..."

GOOS=darwin GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o "dist/${BINARY}-darwin-amd64" ./cmd/wombat
GOOS=darwin GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o "dist/${BINARY}-darwin-arm64" ./cmd/wombat

GOOS=linux GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o "dist/${BINARY}-linux-amd64" ./cmd/wombat
GOOS=linux GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o "dist/${BINARY}-linux-arm64" ./cmd/wombat

GOOS=windows GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o "dist/${BINARY}-windows-amd64.exe" ./cmd/wombat

echo "Building GUI binaries..."

GOOS=darwin GOARCH=amd64 go build -tags gui -ldflags "${LDFLAGS}" -o "dist/${BINARY}-gui-darwin-amd64" ./cmd/wombat
GOOS=darwin GOARCH=arm64 go build -tags gui -ldflags "${LDFLAGS}" -o "dist/${BINARY}-gui-darwin-arm64" ./cmd/wombat

GOOS=linux GOARCH=amd64 go build -tags gui -ldflags "${LDFLAGS}" -o "dist/${BINARY}-gui-linux-amd64" ./cmd/wombat
GOOS=windows GOARCH=amd64 go build -tags gui -ldflags "${LDFLAGS}" -o "dist/${BINARY}-gui-windows-amd64.exe" ./cmd/wombat

echo "Done."
