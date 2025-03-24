package docker

import (
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"github.com/tronprotocol/tron-docker/utils"
)

// envCmd represents the snapshot source command
var envCmd = &cobra.Command{
	Use:   "install-docker",
	Short: "Check and install docker and docker-compose (for Linux and Mac)",
	Long: heredoc.Doc(`
			Check docker and docker-compose installation on Linux and Mac. If docker and docker-compose are not installed, this command will install them for you.

			For Mac: brew install --cask docker

			For Linux: sh get-docker.sh
		`),
	Example: heredoc.Doc(`
			# Check and install docker and docker-compose (for Linux and Mac)
			$ ./trond docker install-docker
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

		fmt.Println("The building progress may take a long time, depending on your network speed.")
		fmt.Println("If you don't specify the flags for building, the default values will be used.")
		fmt.Println("The default result will be: tronprotocol/java-tron:latest")
		fmt.Println("Start building...")
		cmds := []string{
			fmt.Sprintf("./gradlew --no-daemon sourceDocker -PdockerOrgName=%s -PdockerArtifactName=%s -Prelease.releaseVersion=%s", org, artifact, version),
		}
		if err := utils.RunMultipleCommands(strings.Join(cmds, " && "), "./tools/gradlew"); err != nil {
			fmt.Println("Error: ", err)
			return
		}
	},
}

func init() {

	envCmd.Flags().StringP("org", "o", "tronprotocol", "OrgName for the docker image")
	envCmd.Flags().StringP("artifact", "a", "java-tron", "ArtifactName for the docker image")
	envCmd.Flags().StringP("version", "v", "latest", "Release version for the docker image")

	DockerCmd.AddCommand(envCmd)
}
