#!/bin/bash

# Update version numbers in documentation files
# Usage: ./update-docs-version.sh [version]

set -e

VERSION=${1:-$(git describe --tags --abbrev=0 2>/dev/null || echo "v1.0.0")}
VERSION=${VERSION#v}  # Remove 'v' prefix if present

echo "Updating documentation with version ${VERSION}..."

# Update README.md - replace version in download URLs
if [ -f "README.md" ]; then
    sed -i.bak "s/abp-gen_\([0-9.]*\)_/abp-gen_${VERSION}_/g" README.md
    rm -f README.md.bak
    echo "✓ Updated README.md"
fi

# Update examples/schema.json if it has version field
if [ -f "examples/schema.json" ]; then
    # Only update if version field exists
    if grep -q '"version"' examples/schema.json; then
        sed -i.bak "s/\"version\": \"[^\"]*\"/\"version\": \"${VERSION}\"/g" examples/schema.json
        rm -f examples/schema.json.bak
        echo "✓ Updated examples/schema.json"
    fi
fi

# Update any other documentation files that reference version
# Add more files as needed

echo "Documentation updated to version ${VERSION}"

