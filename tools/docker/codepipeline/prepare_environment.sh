#!/bin/bash

TARGET_DIR="/data/tron-node/tron-docker/"
DOCKER_COMPOSE_FILE="docker-compose.fullnode.mail.yml"

cd ${TARGET_DIR} || exit
./trond node run-single stop -t full-main -f ${DOCKER_COMPOSE_FILE}

# Wait and check if container is stopped
max_attempts=30  # 2 minutes total (30 * 10 seconds)
attempt=1

while [ $attempt -le $max_attempts ]; do
    if ! docker-compose -f ${DOCKER_COMPOSE_FILE} ps --services --filter "status=running" | grep -q "tron-node-mainnet"; then
        echo "Container successfully stopped"
        exit 0
    fi
    echo "Attempt $attempt of $max_attempts: Container still running, waiting 10 seconds..."
    sleep 10
    ((attempt++))
done

echo "Failed to stop container after 2 minutes"
exit 1
