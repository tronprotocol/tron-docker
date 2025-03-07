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
			Warning: this command only support configuration for remote nodes, not inlude the local node on the same server. For local node setup, please refer to "./trond node run-single"

			Default environment configuration for node operation:

			Current directory: tron-docker

			The following files are required:

				- Configuration file for private network layout (Please refer to the example configuration file and rewrite it according to your needs)
					./conf/private_net_layout.toml

				- Docker compose file(by default, these exist in the current repository directory)
					single node
						private network witness: ./single_node/docker-compose.witness.private.yml
						private network fullnode: ./single_node/docker-compose.fullnode.private.yml
		`),
	Example: heredoc.Doc(`
			# Check and configure node local environment
			$ ./trond node env-multi
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
	Database struct {
		DatabaseTar string `mapstructure:"database_tar"`
	}
	Nodes []struct {
		NodeIP        string `mapstructure:"node_ip"`
		NodeDirecotry string `mapstructure:"node_direcotry"`
		ConfigFile    string `mapstructure:"config_file"`
		NodeType      string `mapstructure:"node_type"`
		SSHPort       int    `mapstructure:"ssh_port"`
		SSHUser       string `mapstructure:"ssh_user"`
		SSHPassword   string `mapstructure:"ssh_password"`
		SSHKey        string `mapstructure:"ssh_key"` // Optional private key path
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
	fmt.Println("Database Config:")
	fmt.Printf("  Path: %s\n", cfg.Database.DatabaseTar)
	if len(cfg.Database.DatabaseTar) == 0 {
		fmt.Println("Database: not provide database, the chain will start from block 0")
	} else {
		if yes, isDir := utils.PathExists(cfg.Database.DatabaseTar); !yes || isDir {
			fmt.Println("Error: file not exists or not a file:", cfg.Database.DatabaseTar)
			checkFalse = true
		}
	}

	fmt.Println("\nNodes:")
	for i, node := range cfg.Nodes {
		fmt.Printf("  Node %d:\n", i)
		fmt.Printf("    IP: %s\n", node.NodeIP)
		fmt.Printf("    Directory: %s\n", node.NodeDirecotry)
		fmt.Printf("    Config File: %s\n", node.ConfigFile)
		fmt.Printf("    Type: %s\n", node.NodeType)
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
			fmt.Println("Error: file not exists or not a file:", node.ConfigFile)
			checkFalse = true
		}
	}

	return checkFalse, cfg
}

func attemptSCPTransfer(cfg Config) error {
	fmt.Println("\nPerforming SCP Transfers:")

	for i, node := range cfg.Nodes {
		fmt.Printf("  Node %d (%s):\n", i, node.NodeIP)
		fmt.Printf("    Attempting to SCP transfer of %s to %s:%s as %s...\n", node.ConfigFile, node.NodeIP, node.NodeDirecotry, node.SSHUser)
		remotePath := filepath.Join(node.NodeDirecotry, "conf", filepath.Base(node.ConfigFile))
		err := utils.SCPFile(node.NodeIP, node.SSHPort, node.SSHUser, node.SSHPassword, node.SSHKey, node.ConfigFile, remotePath)
		if err != nil {
			fmt.Printf("    SCP transfer failed: %v\n", err)
			return err
		} else {
			fmt.Printf("    SCP transfer succeeded to %s\n", node.NodeIP)
		}

		dockerComposeFile := "./single_node/docker-compose.fullnode.private.yml"
		if node.NodeType == "full" {
			dockerComposeFile = "./single_node/docker-compose.fullnode.private.yml"
		} else if node.NodeType == "sr" {
			dockerComposeFile = "./single_node/docker-compose.witness.private.yml"
		}

		fmt.Printf("    Attempting to SCP transfer of %s to %s:%s as %s...\n", dockerComposeFile, node.NodeIP, node.NodeDirecotry, node.SSHUser)
		remotePath = filepath.Join(node.NodeDirecotry, "single_node", filepath.Base(dockerComposeFile))
		err = utils.SCPFile(node.NodeIP, node.SSHPort, node.SSHUser, node.SSHPassword, node.SSHKey, node.ConfigFile, remotePath)
		if err != nil {
			fmt.Printf("    SCP transfer failed: %v\n", err)
			return err
		} else {
			fmt.Printf("    SCP transfer succeeded to %s\n", node.NodeIP)
		}
	}

	if len(cfg.Database.DatabaseTar) != 0 {
		for i, node := range cfg.Nodes {
			fmt.Printf("  Node %d (%s):\n", i, node.NodeIP)
			fmt.Printf("    Attempting to SCP transfer of %s to %s:%s as %s...\n", cfg.Database.DatabaseTar, node.NodeIP, node.NodeDirecotry, node.SSHUser)

			remotePath := filepath.Join(node.NodeDirecotry, filepath.Base(cfg.Database.DatabaseTar))
			err := utils.SCPFile(node.NodeIP, node.SSHPort, node.SSHUser, node.SSHPassword, node.SSHKey, cfg.Database.DatabaseTar, remotePath)
			if err != nil {
				fmt.Printf("    SCP transfer failed: %v\n", err)
			} else {
				fmt.Printf("    SCP transfer succeeded to %s\n", node.NodeIP)
			}
		}
	}
	return nil
}

func mkdirRemoteDirectory(cfg Config) error {
	fmt.Println("\nPerforming SSH Mkdir:")

	remoteDirs := []string{"logs", "conf", "output-directory", "single_node"}

	for i, node := range cfg.Nodes {
		fmt.Printf("  Node %d (%s):\n", i, node.NodeIP)
		for _, item := range remoteDirs {
			fmt.Printf("    Attempting to create directory %s to %s:%s as %s...\n", item, node.NodeIP, node.NodeDirecotry, node.SSHUser)
			remotePath := filepath.Join(node.NodeDirecotry, item)
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
