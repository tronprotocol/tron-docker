## trond docker build

Build java-tron docker image.

### Synopsis

Build java-tron docker image locally. The master branch of java-tron repository will be built by default, using jdk1.8.0_202.


```
trond docker build [flags]
```

### Examples

```
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

```

### Options

```
  -a, --artifact string   ArtifactName for the docker image (default "java-tron")
  -h, --help              help for build
  -n, --network string    Which code will be used for the docker image, mainnet or nile (default "mainnet")
  -o, --org string        OrgName for the docker image (default "tronprotocol")
  -v, --version string    Release version for the docker image (default "latest")
```

### SEE ALSO

* [trond docker](trond_docker.md)	 - Commands for operating java-tron docker image.
