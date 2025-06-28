#!/bin/bash

# Create the missing project that fern-ginkgo-client expects
echo "Creating project for fern-ginkgo-client..."

curl -X POST http://fern-platform.local:8080/api/project \
  -H "Content-Type: application/json" \
  -d '{
    "projectId": "ef79f688-3b3e-4746-aa14-a8e951f25756",
    "name": "Fern Reporter Test Project",
    "repository": "https://github.com/guidewire-oss/fern-reporter",
    "default_branch": "main"
  }' \
  --connect-timeout 10 \
  --max-time 30

echo ""
echo "Project creation request sent."