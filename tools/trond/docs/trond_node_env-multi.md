## trond node env-multi

Check and configure node environment across multiple nodes.

### Synopsis

Warning: this command only support configuration for remote nodes, not include the local node on the same server. For local node setup, please refer to "./trond node run-single"

Default environment configuration for node operation:

Current directory: tron-docker

The following files are required:

	- Configuration file for private network layout (Please refer to the example configuration file and rewrite it according to your needs)
		./conf/private_net_layout.toml

SECURITY CONFIGURATION:

When deploying to remote nodes, SSH connections are used to transfer configuration files and manage the environment. By default, host key verification is disabled for ease of testing.

For Production Environments:

Enable strict host key checking to prevent man-in-the-middle attacks:

	# Enable strict host key verification
	export TROND_STRICT_HOST_KEY_CHECK=true

	# Optional: Specify custom known_hosts file location
	export TROND_KNOWN_HOSTS_FILE=/path/to/your/known_hosts

	# Then run the command
	./trond node env-multi

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
trond node env-multi [flags]
```

### Examples

```
# Check and configure node local environment (testing mode)
$ ./trond node env-multi

# Check and configure with strict host key verification (production mode)
$ export TROND_STRICT_HOST_KEY_CHECK=true
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

