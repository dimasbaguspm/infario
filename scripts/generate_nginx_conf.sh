#!/usr/bin/env bash

# Generate Nginx configs for all projects in constants.json


SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
CONSTANTS_FILE="$SCRIPT_DIR/../constants.json"
CONF_DIR="$SCRIPT_DIR/../conf.d"

# Set USER to SUDO_USER if running with sudo, otherwise to whoami
if [ -n "$SUDO_USER" ]; then
  USER="$SUDO_USER"
else
  USER=$(whoami)
fi


rm -f "$CONF_DIR"/*.conf
mkdir -p "$CONF_DIR"

TEMPLATE_SERVE_FILE="$SCRIPT_DIR/../templates/nginx.conf.serve.template"
TEMPLATE_STATIC_FILE="$SCRIPT_DIR/../templates/nginx.conf.static.template"

if [ ! -f "$TEMPLATE_SERVE_FILE" ] || [ ! -f "$TEMPLATE_STATIC_FILE" ]; then
  echo "Template files $TEMPLATE_SERVE_FILE or $TEMPLATE_STATIC_FILE not found!"
  exit 1
fi

jq -c '.projects[]' "$CONSTANTS_FILE" | while read -r item; do
  DOMAIN=$(echo $item | jq -r '.domain')
  TYPE=$(echo $item | jq -r '.type')
  PORT=$(echo $item | jq -r '.port // empty')
  IP_HOST=$(jq -r '.host' "$CONSTANTS_FILE")
  APPID=$(echo $item | jq -r '.appId')
  RATELIMIT=$(echo $item | jq -r '.rateLimitPerSecond // 20')
  ZONE_NAME=$(echo $APPID | sed 's/[^a-zA-Z0-9]/_/g')
  PATH_VAL=$(echo $item | jq -r '.path // empty')
  # Robustly replace leading ~/ with /home/${USER}/
  PATH_VAL="${PATH_VAL/#~\//\/home\/${USER}\/}"
  ROOTFILE=$(echo $item | jq -r '.rootFile // empty')
  CONF_FILE="$CONF_DIR/$APPID.conf"

  if [ "$TYPE" = "static" ]; then
    TEMPLATE_FILE="$TEMPLATE_STATIC_FILE"
  else
    TEMPLATE_FILE="$TEMPLATE_SERVE_FILE"
  fi

  sed \
    -e "s/{{DOMAIN}}/$DOMAIN/g" \
    -e "s/{{PORT}}/$PORT/g" \
    -e "s/{{APPID}}/$APPID/g" \
    -e "s/{{ZONE_NAME}}/$ZONE_NAME/g" \
    -e "s/{{RATELIMIT}}/$RATELIMIT/g" \
    -e "s#{{PATH}}#$PATH_VAL#g" \
    -e "s#{{ROOTFILE}}#$ROOTFILE#g" \
    -e "s/{{IP_HOST}}/$IP_HOST/g" \
    "$TEMPLATE_FILE" > "$CONF_FILE"
done
 

echo "Nginx configs generated for all projects using template."
