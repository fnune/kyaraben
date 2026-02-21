# Kyaraben task runner

default:
    @just --list

# Generate TypeScript types from Go
generate-types: ensure
    go run github.com/gzuidhof/tygo@v0.2.20 generate

# Install project dependencies
ensure:
    go mod download
    cd ui && npm ci
    cd site && npm ci

# Run the Electron app in development mode
dev: generate-types _sidecar
    cd ui && npm run dev

# Run all checks (lint + test)
check: ensure lint test
    cd ui && npm run typecheck && npm run lint
    cd site && npm run fmt:check

# Run Go tests
test: ensure
    go test ./...
    cd relay && go test ./...

# Run Go linter
lint: ensure
    golangci-lint run

# Format all code
fmt: ensure
    gofmt -w cmd internal test
    goimports -w -local github.com/fnune/kyaraben cmd internal test
    cd ui && npm run lint:fix
    cd site && npm run fmt

# Build AppImage (dev version with timestamp)
build: ensure generate-types _sidecar
    cd ui && npm run electron:build

# Build AppImage for release (clean version number)
release: ensure generate-types
    RELEASE_BUILD=1 ./scripts/build-sidecar.sh
    cd ui && npm run electron:build

# Run CLI e2e tests in container
e2e: _container-e2e-build
    podman run -it --rm kyaraben-cli-e2e

# Run Playwright UI e2e tests (run 'just build' first)
ui-e2e *args: _extract-appimage _relay-binary
    #!/usr/bin/env bash
    cd ui && \
        KYARABEN_APPIMAGE="$(pwd)/../.sandbox/app/kyaraben-ui" \
        APPDIR="$(pwd)/../.sandbox/app" \
        KYARABEN_RELAY_BINARY="$(pwd)/../.sandbox/relay/relay" \
        ../scripts/run-ui-e2e.sh npx playwright test {{ args }}

# Run Playwright UI e2e tests with interactive UI (run 'just build' first)
ui-e2e-ui *args: _extract-appimage
    #!/usr/bin/env bash
    cd ui && \
        KYARABEN_APPIMAGE="$(pwd)/../.sandbox/app/kyaraben-ui" \
        APPDIR="$(pwd)/../.sandbox/app" \
        ../scripts/run-ui-e2e.sh npx playwright test --ui {{ args }}

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
site-dev: ensure
    cd site && npm run dev

# Build documentation site
site-build: ensure
    cd site && npm run build

# Run relay server in development mode
relay-dev:
    cd relay && go run ./cmd/relay

# Run relay tests
relay-test:
    cd relay && go test ./...

# Build relay container
relay-build:
    podman build -t kyaraben-relay -f relay/Containerfile relay/

# Deploy relay to Koyeb (requires KOYEB_TOKEN)
relay-deploy:
    ./relay/scripts/deploy.sh

# Clean build artifacts
clean:
    rm -f kyaraben
    rm -rf ui/dist ui/dist-electron ui/release ui/binaries

# Clean all sandbox state (chmod needed for read-only paths)
clean-sandbox:
    chmod -R u+w .sandbox 2>/dev/null || true
    rm -rf .sandbox

# Clean all emulator config directories (for development/testing)
clean-emu-configs:
    #!/usr/bin/env bash
    set -euo pipefail

    # XDG config dirs
    config_dirs=(
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

# Deploy AppImage to Steam Deck SD card via SSH
deploy-deck:
    scp ui/release/Kyaraben-*-x86_64.AppImage deck@steamdeck:/run/media/Emulation/External/

# Run an additional kyaraben instance for local sync testing
# Usage: just instance secondary
instance name: _sidecar
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Running kyaraben instance '{{ name }}'"
    echo "  Config: ~/.config/kyaraben-{{ name }}/config.toml"
    echo "  State:  ~/.local/state/kyaraben-{{ name }}/"
    echo ""
    cd ui && npm run build && npm run build:electron && npx electron . -- --instance {{ name }}

# --- Internal targets (prefixed with _) ---

 

_sidecar:
    ./scripts/build-sidecar.sh

_relay-binary:
    #!/usr/bin/env bash
    mkdir -p .sandbox/relay
    cd relay && go build -o ../.sandbox/relay/relay ./cmd/relay

_container-e2e-build:
    podman build -t kyaraben-cli-e2e -f Containerfile.cli-e2e .

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
