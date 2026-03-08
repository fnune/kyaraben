#!/bin/sh
set -e

env -u PORT relay -addr :8081 &
exec caddy run --config /etc/caddy/Caddyfile
