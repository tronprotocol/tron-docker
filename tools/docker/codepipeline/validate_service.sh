#!/bin/bash

# API health check
max_attempts=30  # 5 minutes total (30 * 10 seconds)
attempt=1

while [ $attempt -le $max_attempts ]; do
    response=$(curl -s -X POST http://127.0.0.1:8090/wallet/getblock \
        -H 'Content-Type: application/json; charset=utf-8' \
        -H 'Host: api.trongrid.io' \
        -H 'User-Agent: RapidAPI/4.2.8 (Macintosh; OS X/15.3.1) GCDHTTPRequest' \
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
exit 1
