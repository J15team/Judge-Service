#!/bin/sh
set -e

# Start Docker daemon
dockerd-entrypoint.sh &

# Wait for Docker to be ready
echo "Waiting for Docker daemon..."
while ! docker info >/dev/null 2>&1; do
    sleep 1
done
echo "Docker daemon is ready"

# Build sandbox image
echo "Building sandbox image..."
docker build -t judge-sandbox:latest /sandbox

echo "Starting judge-service..."
exec "$@"
