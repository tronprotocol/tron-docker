## trond node run-multi

Run multi remote java-tron nodes according to the layout configuration file.

### Synopsis

You need to make sure the remote environment is ready before running the node. Run "./trond node env-multi" to init the environment before starting the node.

The following files are required:

	- Configuration file for private network layout (Please refer to the example configuration file and rewrite it according to your needs)
		./conf/private_net_layout.toml

SECURITY CONFIGURATION:

When deploying to remote nodes, SSH connections are used to start and manage docker-compose services. By default, host key verification is disabled for ease of testing.

For Production Environments:

Enable strict host key checking to prevent man-in-the-middle attacks:

	# Enable strict host key verification
	export TROND_STRICT_HOST_KEY_CHECK=true

	# Optional: Specify custom known_hosts file location
	export TROND_KNOWN_HOSTS_FILE=/path/to/your/known_hosts

	# Then run the command
	./trond node run-multi

Setting up known_hosts file:

1. First, manually SSH to each remote server to add it to known_hosts:
	ssh user@remote-server-ip

2. Or use ssh-keyscan to add host keys:
	ssh-keyscan -H remote-server-ip >> ~/.ssh/known_hosts

For Testing/Development:

Host key verification is disabled by default. You'll see a warning message:
	⚠️  WARNING: Host key verification is DISABLED (testing mode)
	⚠️  For production use, set TROND_STRICT_HOST_KEY_CHECK=true

For more details, see SECURITY.md



```
trond node run-multi [flags]
```

### Examples

```
# Run java-tron nodes according to ./conf/private_net_layout.toml (testing mode)
$ ./trond node run-multi

# Run with strict host key verification (production mode)
$ export TROND_STRICT_HOST_KEY_CHECK=true
$ ./trond node run-multi

```

### Options

```
  -h, --help   help for run-multi
```

### SEE ALSO

* [trond node](trond_node.md)	 - Commands for operating java-tron docker node.
* [trond node run-multi stop](trond_node_run-multi_stop.md)	 - Stop multi java-tron node for different networks.
