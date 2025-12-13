#!/bin/bash

# Generate changelog from git commits
# Usage: ./generate-changelog.sh [version] [previous-version]

set -e

VERSION=${1:-$(git describe --tags --abbrev=0 2>/dev/null || echo "v1.0.0")}
PREVIOUS_VERSION=${2:-$(git describe --tags --abbrev=0 HEAD~1 2>/dev/null || echo "")}

# Remove 'v' prefix if present
VERSION=${VERSION#v}
PREVIOUS_VERSION=${PREVIOUS_VERSION#v}

CHANGELOG_FILE="CHANGELOG.md"
TEMP_FILE=$(mktemp)

# Extract date
RELEASE_DATE=$(date +%Y-%m-%d)

# Start building changelog entry
cat > "$TEMP_FILE" <<EOF
## [${VERSION}] - ${RELEASE_DATE}

EOF

# Get commits since previous version
if [ -n "$PREVIOUS_VERSION" ]; then
    COMMITS=$(git log v${PREVIOUS_VERSION}..HEAD --pretty=format:"%s" --no-merges)
    COMPARE_URL="https://github.com/mohamedhabibwork/abp-gen/compare/v${PREVIOUS_VERSION}...v${VERSION}"
else
    COMMITS=$(git log --pretty=format:"%s" --no-merges -20)
    COMPARE_URL="https://github.com/mohamedhabibwork/abp-gen/releases/tag/v${VERSION}"
fi

# Categorize commits
FEATURES=$(echo "$COMMITS" | grep -iE "^feat" || true)
FIXES=$(echo "$COMMITS" | grep -iE "^fix" || true)
DOCS=$(echo "$COMMITS" | grep -iE "^docs" || true)
REFACTOR=$(echo "$COMMITS" | grep -iE "^refactor" || true)
PERF=$(echo "$COMMITS" | grep -iE "^perf" || true)
TEST=$(echo "$COMMITS" | grep -iE "^test" || true)
CHORE=$(echo "$COMMITS" | grep -iE "^chore|^ci|^build" || true)

# Add sections
if [ -n "$FEATURES" ]; then
    echo "### Added" >> "$TEMP_FILE"
    echo "$FEATURES" | sed 's/^feat:/* /' | sed 's/^feat(\(.*\)):/* \1: /' >> "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"
fi

if [ -n "$FIXES" ]; then
    echo "### Fixed" >> "$TEMP_FILE"
    echo "$FIXES" | sed 's/^fix:/* /' | sed 's/^fix(\(.*\)):/* \1: /' >> "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"
fi

if [ -n "$REFACTOR" ]; then
    echo "### Changed" >> "$TEMP_FILE"
    echo "$REFACTOR" | sed 's/^refactor:/* /' | sed 's/^refactor(\(.*\)):/* \1: /' >> "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"
fi

if [ -n "$PERF" ]; then
    echo "### Performance" >> "$TEMP_FILE"
    echo "$PERF" | sed 's/^perf:/* /' | sed 's/^perf(\(.*\)):/* \1: /' >> "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"
fi

if [ -n "$DOCS" ]; then
    echo "### Documentation" >> "$TEMP_FILE"
    echo "$DOCS" | sed 's/^docs:/* /' | sed 's/^docs(\(.*\)):/* \1: /' >> "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"
fi

if [ -n "$TEST" ]; then
    echo "### Tests" >> "$TEMP_FILE"
    echo "$TEST" | sed 's/^test:/* /' | sed 's/^test(\(.*\)):/* \1: /' >> "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"
fi

if [ -n "$CHORE" ]; then
    echo "### Chores" >> "$TEMP_FILE"
    echo "$CHORE" | sed 's/^chore:/* /' | sed 's/^ci:/* /' | sed 's/^build:/* /' >> "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"
fi

# Add compare link
echo "" >> "$TEMP_FILE"
echo "[${VERSION}]: $COMPARE_URL" >> "$TEMP_FILE"
echo "" >> "$TEMP_FILE"

# Prepend to CHANGELOG.md
if [ -f "$CHANGELOG_FILE" ]; then
    # Check if version already exists
    if grep -q "## \[${VERSION}\]" "$CHANGELOG_FILE"; then
        echo "Version ${VERSION} already exists in CHANGELOG.md"
        exit 0
    fi
    
    # Insert after "## [Unreleased]" section
    if grep -q "## \[Unreleased\]" "$CHANGELOG_FILE"; then
        # Find the line number after "## [Unreleased]" and the blank line
        INSERT_LINE=$(awk '/^## \[Unreleased\]/ {found=1; next} found && /^$/ {print NR; exit}' "$CHANGELOG_FILE")
        
        if [ -n "$INSERT_LINE" ]; then
            # Insert the new entry after the blank line following [Unreleased]
            # Use head and tail to split the file, then insert the new content
            {
                head -n "$INSERT_LINE" "$CHANGELOG_FILE"
                cat "$TEMP_FILE"
                tail -n +$((INSERT_LINE + 1)) "$CHANGELOG_FILE"
            } > "${CHANGELOG_FILE}.new"
            mv "${CHANGELOG_FILE}.new" "$CHANGELOG_FILE"
        else
            # Fallback: use awk to read from temp file directly
            awk -v temp_file="$TEMP_FILE" '
                /^## \[Unreleased\]/ {
                    print
                    getline
                    print
                    # Read and print the temp file
                    while ((getline line < temp_file) > 0) {
                        print line
                    }
                    close(temp_file)
                    next
                }
                { print }
            ' "$CHANGELOG_FILE" > "${CHANGELOG_FILE}.new"
            mv "${CHANGELOG_FILE}.new" "$CHANGELOG_FILE"
        fi
    else
        # Prepend to file
        cat "$TEMP_FILE" "$CHANGELOG_FILE" > "${CHANGELOG_FILE}.new"
        mv "${CHANGELOG_FILE}.new" "$CHANGELOG_FILE"
    fi
else
    # Create new changelog
    cat > "$CHANGELOG_FILE" <<EOF
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

EOF
    cat "$TEMP_FILE" >> "$CHANGELOG_FILE"
fi

rm "$TEMP_FILE"
echo "Changelog updated for version ${VERSION}"

