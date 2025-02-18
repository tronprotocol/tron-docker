# Private Network

Here is a quick-start guide for setting up a TRON private network using Docker.

A private chain needs at least one fullnode run by a [Super Representative (SR)](https://tronprotocol.github.io/documentation-en/mechanism-algorithm/sr/) to produce blocks, and any number of fullnodes to synchronize blocks and broadcast transactions.

## Prerequisites

### Minimum hardware requirements
- CPU with 8+ cores
- 16GB RAM
- 100GB free storage space

### Docker

For Docker and Docker Compose installation refer [prerequisites](../README.md#prerequisites).
Then check the Docker resource settings to ensure it has at least 3GB of memory per TRON node.

**Notice:** The actual memory consumption for a FullNode depends heavily on your configurations and use case. Factors such as block generation frequency, transactions per second (TPS), and external API request QPS can significantly impact memory usage. To help you get started quickly, the memory consumption under the following guidance is within **3GB**. However, for a TRON Mainnet Super Representative (SR) FullNode capable of supporting **2000 TPS**, the memory resource per FullNode container should set above **16GB**.


## Quick-Start using Docker
Download the folder [private_net](../private_net) from GitHub. Enter the directory and run the docker-composer.
```
cd ./private_net
docker-compose up -d
```
A TRON private network will be started with one [SR](https://tronprotocol.github.io/documentation-en/mechanism-algorithm/sr/#super-representative) and a normal FullNode.

Check the witness logs by running the command below:
```
docker exec tron-witness1 tail -f ./logs/tron.log
```
Normally, it should show the witness initializing the database and network, then starting to produce blocks every 3 seconds.
```
01:58:06.013 INFO  [DPosMiner] [DB](Manager.java:1546) Generate block 1 begin.
01:58:06.158 INFO  [DPosMiner] [DB](Manager.java:1669) Generate block 1 success, trxs:0, before pendingCount: 0, rePushCount: 0, from pending: 0, rePush: 0, after pendingCount: 0, rePushCount: 0, postponedCount: 0, blockSize: 175 B
01:58:06.162 INFO  [DPosMiner] [net](AdvService.java:200) Ready to broadcast block Num:1,ID:00000000000001206a3e24f26aa0c31033349e2cffab07f741061728a79a55b3
01:58:06.183 INFO  [DPosMiner] [DB](Manager.java:1233) Block num: 1, re-push-size: 0, pending-size: 0, block-tx-size: 0, verify-tx-size: 0
```
It also connects with the other full nodes with the log:

```
Peer stats: channels 1, activePeers 1, active 0, passive 1
```

Check the other fullnode logs by running the command below:
```
docker exec tron-node1 tail -f ./logs/tron.log
```
After initialization, it should show messages about syncing blocks, just following the SR.

**What docker-compose do?**

Check the docker-compose.yml, the two container services use the same TRON image with different configurations.

- `ports`: Used in the tron-witness1 service are exposed for API requests to interact with the TRON private network.

- `command`: Used for java-tron image start-up arguments.
    - `-jvm` is used for Java Virtual Machine parameters, which must be enclosed in double quotes and braces. `"{-Xmx12g -Xms3g}"` sets the maximum and minimum heap size to 12GB and 3GB.
    - `-c` defines the configuration file to use.
    - `-d` defines the database file to use. By mounting a local data directory, it ensures that the block data is persistent.
    - `-w` means to start as a witness. You need to fill the `localwitness` field with private keys in the configuration file. Refer to [**Run as Witness**](https://tronprotocol.github.io/documentation-en/using_javatron/installing_javatron/#startup-a-fullnode-that-produces-blocks).

## Run with customized configure

If you want to add more witness or other syncing fullnodes, you need to make below minimum changes for docker-compose.yml and configuration files.

**Add more services in docker-compose.yml**

Inside the docker-compose.yml, refer to the commented containers `tron-witness2` and `tron-node2`. Make sure the configuration files are changed accordingly, following the details below.

**Common settings**

For all configurations, you need to set `node.p2p.version` to the same value and `node.discovery.enable = true`.
```
node {
 p2p {
    version = 1 # 11111: Mainnet; 20180622: Nile testnet; others for private networks.
  }
  ...
}

node.discovery = {
  enable = true  # you should set this entry value with true if you want your node can be discovered by other node.
  ...
}
```

**Witness setting**

Make sure only one SR witness sets `needSyncCheck = false`, while the rest of the witnesses and other fullnodes set it to `true`. This ensures that there is only one source of truth for block data.
```
block = {
  needSyncCheck = true # only one SR witness set false, the rest all true
  ...
}
```

If you want to add more witnesses:

- First, add the witness private key to the `localwitness` field in the corresponding witness configuration file. If you don't want to use this way of specifying the private key in plain text, you can use the [keystore + password](https://tronprotocol.github.io/documentation-en/using_javatron/installing_javatron/#others) method.
- Then, add initial values to the `genesis.block` for all configuration files. TRON will use this to initialize the genesis block, and nodes with different genesis blocks will be disconnected.

```
localwitness = [
  # public address TCjptjyjenNKB2Y6EwyVT43DQyUUorxKWi
  0ab0b4893c83102ed7be35eee6d50f081625ac75a07da6cb58b1ad2e9c18ce43  # you must enable this value and the witness address are match.
]

genesis.block {
   assets = [ # set account initial balance
   ...
      {
          accountName = "TestE"
          accountType = "AssetIssue"
          address = "TCjptjyjenNKB2Y6EwyVT43DQyUUorxKWi"
          balance = "1000000000000000"
      }
   ]
   witnesses = [ # set witness account initial vote count
    ...
    {
      address: TCjptjyjenNKB2Y6EwyVT43DQyUUorxKWi,
      url = "http://example.com",
      voteCount = 5000
    }
  ]

```

**P2P node discovery setting**

In witness configure file, make sure `node.listen.port` is set for p2p peer discovery.
```
node {
  listen.port = 18888
  ...
}
```

Then, in other configuration files, add witness `container_name:port` to connect to the newly added witness fullnodes.
```
seed.node = {
  ip.list = [
    # used for docker deployment, to connect containers in tron-witness defined in docker-compose.yml
    "tron-witness1:18888",
    "tron-witness2:18888",
    ...
  ]
}
```
### Advanced configuration
Besides the above settings, there are many fields you can modify to customize the behavior of private networks. The configurations that you can change, though not exhaustively listed, include:

- Ethereum compatible virtual machine in `vm = {...}`
- Block settings
- Network details for node discovery and connections, HTTP and RPC services, etc.
- Enable or disable parts of committee-approved proposals
- [Event subscription](https://tronprotocol.github.io/documentation-en/architecture/event/#configure-node)
- [Database configuration](https://tronprotocol.github.io/documentation-en/architecture/database/#database-configuration)

**Notice**: Make sure your changes are consistent among all configuration files, especially for the SRs, as inconsistencies may affect the block generation logic.

For example, you could change these block settings to smaller values to speed up maintenance or proposal logic changes:
```
block = {
  maintenanceTimeInterval = 300000 # 5mins, default is 6 hours
  proposalExpireTime = 600000 # 10mins, default is 3 days
}
```
You could also disable/enable the following committee-approved settings with `0` or `1`:
```
committee = {
  allowCreationOfContracts = 1
  allowAdaptiveEnergy = 0
  allowMultiSign = 1
  allowDelegateResource = 1
  allowSameTokenName = 0
  allowTvmTransferTrc10 = 1
}
```
If you encounter any difficulties or need more customized operations, check the Troubleshooting section below.

### Close Docker application
java-tron supports application shutdown with `kill -15`, which sends a `SIGTERM` signal to the application, allowing it to gracefully shut down. java-tron is also compatible with force shutdown using `kill -9`, which sends a `SIGKILL` signal.

Thus, you can use the command `docker-compose stop/down` or `docker-compose kill` to stop or close all the services.

## Interact with TRON private network
In docker-compose.yml, notice the ports mapping:
```
ports:
- "8090:8090"       # for external HTTP API requests
- "50051:50051"     # for external RPC API requests, for example through wallet-cli
```
After the network runs successfully, you can interact with it using the HTTP API or wallet-cli.

For example, a request to get genesis block info:
```
curl --location 'localhost:8090/wallet/getblock' \
--header 'Content-Type: application/json' \
--data '{
    "id_or_num": "0",
    "detail": true
}'
```
It should return transactions with the type `TransferContract` for asset initialization set in the `genesis.block` in the configuration files. Notice the `parentHash` equals to the value you set in configure file.
```
{
    "blockID": "0000000000000000ad12b5787243011231c77e867e36f54ecb20b4b38ac594b8",
    "block_header": {
        "raw_data": {
            "txTrieRoot": "bf467bbfc436b1690f1c5d0d650ac4012fa2ff304f4e99bd487f90c8718e65ca",
            "witness_address": "41206e65772073797374656d206d75737420616c6c6f77206578697374696e672073797374656d7320746f206265206c696e6b656420746f67657468657220776974686f757420726571756972696e6720616e792063656e7472616c20636f6e74726f6c206f7220636f6f7264696e6174696f6e",
            "parentHash": "957dc2d350daecc7bb6a38f3938ebde0a0c1cedafe15f0edae4256a2907449f6"
        }
    },
    "transactions": [
       {
         "txID": "6a617c0667fd3a710a95d78d212a6e72df3c005ff7d3f53d3765cc81156151ea",
            "raw_data": {
                "contract": [
                    {
                        "parameter": {
                            "value": {
                                "amount": 95000000000000000, // initial balance
                                "owner_address": "3078303030303030303030303030303030303030303030",
                                "to_address": "41928c9af0651632157ef27a2cf17ca72c575a4d21" // hex format = public address TPL66VK2gCXNCD7EJg9pgJRfqcRazjhUZY
                            },
                            "type_url": "type.googleapis.com/protocol.TransferContract"
                        },
                        "type": "TransferContract"
                    }
                ]
            },
            "raw_data_hex": "5a6f0801126b0a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e7472616374123a0a173078303030303030303030303030303030303030303030121541928c9af0651632157ef27a2cf17ca72c575a4d21188080a6adf2bfe0a801"
        },
        ...
    ]

```
If you request block info with a number greater than 0, it should return block_header only, as there are no transactions triggered.

For more API usage, please refer to the [guidance](https://tronprotocol.github.io/documentation-en/getting_started/getting_started_with_javatron/#interacting-with-java-tron-nodes-using-curl).

As you could notice, for now the block produced by SR contains 0 transaction. To send a transaction on TRON network, the data must be signed. To easily trigger a transaction on the TRON network, [wallet-cli](https://tronprotocol.github.io/documentation-en/clients/wallet-cli/) is recommended. Refer to the installation [guidance](https://github.com/tronprotocol/wallet-cli/blob/develop/README.md). Make sure you edit `config.conf` in [src/main/resources](https://github.com/tronprotocol/wallet-cli/blob/develop/src/main/resources/config.conf) as below:
```
fullnode = {
  ip.list = [
    "localhost:50051" # or any value private network hosted
  ]
}
```
Wallet-cli will connect to your local private network. After registering or importing a wallet, you can easily sign and broadcast transactions. Check wallet-cli API usage [here](https://tronprotocol.github.io/documentation-en/clients/wallet-cli-command/#registerwallet).

## Troubleshooting
If you encounter any difficulties, please refer to the [Issue Work Flow](https://tronprotocol.github.io/documentation-en/developers/issue-workflow/#issue-work-flow), then raise an issue on [GitHub](https://github.com/tronprotocol/tron-docker/issues). For general questions, please use [Discord](https://discord.gg/cGKSsRVCGm) or [Telegram](https://t.me/TronOfficialDevelopersGroupEn).

# Advanced
To set up a private network natively, refer to the [Deployment Guide](https://tronprotocol.github.io/documentation-en/using_javatron/private_network/). Make sure you set up all the configuration parameters correctly.
