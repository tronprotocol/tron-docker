#!/bin/bash

TARGET_DIR="/home/ubuntu/yk/mydev/tron-docker/"
DOCKER_COMPOSE_FILE="docker-compose.fullnode.nile.yml"

cp ${DOCKER_COMPOSE_FILE} ${TARGET_DIR}
cd ${TARGET_DIR} || exit
./trond node run-single -t full-nile -f ${DOCKER_COMPOSE_FILE}

# Check if container is running
max_attempts=12  # 2 minutes total (12 * 10 seconds)
attempt=1

while [ $attempt -le $max_attempts ]; do
    if docker-compose -f ${DOCKER_COMPOSE_FILE} ps --services --filter "status=running" | grep -q "tron-node-nile"; then
        echo "Container is running, proceeding to API check"
        break
    fi
    echo "Attempt $attempt of $max_attempts: Container not running yet, waiting 10 seconds..."
    sleep 10
    ((attempt++))
done

if [ $attempt -gt $max_attempts ]; then
    echo "Container failed to start after 2 minutes"
    exit 1
fi
