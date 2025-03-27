package node

import (
	"fmt"
	"log"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"github.com/tronprotocol/tron-docker/utils"
)

var runSingleCmd = &cobra.Command{
	Use:   "run-single",
	Short: "Run single java-tron node for different networks.",
	Long: heredoc.Doc(`
	You need to make sure the local environment is ready before running the node. Run "./trond node env" to check the environment before starting the node.

	The following files are required before starting the node (these files will be need if you want to start any type of java-tron fullnode using the default configuration file):

		- Configuration file(by default, these exist in the current repository directory)
			main network: ./conf/main_net_config.conf
			nile network: ./conf/nile_net_config.conf
			private network: ./conf/private_net_config_*.conf
		- Docker compose file(by default, these exist in the current repository directory)
			main network: ./single_node/docker-compose.fullnode.main.yml
			nile network: ./single_node/docker-compose.fullnode.nile.yml
			private network: ./single_node/docker-compose.witness.private.yml
		- If you want to use a different docker compose file, you can specify it by using the -f flags.


	The following directory will be created after you start any type of java-tron fullnode:

		- Log directory: ./logs/$type
		- Database directory: ./output-directory/$type
		`),
	Example: heredoc.Doc(`
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
		`),
	Run: func(cmd *cobra.Command, args []string) {
		if checkEnvFailed() {
			fmt.Println("Error: local environment check failed, please redownload this repository and try")
			return
		}

		nType, _ := cmd.Flags().GetString("type")
		dockerComposeFile := ""
		switch nType {
		case "full-main":
			dockerComposeFile = "./single_node/docker-compose.fullnode.main.yml"
		case "full-nile":
			dockerComposeFile = "./single_node/docker-compose.fullnode.nile.yml"
		case "witness-private":
			dockerComposeFile = "./single_node/docker-compose.witness.private.yml"
		default:
			fmt.Println("Error: type not supported", nType)
			return
		}

		composeFile, _ := cmd.Flags().GetString("compose-file")
		if composeFile != "" {
			dockerComposeFile = composeFile
		}
		fmt.Printf("Using docker compose file: %s\n", dockerComposeFile)

		if yes, isDir := utils.PathExists(dockerComposeFile); !yes || isDir {
			fmt.Println("Error: file not exists or not a file:", dockerComposeFile)
		}

		fmt.Println("Starting node...")
		fmt.Println("Using docker compose file: ", dockerComposeFile)
		if err := utils.RunComposeServiceOnce(dockerComposeFile); err != nil {
			fmt.Println("Error: ", err)
			return
		}
		fmt.Println("Node started successfully.")

		var directory string
		switch nType {
		case "full-main":
			directory = "/mainnet"
		case "full-nile":
			directory = "/nile"
		case "witness-private":
			directory = "/private"
		}
		fmt.Printf("You can check the log file in ./logs%s directory. Run 'tail -f ./logs%s/tron.log' to check the log.\n", directory, directory)

		fmt.Println("You can also check the docker container's status by running 'docker ps'.")
	},
}

var runSingleStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop single java-tron node for different networks.",
	Long: heredoc.Doc(`
			The following files are required before stopping the node (these files will be need if you started java-tron fullnode using the default configuration file):

				- Docker compose file(by default, these exist in the current repository directory)
					main network: ./single_node/docker-compose.fullnode.main.yml
					nile network: ./single_node/docker-compose.fullnode.nile.yml
					private network: ./single_node/docker-compose.witness.private.yml

				- If you used a different docker compose file, you need to specify it by using the -f flags.
		`),
	Example: heredoc.Doc(`
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
	Run: func(cmd *cobra.Command, args []string) {
		nType, _ := cmd.Flags().GetString("type")

		dockerComposeFile := ""
		switch nType {
		case "full-main":
			dockerComposeFile = "./single_node/docker-compose.fullnode.main.yml"
		case "full-nile":
			dockerComposeFile = "./single_node/docker-compose.fullnode.nile.yml"
		case "witness-private":
			dockerComposeFile = "./single_node/docker-compose.witness.private.yml"
		default:
			fmt.Println("Error: type not supported", nType)
			return
		}

		composeFile, _ := cmd.Flags().GetString("compose-file")
		if composeFile != "" {
			dockerComposeFile = composeFile
		}
		fmt.Printf("Using docker compose file: %s\n", dockerComposeFile)

		if yes, isDir := utils.PathExists(dockerComposeFile); !yes || isDir {
			fmt.Println("Error: file not exists or not a file:", dockerComposeFile)
		}

		fmt.Println("Using docker compose file: ", dockerComposeFile)
		if msg, err := utils.StopDockerCompose(dockerComposeFile); err != nil {
			fmt.Println("Error: ", err)
			return
		} else {
			fmt.Println("Stop done: ", msg)
		}
	},
}

func init() {
	runSingleCmd.AddCommand(runSingleStopCmd)
	NodeCmd.AddCommand(runSingleCmd)

	runSingleCmd.Flags().StringP(
		"type", "t", "",
		"Node type you want to deploy (required, available: full-main, full-nile, witness-private)")

	if err := runSingleCmd.MarkFlagRequired("type"); err != nil {
		log.Fatalf("Error marking type flag as required: %v", err)
	}

	runSingleCmd.Flags().StringP(
		"compose-file", "f", "",
		"The docker compose file you want to use (optional, if not specified, the default file will be used)")

	runSingleStopCmd.Flags().StringP(
		"type", "t", "",
		"Node type you want to deploy (required, available: full-main, full-nile, witness-private)")

	if err := runSingleStopCmd.MarkFlagRequired("type"); err != nil {
		log.Fatalf("Error marking type flag as required: %v", err)
	}

	runSingleStopCmd.Flags().StringP(
		"compose-file", "f", "",
		"The docker compose file you used to start the node before (optional, if not specified, the default file will be used)")
}
