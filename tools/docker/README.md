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

The command below builds the java-tron image for `linux/amd64` by default. Use `-Pplatform=linux/arm64` to build or test the ARM64 image.
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

It will trigger the execution of `task sourceDocker` in [build.gradle](build.gradle). Reviewing the logic, `task sourceDocker` essentially copies the Dockerfile and shell script to a build directory, then runs the docker build. From the logic, you can run `./gradlew sourceDocker` with customised `dockerOrgName`, `dockerArtifactName`, `release.releaseVersion` and `network`.

For example:
```sh
# build java-tron image with customised dockerOrgName, dockerArtifactName and release.releaseVersion
./gradlew --no-daemon sourceDocker -PdockerOrgName=YourOrgName -PdockerArtifactName=ArtifactName -Prelease.releaseVersion=V1.1.0

# build java-tron docker image for nile testnet with specified org, artifact and version.
# Using code from https://github.com/tron-nile-testnet/nile-testnet.git
./gradlew --no-daemon sourceDocker -PdockerOrgName=YourOrgName -PdockerArtifactName=ArtifactName -Prelease.releaseVersion=V1.1.0 -Pnetwork=nile
```

## Test image

Test the java-tron image use the command below.
``` ssh
./gradlew --no-daemon testDocker

# test image with customised dockerOrgName, dockerArtifactName and release.releaseVersion
./gradlew --no-daemon testDocker -PdockerOrgName=YourOrgName -PdockerArtifactName=ArtifactName -Prelease.releaseVersion=V1.1.0

# test image for nile testnet with specified org, artifact and version.
./gradlew --no-daemon testDocker -PdockerOrgName=YourOrgName -PdockerArtifactName=ArtifactName -Prelease.releaseVersion=V1.1.0 -Pnetwork=nile
```
This will trigger the execution of `task testDocker` in [build.gradle](build.gradle). According to this logic, it will run the [test.sh](test.sh) script with the parameter of the Docker image name.

Goss runs in the node container and checks the listening ports and Java process. Once the ports are ready, the runner starts a temporary container from the pinned official Python Alpine image. The first test run needs access to pull this image. The two containers share a network namespace, so the probe can query the node on loopback without publishing a host port.

The [sync probe](tests/01/testSync.py) reads `/wallet/getnodeinfo` every 10 seconds. It passes when `beginSyncNum` exceeds its first valid reading and fails after 1000 seconds without an increase. The runner removes both containers when the test exits. Sync progress appears in the test log and the final Goss JUnit report is written to `tools/docker/reports/01.xml`.

If the runner is killed before cleanup, use the container IDs in the test log to remove the probe first and then the node with `docker rm -f`.

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
