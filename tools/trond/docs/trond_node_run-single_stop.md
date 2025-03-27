## trond node run-single stop

Stop single java-tron node for different networks.

### Synopsis

The following files are required before stopping the node (these files will be need if you started java-tron fullnode using the default configuration file):

	- Docker compose file(by default, these exist in the current repository directory)
		main network: ./single_node/docker-compose.fullnode.main.yml
		nile network: ./single_node/docker-compose.fullnode.nile.yml
		private network: ./single_node/docker-compose.witness.private.yml

	- If you used a different docker compose file, you need to specify it by using the -f flags.

```
trond node run-single stop [flags]
```

### Examples

```
# Stop single java-tron fullnode for main network using the default docker compose file
# The default docker compose file is ./single_node/docker-compose.fullnode.main.yml
$ ./trond node run-single stop -t full-main

# Stop single java-tron fullnode for nile network using the default docker compose file
# The default docker compose file is ./single_node/docker-compose.fullnode.nile.yml
$ ./trond node run-single stop -t full-nile

# Stop single java-tron witness node for private network using the default docker compose file
# The default docker compose file is ./single_node/docker-compose.witness.private.yml
$ ./trond node run-single stop -t witness-private

# Stop single java-tron fullnode for main network using the specified docker compose file
# The docker compose file is ./docker-compose.fullnode.main.yml
$ ./trond node run-single stop -t full-main -f ./docker-compose.fullnode.main.yml

```

### Options

```
  -f, --compose-file string   The docker compose file you used to start the node before (optional, if not specified, the default file will be used)
  -h, --help                  help for stop
  -t, --type string           Node type you want to deploy (required, available: full-main, full-nile, witness-private)
```

### SEE ALSO

* [trond node run-single](trond_node_run-single.md)	 - Run single java-tron node for different networks.
