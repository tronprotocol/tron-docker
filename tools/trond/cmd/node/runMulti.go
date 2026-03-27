package node

import (
	"fmt"
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"github.com/tronprotocol/tron-docker/utils"
	"path/filepath"
)

var runMultiCmd = &cobra.Command{
	Use:   "run-multi",
	Short: "Run multi remote java-tron nodes according to the layout configuration file.",
	Long: heredoc.Doc(`
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

		`),
	Example: heredoc.Doc(`
			# Run java-tron nodes according to ./conf/private_net_layout.toml (testing mode)
			$ ./trond node run-multi

			# Run with strict host key verification (production mode)
			$ export TROND_STRICT_HOST_KEY_CHECK=true
			$ ./trond node run-multi
		`),
	Run: func(cmd *cobra.Command, args []string) {

		cfg, err := loadConfig()
		if err != nil {
			fmt.Printf("Error: failed to load config: %v\n", err)
		}

		for i, node := range cfg.Nodes {
			fmt.Printf("  Node %d:\n", i)
			fmt.Printf("    IP: %s\n", node.NodeIP)
			fmt.Printf("    Directory: %s\n", node.NodeDirectory)
			fmt.Printf("    Config File: %s\n", node.ConfigFile)
			fmt.Printf("    Docker-Config File: %s\n", node.DockerComposeFile)
			fmt.Printf("    SSH Port: %d\n", node.SSHPort)
			fmt.Printf("    SSH User: %s\n", node.SSHUser)

			// Indicate authentication method
			if node.SSHKey != "" {
				fmt.Printf("    Auth Method: SSH Key (%s)\n", node.SSHKey)
			} else if node.SSHPassword != "" {
				fmt.Printf("    Auth Method: Password\n")
			} else {
				fmt.Printf("    Auth Method: SSH Agent\n")
			}

			if err := utils.CheckSSH(node.NodeIP, node.SSHPort, node.SSHUser, node.SSHPassword, node.SSHKey); err != nil {
				fmt.Printf("    SCP IP(%s) Port (%d) Status: %s\n", node.NodeIP, node.SSHPort, err)
			} else {
				fmt.Printf("    SCP IP(%s) Port (%d) Status: %s\n", node.NodeIP, node.SSHPort, "Open")
			}

			dockerComposeFile := filepath.Join(node.NodeDirectory, filepath.Base(node.DockerComposeFile))
			if err := utils.RunRemoteCompose(node.NodeIP, node.SSHPort, node.SSHUser, node.SSHPassword, node.SSHKey, dockerComposeFile, false); err != nil {
				fmt.Printf("    SSH start remote node failed: %v\n", err)
				return
			} else {
				fmt.Printf("    SSH start node succeeded on %s\n", node.NodeIP)
			}
		}
	},
}

var runMultiStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop multi java-tron node for different networks.",
	Long: heredoc.Doc(`
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

		`),
	Example: heredoc.Doc(`
			# Stop multi java-tron node (testing mode)
			$ ./trond node run-multi stop

			# Stop with strict host key verification (production mode)
			$ export TROND_STRICT_HOST_KEY_CHECK=true
			$ ./trond node run-multi stop
		`),
	Run: func(cmd *cobra.Command, args []string) {

		cfg, err := loadConfig()
		if err != nil {
			fmt.Printf("Error: failed to load config: %v\n", err)
		}

		for i, node := range cfg.Nodes {
			fmt.Printf("  Node %d:\n", i)
			fmt.Printf("    IP: %s\n", node.NodeIP)
			fmt.Printf("    Directory: %s\n", node.NodeDirectory)
			fmt.Printf("    Config File: %s\n", node.ConfigFile)
			fmt.Printf("    SSH Port: %d\n", node.SSHPort)
			fmt.Printf("    SSH User: %s\n", node.SSHUser)

			// Indicate authentication method
			if node.SSHKey != "" {
				fmt.Printf("    Auth Method: SSH Key (%s)\n", node.SSHKey)
			} else if node.SSHPassword != "" {
				fmt.Printf("    Auth Method: Password\n")
			} else {
				fmt.Printf("    Auth Method: SSH Agent\n")
			}

			if err := utils.CheckSSH(node.NodeIP, node.SSHPort, node.SSHUser, node.SSHPassword, node.SSHKey); err != nil {
				fmt.Printf("    SCP IP(%s) Port (%d) Status: %s\n", node.NodeIP, node.SSHPort, err)
			} else {
				fmt.Printf("    SCP IP(%s) Port (%d) Status: %s\n", node.NodeIP, node.SSHPort, "Open")
			}

			dockerComposeFile := filepath.Join(node.NodeDirectory, filepath.Base(node.DockerComposeFile))
			if err := utils.RunRemoteCompose(node.NodeIP, node.SSHPort, node.SSHUser, node.SSHPassword, node.SSHKey, dockerComposeFile, true); err != nil {
				fmt.Printf("    SSH stop remote node failed: %v\n", err)
				return
			} else {
				fmt.Printf("    SSH stop node succeeded on %s\n", node.NodeIP)
			}
		}
	},
}

func init() {
	runMultiCmd.AddCommand(runMultiStopCmd)
	NodeCmd.AddCommand(runMultiCmd)
}
