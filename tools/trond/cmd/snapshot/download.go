package snapshot

import (
	"errors"
	"fmt"
	"github.com/tronprotocol/tron-docker/config"
	"log"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"github.com/tronprotocol/tron-docker/utils"
)

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download target backup snapshot to current directory",
	Long: heredoc.Doc(`
			Refer to the snapshot source domain and backup name you input, the available backup snapshot will be downloaded to the local directory.<br>

			Note:
			- because some snapshot sources have multiple snapshot types, you need to specify the type(full, lite) of snapshot you want to download.<br>
			- the snapshot is large, it may need a long time to finish the download, depends on your network performance. You could add 'nohup' to make it continue running even after you log out of your terminal session.
		`),
	Example: heredoc.Doc(`
			# Download target backup snapshot (backup20250205 in 34.143.247.77) to current directory.
			$ nohup ./trond snapshot download -d 34.143.247.77 -b backup20250205 -t lite &
		`),
	Run: func(cmd *cobra.Command, args []string) {
		// Get the flag value
		domain, _ := cmd.Flags().GetString("domain")

		// Get the flag value
		backup, _ := cmd.Flags().GetString("backup")

		// Get the flag value
		nType, _ := cmd.Flags().GetString("type")

		if !utils.CheckDomain(domain) {
			fmt.Println("Error: domain value not supported\nRun \"./trond snapshot source\" to check available items")
			return
		}

		download(domain, backup, nType)
	},
}

var downloadDefaultCmd = &cobra.Command{
	Use:   "default-main",
	Short: "Download latest mainnet lite fullnode snapshot from default source to current directory",
	Long: `This will download the latest snapshot from the default source to the current directory.

 - Default source: 34.143.247.77(Singapore)`,
	Example: heredoc.Doc(`
			# Download latest mainnet lite fullnode snapshot from default source(34.143.247.77) to current directory
			$ nohup ./trond snapshot download default-main &
		`),
	Run: func(cmd *cobra.Command, args []string) {
		// Get the flag value
		domain := "34.143.247.77"

		// Get the flag value
		backup, _ := utils.GetLatestSnapshot(domain, false)
		fmt.Println("Latest backup from 34.143.247.77 is:", backup)
		fmt.Println("Begin downloading...")

		// Get the flag value
		nType := "lite"

		if !utils.CheckDomain(domain) {
			fmt.Println("Error: domain value not supported\nRun \"./trond snapshot source\" to check available items")
			return
		}

		download(domain, backup, nType)
	},
}

var downloadDefaultNileCmd = &cobra.Command{
	Use:   "default-nile",
	Short: "Download latest nile testnet lite fullnode snapshot from default source to local current directory",
	Long: `This will download the latest snapshot from the default source to the current directory.

 - Default source: database.nileex.io`,
	Example: heredoc.Doc(`
			# Download latest nile testnet lite fullnode snapshot from default source(database.nileex.io) to current directory
			$ nohup ./trond snapshot download default-nile &
		`),
	Run: func(cmd *cobra.Command, args []string) {
		// Get the flag value
		domain := "database.nileex.io"

		// Get the flag value
		backup, _ := utils.GetLatestNileSnapshot(domain, false)
		fmt.Println("Latest backup from database.nileex.io is:", backup)
		fmt.Println("Begin downloading...")

		// Get the flag value
		nType := "lite"

		if !utils.CheckDomain(domain) {
			fmt.Println("Error: domain value not supported\nRun \"./trond snapshot source\" to check available items")
			return
		}

		download(domain, backup, nType)
	},
}

func download(domain string, backup string, nType string) {
	networkType, err := config.GetNetworkTypeFromDomain(domain)
	if err != nil {
		fmt.Printf("Abnormal domain: %v\n", err)
		return
	}

	networkDst := fmt.Sprintf("./output-directory/%s", strings.ToLower(string(networkType)))
	databaseDst := fmt.Sprintf("./output-directory/%s/database", strings.ToLower(string(networkType)))

	if folderYes, _ := utils.PathExists(networkDst); !folderYes {
		fmt.Println("Creating directory:", networkDst)
		if err := utils.CreateDir(networkDst, true); err != nil {
			fmt.Println(" - Error creating directory:", err)
			return
		}
	} else if databaseYes, _ := utils.PathExists(databaseDst); databaseYes {
		fmt.Println(fmt.Sprintf("Notice: there already exist database data %s, please delete it if you want to download again.", databaseDst))
		return
	}

	downloadSnapshot, err := downloadFileAndCheckSum(domain, backup, nType)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if err := utils.ExtractTgzWithStatus(downloadSnapshot, "./"); err != nil {
		fmt.Println("Error:", err)
		return
	}

	/* At least according to the domain, move extracted database to the correct subtype directory.
	The database will be used by `./trond node run-single`, check that part for details
	*/
	fmt.Println(fmt.Sprintf("Finally move the database to the corresponding network folder %s", databaseDst))
	srcDatabase := "./output-directory/database"

	// Check whether dst directory already exist, if yes, return warning let user
	// if not exist, create directory
	err = os.Rename(srcDatabase, databaseDst)
	if err != nil {
		fmt.Printf("Error moving directory: %v\n", err)
		return
	}
	fmt.Println("Directory moved to the corresponding network folder successfully")
}

func downloadFileAndCheckSum(domain string, backup string, nType string) (string, error) {
	downloadURLMD5 := utils.GenerateSnapshotMD5DownloadURL(domain, backup, nType)
	if downloadURLMD5 == "" {
		return "", errors.New("Error: --type value not supported, available: full, lite")
	}
	downloadMD5File, err := utils.DownloadFileWithProgress(downloadURLMD5, "")
	if err != nil {
		return "", err
	}
	downloadURL := utils.GenerateSnapshotDownloadURL(domain, backup, nType)
	if downloadURL == "" {
		return "", errors.New("Error: --type value not supported, available: full, lite")
	}
	return utils.DownloadFileWithProgress(downloadURL, downloadMD5File)
}

func init() {
	downloadCmd.AddCommand(downloadDefaultCmd)
	downloadCmd.AddCommand(downloadDefaultNileCmd)
	SnapshotCmd.AddCommand(downloadCmd)

	downloadCmd.Flags().StringP(
		"domain", "d", "",
		"Domain for target snapshot source(required).\nPlease run command \"./trond snapshot source\" to get the available snapshot source domains.")
	downloadCmd.Flags().StringP(
		"backup", "b", "",
		"Backup name(required).\nPlease run command \"./trond snapshot list\" to get the available backup name under target source domains.")
	downloadCmd.Flags().StringP(
		"type", "t", "",
		"Node type of the snapshot(required, available: full, lite).")

	// Mark source and destination as required flags
	if err := downloadCmd.MarkFlagRequired("domain"); err != nil {
		log.Fatalf("Error marking domain flag as required: %v", err)
	}
	if err := downloadCmd.MarkFlagRequired("backup"); err != nil {
		log.Fatalf("Error marking backup flag as required: %v", err)
	}
	if err := downloadCmd.MarkFlagRequired("type"); err != nil {
		log.Fatalf("Error marking type flag as required: %v", err)
	}
}
