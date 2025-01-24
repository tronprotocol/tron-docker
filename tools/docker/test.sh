#!/bin/bash
##
## Copyright contributors to Besu.
##
## Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
## the License. You may obtain a copy of the License at
##
## http://www.apache.org/licenses/LICENSE-2.0
##
## Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
## an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
## specific language governing permissions and limitations under the License.
##
## SPDX-License-Identifier: Apache-2.0
##

export TEST_PATH=./tests
#export GOSS_PATH=$TEST_PATH/goss-linux-${architecture} # TODO. fixed by https://github.com/goss-org/goss/tree/master/extras/dgoss#mac-osx
export GOSS_PATH=$TEST_PATH/goss-linux-amd64
export GOSS_OPTS="$GOSS_OPTS --format junit"
export GOSS_FILES_STRATEGY=cp
DOCKER_IMAGE=$1
DOCKER_FILE="${2:-$PWD/Dockerfile}"

i=0

# Test for normal startup with ports opened
# we test that things listen on the right interface/port, not what interface the advertise
# hence we dont set p2p-host=0.0.0.0 because this sets what its advertising to devp2p; the important piece is that it defaults to listening on all interfaces
GOSS_FILES_PATH=$TEST_PATH/01 \
bash $TEST_PATH/dgoss run --sysctl net.ipv6.conf.all.disable_ipv6=1 $DOCKER_IMAGE \
#-p 8090:8090 -p 8091:8091 -p 18888:18888 -p 18888:18888/udp -p 50051:50051 \
> ./reports/01.xml || i=`expr $i + 1`

exit $i