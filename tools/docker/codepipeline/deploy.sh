#!/bin/bash
cd /home/ubuntu/tron-docker || exit
./trond node run-single stop -t witness-private
./trond node run-single -t witness-private
