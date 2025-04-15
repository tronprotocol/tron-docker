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
  - `query`: Query the latest vote and reward information from the database.
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

## Official Docker image
There are two official Docker images for java-tron:

- `tronprotocol/java-tron`: This image is based on the official [java-tron](https://github.com/tronprotocol/java-tron) repository, used for the Mainnet, and mostly can be used for the Nile testnet too. (**All the demo in this repository are based on this image**)
- `tronnile/java-tron`: This image is based on the [nile-testnet](https://github.com/tron-nile-testnet/nile-testnet) repository which is forked from [java-tron](https://github.com/tronprotocol/java-tron). You need to use this image for the Nile testnet when there is new test release from java-tron and before it was released, especially for the hard fork test on Nile. Information about code release of the Nile testnet, please refer to this [website](https://nileex.io/).

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
    "blockID": "00000000002ceed0bf4668aae518d143dfe1b74eea037a1caed98e0d4ea00de4",
    "block_header": {
        "raw_data": {
            "number": 2944720,
            "txTrieRoot": "d9c4fbe7c210ada548ac54460d717a5fce70d621b1e3da1db425672cbf477ca5",
            "witness_address": "41f29f57614a6b201729473c837e1d2879e9f90b8e",
            "parentHash": "00000000002ceecf692df37b4300746b9c68f7bdfad6d15a74859ceaff852f90",
            "version": 1,
            "timestamp": 1538758257000
        },
        "witness_signature": "445ad3d0585e24b60a9bceec0d97ac70b1558f42b44d24ef62f400a7c2646a271f01b83de985c5796485f74fd69d7b11c50b9ce44ad193fa35e26b317d3e339401"
    },
    "transactions": [
        {
            "signature": [
                "948c3cbe1b48a20661878e591cf4c7fea7d1c63836b1523f5b0b003a290629714bfa4f4e6e6398a8edbca076b9c263c36b226b0f5a046ca3d54098b9e95d1c8e00"
            ],
            "txID": "25e0b74bcd0e7b2691945cce61ccafdf0be879c21fc729e6c25646350b611ad6",
            "raw_data": {
                "contract": [
                    {
                        "parameter": {
                            "value": {
                                "amount": 1,
                                "asset_name": "49504653",
                                "owner_address": "41d13433f53fdf88820c2e530da7828ce15d6585cb",
                                "to_address": "414f93486fa5d2685c70a3dd7e0815769a9fd3f02b"
                            },
                            "type_url": "type.googleapis.com/protocol.TransferAssetContract"
                        },
                        "type": "TransferAssetContract"
                    }
                ],
                "ref_block_bytes": "eece",
                "ref_block_hash": "36ef820a1c8df903",
                "expiration": 1538758311000,
                "timestamp": 1538758251854
            },
            "raw_data_hex": "0a02eece220836ef820a1c8df90340d8f891a9e42c5a700802126c0a32747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e736665724173736574436f6e747261637412360a0449504653121541d13433f53fdf88820c2e530da7828ce15d6585cb1a15414f93486fa5d2685c70a3dd7e0815769a9fd3f02b200170ceaa8ea9e42c"
        },
        ...
    ]
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
We offer a `trond` command-line tool that allows developers to easily initiate features with a single command, enabling the community to quickly engage in TRON network development and interaction.

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
   - **Node monitoring**: Use the [metric_monitor](./metric_monitor) folder.  We provide two solutions: one with enhanced security considerations and another with a simpler setup.

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
