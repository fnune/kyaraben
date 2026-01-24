#!/usr/bin/env bash
# Post-process the Tauri AppImage to fix library/environment issues
#
# Issues fixed:
# 1. Remove bundled wayland libraries (causes symbol conflicts with system Mesa)
# 2. Fix AppRun to set GTK_MODULES="" (prevents loading incompatible system modules)
# 3. Fix GTK_PATH to not include system paths

set -euo pipefail

APPIMAGE="${1:-}"

if [ -z "$APPIMAGE" ] || [ ! -f "$APPIMAGE" ]; then
    echo "Usage: $0 <path-to-appimage>"
    echo "Error: AppImage not found: $APPIMAGE"
    exit 1
fi

echo "Fixing AppImage: $APPIMAGE"

# Create temp directory for extraction
WORK_DIR=$(mktemp -d)
trap "rm -rf $WORK_DIR" EXIT

cd "$WORK_DIR"

# Extract AppImage
chmod +x "$APPIMAGE"
"$APPIMAGE" --appimage-extract >/dev/null 2>&1

SQUASHFS_ROOT="$WORK_DIR/squashfs-root"

if [ ! -d "$SQUASHFS_ROOT" ]; then
    echo "Error: Failed to extract AppImage"
    exit 1
fi

echo "Removing bundled wayland libraries..."
find "$SQUASHFS_ROOT" -name "libwayland-client.so*" -delete -print 2>/dev/null || true
find "$SQUASHFS_ROOT" -name "libwayland-cursor.so*" -delete -print 2>/dev/null || true
find "$SQUASHFS_ROOT" -name "libwayland-egl.so*" -delete -print 2>/dev/null || true
find "$SQUASHFS_ROOT" -name "libwayland-server.so*" -delete -print 2>/dev/null || true

echo "Patching GTK plugin hook..."
GTK_HOOK="$SQUASHFS_ROOT/apprun-hooks/linuxdeploy-plugin-gtk.sh"
if [ -f "$GTK_HOOK" ]; then
    # Remove system GTK paths that cause ABI conflicts
    sed -i 's|export GTK_PATH=.*|export GTK_PATH="${APPDIR}/usr/lib/gtk-3.0"|g' "$GTK_HOOK"
    # Add GTK_MODULES="" to prevent loading system modules
    if ! grep -q 'GTK_MODULES=' "$GTK_HOOK"; then
        echo 'export GTK_MODULES=""' >> "$GTK_HOOK"
    fi
    echo "  Patched: $GTK_HOOK"
fi

echo "Repacking AppImage..."

# Download appimagetool if needed
APPIMAGETOOL="$WORK_DIR/appimagetool"
curl -fsSL "https://github.com/AppImage/appimagetool/releases/download/continuous/appimagetool-x86_64.AppImage" -o "$APPIMAGETOOL"
chmod +x "$APPIMAGETOOL"

# Repack - appimagetool needs APPIMAGE_EXTRACT_AND_RUN in Docker
export APPIMAGE_EXTRACT_AND_RUN=1
ARCH=x86_64 "$APPIMAGETOOL" "$SQUASHFS_ROOT" "$APPIMAGE.fixed" 2>/dev/null

# Replace original
mv "$APPIMAGE.fixed" "$APPIMAGE"

echo "Done! Fixed: $APPIMAGE"
