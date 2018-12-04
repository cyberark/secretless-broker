#!/bin/bash

set -euo pipefail

APPLICATION_URL=$(. ./admin_config.sh; echo ${APPLICATION_URL})

echo "Adding a pet..."
curl \
  -i \
  -d '{"name": "Mr. Snuggles"}' \
  -H "Content-Type: application/json" \
  ${APPLICATION_URL}/pet

echo "Checking the pets..."
curl -i ${APPLICATION_URL}/pets
