## trond node env

Check and configure node local environment

### Synopsis

Default environment configuration for node operation:

Current directory: tron-docker

The following files are required:

	- Configuration file(by default, these exist in the current repository directory)
		main network: ./conf/main_net_config.conf
		nile network: ./conf/nile_net_config.conf
		private network: ./conf/private_net_config_*.conf
	- Docker compose file(by default, these exist in the current repository directory)
		single node
			main network: ./single_node/docker-compose.fullnode.main.yml
			nile network: ./single_node/docker-compose.fullnode.nile.yml
			private network: ./single_node/docker-compose.witness.private.yml
		multiple nodes
			private network: ./private_net/docker-compose.private.yml


```
trond node env [flags]
```

### Examples

```
# Check and configure node local environment
$ ./trond node env

```

### Options

```
  -h, --help   help for env
```

### SEE ALSO

* [trond node](trond_node.md)	 - Commands for operating java-tron docker node.
