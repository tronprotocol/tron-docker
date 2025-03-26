## trond node

Commands for operating java-tron docker node.

### Synopsis

Commands used for operating node, such as:

	1. check and configure node local environment
	2. deploy single Fullnode/SR for different environment(main network, nile network, private network)
	3. stop single node

	** coming soon **
	4. deploy multiple nodes in local single machine
	5. stop multiple nodes in local single machine
	6. deploy multiple nodes in different local machines using ssh(one node on one machine)
	7. deploy wallet-cli

Please refer to the available commands below.


### Examples

```
# Help information for node command
$ ./trond node

# Check and configure node local environment
$ ./trond node env

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
  -h, --help   help for node
```

### SEE ALSO

* [trond](trond.md)	 - Docker automation for TRON nodes
* [trond node env](trond_node_env.md)	 - Check and configure node local environment
* [trond node env-multi](trond_node_env-multi.md)	 - Check and configure node environment across multiple nodes.
* [trond node run-multi](trond_node_run-multi.md)	 - Run multi remote java-tron nodes according to the layout configuration file.
* [trond node run-single](trond_node_run-single.md)	 - Run single java-tron node for different networks.
