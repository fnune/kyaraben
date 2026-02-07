# Screenshots

This directory holds screenshots used in the documentation site.

## Naming convention

Screenshots follow the pattern:

```
v<version>-<page>-<description>.png
```

Examples:

- `v0.1-getting-started-system-picker.png`
- `v0.1-getting-started-apply-progress.png`
- `v0.1-using-doctor-view.png`
- `v0.1-using-preflight-diff.png`
- `v0.1-using-status-output.png`
- `v0.1-sync-status-bar.png`

The version prefix makes it easy to identify which screenshots need re-capturing
after a UI update. When updating screenshots for a new version, add new files with
the new version prefix and update the references in the documentation pages.

## Adding screenshots

Search for `{/* Screenshot:` in the documentation source files to find all
placeholder locations where screenshots should be added. Replace the MDX comment
with a standard image import:

```mdx
import systemPicker from '../../assets/screenshots/v0.1-getting-started-system-picker.png'

<img src={systemPicker.src} alt="The system picker showing available systems grouped by manufacturer" />
```
