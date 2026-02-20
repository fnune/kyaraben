#!/usr/bin/env nix-shell
#!nix-shell -i bash -p imagemagick librsvg curl

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

ESDE_COMMIT="19c93c75d00ce30285b2915fce2283ded181e26a"
ESDE_RAW="https://gitlab.com/es-de/emulationstation-de/-/raw/${ESDE_COMMIT}"

BG_COLOR="#7a7a76"

cd "$SCRIPT_DIR"

echo "Fetching ES-DE assets (commit ${ESDE_COMMIT:0:8})..."
curl -fsSL "$ESDE_RAW/resources/graphics/splash.svg" -o /tmp/esde_splash.svg
curl -fsSL "$ESDE_RAW/es-app/assets/org.es_de.frontend.svg" -o /tmp/esde_icon.svg

echo "Fetching grain texture..."
curl -fsSL "https://www.transparenttextures.com/patterns/noisy.png" -o /tmp/noisy.png

# Grain tile scaling for consistent visual appearance on Steam Deck
#
# Each image displays at a different size on Deck:
#   Grid   (600px source)  -> ~180px display -> 0.30x scale
#   Capsule (460px source) -> ~400px display -> 0.87x scale
#   Hero   (1920px source) -> ~1280px display -> 0.67x scale
#
# To get the same ~200px visual grain on screen:
#   tile_size = target_grain / display_scale
#
# Using hero (300px tile) as baseline:
#   Hero:    300px tile * 0.67 = 201px on screen
#   Grid:    670px tile * 0.30 = 201px on screen (coarser tile)
#   Capsule: 230px tile * 0.87 = 200px on screen (medium tile)

GRAIN_BASE="/tmp/noisy.png"  # 300px original
magick "$GRAIN_BASE" -resize 670x670 /tmp/grain_grid.png
magick "$GRAIN_BASE" -resize 230x230 /tmp/grain_capsule.png

echo "Generating grid.jpg (600x900, 670px grain tile)..."
rsvg-convert -w 350 -h 350 /tmp/esde_icon.svg -o /tmp/icon_grid.png
magick -size 600x900 xc:"$BG_COLOR" \
  \( -size 600x900 tile:/tmp/grain_grid.png -colorspace gray -evaluate multiply 0.7 \) \
  -compose multiply -composite \
  /tmp/icon_grid.png -gravity center -compose src-over -composite \
  -quality 92 grid.jpg

echo "Generating capsule.jpg (460x215, 230px grain tile)..."
rsvg-convert -w 300 -a /tmp/esde_splash.svg -o /tmp/splash_capsule.png
magick -size 460x215 xc:"$BG_COLOR" \
  \( -size 460x215 tile:/tmp/grain_capsule.png -colorspace gray -evaluate multiply 0.7 \) \
  -compose multiply -composite \
  /tmp/splash_capsule.png -gravity center -compose src-over -composite \
  -quality 92 capsule.jpg

echo "Generating hero.jpg (1920x620, 300px grain tile)..."
magick -size 1920x620 xc:"$BG_COLOR" \
  \( -size 1920x620 tile:"$GRAIN_BASE" -colorspace gray -evaluate multiply 0.7 \) \
  -compose multiply -composite \
  -quality 92 hero.jpg

echo "Generating logo.png (transparent, shifted left for Steam alignment)..."
rsvg-convert -w 300 -a /tmp/esde_splash.svg -o /tmp/logo_full.png
magick /tmp/logo_full.png -trim +repage /tmp/logo_trimmed.png
read width height < <(identify -format "%w %h" /tmp/logo_trimmed.png)
pad_v=$((height * 15 / 100))
pad_r=$((width * 30 / 100))  # 30% right, 0% left to shift logo left on screen
magick /tmp/logo_trimmed.png \
  -background transparent \
  -splice 0x${pad_v} \
  -gravity southwest -splice 0x${pad_v} \
  -gravity northeast -splice ${pad_r}x0 \
  logo.png

echo "Done!"
ls -lh *.jpg *.png
