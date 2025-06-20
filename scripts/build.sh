#!/bin/bash
set -e

# This script builds the provider binary for multiple platforms

# Get the version from the environment or use the default
VERSION=${VERSION:-"0.1.0"}

# Set platforms to build for
PLATFORMS=("darwin/amd64" "darwin/arm64" "linux/amd64" "linux/arm64" "windows/amd64")

# Ensure output directory exists
mkdir -p bin

# Build for each platform
for platform in "${PLATFORMS[@]}"; do
    # Split the platform string into OS and architecture
    platform_split=(${platform//\// })
    OS=${platform_split[0]}
    ARCH=${platform_split[1]}
    
    # Set the output binary name based on the OS
    if [ $OS = "windows" ]; then
        output_name="bin/terraform-provider-openai_${VERSION}_${OS}_${ARCH}.exe"
    else
        output_name="bin/terraform-provider-openai_${VERSION}_${OS}_${ARCH}"
    fi
    
    echo "Building for $OS/$ARCH..."
    
    # Build the binary
    GOOS=$OS GOARCH=$ARCH go build -o "${output_name}" -ldflags="-X 'main.version=${VERSION}'"
    
    # Make the binary executable
    if [ $OS != "windows" ]; then
        chmod +x "${output_name}"
    fi
done

echo "Build complete!" 