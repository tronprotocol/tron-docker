## trond node run-multi stop

Stop multi java-tron node for different networks.

### Synopsis

The following configuration files are required:

	- Configuration file(by default, these exist in the current repository directory)
			./conf/private_net_layout.toml

SECURITY CONFIGURATION:

When stopping remote nodes, SSH connections are used to execute docker-compose commands. By default, host key verification is disabled for ease of testing.

For Production Environments:

Enable strict host key checking to prevent man-in-the-middle attacks:

	# Enable strict host key verification
	export TROND_STRICT_HOST_KEY_CHECK=true

	# Optional: Specify custom known_hosts file location
	export TROND_KNOWN_HOSTS_FILE=/path/to/your/known_hosts

	# Then run the command
	./trond node run-multi stop

For Testing/Development:

Host key verification is disabled by default. You'll see a warning message:
	⚠️  WARNING: Host key verification is DISABLED (testing mode)
	⚠️  For production use, set TROND_STRICT_HOST_KEY_CHECK=true

For more details, see SECURITY.md



```
trond node run-multi stop [flags]
```

### Examples

```
# Stop multi java-tron node (testing mode)
$ ./trond node run-multi stop

# Stop with strict host key verification (production mode)
$ export TROND_STRICT_HOST_KEY_CHECK=true
$ ./trond node run-multi stop

```

### Options

```
  -h, --help   help for stop
```

### SEE ALSO

* [trond node run-multi](trond_node_run-multi.md)	 - Run multi remote java-tron nodes according to the layout configuration file.
