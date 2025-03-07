#!/bin/bash
cd /home/ubuntu/tron-docker
./trond node run-single stop -t witness-private
./trond node run-single -t witness-private
