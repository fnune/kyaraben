# NextUI integration

`just build` makes `dist/Kyaraben.pak.zip`. Each release uploads it, and the [NextUI Pak Store](https://github.com/LoveRetro/nextui-pak-store) installs it from there.

## Pak Store

The store reads `pak.json` from the root of `main` on GitHub once an hour, and downloads `releases/download/<version>/<release_filename>`. `just release-create` bumps `version`, so a release reaches the store with no further steps.

The listing breaks if `pak.json` moves, if `release_filename` stops matching the uploaded zip, or if the zip stops holding the pak's contents at its root. Its platforms and `experimental` flag live in `pak.json`. Display name and categories live in the store and change through an issue on its repository.
