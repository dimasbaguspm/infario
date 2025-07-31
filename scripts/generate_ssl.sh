#!/usr/bin/env bash

# Generate SSL certificates for all projects in constants.json using certbot

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
CONSTANTS_FILE="$SCRIPT_DIR/../constants.json"

if ! command -v certbot &> /dev/null; then
  echo "certbot not found. Please install certbot."
  exit 1
fi

jq -c '.[]' "$CONSTANTS_FILE" | while read -r item; do
  SUBDOMAIN=$(echo $item | jq -r '.subdomain')
  TYPE=$(echo $item | jq -r '.type')
  PATH_VAL=$(echo $item | jq -r '.path // empty')
  ROOTFILE=$(echo $item | jq -r '.rootFile // empty')
  if [ "$TYPE" = "static" ]; then
    if [ -z "$PATH_VAL" ] || [ -z "$ROOTFILE" ]; then
      echo "⚠️  Warning: Static site $SUBDOMAIN missing 'path' or 'rootFile' in constants.json. SSL will still be generated."
    fi
  fi
  sudo certbot certonly --standalone -d "$SUBDOMAIN" --non-interactive --agree-tos --register-unsafely-without-email
done

echo "SSL certs generated for all projects."
echo "\nRecommended: Use Certbot (Let's Encrypt) for production."
echo "If you haven't installed certbot, run:"
echo "  sudo apt update && sudo apt install certbot python3-certbot-nginx"
echo "\nMake sure your DNS A records point to your VPS IP for each subdomain."
echo "For automatic renewal, add this to your crontab:"
echo "  0 12 * * * /usr/bin/certbot renew --quiet"
echo "\nNginx should use certificates from /etc/letsencrypt/live/{domain}/fullchain.pem and privkey.pem."
