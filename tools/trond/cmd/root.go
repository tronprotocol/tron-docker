package cmd

import (
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"github.com/tronprotocol/tron-docker/cmd/docker"
	"github.com/tronprotocol/tron-docker/cmd/node"
	"github.com/tronprotocol/tron-docker/cmd/snapshot"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "trond",
	Short: "Docker automation for TRON nodes",
	Long: heredoc.Doc(`
			This tool bundles multiple commands into one, enabling the community to quickly get started with TRON network interaction and development.
		`),
	Example: heredoc.Doc(`
			# Help information for java-tron docker image build and testing command
			$ ./trond docker

			# Help information for database snapshot download related command
			$ ./trond snapshot

			# Help information for TRON node deployment command
			$ ./trond node
		`),
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// Custom front matter function (now accepts a string)
func frontMatter(filename string) string {
	// return fmt.Sprintf("---\ntitle: %s\n---\n", filename)
	return ""
}

// Custom footer linkHandler (now accepts a string)
func linkHandler(filename string) string {
	return filename
}

// genDocsCmd represents the command to generate markdown documentation
var genDocsCmd = &cobra.Command{
	Use:   "gen-docs",
	Short: "Generate markdown documentation for the CLI.",
	Long: heredoc.Doc(`
			This command generates markdown documentation for the CLI commands and subcommands.<br>
			The documentation is saved in the 'docs' directory.<br>
			If the 'docs' directory does not exist, it will be created.
		`),
	Example: heredoc.Doc(`
			# Generate markdown documentation for the CLI
			$ ./trond gen-docs
		`),
	Run: func(cmd *cobra.Command, args []string) {
		docsDir := "./tools/trond/docs"

		// Create docs directory if it doesn't exist
		if _, err := os.Stat(docsDir); os.IsNotExist(err) {
			err := os.Mkdir(docsDir, os.ModePerm)
			if err != nil {
				fmt.Println("Error creating docs directory:", err)
				os.Exit(1)
			}
		}

		rootCmd.DisableAutoGenTag = true
		// Generate Markdown docs
		err := doc.GenMarkdownTreeCustom(rootCmd, docsDir, frontMatter, linkHandler)
		if err != nil {
			fmt.Println("Error generating docs:", err)
			os.Exit(1)
		}

		fmt.Println("Documentation successfully generated in", docsDir)
	},
}

func init() {
	rootCmd.AddCommand(snapshot.SnapshotCmd)
	rootCmd.AddCommand(node.NodeCmd)
	rootCmd.AddCommand(docker.DockerCmd)
	rootCmd.AddCommand(genDocsCmd)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.tron-docker.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
