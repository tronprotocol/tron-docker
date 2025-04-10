package docker

import (
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"github.com/tronprotocol/tron-docker/utils"
)

// buildCmd represents the snapshot source command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build java-tron docker image.",
	Long: heredoc.Doc(`
			Build java-tron docker image locally. The master branch of java-tron repository will be built by default, using jdk1.8.0_202.
		`),
	Example: heredoc.Doc(`
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
		`),
	Run: func(cmd *cobra.Command, args []string) {

		if yes, err := utils.IsJDK1_8(); err != nil || !yes {
			fmt.Println("Error: JDK version should be 1.8")
			return
		}

		// Get the flag value
		org, _ := cmd.Flags().GetString("org")
		artifact, _ := cmd.Flags().GetString("artifact")
		version, _ := cmd.Flags().GetString("version")
		network, _ := cmd.Flags().GetString("network")

		fmt.Println("The building progress may take a long time, depending on your network speed.")
		fmt.Println("If you don't specify the flags for building, the default values will be used.")
		fmt.Println("The default result will be: tronprotocol/java-tron:latest")
		fmt.Println("Start building...")

		cmd1 := fmt.Sprintf("./gradlew --no-daemon sourceDocker -PdockerOrgName=%s -PdockerArtifactName=%s -Prelease.releaseVersion=%s", org, artifact, version)
		if len(network) > 0 {
			cmd1 = fmt.Sprintf("./gradlew --no-daemon sourceDocker -PdockerOrgName=%s -PdockerArtifactName=%s -Prelease.releaseVersion=%s  -Pnetwork=%s", org, artifact, version, network)
		}
		cmds := []string{cmd1}
		if err := utils.RunMultipleCommands(strings.Join(cmds, " && "), "./tools/gradlew"); err != nil {
			fmt.Println("Error: ", err)
			return
		}
	},
}

func init() {

	buildCmd.Flags().StringP("org", "o", "tronprotocol", "OrgName for the docker image")
	buildCmd.Flags().StringP("artifact", "a", "java-tron", "ArtifactName for the docker image")
	buildCmd.Flags().StringP("version", "v", "latest", "Release version for the docker image")
	buildCmd.Flags().StringP("network", "n", "mainnet", "Which code will be used for the docker image, mainnet or nile")

	DockerCmd.AddCommand(buildCmd)
}
