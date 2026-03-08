#!/usr/bin/env bash
set -euo pipefail

: "${KOYEB_TOKEN:?KOYEB_TOKEN must be set}"

APP_NAME="kyaraben-site"
SERVICE_NAME="site"
REGION="fra"

cd "$(dirname "$0")/.."

if ! koyeb app get "$APP_NAME" &>/dev/null; then
    echo "Creating app $APP_NAME..."
    koyeb app create "$APP_NAME"
fi

echo "Uploading archive..."
ARCHIVE_ID=$(koyeb archives create . -o json | jq -r '.archive.id')
echo "Archive ID: $ARCHIVE_ID"

if koyeb service get "$APP_NAME/$SERVICE_NAME" &>/dev/null; then
    echo "Updating service $SERVICE_NAME..."
    koyeb service update "$APP_NAME/$SERVICE_NAME" \
        --archive "$ARCHIVE_ID" \
        --archive-builder docker \
        --archive-docker-dockerfile Containerfile
else
    echo "Creating service $SERVICE_NAME..."
    koyeb service create "$SERVICE_NAME" \
        --app "$APP_NAME" \
        --archive "$ARCHIVE_ID" \
        --archive-builder docker \
        --archive-docker-dockerfile Containerfile \
        --regions "$REGION" \
        --instance-type free \
        --ports 8080:http \
        --routes /:8080 \
        --checks 8080:http:/ \
        --min-scale 0 \
        --max-scale 1
fi

echo "Deployment complete"
koyeb service get "$APP_NAME/$SERVICE_NAME"
