#!/bin/bash

init_and_validate() {
    local env="$1"
    local app="$2"

    [[ "$env" =~ ^(dev|stage|prod)$ ]] || {
        echo "Error: Environment must be dev, stage, or prod"
        exit 1
    }

    [[ "$app" =~ ^(web|api)$ ]] || {
        echo "Error: App must be web or api"
        exit 1
    }

    declare -gA PORTS
    PORTS=(
        [web-dev]=3000
        [web-stage]=3001
        [web-prod]=3002
        [api-dev]=8080
        [api-stage]=8081
        [api-prod]=8082
    )

    declare -gA MEM_LIMITS
    MEM_LIMITS=(
        [dev]=150m
        [stage]=200m
        [prod]=300m
    )

    declare -gA MEM_RESERVATIONS
    MEM_RESERVATIONS=(
        [dev]=96m
        [stage]=128m
        [prod]=200m
    )

    PORT=${PORTS["$app-$env"]}
    MEM_LIMIT=${MEM_LIMITS["$env"]}
    MEM_RESERVATION=${MEM_RESERVATIONS["$env"]}
    CONTAINER_NAME="$app-$env"

    if [[ -z "$PORT" ]]; then
        echo "Error: Configuration not found for $APP in $ENV"
        exit 1
    fi
}

deploy_container() {
    local container_name="$1"
    local image="$2"
    local env="$3"

    local -A INFISICAL_ENVS
    INFISICAL_ENVS=(
        [dev]="development"
        [stage]="staging"
        [prod]="production"
    )

    docker stop "$container_name" 2>/dev/null || echo "Container not running"
    docker rm "$container_name" 2>/dev/null || echo "Container does not exist"

    docker run -d \
        --name "$container_name" \
        --restart unless-stopped \
        -p "$PORT:$PORT" \
        --memory="$MEM_LIMIT" \
        --memory-reservation="$MEM_RESERVATION" \
        -e PORT="$PORT" \
        -e INFISICAL_MACHINE_CLIENT_ID="$INFISICAL_CLIENT_ID" \
        -e INFISICAL_MACHINE_CLIENT_SECRET="$INFISICAL_CLIENT_SECRET" \
        -e PROJECT_ID="$INFISICAL_PROJECT_ID" \
        -e INFISICAL_SECRET_ENV="${INFISICAL_ENVS["$env"]}" \
        "$image"

    echo "Waiting for container to start..."
    SECONDS=0
    TIMEOUT=300

    while [ $SECONDS -lt $TIMEOUT ]; do
        STATUS=$(docker inspect --format='{{.State.Status}}' "$container_name" 2>/dev/null || echo "unknown")

        case "$STATUS" in
            running)
                echo "Container is running!"
                return 0
                ;;
            exited|dead)
                echo "Container exited with code: $(docker inspect --format='{{.State.ExitCode}}' "$container_name")"
                docker logs --tail 20 "$container_name"
                docker stop "$container_name"
                docker rm "$container_name"
                return 1
                ;;
            *)  
                echo "Status: $STATUS (${SECONDS}s elapsed)"
                sleep 5
                ;;
        esac
    done

    echo "✗ Timed out waiting for container to start ($TIMEOUT seconds)"
    docker logs --tail 20 "$container_name"
    docker stop "$container_name"
    docker rm "$container_name"
    return 1
}