#!/bin/bash

set -euo pipefail

go mod tidy

if ! git diff --quiet -- go.mod go.sum; then
    echo "Please run 'go mod tidy' and commit the changes."
    exit 1
fi

echo "go.mod and go.sum are up to date."
exit 0