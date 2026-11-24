package docker

import (
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
)

// buildCmd represents the snapshot source command
var buildCmd = &cobra.Command{
	Use:          "build",
	Short:        "Build java-tron docker image.",
	SilenceUsage: true,
	Long: heredoc.Doc(`
			Build the java-tron Docker image from the master branch. The Dockerfile pins OpenJDK 8 for amd64 and OpenJDK 17 for arm64.
			Use --openjdk-version or set OPENJDK_VERSION to override the Ubuntu package version for one build. The flag takes precedence.
		`),
	Example: heredoc.Doc(`
			# Please ensure that JDK 8 is installed, as it is required to execute the commands below.
			# Build java-tron docker image, default output: tronprotocol/java-tron:latest
			# Using code from https://github.com/tronprotocol/java-tron.git
			$ ./trond docker build

			# Build java-tron docker image with specified org, artifact and version, targeting the amd64 platform
			# Using code from https://github.com/tronprotocol/java-tron.git
			$ ./trond docker build -o tronprotocol -a java-tron -v latest
			$ ./trond docker build -o tronprotocol -a java-tron -v latest -n mainnet

			# Build java-tron docker image for nile testnet with specified org, artifact and version, targeting the amd64 platform
			# Using code from https://github.com/tron-nile-testnet/nile-testnet.git
			$ ./trond docker build -o tronnile -a java-tron -v latest -n nile

			# Build java-tron docker image with specified org, artifact and version, targeting the arm64 platform
			# Using code from https://github.com/tronprotocol/java-tron.git
			$ ./trond docker build -o tronprotocol -a java-tron -v latest
			$ ./trond docker build -o tronprotocol -a java-tron -v latest -n mainnet -p linux/arm64

			# Override the Ubuntu OpenJDK package version for an arm64 build
			$ ./trond docker build -p linux/arm64 --openjdk-version '<full Ubuntu Java 17 package version>'

			# Build java-tron docker image for nile testnet with specified org, artifact and version, targeting the arm64 platform
			# Using code from https://github.com/tron-nile-testnet/nile-testnet.git
			$ ./trond docker build -o tronnile -a java-tron -v latest -n nile -p linux/arm64
		`),
	RunE: func(cmd *cobra.Command, args []string) error {

		// Get the flag value
		org, _ := cmd.Flags().GetString("org")
		artifact, _ := cmd.Flags().GetString("artifact")
		version, _ := cmd.Flags().GetString("version")
		network, _ := cmd.Flags().GetString("network")
		platform, _ := cmd.Flags().GetString("platform")
		openJDKVersion, _ := cmd.Flags().GetString("openjdk-version")

		fmt.Println("The building progress may take a long time, depending on your network speed.")
		fmt.Println("If you don't specify the flags for building, the default values will be used.")
		fmt.Println("The default result will be: tronprotocol/java-tron:latest")
		fmt.Println("Start building...")

		return runGradleDockerTask("sourceDocker", org, artifact, version, network, platform, openJDKVersion)
	},
}

func init() {

	buildCmd.Flags().StringP("org", "o", "tronprotocol", "OrgName for the docker image")
	buildCmd.Flags().StringP("artifact", "a", "java-tron", "ArtifactName for the docker image")
	buildCmd.Flags().StringP("version", "v", "latest", "Release version for the docker image")
	buildCmd.Flags().StringP("network", "n", "mainnet", "Which code will be used for the docker image, mainnet or nile")
	buildCmd.Flags().StringP("platform", "p", "linux/amd64", "Platform for the docker image, linux/amd64 or linux/arm64")
	buildCmd.Flags().String("openjdk-version", "", "Full Ubuntu OpenJDK package version for the target architecture")

	DockerCmd.AddCommand(buildCmd)
}
