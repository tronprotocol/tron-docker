package utils

import (
	"fmt"

	"github.com/tronprotocol/tron-docker/config"
)

func ShowSnapshotDataSourceList() {

	var MainLiteList []config.SnapshotDataSourceItem
	var MainFullList []config.SnapshotDataSourceItem
	var NileList []config.SnapshotDataSourceItem
	for _, list := range config.SnapshotDataSource {
		for _, item := range list {
			if item.NetworkType == config.Network_Mainnet && item.DataType == config.DataType_Lite {
				MainLiteList = append(MainLiteList, item)
			} else if item.NetworkType == config.Network_Mainnet && item.DataType == config.DataType_Full {
				MainFullList = append(MainFullList, item)
			} else if item.NetworkType == config.Network_Nile && item.DataType == config.DataType_All {
				NileList = append(NileList, item)
			}
		}
	}
	displayOrder := map[string][]config.SnapshotDataSourceItem{}
	displayOrder["Nile network - Fullnode/Lite Fullnode Data Source:"] = NileList
	displayOrder["Main network - Lite Fullnode Data Source:"] = MainLiteList
	displayOrder["Main network - Fullnode Data Source:"] = MainFullList

	for title, list := range displayOrder {
		fmt.Println(title)
		for _, item := range list {
			fmt.Printf("  Region: %s\n", item.Region)
			fmt.Printf("    DBType: %s\n", item.DBType)
			fmt.Printf("    Domain: %s\n", item.Domain)
			fmt.Printf("    Description: %s\n\n", item.Description)
		}
	}
}

func CheckDomain(domain string) bool {
	has := false

	for _, items := range config.SnapshotDataSource {
		for _, item := range items {
			if item.Domain == domain {
				return true
			}
		}
	}

	return has
}

func IsNile(domain string) bool {
	has := false

	for mtype, items := range config.SnapshotDataSource {
		for _, item := range items {
			if item.Domain == domain {
				return mtype == config.STNileLevel
			}
		}
	}

	return has
}
