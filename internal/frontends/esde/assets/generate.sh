#!/usr/bin/env nix-shell
#!nix-shell -i bash -p imagemagick librsvg curl

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

ESDE_COMMIT="19c93c75d00ce30285b2915fce2283ded181e26a"
ESDE_RAW="https://gitlab.com/es-de/emulationstation-de/-/raw/${ESDE_COMMIT}"

BG_COLOR="#7a7a76"
BLEND_OPACITY=5

cd "$SCRIPT_DIR"

echo "Fetching ES-DE assets (commit ${ESDE_COMMIT:0:8})..."
curl -fsSL "$ESDE_RAW/resources/graphics/splash.svg" -o /tmp/esde_splash.svg
curl -fsSL "$ESDE_RAW/es-app/assets/org.es_de.frontend.svg" -o /tmp/esde_icon.svg

echo "Generating high-res grain texture..."
magick -size 1920x1920 xc:gray50 +noise Gaussian -blur 0x0.3 -normalize \
  -modulate 100,0 -level 40%,60% /tmp/grain_hires.png

magick /tmp/grain_hires.png -resize 600x600 /tmp/grain_grid.png
magick /tmp/grain_hires.png -resize 460x460 /tmp/grain_capsule.png

echo "Generating hero.jpg (1920x620)..."
magick -size 1920x620 xc:"$BG_COLOR" \
  \( /tmp/grain_hires.png -crop 1920x620+0+0 +repage \) \
  -define compose:args=$BLEND_OPACITY -compose blend -composite \
  -quality 92 hero.jpg

echo "Generating grid.jpg (600x900)..."
rsvg-convert -w 350 -h 350 /tmp/esde_icon.svg -o /tmp/icon_grid.png
magick -size 600x900 xc:"$BG_COLOR" \
  \( -size 600x900 tile:/tmp/grain_grid.png \) \
  -define compose:args=$BLEND_OPACITY -compose blend -composite \
  /tmp/icon_grid.png -gravity center -compose src-over -composite \
  -quality 92 grid.jpg

echo "Generating capsule.jpg (460x215)..."
rsvg-convert -w 300 -a /tmp/esde_splash.svg -o /tmp/splash_capsule.png
magick -size 460x215 xc:"$BG_COLOR" \
  \( -size 460x215 tile:/tmp/grain_capsule.png \) \
  -define compose:args=$BLEND_OPACITY -compose blend -composite \
  /tmp/splash_capsule.png -gravity center -compose src-over -composite \
  -quality 92 capsule.jpg

echo "Generating logo.png (transparent)..."
rsvg-convert -w 400 -a /tmp/esde_splash.svg -o /tmp/logo_full.png
magick /tmp/logo_full.png -trim +repage logo.png

echo "Done!"
ls -lh *.jpg *.png
