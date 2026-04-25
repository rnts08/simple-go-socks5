# Release Process

This document describes how to create and publish a release of go-socks5.

## Prerequisites

- Access to push to the repository
- `gh` CLI installed and authenticated

## Release Checklist

### 1. Update Version (if needed)

Update version in relevant files (debian/changelog, etc.)

### 2. Build the Release

```bash
./build.sh 1.2.3
```

This creates:
- `build/go-socks5_1.2.3_amd64.deb`
- `build/go-socks5_1.2.3_arm64.deb`

### 3. Create GitHub Release

```bash
gh release create v1.2.3 --title "v1.2.3" build/*.deb
```

### 4. Verify Release

Check that:
- [ ] .deb packages uploaded correctly
- [ ] Package installs correctly

### 5. Cleanup

```bash
rm -rf build/
```

## Version Numbering

Follow [Semantic Versioning](https://semver.org/):

```
MAJOR.MINOR.PATCH
```

Examples:
- `1.0.0` - Major release
- `1.0.1` - Patch/bug fix release
- `1.1.0` - Minor feature release

## Rollback Procedure

If something goes wrong:

1. Delete the release from GitHub UI
2. Delete the tag: `git tag -d v1.0.0 && git push origin :refs/tags/v1.0.0`
3. Fix the issue
4. Create new release with incremented version