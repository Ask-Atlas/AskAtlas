#!/bin/bash
set -eu

ENV="$1"
APP="$2"
TAG="$3"

if [[ ! "$ENV" =~ ^(dev|stage|prod)$ ]]; then
    echo "Script missing environment argument (dev, stage, or prod)"
    exit 1
fi

if [[ ! "$APP" =~ ^(web|api)$ ]]; then
    echo "Script missing app argument (web or api)"
    exit 1
fi

if [[ -z "$TAG" ]]; then
    echo "Script missing tag argument"
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
    exit 1
fi

IMAGE="ghcr.io/askatlas/$APP:$TAG"
LOCAL_TAG="$APP:$ENV-latest"
PREVIOUS_TAG="$APP:$ENV-previous"

echo "Deploying $IMAGE to $ENV environment..."

echo "Logging in to GHCR..."
echo "$GHCR_TOKEN" | docker login ghcr.io -u "$GHCR_USERNAME" --password-stdin 2>/dev/null || echo "Already logged in"

echo "Pulling image..."
docker pull "$IMAGE"

if docker image inspect "$LOCAL_TAG" &>/dev/null; then
    echo "Rotating tags: $LOCAL_TAG -> $PREVIOUS_TAG"
    docker tag "$LOCAL_TAG" "$PREVIOUS_TAG"
fi

docker tag "$IMAGE" "$LOCAL_TAG"

echo "Stopping old container..."
docker stop "$CONTAINER_NAME" 2>/dev/null || true
docker rm "$CONTAINER_NAME" 2>/dev/null || true

echo "Starting new container..."
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
    "$IMAGE"

echo "Waiting for health check..."
SECONDS=0
while [ $SECONDS -lt 30 ]; do
    HEALTH=$(docker inspect --format='{{.State.Health.Status}}' "$CONTAINER_NAME" 2>/dev/null || echo "starting")
    if [ "$HEALTH" = "healthy" ]; then
        echo "✓ Container is healthy!"
        break
    fi
    echo "Health status: $HEALTH (${SECONDS}s elapsed)"
    sleep 2
done

echo "Deployment Status:"
docker ps --filter name="$CONTAINER_NAME" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
echo "Resource Usage:"
docker stats "$CONTAINER_NAME" --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}"
echo "✓ Deployment complete!"

exit 0