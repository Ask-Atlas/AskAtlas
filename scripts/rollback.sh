#!/bin/bash
set -eu

ENV="$1"
APP="$2"

if [[ ! "$ENV" =~ ^(dev|stage|prod)$ ]]; then
    echo "Error: Environment must be dev, stage, or prod"
    exit 1
fi

if [[ ! "$APP" =~ ^(web|api)$ ]]; then
    echo "Error: App must be web or api"
    exit 1
fi


declare -A PORTS
PORTS["web-dev"]=3000
PORTS["web-stage"]=3001
PORTS["web-prod"]=3002
PORTS["api-dev"]=8080
PORTS["api-stage"]=8081
PORTS["api-prod"]=8082

declare -A MEM_LIMITS
MEM_LIMITS["dev"]="150m"
MEM_LIMITS["stage"]="200m"
MEM_LIMITS["prod"]="300m"

declare -A MEM_RESERVATIONS
MEM_RESERVATIONS["dev"]="96m"
MEM_RESERVATIONS["stage"]="128m"
MEM_RESERVATIONS["prod"]="200m"

PORT="${PORTS["$APP-$ENV"]}"
MEM_LIMIT="${MEM_LIMITS["$ENV"]}"
MEM_RESERVATION="${MEM_RESERVATIONS["$ENV"]}"
CONTAINER_NAME="$APP-$ENV"

if [[ -z "$PORT" ]]; then
    echo "Error: Configuration not found for $APP in $ENV"
    echo "Available keys: ${!PORTS[@]}"
    exit 1
fi

ROLLBACK_TAG="$APP:$ENV-previous"
LATEST_TAG="$APP:$ENV-latest"

if ! docker image inspect "$ROLLBACK_TAG" &>/dev/null; then
    echo "Rollback image not found ($ROLLBACK_TAG)"
    exit 1
fi

CURRENT_IMAGE=$(docker inspect --format='{{.Config.Image}}' "$CONTAINER_NAME" 2>/dev/null || echo "none")

echo "Rolling back $ENV environment..."
echo "Current image: $CURRENT_IMAGE"
echo "Rollback image: $ROLLBACK_TAG"

echo "Stopping container..."
docker stop "$CONTAINER_NAME" 2>/dev/null || echo "Container not running"
docker rm "$CONTAINER_NAME" 2>/dev/null || echo "Container does not exist"


echo "Starting rollback container..."
docker run -d \
    --name "$CONTAINER_NAME" \
    --restart unless-stopped \
    -p "$PORT:$PORT" \
    --memory="$MEM_LIMIT" \
    --memory-reservation="$MEM_RESERVATION" \
    --cpus="0.25" \
    --health-cmd="curl -f http://localhost:$PORT/health || exit 1" \
    --health-interval=30s \
    --health-timeout=3s \
    --health-start-period=10s \
    --health-retries=3 \
    -e PORT="$PORT" \
    "$ROLLBACK_TAG"

echo "Waiting for container to start..."
SECONDS=0
while [ $SECONDS -lt 30 ]; do
    HEALTH=$(docker inspect --format='{{.State.Health.Status}}' "$CONTAINER_NAME" 2>/dev/null || echo "starting")
    if [ "$HEALTH" = "healthy" ]; then
        echo "✓ Container is healthy!"
        echo "Restoring tags..."
        docker tag "$ROLLBACK_TAG" "$LATEST_TAG"
        break
    fi
    echo "Health status: $HEALTH (${SECONDS}s elapsed)"
    sleep 2
done


echo "Rolled back from: $CURRENT_IMAGE"
echo "Rolled back to:   $ROLLBACK_TAG"
echo "Tags restablized: $LATEST_TAG now points to $ROLLBACK_TAG"
docker ps --filter name="$CONTAINER_NAME" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
echo "✓ Rollback complete!"

exit 0