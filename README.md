# TRON Docker

This repository provides tools and guidance to help the community quickly get started with the TRON network and development.

## Features

### 🚀 Quick start for single FullNode
Easily deploy a single TRON FullNode connected to **Mainnet** or **Nile testnet** using Docker. Follow the instructions to get your node up and running in minutes.

### 🔗 Private chain setup
Set up your own private TRON blockchain network for development and testing. The provided configurations simplify deployment and management, making it ideal for custom use cases.

### 📊 Node monitoring with Prometheus and Grafana
Monitor the health and performance of your TRON nodes with integrated **Prometheus** and **Grafana** services. Real-time metrics and visualizations are just a few steps away.

### 🛠️ Tools
We also provide tools to facilitate the CI and testing process:
- **Gradle Docker**: Automate the build and testing of the `java-tron` Docker image using Gradle.
- **Toolkit**: This package contains a set of database tools for TRON:
  - `mv, move`: Move db to pre-set new path. For example HDD,reduce storage
  expenses.
  - `archive`: A helper to rewrite leveldb manifest.
  - `convert`: Covert leveldb to rocksdb.
  - `lite`: Split lite data for java-tron.
  - `cp, copy`: Quick copy leveldb or rocksdb data.
  - `root`: compute merkle root for tiny db. NOTE: large db may GC overhead
  limit exceeded.
  - `fork`: Modify the database of java-tron for shadow fork testing.
- **Stress Test**: Execute the stress test and evaluate the performance of the `java-tron` fullnode.


## Prerequisites
Please ensure you have the latest versions of Docker and Docker Compose installed by downloading them from the official websites:

- **For Mac:**
  Download Docker from [Docker Desktop for Mac](https://docs.docker.com/docker-for-mac/install/).
  Docker Compose is included in the Docker installation package for Mac.
  - After installation, open Docker application, navigate to Settings \-> Resources \-> Network, and check the option \`Use kernel networking for UDP\`. Then restart Docker to apply.

- **For Linux:**
  Download and install both Docker and Docker Compose plugin from the official websites:
  - Docker: [Install Docker on Linux](https://docs.docker.com/desktop/setup/install/linux/)
  - Docker Compose standalone: [Install Docker Compose on Linux](https://docs.docker.com/compose/install/standalone/)

## Quick Start
To quickly start a java-tron node that connects to the Mainnet, simply use the following Docker command:

```sh
docker run -it --name tron-node -d --memory="16g" -p 8090:8090 -p 50051:50051 tronprotocol/java-tron
```

Alternatively, you can download and run the [docker-compose](single_node/docker-compose-quick-start.yml) file using the command:

```sh
docker-compose -f docker-compose-quick-start.yml up
```

Once the FullNode starts, it will begin to sync blocks with Mainnet from the genesis block. You can use the following API request to check the current synced blocks:
```
Request:
curl --location --request POST 'http://127.0.0.1:8090/wallet/getnowblock'

Response example:
{
    "blockID": "0000000000000a611841e6cd79f99249d7e974e41c0f5b016eb3c64e43fcb01c",
    "block_header": {
        "raw_data": {
            "number": 2657,
            "txTrieRoot": "0000000000000000000000000000000000000000000000000000000000000000",
            "witness_address": "411661f25387370c9cd3a9a5d97e60ca90f4844e7e",
            "parentHash": "0000000000000a6023d563b1a2e1e0eb49cb40226eb553e6dc006e723395d588",
            "timestamp": 1529899509000
        },
        "witness_signature": "1710dc914cb846f595455d2b6ace7470b26cb29c8b3b9424156b1bd8c6fc6488337cb12b5a2e72e4e8d70838e927e0fb4f2c787dc6ad2d6eaf28314c2f75b73300"
    }
}

```
You can also use `docker exec tron-node tail -f ./logs/tron.log` to check the node and syncing status.

For more details on using database snapshort or customized configurations, please refer to the [single_node](single_node/README.md) section.

## Start all features
First, clone the repository:

```sh
git clone https://github.com/tronprotocol/tron-docker.git
cd tron-docker
```

### Start feature by trond tool
We offer a `trond` script that allows developers to easily initiate features with a single command, enabling the community to quickly engage in TRON network development and interaction.

To build the `trond` command-line tool, simply run `build_trond.sh`.
```sh
# this will generate trond in current directory
./build_trond.sh
```
After that, you can explore all the commands it supports.
```
./trond -h
Usage:
  trond [command]

Examples:
# Help information for java-tron Docker image build and testing commands
$ ./trond docker

# Help information for database snapshot download-related commands
$ ./trond snapshot

# Help information for java-tron node deployment commands
$ ./trond node

Available Commands:
  docker      Commands for operating the java-tron Docker image.
  help        Help about any command
  node        Commands for operating the java-tron Docker node.
  snapshot    Commands for obtaining java-tron node snapshots.
```
Check the output of `./trond -h`, which now supports the following features:
- Download database snapshots. It is used to save time by avoiding the need to sync from the genesis block.
- Build and test the java-tron Docker image locally. Note that it is not mandatory for deploying a node, as you can use the Docker image available on the official Docker Hub.
- Deploy a single FullNode for various networks.

For more details on the `trond` command-line tools and how they work, please refer to the [README](./tools/trond/README.md).

### Start the feature individually
To start all available features, or you want more customized operations, navigate to the respective directory and follow the instructions in the corresponding README to start the services:
- **TRON network deployment related:**
   - **Single FullNode**: Use the [single_node](./single_node) folder.
   - **Private TRON network**: Use the [private_net](./private_net) folder.
   - **Node monitoring**: Use the [metric_monitor](./metric_monitor) folder.

- **Tools**:
   - **Gradle Docker**: Automate Docker image builds and testing. Check the [gradle docker](./tools/docker/README.md) documentation.
   - **Toolkit**: Perform a set of database related operations. Follow the [Toolkit guidance](./tools/toolkit/README.md).
   - **Stress Test**: Execute the stress test. Follow the [stress test guidance](./tools/stress_test/README.md).

## Troubleshooting
If you encounter any difficulties, please refer to the [Issue Work Flow](https://tronprotocol.github.io/documentation-en/developers/issue-workflow/#issue-work-flow), then raise an issue on [GitHub](https://github.com/tronprotocol/tron-docker/issues). For general questions, please use [Discord](https://discord.gg/cGKSsRVCGm) or [Telegram](https://t.me/TronOfficialDevelopersGroupEn).

## Contributing

All contributions are welcome. Check [contribution](CONTRIBUTING.md) for more details.

## License

This repository is released under the [LGPLv3 license](https://github.com/tronprotocol/tron-docker/blob/main/LICENSE).
