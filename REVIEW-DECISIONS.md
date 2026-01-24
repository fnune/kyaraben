# Review decisions

This document tracks decisions made when addressing feedback from review
documents. Some items were intentionally deferred or declined.

## Addressed

These items were fixed in this review cycle:

### Architecture (REVIEW-ARCHITECTURE.md)

- Moved flake location from data to state directory (XDG compliance)
- Removed Tauri path artifacts from Electron main.ts and nix/client.go
- Implemented atomic manifest writes using temp file + rename pattern

### Testing (REVIEW-TESTING.md)

- Added config generator unit tests
- Added flake generator unit tests
- Created NixClient interface with FakeClient for testing
- Added race detection flag to CI (`go test -race`)

### Code quality (REVIEW-CODE-QUALITY.md)

- Fixed non-deterministic map iteration in config_writer.go (sorted keys)
- Made build timeout a named constant
- Added component-based logging to nix/client.go

### MVP (REVIEW-MVP.md)

- Created NixClient interface (addresses "no fake Nix client")

### Model (REVIEW-MODEL.md)

- Added ConfigPatch documentation to MODEL.md
- Added Registry documentation to MODEL.md

### Documentation

- Updated CONTRIBUTING.md: Go types as source of truth for daemon protocol
- Added logging section to CONTRIBUTING.md with component-based pattern

## Intentionally not addressed

These items were considered but deferred for the reasons stated.

### Request IDs in protocol (REVIEW-ARCHITECTURE.md)

The daemon currently handles requests sequentially. Adding request IDs would
require protocol version changes, UI updates, and handling concurrent requests.
This is an enhancement rather than a bug fix. The current sequential model
works for the MVP use case.

### nix-portable bundling / AppImage packaging (REVIEW-ARCHITECTURE.md)

This is infrastructure work that requires upstream coordination with
nix-portable releases. The current approach of expecting nix-portable in
specific locations works for development and can be addressed when preparing
for user distribution.

### Binary cache setup (REVIEW-ARCHITECTURE.md)

Setting up Cachix or another binary cache is an infrastructure decision that
should be made when the project has more usage. Current CI builds work without
caching, just slower.

### Config baseline tracking (REVIEW-MVP.md)

The `BaselineHash` field exists in the Manifest type but is never set. This
feature would enable detecting when user configurations drift from what
kyaraben last wrote. This is a feature addition beyond the current scope.
The current three-way merge approach preserves user changes without needing
baseline tracking.

### TypeScript types generated from Go (REVIEW-MVP.md, REVIEW-CODE-QUALITY.md)

Generating TypeScript types from Go would reduce manual duplication but
requires tooling setup (tygo, go-ts-generator, or similar). The current manual
approach works and the protocol surface is small enough that synchronization is
manageable. This should be addressed when the protocol stabilizes and grows.

### Emulator version tracking (REVIEW-MVP.md)

The Version field is always "latest" because extracting actual versions from
Nix outputs requires parsing derivation metadata. This is acceptable for MVP
since kyaraben uses nixpkgs unstable, and users effectively get the latest
version. Real version tracking can be added when there's a need to pin or
display specific versions.

### Home-manager module divergence (REVIEW-MVP.md)

The home-manager module may behave differently from the CLI. Documenting these
differences or unifying behavior requires understanding both codepaths in
detail. This should be addressed when someone actively uses the home-manager
module.

### Provision model assumes file-based checking (REVIEW-MODEL.md)

The current provision checking looks for files at specific paths. More flexible
checking (version validation, checksum verification) could be added, but the
current file-existence check works for BIOS files which are the primary use
case.

### Emulator.NixAttr leaks installation mechanism (REVIEW-MODEL.md)

The NixAttr field exposes that we use Nix for installation. This could be
abstracted behind a more generic "package reference" concept, but the
refactoring cost is not justified when Nix is the only installation mechanism
and no alternatives are planned.

### Table-driven tests (REVIEW-TESTING.md)

Some tests could be converted to table-driven style. The current tests are
readable and pass. Refactoring to table-driven style is a code style
improvement that can be done incrementally.

### Docker layer caching (REVIEW-TESTING.md)

The E2E Dockerfile could be optimized for layer caching. Current build times
are acceptable and CI has GHA caching enabled for Docker builds.
