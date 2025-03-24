## trond docker

Commands for operating java-tron docker image.

### Synopsis

Commands used for operating docker image, such as:

	1. build java-tron docker image locally
	2. test the built image
	3. pull the latest image from officical docker hub

Please refer to the available commands below.


### Examples

```
# Help information for docker command
$ ./trond docker

# Build java-tron docker image, defualt: tronprotocol/java-tron:latest
$ ./trond docker build

# Build java-tron docker image with specified org, artifact and version
$ ./trond docker build -o tronprotocol -a java-tron -v latest

# Test java-tron docker image, defualt: tronprotocol/java-tron:latest
$ ./trond docker test

# Test java-tron docker image with specified org, artifact and version
$ ./trond docker test -o tronprotocol -a java-tron -v latest

```

### Options

```
  -h, --help   help for docker
```

### SEE ALSO

* [trond](trond.md)	 - Docker automation for TRON nodes
* [trond docker build](trond_docker_build.md)	 - Build java-tron docker image.
* [trond docker install-docker](trond_docker_install-docker.md)	 - Check and install docker and docker-compose (for Linux and Mac)
* [trond docker test](trond_docker_test.md)	 - Test java-tron docker image.
