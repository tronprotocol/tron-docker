# Quick Start for Single FullNode
Below are the steps for setting up a FullNode that connects to the TRON Network.

## Prerequisites

### Hardware requirements
Minimum:
- CPU with 8+ cores
- 16GB RAM
- 200GB free storage space to sync the Mainnet with Lite fullnode data snapshot
- 50 MBit/sec download Internet service

Recommended:
- Fast CPU with 16+ cores(32+ cores for a super representative)
- 32GB+ RAM(64GB+ for a super representative)
- High Performance SSD with at least 3TB free space for full data of Maninnet.
- 100+ MB/s download Internet service

### Docker

Please download and install the latest version of Docker from the official Docker website:
* Docker Installation for [Mac](https://docs.docker.com/docker-for-mac/install/)
* Docker Installation for [Windows](https://docs.docker.com/docker-for-windows/install/)
* Docker Installation for [Linux](https://docs.docker.com/desktop/setup/install/linux/)

Then check the Docker resource settings to ensure it has at least 16GB of memory.

## Get Docker image
There are two ways to obtain the TRON image:
- Pull it from the TRON Docker Hub.
- Build it from the java-tron source code.

### Using the official Docker images

The quick start way is to use the official images. Download the official Docker image from [Docker Hub](https://hub.docker.com/r/tronprotocol/java-tron/tags) using the following command:

```
docker pull tronprotocol/java-tron
```
Notice: To ensure your download has not been tampered with, you can use `docker images --digests` to compare the digest with the [officals](https://hub.docker.com/r/tronprotocol/java-tron/tags).

### Build from source code

Building java-tron requires the git package. Clone the repository and switch to the master branch with the following commands:
```
git clone https://github.com/tronprotocol/java-tron.git
cd java-tron
git checkout -t origin/master
```

Then use the following command to navigate to the docker directory and start the build:
```
cd docker
docker build -t tronprotocol/java-tron .
```
Check the Dockerfile for build details. Essentially, Docker will pull the java-tron repository and build it using JDK 1.8.

## Run the container

You can run the following command to start java-tron:
```
docker run -it --name tron -d --memory="16g" \
-p 8090:8090 -p 8091:8091 -p 18888:18888 -p 18888:18888/udp -p 50051:50051 \
tronprotocol/java-tron 
```
The `-p` flag specifies the ports that the container needs to map to the host machine.
`--memory="16g"` sets the memory limit to 16GB, ensuring that the TRON container gets enough memory. 

By default, it will use the [configuration](https://github.com/tronprotocol/java-tron/blob/develop/framework/src/main/resources/config.conf),
which sets the fullNode to connect to the mainnet with genesis block settings in `genesis.block`.
Once the fullnode starts, it will begin to sync blocks with other peers starting from genesis block.

Check the logs using command `docker exec -it tron tail -f ./logs/tron.log`. It will show the fullnode handshaking with peers successfully and then syncing for blocks.
For abnormal cases, please check the troubleshooting section below.

### Run with customized configure
This image also supports customizing some startup parameters. Here is an example for running a FullNode as a witness with a customized configuration file:
```
docker run -it --name tron -d -p 8090:8090 -p 8091:8091 -p 18888:18888 -p 18888:18888/udp -p 50051:50051 --memory="16g" \
           -v /host/path/java-tron/conf:/java-tron/conf \ 
           -v /host/path/java-tron/datadir:/java-tron/data \ 
           tronprotocol/java-tron \
           -jvm "{-Xmx10g -Xms10g}" \
           -c /java-tron/conf/config-localtest.conf \
           -d /java-tron/data \
           -w
```
The `-v` flag specifies the directory that the container needs to map to the host machine.
In the example above, the host file `/host/path/java-tron/conf/config-localtest.conf` will be used. For example, you can refer to the java-tron [config-localtest](https://github.com/tronprotocol/java-tron/blob/develop/framework/src/main/resources/config-localtest.conf).

Flags after `tronprotocol/java-tron` are used for java-tron start-up arguments:
- `-jvm` used for java virtual machine, the parameters must be enclosed in double quotes and braces. `"{-Xmx10g -Xms10g}"` sets the maximum and initial heap size to 10GB.
- `-c` defines the configuration file to use.
- `-d` defines the database file to use. You can mount a directory for `datadir` with snapshots. Please refer to [**Lite-FullNode**](https://tronprotocol.github.io/documentation-en/using_javatron/backup_restore/#_5). This can save time by syncing from a near-latest block number.
- `-w` means to start as a witness. You need to fill the `localwitness` field with private keys in configure file. Refer to the [**Run as Witness**](https://tronprotocol.github.io/documentation-en/using_javatron/installing_javatron/#startup-a-fullnode-that-produces-blocks). If you want to use keystore + password method, make sure the keystore is inside the mounted directory and remove `-d` to interact with console for password input.

Inside the `config-localtest.conf` file `node.p2p.version` is used to set the P2P network id. Only nodes with the same network id can shake hands successfully.
- TRON mainnet: node.p2p.version=11111
- Nile testnet: node.p2p.version = 201910292
- Private network：set to other values

Please note that if you want to switch to a different network, such as Mainnet or Nile, make sure you change the following:

- **Configuration File**:
  - For Mainnet, use [main_net_config.conf](https://github.com/tronprotocol/tron-docker/blob/main/conf/main_net_config.conf).
  - For NileNet, use the configuration file available on this [page](https://nileex.io/join/getJoinPage) or [nile_net_config.conf](https://github.com/tronprotocol/tron-docker/blob/main/conf/nile_net_config.conf).

  The main differences between these two files are:
  - `genesis.block`: Used for initial account asset and witness setup.
  - `seed.node`: Used for P2P nodes discovery.
  - `node.p2p.version`: Differentiates the network.
  - `block.maintenanceTimeInterval` and `block.proposalExpireTime`: Used for TRON core protocol.

- **Data Snapshot**:
  - Ensure that the data snapshot you download corresponds to the correct network.

### Close Docker application
java-tron supports application shutdown with `kill -15`, which sends a `SIGTERM` signal to the application, allowing it to gracefully shut down. java-tron is also compatible with force shutdown using `kill -9`, which sends a `SIGKILL` signal.

Thus, you can use the command `docker stop <container_id>` or `docker kill <container_id>` to close the java-tron container.

## Interact with FullNode
After the fullnode runs successfully, you can interact with it using the HTTP API or wallet-cli. For more details, please refer to [guidance](https://tronprotocol.github.io/documentation-en/getting_started/getting_started_with_javatron/#interacting-with-java-tron-nodes-using-curl).

For example, a request to get block info with a specific number:
```
curl --location 'localhost:8090/wallet/getblock' \
--header 'Content-Type: application/json' \
--data '{
    "id_or_num": "100",
    "detail": true
}'
```
Response:
```
{
    "blockID": "00000000000000644df09e6883a3a7900814f8d78cf47b255b7ed284527a773d",
    "block_header": {
        "raw_data": {
            "number": 100,
            "txTrieRoot": "0000000000000000000000000000000000000000000000000000000000000000",
            "witness_address": "414b4778beebb48abe0bc1df42e92e0fe64d0c8685",
            "parentHash": "0000000000000063ed8544c4c17fc053dfc729e610673c783bcdc3cf0781b07f",
            "timestamp": 1529891811000
        },
        "witness_signature": "277d4440e2feb552b6d2d557ba407f68310887020fcc7ef6e2733286a0d13c703ebf2306293bda9d2ddac09835be67583c736a65494115825b6f4ab6a15f1e0f01"
    }
}
```
**Notice**: Before the local full node has synced with the latest block transactions, requests for account state or transaction information may be outdated or empty.

## Troubleshot
After starting the Docker container, use `docker exec -it tron tail -f ./logs/tron.log` to check if the full node is functioning as expected and to identify any errors when interacting with the full node.

If the following error cases do not cover your issue, please refer to [Issue Work Flow](https://tronprotocol.github.io/documentation-en/developers/issue-workflow/#issue-work-flow), then raise issue in [Github](https://github.com/tronprotocol/tron-docker/issues).

### Error case handling

#### Zero peer connection
If the logs show `Peer stats: all 0, active 0, passive 0`, it means tron node cannot use **P2P Node Discovering Protocol** to find neighbors.
This protocol operates over UDP through port 18888. Therefore, the most likely cause of this issue is a network problem.
Try debugging with the following steps:
- Use the command `docker ps` to check if the ports mapping includes `-p 18888:18888`.
- Verify your local network settings to ensure that port 18888 is not blocked.
- Open the Docker application, navigate to Settings -> Resources -> Network, and check the option 'Use kernel networking for UDP'. Then restart Docker and your container.
