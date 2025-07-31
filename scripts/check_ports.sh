#!/usr/bin/env bash

# Check if assigned ports in constants.json are running
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
CONSTANTS_FILE="$SCRIPT_DIR/../constants.json"

jq -c '.[]' "$CONSTANTS_FILE" | while read -r item; do
  TYPE=$(echo $item | jq -r '.type')
  if [ "$TYPE" != "serve" ]; then
    APPID=$(echo $item | jq -r '.appId')
    SUBDOMAIN=$(echo $item | jq -r '.subdomain')
    echo "ℹ️  Skipping port check for $APPID ($SUBDOMAIN) because type is not 'serve'."
    continue
  fi
  PORT=$(echo $item | jq -r '.port')
  APPID=$(echo $item | jq -r '.appId')
  SUBDOMAIN=$(echo $item | jq -r '.subdomain')

  # Check on host
  if lsof -iTCP:$PORT -sTCP:LISTEN -P | grep -q LISTEN; then
    echo "✅ Port $PORT for $APPID ($SUBDOMAIN) is running on host."
  else
    echo "❌ Port $PORT for $APPID ($SUBDOMAIN) is NOT running on host!"
  fi

  # Check in Docker containers
  CONTAINER_ID=$(docker ps --format '{{.ID}} {{.Ports}} {{.Names}}' | grep ":$PORT" | awk '{print $1}')
  if [ -n "$CONTAINER_ID" ]; then
    # Check if port is listening inside the container
    if docker exec "$CONTAINER_ID" sh -c "netstat -tln | grep :$PORT" >/dev/null 2>&1; then
      echo "✅ Port $PORT for $APPID ($SUBDOMAIN) is listening inside Docker container $CONTAINER_ID."
    else
      echo "❌ Port $PORT for $APPID ($SUBDOMAIN) is NOT listening inside Docker container $CONTAINER_ID!"
    fi
  else
    echo "❌ No Docker container exposing port $PORT for $APPID ($SUBDOMAIN)."
  fi
done
