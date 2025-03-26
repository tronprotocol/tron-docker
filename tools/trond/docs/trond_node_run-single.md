## trond node run-single

Run single java-tron node for different networks.

### Synopsis

You need to make sure the local environment is ready before running the node. Run "./trond node env" to check the environment before starting the node.

The following files are required before starting the node (these files will be need if you want to start any type of java-tron fullnode using the default configuration file):

	- Configuration file(by default, these exist in the current repository directory)
		main network: ./conf/main_net_config.conf
		nile network: ./conf/nile_net_config.conf
		private network: ./conf/private_net_config_*.conf
	- Docker compose file(by default, these exist in the current repository directory)
		main network: ./single_node/docker-compose.fullnode.main.yml
		nile network: ./single_node/docker-compose.fullnode.nile.yml
		private network: ./single_node/docker-compose.witness.private.yml
	- If you want to use a different docker compose file, you can specify it by using the -f flags.


The following directory will be created after you start any type of java-tron fullnode:

	- Log directory: ./logs/$type
	- Database directory: ./output-directory/$type


```
trond node run-single [flags]
```

### Examples

```
# Run single java-tron fullnode for main network using the default docker compose file
# The default docker compose file is ./single_node/docker-compose.fullnode.main.yml
# The default configuration file is ./conf/main_net_config.conf
$ ./trond node run-single -t full-main

# Run single java-tron fullnode for nile network using the default docker compose file
# The default docker compose file is ./single_node/docker-compose.fullnode.nile.yml
# The default configuration file is ./conf/nile_net_config.conf
$ ./trond node run-single -t full-nile

# Run single java-tron witness node for private network using the default docker compose file
# The default docker compose file is ./single_node/docker-compose.witness.private.yml
# The default configuration file is ./conf/private_net_config_*.conf
$ ./trond node run-single -t witness-private

# Run single java-tron fullnode for main network using the specified docker compose file
# The docker compose file is ./docker-compose.fullnode.main.yml
# You need to specify the configuration file in the docker compose file, please refer to the default docker compose files for details
$./trond node run-single -t full-main -f ./docker-compose.fullnode.main.yml

```

### Options

```
  -f, --compose-file string   The docker compose file you want to use (optional, if not specified, the default file will be used)
  -h, --help                  help for run-single
  -t, --type string           Node type you want to deploy (required, available: full-main, full-nile, witness-private)
```

### SEE ALSO

* [trond node](trond_node.md)	 - Commands for operating java-tron docker node.
* [trond node run-single stop](trond_node_run-single_stop.md)	 - Stop single java-tron node for different networks.
