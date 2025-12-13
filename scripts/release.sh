#!/bin/bash

# Release script for abp-gen
# This script automates the process of creating a new version release
# Usage: ./scripts/release.sh [version] [--skip-checks] [--skip-push]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Flags
SKIP_CHECKS=false
SKIP_PUSH=false
VERSION=""

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-checks)
            SKIP_CHECKS=true
            shift
            ;;
        --skip-push)
            SKIP_PUSH=true
            shift
            ;;
        -*)
            echo -e "${RED}Unknown option: $1${NC}"
            echo "Usage: $0 [version] [--skip-checks] [--skip-push]"
            exit 1
            ;;
        *)
            VERSION="$1"
            shift
            ;;
    esac
done

cd "$PROJECT_ROOT"

# Helper functions
info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

success() {
    echo -e "${GREEN}✓${NC} $1"
}

warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

error() {
    echo -e "${RED}✗${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    info "Checking prerequisites..."
    
    if ! command -v git &> /dev/null; then
        error "git is not installed"
        exit 1
    fi
    
    if ! command -v go &> /dev/null; then
        error "go is not installed"
        exit 1
    fi
    
    # Check if we're in a git repository
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        error "Not in a git repository"
        exit 1
    fi
    
    # Check if we're on main/master branch
    CURRENT_BRANCH=$(git branch --show-current)
    if [[ "$CURRENT_BRANCH" != "main" && "$CURRENT_BRANCH" != "master" ]]; then
        warning "Not on main/master branch (currently on: $CURRENT_BRANCH)"
        read -p "Continue anyway? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
    
    # Check for uncommitted changes
    if ! git diff-index --quiet HEAD --; then
        error "You have uncommitted changes. Please commit or stash them first."
        exit 1
    fi
    
    # Check if we're up to date with remote
    git fetch origin
    LOCAL=$(git rev-parse @)
    REMOTE=$(git rev-parse @{u} 2>/dev/null || echo "")
    if [[ -n "$REMOTE" && "$LOCAL" != "$REMOTE" ]]; then
        warning "Your branch is not up to date with remote"
        read -p "Continue anyway? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
    
    success "Prerequisites check passed"
}

# Run pre-release checks
run_checks() {
    if [[ "$SKIP_CHECKS" == "true" ]]; then
        warning "Skipping pre-release checks"
        return
    fi
    
    info "Running pre-release checks..."
    
    # Format code
    info "Formatting code..."
    if ! go fmt ./...; then
        error "Code formatting failed"
        exit 1
    fi
    success "Code formatted"
    
    # Run tests
    info "Running tests..."
    if ! go test ./...; then
        error "Tests failed"
        exit 1
    fi
    success "All tests passed"
    
    # Check for linter (if available)
    if command -v golangci-lint &> /dev/null; then
        info "Running linter..."
        if ! golangci-lint run; then
            error "Linter failed"
            exit 1
        fi
        success "Linter passed"
    else
        warning "golangci-lint not found, skipping lint check"
    fi
    
    # Build to ensure it compiles
    info "Building binary..."
    if ! go build ./cmd/abp-gen; then
        error "Build failed"
        exit 1
    fi
    success "Build successful"
    
    success "All pre-release checks passed"
}

# Get current version
get_current_version() {
    local latest_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
    if [[ -z "$latest_tag" ]]; then
        echo "v0.0.0"
    else
        echo "$latest_tag"
    fi
}

# Validate version format
validate_version() {
    local version=$1
    if [[ ! "$version" =~ ^v?[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)?$ ]]; then
        return 1
    fi
    return 0
}

