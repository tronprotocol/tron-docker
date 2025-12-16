# Tron Node Monitor

A Go tool for monitoring Tron nodes with the following capabilities:

- **Empty Block Monitoring**: Monitor whether nodes produce empty blocks
- **SR Set Change Monitoring**: Monitor changes in the Super Representative (SR) set
- **Prometheus Metrics Export**: Export monitoring data to Prometheus with rich labels

## Features

- Periodically check the latest blocks from nodes to monitor empty block production
- Monitor SR set changes (additions, removals, ordering changes, etc.)
- Expose monitoring data through a single Prometheus metrics endpoint
- Support graceful shutdown
- Designed for monitoring **multiple nodes in one process** via the `node` label in metrics
- **All node and runtime settings are managed via a YAML config file**

## Build

```bash
cd tools/node_monitor
go build -o node_monitor main.go
```

## Configuration (YAML)

Example `node_monitor.yml`:

```yaml
metrics_addr: "0.0.0.0:9090"   # listen address for the metrics server
interval: "10s"               # monitoring check interval
nodes:
  - label: "tron-node1"
    url: "http://tron-node1:8090"
  - label: "tron-node2"
    url: "http://tron-node2:8090"
  - label: "tron-node3"
    url: "http://tron-node3:8090"
```

- `metrics_addr`: Address where the `/metrics` and `/health` endpoints are exposed.
- `interval`: How often each node is checked (Go duration format, e.g. `5s`, `1m`).
- `nodes`: List of Tron nodes to monitor.
  - `label`: Logical node label used as the `node` label in Prometheus metrics.
  - `url`: Tron node HTTP API URL (e.g. `http://host:8090`).

## Run

```bash
./node_monitor -config node_monitor.yml
```

This will:
- Start monitors for all nodes defined under `nodes:` in the config file
- Use the `label` as the `node` label in Prometheus metrics
- Expose all metrics at `http://<host>:9090/metrics` (or the `metrics_addr` you configure)

## Command Line Arguments

- `-config`: **Required**. Path to YAML config file. All nodes and runtime settings are read from this file.

## Prometheus Metrics

After the service starts, you can access the following endpoints:

- `/metrics`: Prometheus metrics endpoint
- `/health`: Health check endpoint

Key labels exposed include:

- `node`: logical node label (from config `label`)
- `witness_address`, `block_number`, `block_hash` for empty-block metrics
- `address`, `url`, `change_type` for SR-related metrics

### Prometheus Configuration

Example Prometheus scrape configuration for one monitor service:

```yaml
scrape_configs:
  - job_name: 'tron-node-monitor'
    static_configs:
      - targets: ['monitor-node:9090']
```

Metrics for different Tron nodes are separated by the `node` label, so dashboards and alerts can filter or group by it.

## License

Same as the tron-docker project
