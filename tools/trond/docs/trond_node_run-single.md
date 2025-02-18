## trond node run-single

Run single java-tron node for different networks.

### Synopsis

You need to make sure the local environment is ready before running the node. Run "./trond node env" to check the environment before starting the node.

The following files are required:

	- Configuration file(by default, these exist in the current repository directory)
		main network: ./conf/main_net_config.conf
		nile network: ./conf/nile_net_config.conf
		private network: ./conf/private_net_config.conf
	- Docker compose file(by default, these exist in the current repository directory)
		main network: ./single_node/docker-compose.fullnode.main.yml
		nile network: ./single_node/docker-compose.fullnode.nile.yml
		private network: ./single_node/docker-compose.witness.private.yml


The following directory will be created after you start any type of java-tron fullnode:

	- Log directory: ./logs/$type
	- Database directory: ./output-directory/$type


```
trond node run-single [flags]
```

### Examples

```
# Run single java-tron fullnode for main network
$ ./trond node run-single -t full-main

# Run single java-tron fullnode for nile network
$ ./trond node run-single -t full-nile

# Run single java-tron witness node for private network
$ ./trond node run-single -t witness-private

```

### Options

```
  -h, --help          help for run-single
  -t, --type string   Node type you want to deploy (required, available: full-main, full-nile, witness-private)
```

### SEE ALSO

* [trond node](trond_node.md)	 - Commands for operating java-tron docker node.
* [trond node run-single stop](trond_node_run-single_stop.md)	 - Stop single java-tron node for different networks.
