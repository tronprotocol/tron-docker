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
	6. deploy multiple nodes in different local machines with ssh(one node on one machine)
	7. deploy wallet-cli

Please refer to the available commands below.


### Examples

```
# Help information for node command
$ ./trond node

# Check and configure node local environment
$ ./trond node env

# Run single java-tron fullnode for main network
$ ./trond node run-single -t full-main

# Stop
$ ./trond node run-single stop -t full-main

# Run single java-tron fullnode for nile network
$ ./trond node run-single -t full-nile

# Stop
$ ./trond node run-single stop -t full-nile

# Run single java-tron witness node for private network
$ ./trond node run-single -t witness-private

# Stop
$ ./trond node run-single stop -t witness-private

```

### Options

```
  -h, --help   help for node
```

### SEE ALSO

* [trond](trond.md)	 - Docker automation for TRON nodes
* [trond node env](trond_node_env.md)	 - Check and configure node local environment
* [trond node run-single](trond_node_run-single.md)	 - Run single java-tron node for different networks.
