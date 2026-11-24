## trond docker build

Build java-tron docker image.

### Synopsis

Build the java-tron Docker image from the master branch. The Dockerfile pins OpenJDK 8 for amd64 and OpenJDK 17 for arm64.
Use --openjdk-version or set OPENJDK_VERSION to override the Ubuntu package version for one build. The flag takes precedence.


```
trond docker build [flags]
```

### Examples

```
# Please ensure that JDK 8 is installed, as it is required to execute the commands below.
# Build java-tron docker image, default output: tronprotocol/java-tron:latest
# Using code from https://github.com/tronprotocol/java-tron.git
$ ./trond docker build

# Build java-tron docker image with specified org, artifact and version, targeting the amd64 platform
# Using code from https://github.com/tronprotocol/java-tron.git
$ ./trond docker build -o tronprotocol -a java-tron -v latest
$ ./trond docker build -o tronprotocol -a java-tron -v latest -n mainnet

# Build java-tron docker image for nile testnet with specified org, artifact and version, targeting the amd64 platform
# Using code from https://github.com/tron-nile-testnet/nile-testnet.git
$ ./trond docker build -o tronnile -a java-tron -v latest -n nile

# Build java-tron docker image with specified org, artifact and version, targeting the arm64 platform
# Using code from https://github.com/tronprotocol/java-tron.git
$ ./trond docker build -o tronprotocol -a java-tron -v latest
$ ./trond docker build -o tronprotocol -a java-tron -v latest -n mainnet -p linux/arm64

# Override the Ubuntu OpenJDK package version for an arm64 build
$ ./trond docker build -p linux/arm64 --openjdk-version '<full Ubuntu Java 17 package version>'

# Build java-tron docker image for nile testnet with specified org, artifact and version, targeting the arm64 platform
# Using code from https://github.com/tron-nile-testnet/nile-testnet.git
$ ./trond docker build -o tronnile -a java-tron -v latest -n nile -p linux/arm64

```

### Options

```
  -a, --artifact string          ArtifactName for the docker image (default "java-tron")
  -h, --help                     help for build
  -n, --network string           Which code will be used for the docker image, mainnet or nile (default "mainnet")
      --openjdk-version string   Full Ubuntu OpenJDK package version for the target architecture
  -o, --org string               OrgName for the docker image (default "tronprotocol")
  -p, --platform string          Platform for the docker image, linux/amd64 or linux/arm64 (default "linux/amd64")
  -v, --version string           Release version for the docker image (default "latest")
```

### SEE ALSO

* [trond docker](trond_docker.md)	 - Commands for operating java-tron docker image.
