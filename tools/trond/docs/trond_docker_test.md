## trond docker test

Test java-tron docker image.

### Synopsis

Test java-tron docker image locally. If no flags are provided, "tronprotocol/java-tron:latest" image will be tested.

The test includes the following tasks:

	1. Perform port checks
	2. Verify whether block synchronization is functioning normally
	3. Automatically remove the test container it starts. If, for any reason, this process is manually killed, you can find the container ID from the log message "Container ID: xxxxxx" and manually remove the test container using "docker rm -f xxxxxx".

It is used to test TRON Mainnet block synchronization within the docker container. It has a timeout limit of 1000 seconds and may encounter the error “ERROR: goss_wait.yaml never passed” due to network issues or timeouts. If this happens, please check your network status and try again.


```
trond docker test [flags]
```

### Examples

```
# Build java-tron docker image, defualt: tronprotocol/java-tron:latest
$ ./trond docker test

# Build java-tron docker image with specified org, artifact and version
$ ./trond docker test -o tronprotocol -a java-tron -v latest

```

### Options

```
  -a, --artifact string   ArtifactName for the docker image (default "java-tron")
  -h, --help              help for test
  -o, --org string        OrgName for the docker image (default "tronprotocol")
  -v, --version string    Release version for the docker image (default "latest")
```

### SEE ALSO

* [trond docker](trond_docker.md)	 - Commands for operating java-tron docker image.
