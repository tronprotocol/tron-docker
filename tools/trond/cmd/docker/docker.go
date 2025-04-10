package docker

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
)

// DockerCmd represents the docker command
var DockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Commands for operating java-tron docker image.",
	Long: heredoc.Doc(`
			Commands used for operating docker image, such as:

				1. build java-tron docker image locally
				2. test the built image
				3. pull the latest image from officical docker hub

			Please refer to the available commands below.
		`),
	Example: heredoc.Doc(`
			# Help information for docker command
			$ ./trond docker

			# Check and install docker and docker-compose (for Linux and Mac)
			$ ./trond docker install-docker

			# Please ensure that JDK 8 is installed, as it is required to execute the commands below.
			# Build java-tron docker image, default output: tronprotocol/java-tron:latest
			# Using code from https://github.com/tronprotocol/java-tron.git
			$ ./trond docker build

			# Build java-tron docker image with specified org, artifact and version
			# Using code from https://github.com/tronprotocol/java-tron.git
			$ ./trond docker build -o tronprotocol -a java-tron -v latest
			$ ./trond docker build -o tronprotocol -a java-tron -v latest -n mainnet

			# Build java-tron docker image for nile testnet with specified org, artifact and version
			# Using code from https://github.com/tron-nile-testnet/nile-testnet.git
			$ ./trond docker build -o tronnile -a java-tron -v latest -n nile

			# Test java-tron docker image, defualt: tronprotocol/java-tron:latest
			$ ./trond docker test

			# Test java-tron docker image with specified org, artifact and version
			$ ./trond docker test -o tronprotocol -a java-tron -v latest
			$ ./trond docker test -o tronnile -a java-tron -v latest -n nile
		`),
}

func init() {
}
