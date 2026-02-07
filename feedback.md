# Feedback

## Low-hanging fruit

- DuckStation onboarding wizard: needs a default config to prevent the wizard from appearing on first launch
- Dolphin autoupdate prompt: needs a default config to disable the built-in autoupdate mechanism
- UI ROM/BIOS folder access: add buttons to PathsModal for per-system ROM and BIOS folders (`~/Emulation/roms/{systemId}/`, `~/Emulation/bios/{systemId}/`)
- Fix frontend e2e test: version resolution in test environment doesn't match real app, causing Apply button visibility check to be skipped after Done click

## Important

- Automatic self-updates: users currently must manually download new AppImages
- Garbage collection: nix store grows unbounded, need a way to trigger cleanup via nix-portable and show space freed
- ES-DE as non-Steam application: add to Steam for Steam Deck game mode launch
- RetroArch download size display: misleading because it shows AppImage size for each core, not accounting for shared package. Total shown is wildly inaccurate when enabling multiple RetroArch cores
- "Discard changes" button UX: confusing when config differs from manifest but user didn't make changes in this session. Consider renaming to "Reset to installed state" or only showing when user made UI modifications

## Nice to have

- Version tracking script: programmatic way to check for new emulator versions, compare against versions.toml, create PRs when updates available
- Emulator health check in doctor: verify installed emulators work (check binaries exist, wrapper scripts valid, AppImage integrity)
- Backend preview command: move apply status logic (shared emulators, will-install/update/uninstall) from frontend to backend `CommandTypeApplyPreview`
- UI show all available emulators: when a system is enabled, show all available emulators with checkboxes instead of a dropdown that switches
- Sync UI: dedicated UI for pairing devices, monitoring sync status (currently CLI/config only)
- CLI review step: offer interactive review before overwriting user-modified managed keys, add `--dry-run` flag

## Probably not worth it

- Reconsider flake generations: each apply creates new generation directory and lock file with warning. Works fine, unclear benefit from changing
- Reconsider the manifest: could derive state from filesystem instead of tracking in manifest, but risky refactor for unclear benefit
- Controller configuration abstraction: every emulator handles this differently, massive scope
- Home-manager module: very niche audience, unclear if AppImages work on NixOS
- Reduce download size: 190MB is acceptable for one-time download, effort better spent elsewhere
