#!/usr/bin/env bash
set -e

# Compose and env file for production

COMPOSE_FILE="docker-compose.yml"

print_env_info() {
  echo "🚀 Infario Production Environment"
  echo "📄 Compose file: $COMPOSE_FILE"
  echo " Docker Compose command: docker compose -f $COMPOSE_FILE"
  echo ""
}

if [ "$1" != "" ] && [ "$1" != "--help" ] && [ "$1" != "-h" ]; then
  print_env_info
  if [ ! -f "$COMPOSE_FILE" ]; then
    echo "❌ Error: Compose file '$COMPOSE_FILE' not found!"
    exit 1
  fi
fi

case "$1" in
  up)
    echo "🚀 Starting services..."
    docker compose -f $COMPOSE_FILE up -d
    echo "🎉 Production environment started successfully!"
    ;;
  down)
    echo "🛑 Stopping services..."
    docker compose -f $COMPOSE_FILE down
    echo "✅ Services stopped."
    ;;
  clean-cache)
    echo "⚠️  WARNING: This will remove all containers, images, and volumes!"
    read -p "Are you sure you want to continue? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
      echo "❌ Clean cache cancelled"
      exit 1
    fi
    docker compose -f $COMPOSE_FILE down --volumes --remove-orphans
    docker compose -f $COMPOSE_FILE rm -f -v
    docker compose -f $COMPOSE_FILE config --images | while read -r image; do
      if [ -n "$image" ]; then
        docker rmi -f "$image" 2>/dev/null || echo "⚠️  Image $image not found or already removed"
      fi
    done
    docker image prune -f
    docker builder prune -f
    docker volume prune -f
    docker compose -f $COMPOSE_FILE build --no-cache
    echo "✅ Cache cleaned and all services rebuilt"
    ;;
  logs)
    SERVICE=${2:-""}
    if [ -n "$SERVICE" ]; then
      echo "📋 Showing logs for service: $SERVICE"
      docker compose -f $COMPOSE_FILE logs -f $SERVICE
    else
      echo "📋 Showing logs for all services"
      docker compose -f $COMPOSE_FILE logs -f
    fi
    ;;
  exec)
    SERVICE=${2:-""}
    shift 2
    if [ -z "$SERVICE" ]; then
      echo "❌ Error: Please specify a service to exec into."
      exit 1
    fi
    docker compose -f $COMPOSE_FILE exec $SERVICE "$@"
    ;;
  memory-monitor)
    echo "📊 Docker container memory usage (snapshot):"
    docker stats --no-stream
    echo "🖥️  System memory usage (free -h):"
    free -h
    ;;
  *)
    echo "Usage: $0 {up|down|clean-cache|logs [service]|exec <service> <cmd>|memory-monitor}"
    echo ""
    echo "Commands:"
    echo "  up                - Start all services (only needed for 'serve' type apps)"
    echo "  down              - Stop all services"
    echo "  clean-cache       - Remove containers/images/volumes and rebuild"
    echo "  logs [service]    - Show logs (all or specific service)"
    echo "  exec <service> <cmd> - Exec into a service container"
    echo "  memory-monitor    - Show memory usage for containers and system"
    echo ""
    echo "Note: Static sites (type=static) are served directly by Nginx and do not require a backend port or container."
    print_env_info
    exit 1
    ;;
esac
