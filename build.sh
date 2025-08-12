#!/bin/bash
mkdir -p builds

echo "Building for different platforms..."

# Windows 64-bit
GOOS=windows GOARCH=amd64 go build -o builds/accessibility-scanner-windows-amd64.exe main.go
echo "✅ Windows 64-bit built"

# Windows 32-bit  
GOOS=windows GOARCH=386 go build -o builds/accessibility-scanner-windows-386.exe main.go
echo "✅ Windows 32-bit built"

# macOS 64-bit (Intel)
GOOS=darwin GOARCH=amd64 go build -o builds/accessibility-scanner-macos-amd64 main.go
echo "✅ macOS Intel built"

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o builds/accessibility-scanner-macos-arm64 main.go
echo "✅ macOS Apple Silicon built"

# Linux 64-bit
GOOS=linux GOARCH=amd64 go build -o builds/accessibility-scanner-linux-amd64 main.go
echo "✅ Linux 64-bit built"

# Linux 32-bit
GOOS=linux GOARCH=386 go build -o builds/accessibility-scanner-linux-386 main.go
echo "✅ Linux 32-bit built"

# Linux ARM64 (for servers/Raspberry Pi)
GOOS=linux GOARCH=arm64 go build -o builds/accessibility-scanner-linux-arm64 main.go
echo "✅ Linux ARM64 built"

echo "🎉 All builds completed in ./builds/ directory"
ls -la builds/