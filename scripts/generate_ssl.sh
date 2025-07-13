#!/usr/bin/env bash

# Generate SSL certificates for all projects in constants.json using certbot

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
CONSTANTS_FILE="$SCRIPT_DIR/../constants.json"
SSL_DIR="/etc/nginx/ssl"

if ! command -v certbot &> /dev/null; then
  echo "certbot not found. Please install certbot."
  exit 1
fi

jq -c '.[]' "$CONSTANTS_FILE" | while read -r item; do
  SUBDOMAIN=$(echo $item | jq -r '.subdomain')
  APPID=$(echo $item | jq -r '.appId')

  sudo mkdir -p "$SSL_DIR/$APPID"
  sudo certbot certonly --standalone -d "$SUBDOMAIN" --non-interactive --agree-tos --register-unsafely-without-email --key-path "$SSL_DIR/$APPID/privkey.pem" --fullchain-path "$SSL_DIR/$APPID/fullchain.pem"
done

echo "SSL certs generated for all projects."
echo "\nRecommended: Use Certbot (Let's Encrypt) for production."
echo "If you haven't installed certbot, run:"
echo "  sudo apt update && sudo apt install certbot python3-certbot-nginx"
echo "\nMake sure your DNS A records point to your VPS IP for each subdomain."
echo "For automatic renewal, add this to your crontab:"
echo "  0 12 * * * /usr/bin/certbot renew --quiet"
