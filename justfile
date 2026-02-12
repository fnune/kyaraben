# Kyaraben task runner

default:
    @just --list

# Generate TypeScript types from Go
generate-types:
    go run github.com/gzuidhof/tygo@v0.2.20 generate

# Run the Electron app in development mode
dev: _ensure-ui-deps generate-types _sidecar
    cd ui && npm run dev

# Run all checks (lint + test)
check: lint test
    cd ui && npm run typecheck && npm run lint

# Run Go tests
test:
    go test ./...

# Run Go linter
lint:
    golangci-lint run

# Format all code
fmt:
    gofmt -w .
    goimports -w -local github.com/fnune/kyaraben .
    cd ui && npm run lint:fix
    cd site && npm run fmt

# Build release AppImage
build: _ensure-ui-deps generate-types _sidecar
    cd ui && npm run electron:build

# Run CLI e2e tests in container (nix builds)
e2e: _container-e2e-build
    podman run -it --rm kyaraben-nix-e2e

# Run Playwright UI e2e tests in container (headless)
ui-e2e: _container-electron-e2e-build
    podman run --ipc=host --rm kyaraben-electron-e2e

# Run Playwright UI e2e tests with interactive UI (run 'just build' first)
ui-e2e-ui *args: _extract-appimage
    #!/usr/bin/env bash
    cd ui && KYARABEN_APPIMAGE="$(pwd)/../.sandbox/app/kyaraben-ui" \
        APPDIR="$(pwd)/../.sandbox/app" \
        npx playwright test --ui {{ args }}

# Run app in sandbox container for manual testing (persistent state)
sandbox: build _container-sandbox-build _extract-appimage
    #!/usr/bin/env bash
    mkdir -p .sandbox/home
    podman run -it --rm \
        --userns=keep-id \
        -e WAYLAND_DISPLAY=$WAYLAND_DISPLAY \
        -e XDG_RUNTIME_DIR=/run/user/$(id -u) \
        -e DBUS_SESSION_BUS_ADDRESS="unix:path=/run/user/$(id -u)/bus" \
        -v $XDG_RUNTIME_DIR:/run/user/$(id -u) \
        -v /run/dbus/system_bus_socket:/run/dbus/system_bus_socket:ro \
        -v "$(pwd)/.sandbox/app:/app:ro" \
        -v "$(pwd)/.sandbox/home:/home/sandbox" \
        --device /dev/dri \
        --security-opt label=disable \
        kyaraben-sandbox

# Run site development server
site-dev: _ensure-site-deps
    cd site && npm run dev

# Build documentation site
site-build: _ensure-site-deps
    cd site && npm run build

# Clean build artifacts
clean:
    rm -f kyaraben
    rm -rf ui/dist ui/dist-electron ui/release ui/binaries

# Clean all sandbox state (chmod needed because nix store paths are read-only)
clean-sandbox:
    chmod -R u+w .sandbox 2>/dev/null || true
    rm -rf .sandbox

# Clean all emulator config directories (for development/testing)
clean-emu-configs:
    #!/usr/bin/env bash
    set -euo pipefail

    # XDG config dirs
    config_dirs=(
        "$HOME/.config/azahar"
        "$HOME/.config/Cemu"
        "$HOME/.config/dolphin-emu"
        "$HOME/.config/duckstation"
        "$HOME/.config/flycast"
        "$HOME/.config/melonDS"
        "$HOME/.config/mgba"
        "$HOME/.config/PCSX2"
        "$HOME/.config/ppsspp"
        "$HOME/.config/retroarch"
        "$HOME/.config/rpcs3"
    )

    # XDG data dirs (for emulators using symlinks)
    data_dirs=(
        "$HOME/.local/share/Cemu"
        "$HOME/.local/share/dolphin-emu"
    )

    # Frontend dirs
    frontend_dirs=(
        "$HOME/ES-DE"
    )

    all_dirs=("${config_dirs[@]}" "${data_dirs[@]}" "${frontend_dirs[@]}")

    echo "Emulator directories that will be removed:"
    echo
    echo "Config (~/.config/):"
    found=0
    for dir in "${config_dirs[@]}"; do
        if [ -d "$dir" ]; then
            size=$(du -sh "$dir" 2>/dev/null | cut -f1)
            echo "  [EXISTS] $dir ($size)"
            found=$((found + 1))
        fi
    done

    echo
    echo "Data (~/.local/share/):"
    for dir in "${data_dirs[@]}"; do
        if [ -d "$dir" ] || [ -L "$dir" ]; then
            size=$(du -sh "$dir" 2>/dev/null | cut -f1 || echo "symlink")
            echo "  [EXISTS] $dir ($size)"
            found=$((found + 1))
        fi
    done

    echo
    echo "Frontends:"
    for dir in "${frontend_dirs[@]}"; do
        if [ -d "$dir" ]; then
            size=$(du -sh "$dir" 2>/dev/null | cut -f1)
            echo "  [EXISTS] $dir ($size)"
            found=$((found + 1))
        fi
    done
    echo

    if [ $found -eq 0 ]; then
        echo "No directories found to remove."
        exit 0
    fi

    echo "This will remove $found directories/symlinks."
    read -p "Continue? [y/N] " confirm
    if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
        echo "Aborted."
        exit 1
    fi

    for dir in "${all_dirs[@]}"; do
        if [ -d "$dir" ] || [ -L "$dir" ]; then
            echo "Removing $dir..."
            rm -rf "$dir"
        fi
    done

    echo "Done."

# --- Internal targets (prefixed with _) ---

_ensure-ui-deps:
    cd ui && npm ci

_ensure-site-deps:
    cd site && npm ci

_sidecar:
    ./scripts/build-sidecar.sh

_container-e2e-build:
    podman build -t kyaraben-nix-e2e -f Containerfile.nix-e2e .

_container-electron-e2e-build:
    podman build -t kyaraben-electron-e2e -f Containerfile.electron-e2e .

_container-sandbox-build:
    podman build -t kyaraben-sandbox -f Containerfile.sandbox \
        --build-arg USER_ID=$(id -u) --build-arg GROUP_ID=$(id -g) .

_extract-appimage:
    #!/usr/bin/env bash
    appimage=$(realpath ui/release/Kyaraben-*-x86_64.AppImage 2>/dev/null | head -1)
    if [ -z "$appimage" ]; then
        echo "AppImage not found. Run 'just build' first."
        exit 1
    fi
    rm -rf .sandbox/app
    mkdir -p .sandbox/app
    (cd .sandbox/app && "$appimage" --appimage-extract > /dev/null)
    mv .sandbox/app/squashfs-root/* .sandbox/app/
    rm -rf .sandbox/app/squashfs-root
