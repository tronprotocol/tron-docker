## trond docker

Commands for operating java-tron docker image.

### Synopsis

Commands used for operating docker image, such as:

	1. build java-tron docker image locally
	2. test the built image

Please refer to the available commands below.


### Examples

```
# Help information for docker command
$ ./trond docker

# Check and install docker and docker-compose (for Linux and Mac)
$ ./trond docker install-docker

# Please ensure that JDK 8 is installed, as it is required to execute the commands below.
# Build java-tron docker image, default output: tronprotocol/java-tron:latest
# Using code from https://github.com/tronprotocol/java-tron.git
$ ./trond docker build

# Build java-tron docker image with specified org, artifact and version
# Using code from https://github.com/tronprotocol/java-tron.git
$ ./trond docker build -o tronprotocol -a java-tron -v latest
$ ./trond docker build -o tronprotocol -a java-tron -v latest -n mainnet

# Build java-tron docker image for nile testnet with specified org, artifact and version
# Using code from https://github.com/tron-nile-testnet/nile-testnet.git
$ ./trond docker build -o tronnile -a java-tron -v latest -n nile

# Test java-tron docker image, defualt: tronprotocol/java-tron:latest
$ ./trond docker test

# Test java-tron docker image with specified org, artifact and version
$ ./trond docker test -o tronprotocol -a java-tron -v latest
$ ./trond docker test -o tronnile -a java-tron -v latest -n nile

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
