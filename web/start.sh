#!/bin/sh
# from https://infisical.com/docs/integrations/platforms/docker
export INFISICAL_TOKEN=$(infisical login --method=universal-auth --client-id=$INFISICAL_MACHINE_CLIENT_ID --client-secret=$INFISICAL_MACHINE_CLIENT_SECRET --plain --silent)
exec infisical run --token $INFISICAL_TOKEN --projectId $PROJECT_ID --env $INFISICAL_SECRET_ENV -- node server.js