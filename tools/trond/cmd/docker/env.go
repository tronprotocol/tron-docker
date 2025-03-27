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

			The tools/docker/docker_env/check-install-docker.sh script will be used to check and install docker and docker-compose.
		`),
	Example: heredoc.Doc(`
			# Check and install docker and docker-compose (for Linux and Mac)
			$ ./trond docker install-docker
		`),
	Run: func(cmd *cobra.Command, args []string) {
		cmds := []string{
			"./check-install-docker.sh",
		}
		if err := utils.RunMultipleCommands(strings.Join(cmds, " && "), "./tools/docker/docker_env"); err != nil {
			fmt.Println("Error: ", err)
			return
		}
	},
}

func init() {
	DockerCmd.AddCommand(envCmd)
}
