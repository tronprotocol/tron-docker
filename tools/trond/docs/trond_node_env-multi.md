## trond node env-multi

Check and configure node environment across multiple nodes.

### Synopsis

Warning: this command only support configuration for remote nodes, not inlude the local node on the same server. For local node setup, please refer to "./trond node run-single"

Default environment configuration for node operation:

Current directory: tron-docker

The following files are required:

	- Configuration file for private network layout (Please refer to the example configuration file and rewrite it according to your needs)
		./conf/private_net_layout.toml

	- Docker compose file(by default, these exist in the current repository directory)
		single node
			private network witness: ./single_node/docker-compose.witness.private.yml
			private network fullnode: ./single_node/docker-compose.fullnode.private.yml


```
trond node env-multi [flags]
```

### Examples

```
# Check and configure node local environment
$ ./trond node env-multi

```

### Options

```
  -h, --help   help for env-multi
```

### SEE ALSO

* [trond node](trond_node.md)	 - Commands for operating java-tron docker node.
