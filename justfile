# Kyaraben task runner

test_dir := ".test"

default:
    @just --list

# Generate TypeScript types from Go (disable CGO - tygo doesn't need it)
generate-types: ensure
    CGO_ENABLED=0 go run github.com/gzuidhof/tygo@v0.2.20 generate

# Install project dependencies
ensure:
    go mod download
    cd ui && npm ci
    cd site && npm ci
    pre-commit install --install-hooks -t pre-commit -t commit-msg

# Run the Electron app in development mode
dev: generate-types _sidecar
    cd ui && npm run dev

# Run all checks (lint + test)
check: ensure
    cd ui && npm run typecheck && npm run lint
    cd site && npm run fmt:check
    golangci-lint run
    cd relay && golangci-lint run
    go test -short ./...
    cd relay && go test -short ./...

# Run Go tests
test: ensure
    go test ./...
    cd relay && go test ./...

# Check for package version updates
check-versions:
    go run ./cmd/check-versions

# Run Go linter
lint: ensure
    golangci-lint run
    cd relay && golangci-lint run

# Format all code
fmt: ensure
    gofmt -w cmd internal test
    goimports -w -local github.com/fnune/kyaraben cmd internal test
    cd relay && gofmt -w .
    cd relay && goimports -w -local github.com/fnune/kyaraben/relay .
    cd ui && npm run lint:fix
    cd site && npm run fmt

# Build AppImage (dev version with timestamp)
build: ensure generate-types _sidecar
    cd ui && npm run electron:build

# Build AppImage for release (clean version number)
release: ensure generate-types
    RELEASE_BUILD=1 ./scripts/build-sidecar.sh
    cd ui && npm run electron:build

# Test release build locally without publishing
release-test: ensure generate-types
    RELEASE_BUILD=1 ./scripts/build-sidecar.sh
    goreleaser release --snapshot --clean
    cd ui && npm run electron:build

# Create and push a release tag (triggers CI release)
release-create version:
    #!/usr/bin/env bash
    set -euo pipefail
    if ! git diff --quiet || ! git diff --cached --quiet; then
        echo "Working directory not clean. Commit or stash changes first."
        exit 1
    fi
    if ! git diff --quiet HEAD origin/main; then
        echo "Local main differs from origin. Push first."
        exit 1
    fi

    echo "Generating changelog for v{{ version }}..."
    git-cliff --tag "v{{ version }}" -o CHANGELOG.md

    git add CHANGELOG.md
    git commit -m "chore(release): update changelog for v{{ version }}"
    git push origin main

    git tag "v{{ version }}"
    git push origin "v{{ version }}"
    echo "Tag v{{ version }} pushed. Watch: https://github.com/fnune/kyaraben/actions"

# Delete a release (GitHub Release + git tag)
release-delete version:
    gh release delete "v{{ version }}" --yes || true
    git tag -d "v{{ version }}" || true
    git push origin --delete "v{{ version }}" || true

# Run CLI e2e tests in container
e2e: _container-e2e-build
    podman run -it --rm kyaraben-cli-e2e

# Run Playwright UI e2e tests (run 'just build' first)
ui-e2e *args: _extract-appimage _relay-binary
    #!/usr/bin/env bash
    cd ui && \
        KYARABEN_APPIMAGE="$(pwd)/../{{ test_dir }}/app/kyaraben-ui" \
        APPDIR="$(pwd)/../{{ test_dir }}/app" \
        KYARABEN_RELAY_BINARY="$(pwd)/../{{ test_dir }}/relay/relay" \
        ELECTRON_OZONE_PLATFORM_HINT=x11 \
        xvfb-run --auto-servernum ../scripts/run-ui-e2e.sh npx playwright test {{ args }}

# Run Playwright UI e2e tests with interactive UI (run 'just build' first)
ui-e2e-ui *args: _extract-appimage
    #!/usr/bin/env bash
    cd ui && \
        KYARABEN_APPIMAGE="$(pwd)/../{{ test_dir }}/app/kyaraben-ui" \
        APPDIR="$(pwd)/../{{ test_dir }}/app" \
        ELECTRON_OZONE_PLATFORM_HINT=auto \
        ../scripts/run-ui-e2e.sh npx playwright test --ui {{ args }}

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

# Build combined container (relay + site)
deploy-build:
    podman build -t kyaraben -f deploy/Containerfile .

# Deploy to Koyeb (requires KOYEB_TOKEN)
deploy:
    ./deploy/scripts/deploy.sh

# Clean build artifacts
clean:
    rm -f kyaraben
    rm -rf ui/dist ui/dist-electron ui/release ui/binaries {{ test_dir }}

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

# Build CLI binary for current platform
build-cli:
    ./scripts/build-sidecar.sh

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
    mkdir -p {{ test_dir }}/relay
    cd relay && go build -o ../{{ test_dir }}/relay/relay ./cmd/relay

_container-e2e-build:
    podman build -t kyaraben-cli-e2e -f Containerfile.cli-e2e .

_extract-appimage:
    #!/usr/bin/env bash
    appimage=$(realpath ui/release/Kyaraben-*-x86_64.AppImage 2>/dev/null | head -1)
    if [ -z "$appimage" ]; then
        echo "AppImage not found. Run 'just build' first."
        exit 1
    fi
    rm -rf {{ test_dir }}/app
    mkdir -p {{ test_dir }}/app
    (cd {{ test_dir }}/app && "$appimage" --appimage-extract > /dev/null)
    mv {{ test_dir }}/app/squashfs-root/* {{ test_dir }}/app/
    rm -rf {{ test_dir }}/app/squashfs-root
