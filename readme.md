
# Infario (Personal Deployment)

Infario is my personal deployment for the "personal-prod" environment (by dimas).


This project automates the setup and management of Nginx for multiple projects, including:
- Forwarded port configuration
- Proxy rules
- SSL certificate management
- Automated Nginx reloads


## Domains
List of domains managed by Infario (see `constants.json` for details):
- [versaur](versaur.dimasbaguspm.com)
- [spenicle](spenicle.dimasbaguspm.com)
- [spenicle-api](spenicle-api.dimasbaguspm.com)
  
Each project's `subdomain` is defined in `constants.json`.



## Structure
- `templates/nginx.conf.serve.template`: Nginx config template for reverse-proxy (type=serve) projects
- `templates/nginx.conf.static.template`: Nginx config template for static (type=static) projects, with SPA routing support
- `conf.d/`: Per-project Nginx configs (e.g., `project1.conf`)
- `scripts/`: Automation scripts for config and SSL generation, reload
- `ssl/`: SSL certificates and keys (mounted to `/etc/nginx/ssl` in container)
- `constants.json`: Centralized project definitions (subdomain, port, appId, rateLimitPerSecond)

## Usage Flow

1. **Define Projects**
   - Add your projects to `constants.json` with `subdomain`, `port`, `appId`, and optionally `rateLimitPerSecond` for custom rate limiting.
   - Ensure each domain/subdomain is correctly listed and DNS records point to your VPS.

2. **Generate SSL Certificates**
   - Run:
     ```bash
     sudo ./scripts/generate_ssl.sh
     ```
   - This uses certbot to generate SSL certs for each subdomain (see `generate_ssl.sh`). Certs are placed in `/etc/nginx/ssl/{appId}/`.
   - Make sure certbot is installed and DNS is set up before running.

3. **Generate Nginx Configs**
   - Run:
     ```bash
     ./scripts/generate_nginx_conf.sh
     ```
   - This generates a config for each project in `conf.d/{appId}.conf` using the template and your project info, including unique rate limiting per subdomain (see `generate_nginx_conf.sh`).

4. **Check Assigned Ports**
   - Run:
     ```bash
     ./scripts/check_ports.sh
     ```
   - This verifies that each port in `constants.json` is running before starting Nginx.

5. **Build and Run Nginx in Docker**
   - Build and start the container:
     ```bash
     docker compose up --build -d
     ```
   - This uses your generated configs, SSL certs, and mounts the `/deployments` directory for static sites.
   - Any changes to files in `/deployments` on the host are immediately reflected inside the container (no restart needed for static file updates).


## Example
- See `constants.json` for a sample project definition, including `rateLimitPerSecond`.
- See `conf.d/` for generated configs.
- SSL certs are stored in `/etc/nginx/ssl/{appId}/`.
- Static deployments are stored in `/deployments/{your-app}/` and are mounted read-only into the Nginx container.

## Notes
- Make sure your DNS records point to your VPS for each subdomain.
- For automatic SSL renewal, add to crontab:
  ```bash
  0 12 * * * /usr/bin/certbot renew --quiet
  ```
- All static deployments in `/deployments` (and subdirectories) are accessible to Nginx. No need to add extra mounts for subfolders.
- Ensure permissions for `/deployments` and its contents allow world-read (directories: 755, files: 644) so Nginx can serve them.
