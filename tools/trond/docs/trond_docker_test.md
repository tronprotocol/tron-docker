## trond docker test

Test java-tron docker image.

### Synopsis

Test java-tron docker image locally. If no flags are provided, "tronprotocol/java-tron:latest" image will be tested.

The test includes the following tasks:

	1. Perform port checks
	2. Verify whether block synchronization is functioning normally
	3. Remove the sync probe and node containers on exit. If the runner is killed before cleanup, use their logged IDs to remove the probe first and then the node with "docker rm -f".

A temporary Python container shares the node's network and checks for an increase in block height. The sync check times out after 1000 seconds. If it times out, check the node's peer connections.


```
trond docker test [flags]
```

### Examples

```
# Test java-tron docker image, defualt: tronprotocol/java-tron:latest
$ ./trond docker test

# Test java-tron docker image for amd64 with specified org, artifact and version
$ ./trond docker test -o tronprotocol -a java-tron -v latest
$ ./trond docker test -o tronnile -a java-tron -v latest -n nile

# Test java-tron docker image for arm64 with specified org, artifact and version
$ ./trond docker test -o tronprotocol -a java-tron -v latest -p linux/arm64
$ ./trond docker test -o tronnile -a java-tron -v latest -n nile -p linux/arm64

```

### Options

```
  -a, --artifact string   ArtifactName for the docker image (default "java-tron")
  -h, --help              help for test
  -n, --network string    Which code will be used for the docker image, mainnet or nile (default "mainnet")
  -o, --org string        OrgName for the docker image (default "tronprotocol")
  -p, --platform string   Platform for the docker image, linux/amd64 or linux/arm64 (default "linux/amd64")
  -v, --version string    Release version for the docker image (default "latest")
```

### SEE ALSO

* [trond docker](trond_docker.md)	 - Commands for operating java-tron docker image.
