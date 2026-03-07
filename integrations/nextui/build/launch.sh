#!/bin/sh
PAK_DIR="$(dirname "$0")"
export PAK_PATH="$PAK_DIR"

exec "$PAK_DIR/kyaraben-nextui"
