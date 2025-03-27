package docker

import (
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"github.com/tronprotocol/tron-docker/utils"
)

// testCmd represents the snapshot source command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test java-tron docker image.",
	Long: heredoc.Doc(`
			Test java-tron docker image locally. If no flags are provided, "tronprotocol/java-tron:latest" image will be tested.

			The test includes the following tasks:

				1. Perform port checks
				2. Verify whether block synchronization is functioning normally
				3. Automatically remove the test container it starts. If, for any reason, this process is manually killed, you can find the container ID from the log message "Container ID: xxxxxx" and manually remove the test container using "docker rm -f xxxxxx".

			It is used to test TRON Mainnet block synchronization within the docker container. It has a timeout limit of 1000 seconds and may encounter the error “ERROR: goss_wait.yaml never passed” due to network issues or timeouts. If this happens, please check your network status and try again.
		`),
	Example: heredoc.Doc(`
			# Build java-tron docker image, defualt: tronprotocol/java-tron:latest
			$ ./trond docker test

			# Build java-tron docker image with specified org, artifact and version
			$ ./trond docker test -o tronprotocol -a java-tron -v latest
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

		fmt.Println("If you don't specify the flags for building, the default values will be used.")
		fmt.Println("The default result will be: tronprotocol/java-tron:latest")
		fmt.Println("Start testing...")
		cmds := []string{
			fmt.Sprintf("./gradlew --no-daemon testDocker -PdockerOrgName=%s -PdockerArtifactName=%s -Prelease.releaseVersion=%s", org, artifact, version),
		}
		if err := utils.RunMultipleCommands(strings.Join(cmds, " && "), "./tools/gradlew"); err != nil {
			fmt.Println("Error: ", err)
		}
	},
}

func init() {

	testCmd.Flags().StringP("org", "o", "tronprotocol", "OrgName for the docker image")
	testCmd.Flags().StringP("artifact", "a", "java-tron", "ArtifactName for the docker image")
	testCmd.Flags().StringP("version", "v", "latest", "Release version for the docker image")

	DockerCmd.AddCommand(testCmd)
}
