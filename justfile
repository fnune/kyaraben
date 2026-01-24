# Kyaraben task runner

# Default: show available tasks
default:
    @just --list

# Build the CLI
build:
    go build -o kyaraben ./cmd/kyaraben

# Run all tests
test:
    go test ./...

# Run Go linter
lint:
    golangci-lint run

# Format Go code
fmt:
    gofmt -w .
    goimports -w -local github.com/fnune/kyaraben .

# Clean build artifacts
clean:
    rm -f kyaraben

# UI: install dependencies
ui-install:
    cd ui && npm ci

# UI: run dev server (frontend only)
ui-dev:
    cd ui && npm run dev

# UI: run Tauri dev (full app with sidecar)
ui-tauri-dev: sidecar
    cd ui && npm run tauri dev

# UI: build for production
ui-build:
    cd ui && npm run build

# UI: lint
ui-lint:
    cd ui && npm run lint

# UI: fix lint issues
ui-lint-fix:
    cd ui && npm run lint:fix

# Build Go CLI as Tauri sidecar
sidecar:
    ./scripts/build-sidecar.sh

# UI: build Tauri app (release)
ui-tauri-build: sidecar
    cd ui && npm run tauri build

# UI: run E2E tests (requires built Tauri app + tauri-driver)
ui-test-e2e:
    cd ui && npm run test:e2e

# Build dev container
container-build:
    podman build -t kyaraben-dev -f Containerfile.dev .

# Run dev container (interactive)
container-run:
    podman run -it --rm -v "$(pwd):/workspace:Z" kyaraben-dev

# Build CLI E2E container (with Nix)
container-e2e-build:
    podman build -t kyaraben-nix-e2e -f Containerfile.nix-e2e .

# Run CLI E2E tests in container
container-e2e:
    podman run -it --rm kyaraben-nix-e2e

# Build Tauri E2E container (full UI tests)
container-tauri-e2e-build:
    podman build -t kyaraben-tauri-e2e -f Containerfile.tauri-e2e .

# Run Tauri E2E tests in container
container-tauri-e2e:
    podman run -it --rm kyaraben-tauri-e2e

# All checks (lint + test)
check: lint test ui-lint
