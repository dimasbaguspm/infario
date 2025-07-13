# Nginx Dockerfile for automation setup
FROM nginx:stable-alpine

# Copy per-project configs
COPY conf.d/ /etc/nginx/conf.d/

# Copy SSL certificates
# SSL certificates are mounted from the host using a Docker volume (see docker-compose.yml)

# Expose HTTP and HTTPS ports
EXPOSE 80 443

# Entrypoint (default: nginx)
CMD ["nginx", "-g", "daemon off;"]
