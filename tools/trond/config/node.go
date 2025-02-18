package config

import (
	"errors"
	"fmt"
)

type SnapshotType int

const (
	STFullLevelNA SnapshotType = iota // 1
	STFullLevelSG
	STFullLevelNAWithAccountHistory
	STFullRocksSG
	STLiteLevelSG
	STNileLevel
)

type NetworkType string
type DataType string

const (
	Network_Mainnet NetworkType = "Mainnet"
	Network_Nile    NetworkType = "Nile"

	DataType_Full DataType = "Fullnode Data Source"
	DataType_Lite DataType = "Lite Fullnode Data Source"
	DataType_All  DataType = "Fullnode/Lite Fullnode Data Source"
)

type SnapshotDataSourceItem struct {
	DataType    DataType
	DBType      string
	Region      string
	Domain      string
	DownloadURL string
	Description string
	NetworkType NetworkType
}

var SnapshotDataSource = map[SnapshotType]map[string]SnapshotDataSourceItem{
	STFullLevelNA: {
		"34.86.86.229": {
			DataType:    DataType_Full,
			DBType:      "LevelDB",
			Region:      "America",
			Domain:      "34.86.86.229",
			Description: "Exclude internal transactions (About 2094G on 25-Jan-2025)",
			NetworkType: Network_Mainnet,
		},
	},
	STFullLevelSG: {
		"34.143.247.77": {
			DataType:    DataType_Full,
			DBType:      "LevelDB",
			Region:      "Singapore",
			Domain:      "34.143.247.77",
			Description: "Exclude internal transactions (About 2093G on 24-Jan-2025)",
			NetworkType: Network_Mainnet,
		},
		"35.247.128.170": {
			DataType:    DataType_Full,
			DBType:      "LevelDB",
			Region:      "Singapore",
			Domain:      "35.247.128.170",
			Description: "Include internal transactions (About 2278G on 24-Jan-2025)",
			NetworkType: Network_Mainnet,
		},
	},
	STFullLevelNAWithAccountHistory: {
		"34.48.6.163": {
			DataType:    DataType_Full,
			DBType:      "LevelDB",
			Region:      "America",
			Domain:      "34.48.6.163",
			Description: "Exclude internal transactions, include account history TRX balance (About 2627G on 24-Jan-2025)",
			NetworkType: Network_Mainnet,
		},
	},
	STFullRocksSG: {
		"35.197.17.205": {
			DataType:    DataType_Full,
			DBType:      "RocksDB",
			Region:      "America",
			Domain:      "35.197.17.205",
			Description: "Exclude internal transactions (About 2067G on 24-Jan-2025)",
			NetworkType: Network_Mainnet,
		},
	},
	STLiteLevelSG: {
		"34.143.247.77": {
			DataType:    DataType_Lite,
			DBType:      "LevelDB",
			Region:      "Singapore",
			Domain:      "34.143.247.77",
			Description: "(About 46G on 24-Jan-2025)",
			NetworkType: Network_Mainnet,
		},
	},
	STNileLevel: {
		"database.nileex.io": {
			DataType:    DataType_All,
			DBType:      "LevelDB",
			Region:      "Singapore",
			Domain:      "database.nileex.io",
			DownloadURL: "https://nile-snapshots.s3-accelerate.amazonaws.com",
			Description: "Fullnode/Lite Fullnode (About 30G on 24-Jan-2025)",
			NetworkType: Network_Nile,
		},
	},
}

// Define custom error type for "not supported" errors
type NotSupportedError struct {
	Name  string
	Value string
}

func (e *NotSupportedError) Error() string {
	return fmt.Sprintf("%s '%s' is not supported", e.Name, e.Value)
}

func GetNetworkTypeFromDomain(domain string) (NetworkType, error) {
	for _, list := range SnapshotDataSource {
		for _, item := range list {
			if item.Domain == domain {
				return item.NetworkType, nil
			}
		}
	}
	return "", errors.New("Domain not found") // default Mainnet, as it has the most domain
}
