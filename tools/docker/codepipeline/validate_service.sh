#!/bin/bash

TARGET_DIR="/data/tron-node/tron-docker/"
DOCKER_COMPOSE_FILE="docker-compose.fullnode.mail.yml"

# API health check
max_attempts=30  # 5 minutes total (30 * 10 seconds)
attempt=1

while [ $attempt -le $max_attempts ]; do
    response=$(curl -s -X POST http://127.0.0.1:8090/wallet/getblock \
        -H 'Content-Type: application/json; charset=utf-8' \
        -d '{"detail":false}')

    if [ -n "$response" ]; then
        echo "Node API is responding successfully"
        exit 0
    fi

    echo "Attempt $attempt of $max_attempts: API not ready, retrying in 10 seconds..."
    sleep 10
    ((attempt++))
done

echo "Node API failed to respond after 5 minutes"

cd ${TARGET_DIR} || exit
./trond node run-single stop -t full-main -f ${DOCKER_COMPOSE_FILE}

exit 1
