# TRON Docker

This repository provides a quick and easy way to set up a single node, private chain, and monitor the status of nodes for the TRON Network using Docker.

## Features

### Quick start for single FullNode

This repository includes Docker configurations to quickly start a single TRON FullNode connected to the Mainnet or NileNet. Simply follow the instructions to get your node up and running in no time.

### Private chain setup

You can also use this repository to set up a private TRON blockchain network. This is useful for development and testing purposes. The provided configurations make it straightforward to deploy and manage your own private chain.

## Getting Started

### Prerequisites

- Docker
- Docker Compose

### Installation

1. **Clone the repository:**
   ```sh
   git clone https://github.com/tronprotocol/tron-docker.git
   cd tron-docker
   ```

2. **Start the services:**
   Navigate to the corresponding directory and follow the instructions in the respective README. Then you can easily start the services.
   - To start a single FullNode, use the folder [single_node](./single_node).
   - To set up a private TRON network, use the folder [private_net](./private_net).

## Troubleshooting
If you encounter any difficulties, please refer to the [Issue Work Flow](https://tronprotocol.github.io/documentation-en/developers/issue-workflow/#issue-work-flow), then raise an issue on [GitHub](https://github.com/tronprotocol/tron-docker/issues). For general questions, please use [Discord](https://discord.gg/cGKSsRVCGm) or [Telegram](https://t.me/TronOfficialDevelopersGroupEn).

# License

This repository is released under the [LGPLv3 license](https://github.com/tronprotocol/tron-docker/blob/main/LICENSE).
