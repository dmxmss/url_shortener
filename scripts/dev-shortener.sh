#!/usr/bin/env sh
set -eu

curl -s -X POST "${API_URL:-http://localhost:8080}/api/shorten" \
  -H 'content-type: application/json' \
  -d '{"url":"https://example.com/very/long/url"}'

