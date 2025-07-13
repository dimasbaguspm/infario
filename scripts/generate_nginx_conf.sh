#!/usr/bin/env bash

# Generate Nginx configs for all projects in constants.json


SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
CONSTANTS_FILE="$SCRIPT_DIR/../constants.json"
CONF_DIR="$SCRIPT_DIR/../conf.d"
TEMPLATE_FILE="$SCRIPT_DIR/../templates/nginx.conf.template"

if [ ! -f "$TEMPLATE_FILE" ]; then
  echo "Template file $TEMPLATE_FILE not found!"
  exit 1
fi

jq -c '.[]' "$CONSTANTS_FILE" | while read -r item; do
  SUBDOMAIN=$(echo $item | jq -r '.subdomain')
  PORT=$(echo $item | jq -r '.port')
  APPID=$(echo $item | jq -r '.appId')
  RATELIMIT=$(echo $item | jq -r '.rateLimitPerSecond // 20')
  ZONE_NAME=$(echo $APPID | sed 's/[^a-zA-Z0-9]/_/g')
  CONF_FILE="$CONF_DIR/$APPID.conf"
  sed \
    -e "s/{{SUBDOMAIN}}/$SUBDOMAIN/g" \
    -e "s/{{PORT}}/$PORT/g" \
    -e "s/{{APPID}}/$APPID/g" \
    -e "s/{{ZONE_NAME}}/$ZONE_NAME/g" \
    -e "s/{{RATELIMIT}}/$RATELIMIT/g" \
    "$TEMPLATE_FILE" > "$CONF_FILE"
done

echo "Nginx configs generated for all projects using template."
