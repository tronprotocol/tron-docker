package node

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
)

// NodeCmd represents the node command
var NodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Commands for operating java-tron docker node.",
	Long: heredoc.Doc(`
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
		`),
	Example: heredoc.Doc(`
			# Help information for node command
			$ ./trond node

			# Check and configure node local environment. Make sure run this first before starting the node.
			$ ./trond node env

			# Run single java-tron fullnode for main network using the default docker compose file
			# The default docker compose file is ./single_node/docker-compose.fullnode.main.yml
			# The default configuration file is ./conf/main_net_config.conf
			$ ./trond node run-single -t full-main

			# Run single java-tron fullnode for nile network using the default docker compose file
			# The default docker compose file is ./single_node/docker-compose.fullnode.nile.yml
			# The default configuration file is ./conf/nile_net_config.conf
			$ ./trond node run-single -t full-nile

			# Run single java-tron witness node for private network using the default docker compose file
			# The default docker compose file is ./single_node/docker-compose.witness.private.yml
			# The default configuration file is ./conf/private_net_config_*.conf
			$ ./trond node run-single -t witness-private

			# Run single java-tron fullnode for main network using the specified docker compose file
			# The docker compose file is ./docker-compose.fullnode.main.yml
			# You need to specify the configuration file in the docker compose file, please refer to the default docker compose files for details
			$./trond node run-single -t full-main -f ./docker-compose.fullnode.main.yml

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
		`),
}

func init() {
}
