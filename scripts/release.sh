#!/bin/bash
set -e

# This script helps prepare a new release of the provider

# Check if a version argument was provided
if [ $# -ne 1 ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 v0.1.0"
    exit 1
fi

VERSION=$1

# Remove the 'v' prefix if present
if [[ $VERSION == v* ]]; then
    VERSION_NUM=${VERSION:1}
else
    VERSION_NUM=$VERSION
    VERSION="v$VERSION"
fi

echo "Preparing release for $VERSION"

# Update version in the code
echo "Updating version in code..."
# Update version in any files that need it

# Run tests
echo "Running tests..."
make test

# Create a git tag
echo "Creating git tag..."
git tag -a "$VERSION" -m "Release $VERSION"
echo "Tag created. To push the tag, run: git push origin $VERSION"

# Build binaries
echo "Building binaries..."
VERSION="$VERSION_NUM" scripts/build.sh

echo "Release preparation complete for $VERSION"
echo "Next steps:"
echo "1. Push the tag: git push origin $VERSION"
echo "2. Create a GitHub release with the built binaries"
echo "3. Update the provider documentation if needed" 