## trond node run-multi stop

Stop single java-tron node for different networks.

### Synopsis

The following configuration files are required:

	- Configuration file(by default, these exist in the current repository directory)
		main network: ./conf/main_net_config.conf
		nile network: ./conf/nile_net_config.conf
		private network: ./conf/private_net_config_*.conf
	- Docker compose file(by default, these exist in the current repository directory)
		main network: ./single_node/docker-compose.fullnode.main.yml
		nile network: ./single_node/docker-compose.fullnode.nile.yml
		private network: ./single_node/docker-compose.witness.private.yml


```
trond node run-multi stop [flags]
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
  -h, --help   help for stop
```

### SEE ALSO

* [trond node run-multi](trond_node_run-multi.md)	 - Run multi remote java-tron nodes according to the layout configuration file.
