# Release Process

This document describes the process for creating releases of abp-gen.

## Prerequisites

- Write access to the repository
- GoReleaser installed (or use GitHub Actions)
- Git configured with proper credentials

## Release Checklist

Before creating a release:

- [ ] All tests pass (`go test ./...`)
- [ ] Code is formatted (`go fmt ./...`)
- [ ] Linter passes (`golangci-lint run`)
- [ ] Documentation is updated
- [ ] CHANGELOG.md is updated
- [ ] Version numbers are updated (if needed)
- [ ] Breaking changes are documented

## Version Numbering

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR** version for incompatible API changes
- **MINOR** version for new functionality in a backwards compatible manner
- **PATCH** version for backwards compatible bug fixes

Examples:
- `v1.0.0` - Initial release
- `v1.1.0` - New features, backwards compatible
- `v1.1.1` - Bug fixes
- `v2.0.0` - Breaking changes

## Creating a Release

### 1. Prepare the Release

```bash
# Ensure you're on main branch and up to date
git checkout main
git pull origin main

# Run tests
go test ./...

# No need to manually update CHANGELOG.md - it will be auto-generated!
```

### 2. Create and Push Tag

```bash
# Create annotated tag
git tag -a v1.0.0 -m "Release v1.0.0"

# Push tag to trigger release workflow
git push origin v1.0.0
```

### 3. GitHub Actions Release

When you push a tag starting with `v`, GitHub Actions will automatically:

1. **Generate Changelog**: Extract commits since last release and update CHANGELOG.md
2. **Update Documentation**: Update version numbers in README.md and other docs
3. **Commit Changes**: Automatically commit changelog and doc updates
4. **Build Binaries**: Run GoReleaser to build for all platforms
5. **Create Release**: Create GitHub release with auto-generated notes
6. **Upload Artifacts**: Upload binaries, checksums, and packages

### 4. Verify Release

- Check GitHub Releases page
- Verify all platform binaries are uploaded
- Test downloading and running a binary
- Verify checksums are correct

## Manual Release (Alternative)

If you need to create a release manually:

```bash
# Install GoReleaser
go install github.com/goreleaser/goreleaser@latest

# Create a snapshot release (dry-run)
goreleaser release --snapshot

# Create actual release
GITHB_TOKEN=your_token goreleaser release
```

## Release Notes

Release notes are automatically generated from:

1. CHANGELOG.md entries
2. Git commits since last tag
3. Pull requests merged since last tag

To customize release notes, edit the release on GitHub after it's created.

## Post-Release Tasks

After a release:

- [ ] Announce on social media (if applicable)
- [ ] Update documentation site (if applicable)
- [ ] Notify users of breaking changes
- [ ] Monitor for issues or bugs

## Hotfix Releases

For urgent bug fixes:

1. Create hotfix branch from main
2. Fix the bug
3. Update CHANGELOG.md
4. Create patch version tag (e.g., `v1.0.1`)
5. Merge hotfix to main
6. Push tag to trigger release

## Pre-Release Testing

Before tagging a release, test:

```bash
# Build locally
go build ./cmd/abp-gen

# Test basic commands
./abp-gen --version
./abp-gen --help
./abp-gen init
./abp-gen generate --dry-run

# Run all tests
go test -v -race ./...

# Test on different platforms (if possible)
GOOS=linux GOARCH=amd64 go build ./cmd/abp-gen
GOOS=windows GOARCH=amd64 go build ./cmd/abp-gen
GOOS=darwin GOARCH=arm64 go build ./cmd/abp-gen
```

## Troubleshooting

### Release Workflow Fails

- Check GitHub Actions logs
- Verify GITHB_TOKEN has proper permissions
- Ensure tag format is correct (`v*`)
- Check GoReleaser configuration

### Missing Binaries

- Verify all platforms are configured in `.goreleaser.yml`
- Check build logs for errors
- Ensure GoReleaser version is up to date

### Release Notes Not Generated

- Verify CHANGELOG.md is updated
- Check git tag format
- Review GoReleaser changelog configuration

## Questions?

If you have questions about the release process, open an issue or contact maintainers.

