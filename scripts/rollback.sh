#!/bin/bash
set -eu

source "$(dirname "$0")/deploy-common.sh"

ENV="$1"
APP="$2"

init_and_validate "$ENV" "$APP"

ROLLBACK_TAG="$APP:$ENV-previous"
LATEST_TAG="$APP:$ENV-latest"

docker image inspect "$ROLLBACK_TAG" &>/dev/null || {
    echo "Rollback image not found ($ROLLBACK_TAG)"
    exit 1
}

CURRENT_IMAGE=$(docker inspect --format='{{.Config.Image}}' "$CONTAINER_NAME" 2>/dev/null || echo "none")

echo "Rolling back $ENV environment..."
echo "Current image: $CURRENT_IMAGE"
echo "Rollback image: $ROLLBACK_TAG"

deploy_container "$CONTAINER_NAME" "$ROLLBACK_TAG" || {
    echo "Rollback failed"
    exit 1
}

docker tag "$ROLLBACK_TAG" "$LATEST_TAG"
docker ps --filter name="$CONTAINER_NAME" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
docker stats "$CONTAINER_NAME" --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}"
echo "✓ Rollback complete!"

exit 0