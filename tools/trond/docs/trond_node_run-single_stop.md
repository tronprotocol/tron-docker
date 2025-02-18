## trond node run-single stop

Stop single java-tron node for different networks.

### Synopsis

The following configuration files are required:

	- Configuration file(by default, these exist in the current repository directory)
		main network: ./conf/main_net_config.conf
		nile network: ./conf/nile_net_config.conf
		private network: ./conf/private_net_config.conf
	- Docker compose file(by default, these exist in the current repository directory)
		main network: ./single_node/docker-compose.fullnode.main.yml
		nile network: ./single_node/docker-compose.fullnode.nile.yml
		private network: ./single_node/docker-compose.witness.private.yml


```
trond node run-single stop [flags]
```

### Examples

```
# Stop single java-tron fullnode for main network
$ ./trond node run-single stop -t full-main

# Stop single java-tron fullnode for nile network
$ ./trond node run-single stop -t full-nile

# Stop single java-tron witness node for private network
$ ./trond node run-single stop -t witness-private

```

### Options

```
  -h, --help          help for stop
  -t, --type string   Node type you want to deploy (required, available: full-main, full-nile, witness-private)
```

### SEE ALSO

* [trond node run-single](trond_node_run-single.md)	 - Run single java-tron node for different networks.
