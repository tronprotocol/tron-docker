## trond node run-multi

Run multi remote java-tron nodes according to the layout configuration file.

### Synopsis

You need to make sure the remote environment is ready before running the node. Run "./trond node env-multi" to init the environment before starting the node.

The following files are required:

	- Configuration file for private network layout (Please refer to the example configuration file and rewrite it according to your needs)
		./conf/private_net_layout.toml



```
trond node run-multi [flags]
```

### Examples

```
# Run java-tron nodes according to ./conf/private_net_layout.toml
$ ./trond node run-multi

```

### Options

```
  -h, --help   help for run-multi
```

### SEE ALSO

* [trond node](trond_node.md)	 - Commands for operating java-tron docker node.
* [trond node run-multi stop](trond_node_run-multi_stop.md)	 - Stop single java-tron node for different networks.
