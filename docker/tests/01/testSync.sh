#!/bin/bash

last_syncNum=0

response=$(curl -X GET http://127.0.0.1:8090/wallet/getnodeinfo)

# parse response
result=$(echo "$response" | jq -r '.beginSyncNum')
echo "result 1: $result, response 1: $response"

if [ ! -z "$result" ]; then
      last_sync=$(printf "%d" "$result")
      last_syncNum=$last_sync
      echo "TRON node first sync: $last_sync"

else
      alert="TRON node first sync error: : $response"
      echo $alert
      exit 1
fi

# interval as 10s
monitor_interval=10
# try 100 times
count=100
second_syncNum=0

check_sync_process() {
  response=$(curl -X GET http://127.0.0.1:8090/wallet/getnodeinfo)

  # parse response
  result=$(echo "$response" | jq -r '.beginSyncNum')
  echo "result 2: $result, response 2: $response"

  if [ ! -z "$result" ]; then
        last_sync=$(printf "%d" "$result")
        second_syncNum=$last_sync
        echo "TRON node second sync: $last_sync"
  else
        alert="TRON node second sync error: : $response"
        echo $alert
        exit 1
  fi
}

for((i=1;i<=$count;i++)); do
  echo "try i: $i"
  sleep $monitor_interval
  check_sync_process
  if [ $second_syncNum -gt $last_syncNum ]; then
      echo "sync increased"
      exit 0
  fi
done
echo "sync not increased"
exit 1