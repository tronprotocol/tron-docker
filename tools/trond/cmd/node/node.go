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
		`),
}

func init() {
}
