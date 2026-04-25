# Release Process

This document describes how to create and publish a release of go-socks5.

## Prerequisites

- Access to push to the repository
- GPG key configured for signing (optional, for official releases)

## Release Checklist

### 1. Update Version

Update the version in these files:

| File | What to Update |
|------|----------------|
| `debian/changelog` | Version number and date |
| `debian/control` | Version if needed |
| `README.md` | Update version if shown |

### 2. Test the Build

```bash
# Build the .deb locally to verify
make build-deb
```

This will create:
- `../go-socks5_{version}_amd64.deb`
- `../go-socks5_{version}.dsc`
- `../go-socks5_{version}_amd64.buildinfo`
- `../go-socks5_{version}_amd64.changes`
- `../go-socks5_{version}.tar.xz`

### 3. Tag the Release

```bash
# Create and push the tag
git tag v1.0.0
git push origin v1.0.0
```

Or for a release candidate:
```bash
git tag v1.0.0-rc1
git push origin v1.0.0-rc1
```

### 4. GitHub Action Builds .deb

The GitHub Action (`.github/workflows/deb.yml`) automatically:
1. Triggers on any `v*` tag
2. Builds the .deb package
3. Attaches it to the GitHub Release

### 5. Verify Release

Check that:
- [ ] Tag pushed successfully
- [ ] GitHub Action completed without errors
- [ ] .deb attached to GitHub Release
- [ ] Package installs correctly

## Version Numbering

Follow [Semantic Versioning](https://semver.org/):

```
MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]
```

Examples:
- `1.0.0` - Major release
- `1.0.1` - Patch/bug fix release
- `1.1.0` - Minor feature release
- `1.0.0-rc1` - Release candidate
- `1.0.0-alpha1` - Alpha release

## Manual Build (Without GitHub)

If you need to build locally:

```bash
# Install build dependencies
sudo apt-get install -y dh-make debhelper fakeroot

# Build the package
dpkg-buildpackage -us -uc

# The .deb will be in the parent directory
ls -la ../go-socks5_*.deb
```

## Post-Release

After a successful release:

1. Update `debian/changelog` for next version
2. Update any version references in documentation
3. Announce the release

## Rollback Procedure

If something goes wrong:

1. Delete the tag: `git tag -d v1.0.0 && git push origin :refs/tags/v1.0.0`
2. Delete the GitHub Release from the UI
3. Fix the issue
4. Create a new tag with incremented version (e.g., `v1.0.1`)