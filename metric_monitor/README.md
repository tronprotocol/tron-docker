# TRON Node Metrics Monitoring
Starting from the GreatVoyage-4.5.1 (Tertullian) version, the node provides a series of interfaces compatible with the Prometheus protocol, allowing the node deployer to monitor the health status of the node more conveniently.

Below, we provide a guide on using metrics to monitor the TRON node status. Then, we list all available metrics.

## Prerequisites
### Docker

For Docker and Docker Compose installation refer [prerequisites](../README.md#prerequisites).

Then check the Docker resource settings to ensure it has at least 16GB of memory per FullNode container.

## Quick start
Download the `tron-docker` repository, enter the `metric` directory, and start the services defined in [docker-compose.yml](./docker-compose.yml) using the following command:

```sh
docker-compose up -d
```
It will start a TRON FullNode that connects to the Mainnet, along with Prometheus and Grafana services. Note that in [main_net_config.conf](../conf/main_net_config.conf), it contains the configuration below to enable metrics.
```
metrics{
  prometheus{
    enable=true
    port="9527"
  }
}
```

### Prometheus service
The Prometheus service will use the configuration file [prometheus.yml](metric_conf/prometheus.yml). It uses the configuration below to add targets for monitoring.
```
- targets:
    - tron-node1:9527 # use container name
  labels:
    group: group-tron
    instance: fullnode-01
```

You can view the running status of the Prometheus service at `http://localhost:9090/`. Click on "Status" -> "Configuration" to check whether the configuration file used by the container is correct.
![image](../images/prometheus_configuration.png)

If you want to monitor more nodes, simply add more targets following the same format. Click on "Status" -> "Targets" to view the status of each monitored java-tron node.
![image](../images/prometheus_targets.png)
**Notice**: If you want to check metrics, please use `http://localhost:9527/metrics` on host machine instead of `http://tron-node1:9527/metrics`, as the latter is used for container access inside Docker.

### Grafana service
After startup, you can log in to the Grafana web UI through [http://localhost:3000/](http://localhost:3000/). The initial username and password are both `admin`. After logging in, change the password according to the prompts, and then you can enter the main interface.

Click the **Connections** on the left side of the main page and select "Data Sources" to configure Grafana data sources. Enter the ip and port of the prometheus service in URL with `http://prometheus:9090`.
![image](../images/grafana_data_source.png)

#### Import dashboard
For the convenience of java-tron node deployers, the TRON community provides a comprehensive dashboard configuration file [grafana_dashboard_tron_server.json](metric_conf/grafana_dashboard_tron_server.json).
Click the Grafana dashboards icon on the left, then select "New" and "Import", then click "Upload JSON file" to import the downloaded dashboard configuration file:
![image](../images/grafana_dashboard.png)

Then you can see the following monitors displaying the running status of the nodes in real time:
![image](../images/grafana_dashboard_monitoring.png)

Below, we will introduce all supported metrics to facilitate customizing your dashboard.

## All metrics
As you can see from above Grafana dashboard or http://localhost:9527/metrics, the available metrics are categorized into the following:

- Blockchain status
- Node system status
- Block and transaction status
- Network peer status
- API information
- Database information
- JVM status

### Blockchain status

- `tron:header_time`: The latest block time of java-tron on this node
- `tron:header_height`: The latest block height of java-tron on this node
- `tron:miner_total`: Used to display the blocks produced by a certain SR

### Node system status
Metric of specific container:
- `process_cpu_load`: Process CPU load
- `process_cpu_seconds_total`
- `process_max_fds`: Maximum number of open file descriptors
- `process_open_fds`: Number of open file descriptors
- `process_resident_memory_bytes`: Resident memory size in bytes
- `process_start_time_seconds`: Start time of the process since unix epoch in seconds
- `process_virtual_memory_bytes`

Metric for docker resources:
- `system_available_cpus`: System available cpus
- `system_cpu_load`: System CPU load
- `system_free_physical_memory_bytes`: System free physical memory bytes
- `system_free_swap_spaces_bytes`: System free swap spaces
- `system_load_average`: System CPU load average
- `system_total_physical_memory_bytes`: System total physical memory bytes
- `system_total_swap_spaces_bytes`: System free swap spaces bytes

Follow the example dashboard to add more panels.
![image](../images/metric_system.png)

### Block status

Used to check the block process performance from TronNetDelegate:
- `tron:block_process_latency_seconds_bucket`: Cumulative counters
- `tron:block_process_latency_seconds_count`: Count of events
- `tron:block_process_latency_seconds_sum`: Total sum of all observed values

Used to check the block processing latency from the Manager, which is invoked by TronNetDelegate:
- `tron:block_push_latency_seconds_bucket`: Cumulative counters
- `tron:block_push_latency_seconds_count`: Count of events
- `tron:block_push_latency_seconds_sum`: Total sum of all observed values

When handling the above block push logic, TRON's processing logic needs to acquire a synchronization lock. The `lock_acquire_latency_seconds_x` metric is used to indicate the latency.
- `tron:lock_acquire_latency_seconds_bucket`: Cumulative counters
- `tron:lock_acquire_latency_seconds_count`: Count of events
- `tron:lock_acquire_latency_seconds_sum`: Total sum of all observed values

Used to check the block latency received from peers and not from sync requests:
- `tron:block_fetch_latency_seconds_bucket`: Cumulative counters
- `tron:block_fetch_latency_seconds_count`: Count of events
- `tron:block_fetch_latency_seconds_sum`: Total sum of all observed values
- `tron:block_receive_delay_seconds_bucket/count/sum`

Verify the latency of all transactions' signatures when processing a block:
- `tron:verify_sign_latency_seconds_bucket`: Cumulative counters for
- `tron:verify_sign_latency_seconds_count`: Count of events
- `tron:verify_sign_latency_seconds_sum`: Total sum of all observed values

Check the usage from dashboard panel (enter edit mode), or by searching in [grafana_dashboard_tron_server.json](metric_conf/grafana_dashboard_tron_server.json).
![image](../images/metric_block_latency.png)

### Transaction status
- `tron:manager_queue_size`: The Manager Queue Size for pending/popped/queued/repush transaction types.
- `tron:tx_cache`: TRON tx cache put action event.

Average transaction processing time:
- `tron:process_transaction_latency_seconds_bucket`: Cumulative counters
- `tron:process_transaction_latency_seconds_count`: Count of event
- `tron:process_transaction_latency_seconds_sum`: Total sum of all observed values

### Network peer status

TRON peers info and abnormal statistic metrics:
- `tron:peers`
- `tron:p2p_disconnect_total`
- `tron:p2p_error_total`

The latency exceeds 50ms to process a message from a peer will be logged by the below metrics:
- `tron:message_process_latency_seconds_bucket`: Cumulative counters for
- `tron:message_process_latency_seconds_count`: Count of events
- `tron:message_process_latency_seconds_sum`: Total sum of all observed values

![image](../images/metric_network_message_process_latency.png)
Currently, the possible message types are: `P2P_PING`, `P2P_PONG`, `P2P_HELLO`, `P2P_DISCONNECT`, `SYNC_BLOCK_CHAIN`, `BLOCK_CHAIN_INVENTORY`, `INVENTORY`, `FETCH_INV_DATA`, `BLOCK`, `TRXS`, `PBFT_COMMIT_MSG`.
Check [node-connection](https://tronprotocol.github.io/documentation-en/developers/code-structure/#node-connection) for detail explanation of above types.

TCP/UDP network data traffic statistics:
- `tron:tcp_bytes_bucket`：Cumulative counters
- `tron:tcp_bytes_count`：Count of events
- `tron:tcp_bytes_sum`：Total sum of all observed values
- `tron:udp_bytes_bucket/count/sum`

![image](../images/metric_network_tcp_udp.png)


### API information

Http request data traffic statistics:
- `tron:http_bytes_bucket`: Cumulative counters
- `tron:http_bytes_count`:Count of events
- `tron:http_bytes_sum`: Total sum of all observed values

Http/GRPC request latency metrics:
- `tron:http_service_latency_seconds_bucket`: Cumulative counters
- `tron:http_service_latency_seconds_count`: Count of events
- `tron:http_service_latency_seconds_sum`: Total sum of all observed values
- `tron:grpc_service_latency_seconds_bucket/count/sum`
- `tron:internal_service_latency_seconds_bucket/count/sum`

Example:
![image](../images/metric_API.png)

### Database information

TRON blockchain storage chooses to use LevelDB, which is developed by Google and proven successful with many companies and projects. These below db related metrics all have filters with `db` name and `level`.
- `tron:db_size_bytes`
- `tron:guava_cache_hit_rate`: Hit rate of a guava cache.
- `tron:guava_cache_request`: Request of a guava cache.
- `tron:guava_cache_eviction_count`: Eviction count of a guava cache.
- `tron:db_sst_level`: Related with LevelDB SST file compaction.

Currently, for `db` values of above metrics TRON has below possible objects:
- accountid-index
- abi
- account
- votes
- proposal
- witness
- code
- recent-transaction
- exchange-v2
- market_pair_to_price
- trans
- contract
- storage-row
- block
- exchange
- DelegatedResource
- tree-block-index
- balance-trace
- market_pair_price_to_order
- asset-issue
- transactionHistoryStore
- IncrementalMerkleTree
- delegation
- transactionRetStore
- account-index
- market_order
- witness_schedule
- nullifier
- DelegatedResourceAccountIndex
- properties
- common
- block-index
- accountTrie
- contract-state
- account-trace
- market_account
- recent-block
- asset-issue-v2
- section-bloom
- tmp

### JVM status

**JVM basic info**

- `jvm_info`: Basic JVM info with version
- `jvm_classes_currently_loaded`: The number of classes that are currently loaded in the JVM
- `jvm_classes_loaded_total`
- `jvm_classes_unloaded_total`

**JVM thread related**

* `jvm_threads_current`
* `jvm_threads_daemon`
* `jvm_threads_peak`
* `jvm_threads_started_total`
* `jvm_threads_deadlocked`
* `jvm_threads_deadlocked_monitor`
* `jvm_threads_state`{state="RUNNABLE"/"TERMINATED"/"TIMED_WAITING"/"NEW"/"WAITING"/"BLOCKED"}

**JVM garbage collection**

* `jvm_gc_collection_seconds_count`: Count of JVM garbage collector event
* `jvm_gc_collection_seconds_sum`: Total sum of observed values

**JVM memory related**

* `jvm_buffer_pool_capacity_bytes`: Bytes capacity of a given JVM buffer pool
* `jvm_buffer_pool_used_buffers`: Used buffers of a given JVM buffer pool
* `jvm_buffer_pool_used_bytes`: Used bytes of a given JVM buffer pool
* `jvm_memory_bytes_committed`: Committed (bytes) of a given JVM memory area
* `jvm_memory_bytes_init`: Initial bytes of a given JVM memory area
* `jvm_memory_bytes_max`: Max (bytes) of a given JVM memory area
* `jvm_memory_bytes_used`: Used bytes of a given JVM memory area
* `jvm_memory_objects_pending_finalization`: The number of objects waiting in the finalizer queue
* `jvm_memory_pool_allocated_bytes_total`
* `jvm_memory_pool_bytes_committed`: Committed bytes of a given JVM memory pool
* `jvm_memory_pool_bytes_init`: Initial bytes of a given JVM memory pool
* `jvm_memory_pool_bytes_max`: Max bytes of a given JVM memory pool
* `jvm_memory_pool_bytes_used`: Used bytes of a given JVM memory pool
* `jvm_memory_pool_collection_committed_bytes`: Committed after last collection bytes of a given JVM memory pool
* `jvm_memory_pool_collection_init_bytes`: Initial after last collection bytes of a given JVM memory pool
* `jvm_memory_pool_collection_max_bytes`: Max bytes after last collection of a given JVM memory pool
* `jvm_memory_pool_collection_used_bytes`: Used bytes after last collection of a given JVM memory pool

### Other metrics
Beside above metrics, there are also metrics to measure the duration of a scrape process, which is useful for monitoring and understanding the performance of your Prometheus server and the targets it scrapes.
- `scrape_duration_seconds`: It measures the time taken (in seconds) for Prometheus to scrape a target. This includes the entire process of making an HTTP request to the target, receiving the response, and processing the metrics.
- `scrape_samples_post_metric_relabeling`
- `scrape_samples_scraped`
- `scrape_series_added`
