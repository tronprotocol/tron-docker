## Stress Test Tool
The stress testing tool is designed to evaluate the performance of the `java-tron` fullnode.
It can generate a large volume of transactions, store them locally, and broadcast them to the test network.
Finally, it provides a TPS (Transactions Per Second) report as the stress test result.

### Build the stress test tool
To build the stress test tool, we need to execute the following commands:
```shell script
# clone the tron-docker
git clone https://github.com/tronprotocol/tron-docker.git
# enter the directory
cd tron-docker/tools/gradlew
# compile the stress test tool
./gradlew :stress-test:build
# execute full command
java -jar ../stress_test/build/libs/stresstest.jar help
```
The stress test tool includes four components:
- `collect`: Collect the address list from the account database.
- `generate`: Generate plenty of transactions used the in the stress test.
- `broadcast`: Broadcast the transactions and compute the TPS.
- `statistic`: Compute the TPS from specified block range.

All the configurations of the components are placed in the `stress.conf`, please refer [stress.conf](./src/main/resources/stress.conf)
as an example.

### Collect Address List
`collect` subcommand is used to collect the addresses from the database, which act as `to` addresses when generating
the transactions. The corresponding configuration is:
```
collectAddress = {
  total = 1000000
  dbPath = "/path/to/output-directory"
}
```
- `total`: denotes the total addresses number we need to collect.
- `dbPath`: denotes the database path

Then we can execute the following `collect` subcommand:

```shell
# execute full command
java -jar /path/to/stresstest.jar collect -c /path/to/stress.conf
# check the log
tail -f logs/stress_test.log
```
The collected addresses are stored in `address-list.csv` file of the current directory.

### Generate the transactions
`generate` subcommand is used to generate plenty of transactions used for the stress test.
The corresponding configuration is:
```
generateTx = {
  enable = true
  totalTxCnt = 600000
  singleTaskTxCount = 100000
  txType = {
    transfer = 60
    transferTrc10 = 10
    transferTrc20 = 30
  }
  updateRefUrl = "127.0.0.1:50051"
  // TRY18iTFy6p8yhWiCt1dhd2gz2c15ungq3
  privateKey = "aab926e86a17f0f46b4d22e61725edd5770a5b0fbdabb04b0f46ee499b1e34f2"
  addressListFile = "/path/to/address-list.csv"
  trc10Id = 1000001
  trc20Address = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
}
```
Here is the introduction for the configuration options:

`enable`: configure whether to generate the transactions;

`totalTxCnt`: configure the total generated transactions count;

`singleTaskTxCount`: configure the transaction count for single task;

`txType`: configure the generated transaction type and proportion. Currently, the supported transaction type
including `transfer`, `transferTrc10`, `transferTrc20`. The sum of all transaction type proportion must be equal 100;

`updateRefUrl`: configure the url which is used to update the `refBlockNum` and `refBlockNum` when generating the transactions;

`privateKey`: configure the private key used to sign the transactions;

`addressListFile`: configure the file path of receiver address list used to build the transactions;

`trc10Id`: configure the TRC10 id used to build the `transferTrc10` transactions;

`trc20ContractAddress`: configure the TRC20 contract address used to build the `transferTrc20` transactions;

Then we can execute the following `generate` subcommand:

```shell
# execute full command
nohup java -jar /path/to/stresstest.jar generate -c /path/to/stress.conf >> start.log 2>&1 &
# check the log
tail -f logs/stress_test.log
```
The generated transactions are stored in the `generate-tx*.csv` files in current `stress-test-output` directory.

*Note*: the expiration time of the generated transactions is 24 hours, which means you need to broadcast the stored transactions in 24 hours.

### Broadcast the transactions
`broadcast` subcommand is used to broadcast the transactions and compute the TPS for the stress test.
The corresponding configuration is:
```
broadcastTx = {
  generateTx = true
  relayTx = false
  tpsLimit = 3000
  saveTxId = true
}
```
- `generateTx`: configure whether to broadcast the generated the transactions;

