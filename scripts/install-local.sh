#!/bin/bash
set -e

# This script installs the provider locally for testing purposes

# Get the version from the environment or use the default
VERSION=${VERSION:-"0.1.0"}

# Determine OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
if [[ "$OS" == "darwin" ]]; then
    if [[ $(uname -m) == "arm64" ]]; then
        ARCH="arm64"
    else
        ARCH="amd64"
    fi
elif [[ "$OS" == "linux" ]]; then
    ARCH=$(uname -m)
    if [[ "$ARCH" == "x86_64" ]]; then
        ARCH="amd64"
    elif [[ "$ARCH" == "aarch64" ]]; then
        ARCH="arm64"
    fi
else
    echo "Unsupported OS: $OS"
    exit 1
fi

echo "Installing provider for $OS/$ARCH..."

# Create plugin directory structure
TF_PLUGIN_DIR="$HOME/.terraform.d/plugins/registry.terraform.io/fjcorp/openai/${VERSION}/${OS}_${ARCH}"
mkdir -p "$TF_PLUGIN_DIR"

# Build the provider
echo "Building provider..."
go build -o terraform-provider-openai

# Copy the provider to the plugin directory
echo "Copying provider to $TF_PLUGIN_DIR..."
cp terraform-provider-openai "$TF_PLUGIN_DIR/"

echo "Installation complete. Provider installed at: $TF_PLUGIN_DIR/terraform-provider-openai"
echo "You can now use the provider in your Terraform configurations with:"
echo ""
echo 'terraform {'
echo '  required_providers {'
echo '    openai = {'
echo '      source  = "fjcorp/openai"'
echo '      version = "'"$VERSION"'"'
echo '    }'
echo '  }'
echo '}' 