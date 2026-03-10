## trond node env-multi

Check and configure node environment across multiple nodes.

### Synopsis

Warning: this command only support configuration for remote nodes, not include the local node on the same server. For local node setup, please refer to "./trond node run-single"

Default environment configuration for node operation:

Current directory: tron-docker

The following files are required:

	- Configuration file for private network layout (Please refer to the example configuration file and rewrite it according to your needs)
		./conf/private_net_layout.toml


```
trond node env-multi [flags]
```

### Examples

```
# Check and configure node local environment
$ ./trond node env-multi

# Use the scp command to copy files and synchronize databases between multiple nodes:
$ scp -P 2222 local_file.txt remote_user@192.168.1.100:/home/user/

```

### Options

```
  -h, --help   help for env-multi
```

### SEE ALSO

* [trond node](trond_node.md)	 - Commands for operating java-tron docker node.
