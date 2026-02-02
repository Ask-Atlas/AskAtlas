#!/bin/bash
set -eu

source "$(dirname "$0")/deploy-common.sh"

[[ -n "${3:-}" ]] || {
    echo "Error: Tag argument is missing"
    exit 1
}

ENV="$1"
APP="$2"
TAG="$3"

init_and_validate "$ENV" "$APP"

IMAGE="ghcr.io/ask-atlas/$APP:$TAG"
LOCAL_TAG="$APP:$ENV-latest"
PREVIOUS_TAG="$APP:$ENV-previous"

echo "Deploying $IMAGE to $ENV environment..."

echo "Logging in to GHCR..."
echo "$GHCR_TOKEN" | docker login ghcr.io -u "$GHCR_USERNAME" --password-stdin 2>/dev/null || echo "Already logged in"

echo "Pulling image..."
docker pull "$IMAGE"
docker image inspect "$LOCAL_TAG" &>/dev/null && {
    echo "Rotating tags: $LOCAL_TAG -> $PREVIOUS_TAG"
    docker tag "$LOCAL_TAG" "$PREVIOUS_TAG"
}
docker tag "$IMAGE" "$LOCAL_TAG"

deploy_container "$CONTAINER_NAME" "$IMAGE" "$ENV" || {
    echo "Deployment failed"
    exit 1
}

docker image prune -f --filter "dangling=true"
docker ps --filter name="$CONTAINER_NAME" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
docker stats "$CONTAINER_NAME" --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}"
echo "✓ Deployment complete!"

exit 0