- `relayTx`: configure whether to broadcast the relayed transactions;

- `broadcastUrl`: configure the broadcast url list;

- `tpsLimit`: configure the maximum broadcast transactions per second;

- `saveTxId`: configure whether to save the transaction id of the broadcast transactions.

*Note*: we can use the [dbfork](../toolkit/DbFork.md) tool to get enough `TRX/TRC10/TRC20` balances of address corresponding
  to the `privateKey` for the stress test.

Then we can execute the following `generate` subcommand:

```shell
# execute full command
nohup java -jar ../stress_test/build/libs/stresstest.jar broadcast -c /path/to/stress.conf \
--fn-config /path/to/config.conf -d /path/to/output-directory >> start.log 2>&1 &
# check the log
tail -f logs/stress_test.log
```
- `--fn-config`: configure the `java-tron` network we need to connect, please refer [config.conf](./src/main/resources/config.conf)
  as an example. Make sure to configure the correct `p2p.version` and `seed.node` options.

- `-d`: configure the `java-tron` database. Before executing the broadcasting process, `broadcast` component
  needs to sync with the connected network. So we'd better copy the database from the connected network.

If you set `saveTxId = true`, the broadcast transactions ids will be stored
in the `broadcast-txID*.csv` files in current `stress-test-output` directory.

After broadcasting all the transactions, it will generate the `stress-test-output/broadcast-generate-result`
file to report the stress-test statistic result. For example:
```
Stress test report:
broadcast tps limit: 3000
statistic block range: startBlock: 67926067, endBlock: 67926133
total generate tx count: 600000, total broadcast tx count: 580862, tx on chain rate: 0.968103
cost time: 3.300000 minutes
max block size: 9615
min block size: 3001
tps: 2933.646484
miss block rate: 0.000000
```
The above the result shows the stress test TPS has reached `2933`.

## Relay and Broadcast
If you want to relay the transactions from other network, you need to set `relayTrx.enable = true` and
other related the parameters：
```
relayTx = {
  enable = false
  url = "grpc.trongrid.io:50051"
  startBlockNumber = 59720000
  endBlockNumber = 59720500
}
```
- `enable`: configure whether to relay the transactions from other network and save them locally.
- `url`: configure the url to indicate the network the relayed transactions come from;
- `startBlockNumber`: configure the start block number of range for the relayed transactions;
- `endBlockNumber`: configure the end block number of range for the relayed transactions;

Then we can execute the `generate` subcommand.
```shell
# execute full command
nohup java -jar /path/to/stresstest.jar generate -c /path/to/stress.conf >> start.log 2>&1 &
# check the log
tail -f logs/stress_test.log
```
The relayed transactions will be stored in the `relay-tx.csv` file in current `stress-test-output` directory.

To broadcast the relayed transactions, we need set `relayTx = true` and execute the `broadcast` subcommand,
which will broadcast the transactions stored in the `relay-tx.csv` file.

*Warn*: Most of the relayed transactions may be illegal in the stress test network. You need to change the
transaction verification condition in `java-tron` source code to replay the transactions.

## TPS statistic
If the above the `broadcast` subcommand doesn't output `broadcast-generate-result` statistic result,
we can still compute the TPS from specified block range by executing the `statistic` subcommand:

```
statistic = {
  url = "127.0.0.1:50051"
  startBlockNumber = 68599177
  endBlockNumber = 68599689
}
```
- `url`: configure the url to indicate the network for the TPS statistic;
- `startBlockNumber`: configure the start block number of range for the TPS statistic;
- `endBlockNumber`: configure the end block number of range for the TPS statistic;

Then we can execute the `statistic` subcommand.
```
# execute full command
java -jar ../stress_test/build/libs/stresstest.jar statistic -c /path/to/stress.conf -o tps-statistic-result
# check the log
tail -f logs/stress_test.log
```
The TPS statistic result will be saved in the `tps-statistic-result` file of current directory, which is same
with the `broadcast` subcommand statistic result.
