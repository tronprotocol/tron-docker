# trond

`trond` is a command-line tool written in Golang using the [Cobra](https://github.com/spf13/cobra) framework, which is used for creating powerful modern CLI applications.
`trond` simplifies the configuration and operation of various `docker-tron` features, allowing you to start various features with a single command.

## Prerequisites
- Shell Environment: Required for running shell scripts.
- **Go Language (v1.21.4+)**
    - Official download: <https://go.dev/dl>
    - Note: If Go is not installed, the [build_trond.sh](../../build_trond.sh) script will attempt to install it automatically, but it is recommended to configure it manually in advance.
- **Docker and Docker Compose**
    - Minimum version requirements:
        - Docker Engine ≥ 20.10.13
        - Docker Compose ≥ 1.29.2
    - Official installation guides: Check [prerequiesites](../../README.md#prerequisites)
- **Python 3.11.0+**
    - Official download: <https://www.python.org/downloads/release/python-3110/>

## Installation

1. Clone the `tron-docker` repository from GitHub and navigate to the repository directory:

    ```sh
    git clone https://github.com/tronprotocol/tron-docker.git
    cd tron-docker/
    ```

2. Build the `trond` binary. Run the following command to generate the trond executable:

    ```sh
    # this will generate trond in current directory
    ./build_trond.sh
    ```

## Usage
You can explore all the commands supported by trond using the following command:
```
./trond -h
```

### Available Commands
Now it supports the following commands:
- `./trond snapshot`
  - These commands are used to manage and download database snapshots, allowing you to start a node without syncing from the genesis block
- `./trond docker`
  - Build and test the java-tron Docker image locally. Note that this is optional for deploying a node, as you can use the official Docker image from Docker Hub.
- `./trond node`
  - Deploy java-tron node for various networks.

### Examples

Here we show examples of downloading the mainnet database snapshot, then deploy a java-tron node based on it.
You need to run these commands in `tron-docker` root directory.

#### 1. Download the Mainnet database snapshot

First, download the latest Mainnet database lite snapshot from the default source (`34.143.247.77`) to the current directory:

```
./trond snapshot download default-main
```

After the download completes, the database will be extracted to `./output-directory/mainnet/database` of current directory.

Notice: The snapshot is large(46G on 24-Jan-2025). It may take above 1 hour to finish. You could add `nohup` to make it continue running even after you log out of your terminal session.
The full command will be `nohup ./trond snapshot download default-main &`

#### 2. Check configure files
Next, check the configuration files and Docker compose files required for deploying a node:
```
./trond node env
```
As running a mainnet node,
you need file [docker-compose.fullnode.main.yml](../../single_node/docker-compose.fullnode.main.yml) and [main_net_config.conf](../../conf/main_net_config.conf).

#### 3. Deploy a java-tron node on Mainnet

Then, start a node connecting to the Mainnet:

```
./trond node run-single -t full-main
```

This command will trigger execution of the [docker-compose](../../single_node/docker-compose.fullnode.main.yml).
It will use the database at `./output-directory/mainnet/database` of the command trigger directory.
If the directory is empty, the node will sync transaction data from the genesis block. If you don’t need the snapshot, you can skip the download step.

To view the node logs, you can access them at `./logs/mainnet`, or you can use the command `docker exec tron-node tail -f ./logs/tron.log`.

#### 4. Stop the java-tron node
As java-tron service supports application shutdown with `kill -15`, which sends a `SIGTERM` signal to the application, allowing it to gracefully shut down. java-tron is also compatible with force shutdown using `kill -9`, which sends a `SIGKILL` signal.
You could stop the node in multiple ways.

Here is an example of stopping the node using `trond`:
```
./trond node run-single stop -t full-main
```


For more detailed usage instructions, refer to the help command or the [command documentation](./docs/trond.md).

## TroubleShooting
If you have any issues starting a java-tron node, please refer to the corresponding [TroubleShooting](../../single_node/README.md#troubleshot) guide.

This tool has been tested on **macOS** and **Linux** only. If you encounter any other issues, please report them on [GitHub](https://github.com/tronprotocol/tron-docker/issues).
