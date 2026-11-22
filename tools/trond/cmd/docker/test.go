package docker

import (
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
)

// testCmd represents the snapshot source command
var testCmd = &cobra.Command{
	Use:          "test",
	Short:        "Test java-tron docker image.",
	SilenceUsage: true,
	Long: heredoc.Doc(`
			Test java-tron docker image locally. If no flags are provided, "tronprotocol/java-tron:latest" image will be tested.

			The test includes the following tasks:

				1. Perform port checks
				2. Verify whether block synchronization is functioning normally
				3. Remove the sync probe and node containers on exit. If the runner is killed before cleanup, use their logged IDs to remove the probe first and then the node with "docker rm -f".

			A temporary Python container shares the node's network and checks for an increase in block height. The sync check times out after 1000 seconds. If it times out, check the node's peer connections.
		`),
	Example: heredoc.Doc(`
			# Test java-tron docker image, defualt: tronprotocol/java-tron:latest
			$ ./trond docker test

			# Test java-tron docker image for amd64 with specified org, artifact and version
			$ ./trond docker test -o tronprotocol -a java-tron -v latest
			$ ./trond docker test -o tronnile -a java-tron -v latest -n nile

			# Test java-tron docker image for arm64 with specified org, artifact and version
			$ ./trond docker test -o tronprotocol -a java-tron -v latest -p linux/arm64
			$ ./trond docker test -o tronnile -a java-tron -v latest -n nile -p linux/arm64
		`),
	RunE: func(cmd *cobra.Command, args []string) error {

		/* if yes, err := utils.IsJDK1_8(); err != nil || !yes {
			fmt.Println("Error: JDK version should be 1.8")
			return
		} */

		// Get the flag value
		org, _ := cmd.Flags().GetString("org")
		artifact, _ := cmd.Flags().GetString("artifact")
		version, _ := cmd.Flags().GetString("version")
		network, _ := cmd.Flags().GetString("network")
		platform, _ := cmd.Flags().GetString("platform")

		fmt.Println("If you don't specify the flags for building, the default values will be used.")
		fmt.Println("The default result will be: tronprotocol/java-tron:latest")
		fmt.Println("Start testing...")

		return runGradleDockerTask("testDocker", org, artifact, version, network, platform)
	},
}

func init() {

	testCmd.Flags().StringP("org", "o", "tronprotocol", "OrgName for the docker image")
	testCmd.Flags().StringP("artifact", "a", "java-tron", "ArtifactName for the docker image")
	testCmd.Flags().StringP("version", "v", "latest", "Release version for the docker image")
	testCmd.Flags().StringP("network", "n", "mainnet", "Which code will be used for the docker image, mainnet or nile")
	testCmd.Flags().StringP("platform", "p", "linux/amd64", "Platform for the docker image, linux/amd64 or linux/arm64")

	DockerCmd.AddCommand(testCmd)
}
