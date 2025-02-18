# trond

`trond` is a command-line tool written in Golang using the [Cobra](https://github.com/spf13/cobra) framework, which is used for creating powerful modern CLI applications.
`trond` simplifies the configuration and operation of various `docker-tron` features, allowing you to start various features with a single command.

## Prerequisites
- Shell Environment: Required for running shell scripts.
- Golang Build Environment: The [./build_trond.sh](../../build_trond.sh) script will attempt to install the necessary Golang environment.

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
- `/trond docker`
  - Build and test the java-tron Docker image locally. Note that this is optional for deploying a node, as you can use the official Docker image from Docker Hub.
- `./trond node`
  - Deploy java-tron node for various networks.

### Examples

Here we show examples of downloading the mainnet database snapshot, then deploy a java-tron node based on it.
It is recommended to run these commands in `tron-docker` root directory.

#### 1. Download the Mainnet database snapshot

First, download the latest mainnet lite fullnode snapshot from the default source (`34.143.247.77`) to the current directory:

```
nohup ./trond snapshot download default-main &
```

After the download completes, the database will be extracted to `./output-directory/mainnet/database` of current directory.

#### 2. Deploy a java-tron node on Mainnet

Next, start a node connecting to the Mainnet:

```
./trond node run-single -t full-main &
```

This command will trigger execution of the [docker-compose](../../single_node/docker-compose.fullnode.main.yml). It will use the database at `./output-directory/mainnet/database`.
If the directory is empty, the node will sync transaction data from the genesis block. If you don’t need the snapshot, you can skip the download step.

To view the node logs, you can access them at `./logs/mainnet`, or you can use the command `docker exec tron-node tail -f ./logs/tron.log`.

---

For more detailed usage instructions, refer to the help command or the [command documentation](./docs/trond.md).

---

## Note

This tool has been tested on **macOS** and **Linux** only. If you encounter any issues, please report them on [GitHub](https://github.com/tronprotocol/tron-docker/issues).
