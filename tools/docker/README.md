# Gradle Docker
This document provides guidance on using Gradle to automate the build and test processes for java-tron docker image. You can customize the corresponding scripts to perform specific actions or integrate them with your existing Continuous Integration (CI) setup.

If you encounter any problems during the build or testing process, please refer to the troubleshooting section.

## Prerequisites

- Docker
- JDK 8 (required by Gradle)

Follow the [getting-started](https://github.com/tronprotocol/tron-docker/blob/main/README.md#getting-started) guide to download Docker and the tron-docker repository. Then, navigate to the gradlew directory.
```
cd ./tools/gradlew
```
Container testing uses the [Goss](https://github.com/goss-org/goss/blob/v0.4.9/README.md) tool, a YAML-based testing framework for validating service states. While no additional installation is required, it is beneficial to learn the basic usage of [Goss with container](https://goss.readthedocs.io/en/stable/container_image/) to help you better understand the following testing scripts.

## Build image

The command below will trigger the build process for java-tron image. Now we only support platform:`linux/amd64`.
```
./gradlew --no-daemon sourceDocker
```

The compilation process may take above 30 minutes, depending on your network conditions. Once it successfully completes, you will be able to see the generated image.
```
$ docker images
REPOSITORY               TAG        IMAGE ID       CREATED          SIZE
tronprotocol/java-tron   1.0.0      76702facd55e   23 seconds ago   549MB
```

### What ./gradlew sourceDocker do?

It will trigger the execution of `task sourceDocker` in [build.gradle](build.gradle). Reviewing the logic, `task sourceDocker` essentially copies the Dockerfile and shell script to a build directory, then runs the docker build. From the logic, you can run `./gradlew sourceDocker` with customised `dockerOrgName`, `dockerArtifactName`, and `release.releaseVersion`.

For example:
```
./gradlew --no-daemon sourceDocker -PdockerOrgName=yourOrgName -PdockerArtifactName=test -Prelease.releaseVersion=V1.1.0
```
## Test image

Test the java-tron image use the command below.
```
./gradlew --no-daemon testDocker
```
This will trigger the execution of `task testDocker` in [build.gradle](build.gradle). According to this logic, it will run the [test.sh](test.sh) script with the parameter of the Docker image name.

The test.sh script sets up the Goss environment and then invokes the [dgoss shell](tests/dgoss), which will copy the test cases from [tests/01](tests/01) folder to the Docker container and execute the test validations.

Currently, there are three test files:

- The `goss.yaml` and `goss_wait.yaml` files are used to perform port checks.

- The `testSync.sh` script is used to verify whether block synchronization is functioning normally. It will call `http://127.0.0.1:8090/wallet/getnodeinfo` 100 times. As long as `beginSyncNum` changes from the genesis block 0 to a larger number, the test will pass.

Successful execution will output the following content:
```
$ ./gradlew --no-daemon testDocker

To honour the JVM settings for this build a single-use Daemon process will be forked. See https://docs.gradle.org/7.6.4/userguide/gradle_daemon.html#sec:disabling_the_daemon.
Daemon will be stopped at the end of the build

> Task :docker:sourceDocker
Building for default linux/amd64 platform
#0 building with "desktop-linux" instance using docker driver

#1 [internal] load build definition from Dockerfile
#1 transferring dockerfile: 2.32kB done
#1 DONE 0.0s

#2 [internal] load metadata for docker.io/library/ubuntu:24.04
#2 DONE 3.9s

#3 [internal] load .dockerignore
#3 transferring context: 2B done
#3 DONE 0.0s

#4 [1/5] FROM docker.io/library/ubuntu:24.04@sha256:80dd3c3b9c6cecb9f1667e9290b3bc61b78c2678c02cbdae5f0fea92cc6734ab
#4 DONE 0.0s

#5 [internal] load build context
#5 transferring context: 160B done
#5 DONE 0.0s

#6 [2/5] RUN apt-get update -o Acquire::BrokenProxy=true -o Acquire::http::No-Cache=true -o Acquire::http::Pipeline-Depth=0  &&   apt-get --quiet --yes install git wget 7zip curl jq &&   wget -P /usr/local https://github.com/frekele/oracle-java/releases/download/8u202-b08/jdk-8u202-linux-x64.tar.gz   && echo "0029351f7a946f6c05b582100c7d45b7 /usr/local/jdk-8u202-linux-x64.tar.gz" | md5sum -c   && tar -zxf /usr/local/jdk-8u202-linux-x64.tar.gz -C /usr/local  && rm /usr/local/jdk-8u202-linux-x64.tar.gz   && export JAVA_HOME=/usr/local/jdk1.8.0_202   && export CLASSPATH=$JAVA_HOME/lib/dt.jar:$JAVA_HOME/lib/tools.jar   && export PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:$JAVA_HOME/bin   && echo "git clone"   && mkdir -p /tron-build   && cd /tron-build   && git clone https://github.com/tronprotocol/java-tron.git   && cd java-tron   && git checkout master   && ./gradlew build -x test   && cd build/distributions   && 7z x -y java-tron-1.0.0.zip    && mv java-tron-1.0.0 /java-tron   && rm -rf /tron-build   && rm -rf ~/.gradle   && mv /usr/local/jdk1.8.0_202/jre /usr/local   && rm -rf /usr/local/jdk1.8.0_202   apt-get clean  &&   rm -rf /var/cache/apt/archives/* /var/cache/apt/archives/partial/*  &&   rm -rf /var/lib/apt/lists/*
#6 CACHED

#7 [3/5] RUN wget -P /java-tron/config https://raw.githubusercontent.com/tronprotocol/tron-deployment/master/main_net_config.conf
#7 CACHED

#8 [4/5] COPY docker-entrypoint.sh /java-tron/bin
#8 CACHED

#9 [5/5] WORKDIR /java-tron
#9 CACHED

#10 exporting to image
#10 exporting layers done
#10 writing image sha256:da229c38032f4c49c8199dcb9f1632f3312ee89b1d8d75f08700f43d35d6fd13 done
#10 naming to docker.io/tronprotocol/java-tron:1.0.0 done
#10 DONE 0.0s

View build details: docker-desktop://dashboard/build/desktop-linux/desktop-linux/gyq5fuhe324kx4qd9judic3jl

> Task :docker:testDocker
INFO: Run Docker tests
INFO: Setting up test dir
INFO: Setup complete
INFO: Creating docker container
WARNING: The requested image's platform (linux/amd64) does not match the detected host platform (linux/arm64/v8) and no specific platform was requested
INFO: Copy goss files into container
INFO: Starting docker container
INFO: Container ID: 080c92c3e40b575999df863d05eb0ef85899477b1188951fd4506697372d70c1
INFO: Found goss_wait.yaml, waiting for it to pass before running tests
INFO: Sleeping for 1.0
INFO: Running Tests
<?xml version="1.0" encoding="UTF-8"?>
<testsuite name="goss" errors="0" tests="1" failures="0" skipped="0" time="0.000" timestamp="2025-01-23T06:25:42Z">
<testcase name="Process java running" time="0.000">
<system-out>Process: java: running: matches expectation: true</system-out>
</testcase>
</testsuite>

BUILD SUCCESSFUL in 47s
2 actionable tasks: 2 executed
```
It will automatically remove the test container it starts. If, for any reason, this process is manually killed, you can find the container ID from the log message "Container ID: xxxxxx" and manually remove the test container using `docker rm -f xxxxxx`.

## Troubleshooting

If you encounter the following errors while building the image:

```
800.3 Cloning into 'java-tron'...
1187.3 error: RPC failed; curl 92 HTTP/2 stream 5 was not closed cleanly: CANCEL (err 8)
1187.3 error: xxxx bytes of body are still expected
1187.3 fetch-pack: unexpected disconnect while reading sideband packet
1187.3 fatal: early EOF
1187.3 fatal: fetch-pack: invalid index-pack output
```

Adjust your Git HTTP post buffer to a larger size by using the command below:
```
git config --global http.postBuffer 5242880000  # Sets it to 5GB
```

For other issues, they may be caused by poor network conditions. You could try switching to a different network or VPN. If the issue still cannot be resolved, please refer to [Issue Work Flow](https://tronprotocol.github.io/documentation-en/developers/issue-workflow/#issue-work-flow), then raise issue in [Github](https://github.com/tronprotocol/tron-docker/issues).
