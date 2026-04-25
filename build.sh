#!/bin/bash
set -e

VERSION=${1:-$(git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//')}
if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version> or run from a git repo with tags"
    exit 1
fi

echo "Building go-socks5 v${VERSION}"

# Clean
rm -rf build/
mkdir -p build/

# Build amd64
echo "Building amd64..."
GOARCH=amd64 CGO_ENABLED=0 go build -o build/go-socks5-amd64 .

# Build arm64
echo "Building arm64..."
GOARCH=arm64 CGO_ENABLED=0 go build -o build/go-socks5-arm64 .

# Create amd64 package
echo "Creating amd64.deb..."
mkdir -p build/pkg-amd64/DEBIAN
mkdir -p build/pkg-amd64/usr/bin
mkdir -p build/pkg-amd64/lib/systemd/system
mkdir -p build/pkg-amd64/etc/default
cp build/go-socks5-amd64 build/pkg-amd64/usr/bin/go-socks5
cp debian/go-socks5.service build/pkg-amd64/lib/systemd/system/
cp debian/default build/pkg-amd64/etc/default/go-socks5
cp debian/postinst build/pkg-amd64/DEBIAN/
cp debian/prerm build/pkg-amd64/DEBIAN/
printf "Package: go-socks5\nVersion: %s\nArchitecture: amd64\nDepends: libc6 (>= 2.17), ca-certificates\nDescription: SOCKS5 proxy server\n" "$VERSION" > build/pkg-amd64/DEBIAN/control
dpkg-deb --build build/pkg-amd64/ "build/go-socks5_${VERSION}_amd64.deb"
rm -rf build/pkg-amd64

# Create arm64 package
echo "Creating arm64.deb..."
mkdir -p build/pkg-arm64/DEBIAN
mkdir -p build/pkg-arm64/usr/bin
mkdir -p build/pkg-arm64/lib/systemd/system
mkdir -p build/pkg-arm64/etc/default
cp build/go-socks5-arm64 build/pkg-arm64/usr/bin/go-socks5
cp debian/go-socks5.service build/pkg-arm64/lib/systemd/system/
cp debian/default build/pkg-arm64/etc/default/go-socks5
cp debian/postinst build/pkg-arm64/DEBIAN/
cp debian/prerm build/pkg-arm64/DEBIAN/
printf "Package: go-socks5\nVersion: %s\nArchitecture: arm64\nDepends: libc6 (>= 2.17), ca-certificates\nDescription: SOCKS5 proxy server\n" "$VERSION" > build/pkg-arm64/DEBIAN/control
dpkg-deb --build build/pkg-arm64/ "build/go-socks5_${VERSION}_arm64.deb"
rm -rf build/pkg-arm64

echo "Done! Packages:"
ls -lh build/*.deb

echo ""
echo "Upload with:"
echo "  gh release create v${VERSION} --title \"v${VERSION}\" build/*.deb"