# Get version from user
get_version() {
    if [[ -n "$VERSION" ]]; then
        # Remove 'v' prefix if present, we'll add it back
        VERSION=${VERSION#v}
        if ! validate_version "v$VERSION"; then
            error "Invalid version format: $VERSION"
            error "Version must follow semantic versioning (e.g., 1.0.0, 1.2.3)"
            exit 1
        fi
        VERSION="v$VERSION"
    else
        local current_version=$(get_current_version)
        info "Current version: $current_version"
        
        echo ""
        echo "What type of release is this?"
        echo "1) Patch (bug fixes) - e.g., 1.0.0 -> 1.0.1"
        echo "2) Minor (new features) - e.g., 1.0.0 -> 1.1.0"
        echo "3) Major (breaking changes) - e.g., 1.0.0 -> 2.0.0"
        echo "4) Custom version"
        read -p "Select option (1-4): " -n 1 -r
        echo
        
        case $REPLY in
            1)
                # Patch version
                if [[ "$current_version" == "v0.0.0" ]]; then
                    VERSION="v0.0.1"
                else
                    local major=$(echo "$current_version" | sed 's/v\([0-9]*\)\.[0-9]*\.[0-9]*/\1/')
                    local minor=$(echo "$current_version" | sed 's/v[0-9]*\.\([0-9]*\)\.[0-9]*/\1/')
                    local patch=$(echo "$current_version" | sed 's/v[0-9]*\.[0-9]*\.\([0-9]*\)/\1/')
                    patch=$((patch + 1))
                    VERSION="v${major}.${minor}.${patch}"
                fi
                ;;
            2)
                # Minor version
                if [[ "$current_version" == "v0.0.0" ]]; then
                    VERSION="v0.1.0"
                else
                    local major=$(echo "$current_version" | sed 's/v\([0-9]*\)\.[0-9]*\.[0-9]*/\1/')
                    local minor=$(echo "$current_version" | sed 's/v[0-9]*\.\([0-9]*\)\.[0-9]*/\1/')
                    minor=$((minor + 1))
                    VERSION="v${major}.${minor}.0"
                fi
                ;;
            3)
                # Major version
                if [[ "$current_version" == "v0.0.0" ]]; then
                    VERSION="v1.0.0"
                else
                    local major=$(echo "$current_version" | sed 's/v\([0-9]*\)\.[0-9]*\.[0-9]*/\1/')
                    major=$((major + 1))
                    VERSION="v${major}.0.0"
                fi
                ;;
            4)
                read -p "Enter version (e.g., 1.2.3): " VERSION
                VERSION=${VERSION#v}  # Remove 'v' if present
                if ! validate_version "v$VERSION"; then
                    error "Invalid version format"
                    exit 1
                fi
                VERSION="v$VERSION"
                ;;
            *)
                error "Invalid option"
                exit 1
                ;;
        esac
    fi
    
    # Check if version already exists
    if git rev-parse "$VERSION" >/dev/null 2>&1; then
        error "Tag $VERSION already exists"
        exit 1
    fi
    
    info "New version: $VERSION"
}

# Update changelog
update_changelog() {
    info "Updating changelog..."
    local previous_version=$(get_current_version)
    
    if [[ "$previous_version" == "v0.0.0" ]]; then
        previous_version=""
    fi
    
    if ! bash "$SCRIPT_DIR/generate-changelog.sh" "$VERSION" "$previous_version"; then
        error "Failed to update changelog"
        exit 1
    fi
    
    success "Changelog updated"
}

# Update documentation
update_documentation() {
    info "Updating documentation..."
    
    if ! bash "$SCRIPT_DIR/update-docs-version.sh" "$VERSION"; then
        error "Failed to update documentation"
        exit 1
    fi
    
    success "Documentation updated"
}

# Create git tag
create_tag() {
    info "Creating git tag: $VERSION"
    
    # Check if there are changes to commit
    if ! git diff-index --quiet HEAD --; then
        info "Staging changes..."
        git add CHANGELOG.md README.md examples/schema.json 2>/dev/null || true
        
        info "Committing changes..."
        git commit -m "chore: prepare release $VERSION"
        success "Changes committed"
    fi
    
    # Create annotated tag
    read -p "Enter release message (or press Enter for default): " TAG_MESSAGE
    if [[ -z "$TAG_MESSAGE" ]]; then
        TAG_MESSAGE="Release $VERSION"
    fi
    
    if ! git tag -a "$VERSION" -m "$TAG_MESSAGE"; then
        error "Failed to create tag"
        exit 1
    fi
    
    success "Tag $VERSION created"
}

# Push changes
push_changes() {
    if [[ "$SKIP_PUSH" == "true" ]]; then
        warning "Skipping push (use 'git push origin $VERSION' to push manually)"
        return
    fi
    
    info "Pushing changes to remote..."
    
    # Push commits first
    if git rev-parse --verify HEAD@{upstream} >/dev/null 2>&1; then
        if ! git push origin HEAD; then
            error "Failed to push commits"
            exit 1
        fi
        success "Commits pushed"
    fi
    
    # Push tag
    if ! git push origin "$VERSION"; then
        error "Failed to push tag"
        exit 1
    fi
    
    success "Tag $VERSION pushed to remote"
}

# Main execution
main() {
    echo "=========================================="
    echo "  abp-gen Release Script"
    echo "=========================================="
    echo ""
    
    check_prerequisites
    run_checks
    get_version
    
    echo ""
    info "Preparing release $VERSION..."
    echo ""
    
    update_changelog
    update_documentation
    create_tag
    push_changes
    
    echo ""
    echo "=========================================="
    success "Release $VERSION prepared successfully!"
    echo "=========================================="
    echo ""
    info "Next steps:"
    echo "  1. GitHub Actions will automatically build and publish the release"
    echo "  2. Monitor the release workflow: https://github.com/mohamedhabibwork/abp-gen/actions"
    echo "  3. Verify the release: https://github.com/mohamedhabibwork/abp-gen/releases"
    echo ""
    
    if [[ "$SKIP_PUSH" == "true" ]]; then
        warning "Remember to push the tag manually:"
        echo "  git push origin $VERSION"
        echo ""
    fi
}

# Run main function
main

