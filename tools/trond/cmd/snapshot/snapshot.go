package snapshot

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
)

// SnapshotCmd represents the snapshot command
var SnapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Commands for getting java-tron node snapshots.",
	Long: heredoc.Doc(`
		Commands used for downloading node's snapshot, such as:<br>
			1. show available snapshot source<br>
			2. list available snapshots in target source<br>
			3. download target snapshot
	`),
	Example: heredoc.Doc(`
			# Help information for snapshot command
			$ ./trond snapshot

			# Show available snapshot source
			$ ./trond snapshot source

			# List available snapshots of target source domain 34.143.247.77
			$ ./trond snapshot list -d 34.143.247.77

			# Download target backup snapshot (backup20250205 in 34.143.247.77) to current directory
			$ nohup ./trond snapshot download -d 34.143.247.77 -b backup20250205 -t lite &

			# Download latest mainnet lite fullnode snapshot from default source(34.143.247.77) to current directory
			$ nohup ./trond snapshot download default-main &

			# Download latest nile testnet lite fullnode snapshot from default source(database.nileex.io) to current directory
			$ nohup ./trond snapshot download default-nile &
		`),
}

func init() {
}
