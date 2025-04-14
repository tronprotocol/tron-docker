package node

import (
	"fmt"
	"path/filepath"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tronprotocol/tron-docker/utils"
)

var envMultiCmd = &cobra.Command{
	Use:   "env-multi",
	Short: "Check and configure node environment across multiple nodes.",
	Long: heredoc.Doc(`
			Warning: this command only support configuration for remote nodes, not include the local node on the same server. For local node setup, please refer to "./trond node run-single"

			Default environment configuration for node operation:

			Current directory: tron-docker

			The following files are required:

				- Configuration file for private network layout (Please refer to the example configuration file and rewrite it according to your needs)
					./conf/private_net_layout.toml
		`),
	Example: heredoc.Doc(`
			# Check and configure node local environment
			$ ./trond node env-multi

			# Use the scp command to copy files and synchronize databases between multiple nodes:
			$ scp -P 2222 local_file.txt remote_user@192.168.1.100:/home/user/
		`),
	Run: func(cmd *cobra.Command, args []string) {
		if checkFalse, cfg := checkEnvForMulti(); checkFalse {
			fmt.Println("Error: configuration file check failed, please check it and try again")
			return
		} else {
			if err := mkdirRemoteDirectory(cfg); err != nil {
				fmt.Println("Error: create directory on remote node failed,", err)
			} else {
				if err := attemptSCPTransfer(cfg); err != nil {
					fmt.Println("Error: scp file to remote node failed,", err)
				}
			}
		}
	},
}

// Config struct to match TOML
type Config struct {
	Nodes []struct {
		NodeIP            string `mapstructure:"node_ip"`
		NodeDirectory     string `mapstructure:"node_directory"`
		ConfigFile        string `mapstructure:"config_file"`
		DockerComposeFile string `mapstructure:"docker_compose_file"`
		SSHPort           int    `mapstructure:"ssh_port"`
		SSHUser           string `mapstructure:"ssh_user"`
		SSHPassword       string `mapstructure:"ssh_password"`
		SSHKey            string `mapstructure:"ssh_key"` // Optional private key path
	}
}

func loadConfig() (Config, error) {
	var cfg Config
	viper.SetConfigName("private_net_layout")
	viper.SetConfigType("toml")
	viper.AddConfigPath("./conf")

	if err := viper.ReadInConfig(); err != nil {
		return cfg, fmt.Errorf("error reading config: %v", err)
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("error unmarshaling config: %v", err)
	}
	return cfg, nil
}

func checkEnvForMulti() (bool, Config) {
	checkFalse := false

	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("Error: failed to load config: %v\n", err)
		checkFalse = true
	}

	// Access config
	fmt.Println("\nNodes:")
	for i, node := range cfg.Nodes {
		fmt.Printf("  Node %d:\n", i)
		fmt.Printf("    IP: %s\n", node.NodeIP)
		fmt.Printf("    Directory: %s\n", node.NodeDirectory)
		fmt.Printf("    Config File: %s\n", node.ConfigFile)
		fmt.Printf("    Docker Compose File: %s\n", node.DockerComposeFile)
		fmt.Printf("    SSH Port: %d\n", node.SSHPort)
		fmt.Printf("    SSH User: %s\n", node.SSHUser)

		// Indicate authentication method
		if node.SSHKey != "" {
			fmt.Printf("    Auth Method: SSH Key (%s)\n", node.SSHKey)
		} else if node.SSHPassword != "" {
			fmt.Printf("    Auth Method: Password\n")
		} else {
			fmt.Printf("    Auth Method: SSH Agent\n")
		}

		if err := utils.CheckSSH(node.NodeIP, node.SSHPort, node.SSHUser, node.SSHPassword, node.SSHKey); err != nil {
			fmt.Printf("    SCP IP(%s) Port (%d) Status: %s\n", node.NodeIP, node.SSHPort, err)
		} else {
			fmt.Printf("    SCP IP(%s) Port (%d) Status: %s\n", node.NodeIP, node.SSHPort, "Open")
		}

		if yes, isDir := utils.PathExists(node.ConfigFile); !yes || isDir {
			fmt.Println("Error: config file not exists or not a file:", node.ConfigFile)
			checkFalse = true
		}
		if yes, isDir := utils.PathExists(node.DockerComposeFile); !yes || isDir {
			fmt.Println("Error: docker-compose file not exists or not a file:", node.ConfigFile)
			checkFalse = true
		}
	}

	return checkFalse, cfg
}

func attemptSCPTransfer(cfg Config) error {
	fmt.Println("\nPerforming SCP Transfers:")

	for i, node := range cfg.Nodes {
		fmt.Printf("  Node %d (%s):\n", i, node.NodeIP)
		fmt.Printf("    Attempting to SCP transfer of %s to %s:%s as %s...\n", node.ConfigFile, node.NodeIP, node.NodeDirectory, node.SSHUser)
		remotePath := filepath.Join(node.NodeDirectory, "conf", filepath.Base(node.ConfigFile))
		err := utils.SCPFile(node.NodeIP, node.SSHPort, node.SSHUser, node.SSHPassword, node.SSHKey, node.ConfigFile, remotePath)
		if err != nil {
			fmt.Printf("    SCP transfer failed: %v\n", err)
			return err
		} else {
			fmt.Printf("    SCP transfer succeeded to %s\n", node.NodeIP)
		}

		fmt.Printf("    Attempting to SCP transfer of %s to %s:%s as %s...\n", node.DockerComposeFile, node.NodeIP, node.NodeDirectory, node.SSHUser)
		remotePath = filepath.Join(node.NodeDirectory, filepath.Base(node.DockerComposeFile))
		err = utils.SCPFile(node.NodeIP, node.SSHPort, node.SSHUser, node.SSHPassword, node.SSHKey, node.DockerComposeFile, remotePath)
		if err != nil {
			fmt.Printf("    SCP transfer failed: %v\n", err)
			return err
		} else {
			fmt.Printf("    SCP transfer succeeded to %s\n", node.NodeIP)
		}
	}

	return nil
}

func mkdirRemoteDirectory(cfg Config) error {
	fmt.Println("\nPerforming SSH Mkdir:")

	remoteDirs := []string{"logs", "conf", "output-directory"}

	for i, node := range cfg.Nodes {
		fmt.Printf("  Node %d (%s):\n", i, node.NodeIP)
		for _, item := range remoteDirs {
			fmt.Printf("    Attempting to create directory %s to %s:%s as %s...\n", item, node.NodeIP, node.NodeDirectory, node.SSHUser)
			remotePath := filepath.Join(node.NodeDirectory, item)
			err := utils.SSHMkdirIfNotExist(node.NodeIP, node.SSHPort, node.SSHUser, node.SSHPassword, node.SSHKey, remotePath)
			if err != nil {
				fmt.Printf("    Create directory failed: %v\n", err)
				return err
			} else {
				fmt.Printf("    Create directory succeeded to %s\n", node.NodeIP)
			}
		}
	}

	return nil
}

func init() {
	NodeCmd.AddCommand(envMultiCmd)
}
