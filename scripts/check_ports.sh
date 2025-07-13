#!/usr/bin/env bash

# Check if assigned ports in constants.json are running
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
CONSTANTS_FILE="$SCRIPT_DIR/../constants.json"

jq -c '.[]' "$CONSTANTS_FILE" | while read -r item; do
  PORT=$(echo $item | jq -r '.port')
  APPID=$(echo $item | jq -r '.appId')
  SUBDOMAIN=$(echo $item | jq -r '.subdomain')

  if lsof -iTCP:$PORT -sTCP:LISTEN -P | grep -q LISTEN; then
    echo "✅ Port $PORT for $APPID ($SUBDOMAIN) is running."
  else
    echo "❌ Port $PORT for $APPID ($SUBDOMAIN) is NOT running!"
  fi

done
