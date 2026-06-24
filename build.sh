#!/bin/bash
set -e

source ~/.bashrc 2>/dev/null || true
source ~/.profile 2>/dev/null || true
export PATH="$HOME/go/bin:/usr/local/go/bin:$PATH"

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BUILD_DIR="$SCRIPT_DIR/compiled"

mkdir -p "$BUILD_DIR"
cd "$SCRIPT_DIR"

VERSION=$(date +"%Y.%-m.%-d.%H%M")
echo "=> Version: $VERSION"

LDFLAGS="-X 'craftopiamc-launcher/core.AppVersion=$VERSION'"

# ---- Convert icon ----
ICON_SRC="$SCRIPT_DIR/assets/appicon.png"
if [ -f "$ICON_SRC" ]; then
    echo "=> Converting icons..."
    convert "$ICON_SRC" -resize 256x256 "$SCRIPT_DIR/modules/tray-icon.png" 2>/dev/null && echo "   tray-icon.png: done"
    convert "$ICON_SRC" -resize 256x256 "$SCRIPT_DIR/modules/tray_icon.ico" 2>/dev/null && echo "   tray_icon.ico: done"
    convert "$ICON_SRC" -resize 256x256 "$SCRIPT_DIR/build/windows/icon.ico" 2>/dev/null && echo "   build/windows/icon.ico: done"
    convert "$ICON_SRC" -resize 256x256 "$SCRIPT_DIR/assets/appicon.png" 2>/dev/null && echo "   appicon.png: done"
fi

# ---- LINUX ----
rm -rf "$SCRIPT_DIR/build/bin"
echo "=> Building Linux (amd64)..."
wails build -clean -platform linux/amd64 -ldflags "$LDFLAGS" -o "launcher"
LINUX_BUILT="$SCRIPT_DIR/build/bin/launcher"
if [ -f "$LINUX_BUILT" ]; then
    rm -f "$BUILD_DIR/launcher"
    cp "$LINUX_BUILT" "$BUILD_DIR/launcher"
    echo "=> Linux: compiled/launcher"
fi

# ---- WINDOWS ----
rm -rf "$SCRIPT_DIR/build/bin"
echo "=> Building Windows (amd64)..."
wails build -clean -platform windows/amd64 -ldflags "$LDFLAGS" -o "launcher.exe"
WIN_BUILT="$SCRIPT_DIR/build/bin/launcher.exe"
if [ -f "$WIN_BUILT" ]; then
    rm -f "$BUILD_DIR/launcher.exe"
    cp "$WIN_BUILT" "$BUILD_DIR/launcher.exe"
    echo "=> Windows: compiled/launcher.exe"
fi

rm -rf "$SCRIPT_DIR/build/bin" 2>/dev/null
rm -rf "$SCRIPT_DIR/build/frontend" 2>/dev/null
echo "=> Done! Version: $VERSION"
