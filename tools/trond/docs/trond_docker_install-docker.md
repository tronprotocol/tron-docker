## trond docker install-docker

Check and install docker and docker-compose (for Linux and Mac)

### Synopsis

Check docker and docker-compose installation on Linux and Mac. If docker and docker-compose are not installed, this command will install them for you.

For Mac: brew install --cask docker

For Linux: sh get-docker.sh


```
trond docker install-docker [flags]
```

### Examples

```
# Check and install docker and docker-compose (for Linux and Mac)
$ ./trond docker install-docker

```

### Options

```
  -a, --artifact string   ArtifactName for the docker image (default "java-tron")
  -h, --help              help for install-docker
  -o, --org string        OrgName for the docker image (default "tronprotocol")
  -v, --version string    Release version for the docker image (default "latest")
```

### SEE ALSO

* [trond docker](trond_docker.md)	 - Commands for operating java-tron docker image.
