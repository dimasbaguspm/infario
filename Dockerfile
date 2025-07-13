# Nginx Dockerfile for automation setup
FROM nginx:stable-alpine

# Copy main nginx config template (generated per-project configs are included via conf.d)
COPY conf.d/ /etc/nginx/conf.d/

# Copy SSL certificates
COPY ssl/ /etc/nginx/ssl/

# Expose HTTP and HTTPS ports
EXPOSE 80 443

# Entrypoint (default: nginx)
CMD ["nginx", "-g", "daemon off;"]
