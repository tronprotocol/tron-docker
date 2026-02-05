# TRON Private Chain Development Tools

Build and run a TRON private chain from source code for development and testing purposes. This tool provides a complete Docker-based solution to clone, compile, and deploy java-tron nodes.

## Features

- ✅ Automatic source code cloning from Git repository
- ✅ Multi-stage Docker build for optimized image size
- ✅ Architecture auto-detection (ARM64/AMD64)
- ✅ Automatic deployment of SR + Fullnode private chain
- ✅ Built-in Prometheus metrics support
- ✅ Customizable repository and branch selection
- ✅ Docker build cache for faster rebuilds

## Prerequisites

### Docker

For Docker and Docker Compose installation, refer to [prerequisites](../../README.md#prerequisites).

Ensure Docker has at least:
- **8GB memory** per node container
- **20GB disk space** for source compilation
- **BuildKit enabled** (Docker 18.09+ default)

### System Requirements

| Architecture | JDK Version | Status |
|-------------|-------------|--------|
| ARM64 (Apple Silicon M1/M2/M3) | JDK 17 | ✅ Auto-detected |
| AMD64 (Intel x86_64) | JDK 8 | ✅ Auto-detected |

The Dockerfile automatically detects your system architecture and uses the appropriate JDK version.

## Quick Start

Download the `tron-docker` repository, enter the [tools/dev](./) directory, and start the services using the following commands:

```bash
cd tools/dev

# Use default configuration (master branch)
docker-compose -f docker-compose-build.yml up --build -d
```

This will:
1. Clone the java-tron repository from GitHub
2. Compile the source code using Gradle
3. Build a Docker image with the compiled binary
4. Start two nodes: one SR (witness) and one Fullnode
5. Expose Fullnode API on port 8090 and metrics on port 9527

### Verify the deployment

After startup (approximately 8-15 minutes for first build), you can verify the nodes are running:

```bash
# Check node status
curl http://localhost:8090/wallet/getnodeinfo | jq

# Check block height (should be increasing)
curl http://localhost:8090/wallet/getnowblock | jq '.block_header.raw_data.number'

# Check metrics
curl http://localhost:9527/metrics | grep tron
```

## Configuration

### Using environment variables

Copy the template configuration file and customize it:

```bash
# Create configuration file from template
cp env.template .env

# Edit configuration
vim .env
```

Available configuration options in `.env`:

```bash
# ============================================
# Java-Tron Source Configuration
# ============================================
# Git repository URL (optional, defaults to official repo, also can build from other repository)
#REPO=https://github.com/tronprotocol/java-tron.git

# Branch name (optional, defaults to master)
#BRANCH=master

# Example: Use your own fork
#REPO=https://github.com/your-name/java-tron.git
#BRANCH=develop

# ============================================
# Port Configuration
# ============================================
# Fullnode ports (exposed by default)
#FULLNODE_HTTP_PORT=8090
#FULLNODE_SOLIDITY_PORT=8091
#FULLNODE_GRPC_PORT=50051
#FULLNODE_METRICS_PORT=9527

# SR ports (not exposed by default, uncomment if needed)
#SR_HTTP_PORT=8090
#SR_SOLIDITY_PORT=8091
#SR_GRPC_PORT=50051

# ============================================
# Resource Configuration
# ============================================
# Memory limits for nodes
#SR_MEMORY=8g
#FULLNODE_MEMORY=8g
```

### Using custom repository or branch

You can build from a specific branch or your own fork:

```bash
# Build from develop branch
BRANCH=develop docker-compose -f docker-compose-build.yml up --build -d

# Build from your fork
REPO=https://github.com/your-name/java-tron.git BRANCH=feature-branch \
  docker-compose -f docker-compose-build.yml up --build -d
```

## Node Information

After deployment, the following services are available:

### Fullnode (exposed to host)
- **HTTP API**: `http://localhost:8090`
- **Solidity HTTP API**: `http://localhost:8091`
- **gRPC API**: `localhost:50051`
- **Metrics API**: `http://localhost:9527/metrics`

### SR Node (internal only by default)
- **Hostname**: `tron-witness1` (accessible within Docker network)
- **Role**: Block production
- **Ports**: Not exposed to host (optional, see configuration)

**Note**: The SR node is not exposed to the host by default. It only handles block production while the Fullnode provides API services. To expose SR ports, uncomment the `ports` section in `docker-compose-build.yml`.

## Common Commands

### Build and deployment

```bash
# Build and start in background
docker-compose -f docker-compose-build.yml up --build -d

# Build without cache (full rebuild)
docker-compose -f docker-compose-build.yml build --no-cache
docker-compose -f docker-compose-build.yml up -d

# Rebuild only when source changes
docker-compose -f docker-compose-build.yml up --build -d
```

### Monitoring

Node logs are persisted in Docker volumes and can be accessed even when containers are stopped:

```bash
# View Fullnode tron.log (last 100 lines)
docker run --rm -v tron-dev-fullnode-logs:/logs alpine tail -n 100 /logs/tron.log

# View SR tron.log (last 100 lines)
docker run --rm -v tron-dev-sr-logs:/logs alpine tail -n 100 /logs/tron.log

# Follow Fullnode tron.log in real-time
docker run --rm -v tron-dev-fullnode-logs:/logs alpine tail -f /logs/tron.log

# Follow SR tron.log in real-time
docker run --rm -v tron-dev-sr-logs:/logs alpine tail -f /logs/tron.log

# List all log files
docker run --rm -v tron-dev-fullnode-logs:/logs alpine ls -lah /logs

# Search for errors in Fullnode logs
docker run --rm -v tron-dev-fullnode-logs:/logs alpine grep -i "error" /logs/tron.log

# Copy logs to local directory
mkdir -p ~/tron-logs
docker run --rm -v tron-dev-fullnode-logs:/logs -v ~/tron-logs:/backup \
  alpine cp -r /logs/. /backup/fullnode/
docker run --rm -v tron-dev-sr-logs:/logs -v ~/tron-logs:/backup \
  alpine cp -r /logs/. /backup/sr/
```

**Alternative: Use docker-compose logs for container output**

```bash
# View Docker container logs (stdout/stderr)
docker-compose -f docker-compose-build.yml logs -f tron-fullnode
docker-compose -f docker-compose-build.yml logs -f tron-sr

# Check container status
docker-compose -f docker-compose-build.yml ps

# Check resource usage
docker stats tron-dev-sr tron-dev-fullnode
```

### Cleanup

```bash
# Stop containers
docker-compose -f docker-compose-build.yml down

# Stop and remove volumes (deletes all blockchain data)
docker-compose -f docker-compose-build.yml down -v

# Remove images as well
docker-compose -f docker-compose-build.yml down --rmi all

# Clean build cache
docker builder prune -f
```

## Monitoring with Metrics

The nodes expose Prometheus-compatible metrics on port 9527. You can integrate with Prometheus and Grafana for monitoring.

### Available metrics

```bash
# View all metrics
curl http://localhost:9527/metrics

# Filter specific metrics
curl http://localhost:9527/metrics | grep tron_header_height
curl http://localhost:9527/metrics | grep tron_peers
```

For more information about available metrics, see [metric_monitor README](../../metric_monitor/README.md#all-metrics).

### Integration with Prometheus

You can add the Fullnode as a target in your Prometheus configuration:

```yaml
scrape_configs:
  - job_name: 'tron-private-chain'
    static_configs:
      - targets: ['localhost:9527']
        labels:
          node_type: 'fullnode'
          environment: 'dev'
```

## File Structure

| File | Description |
|------|-------------|
| `Dockerfile.build` | Multi-stage Dockerfile for building from source |
| `docker-compose-build.yml` | Docker Compose configuration for deployment |
| `env.template` | Environment variable template |
| `DOCKER-COMPOSE-GUIDE.md` | 📖 Detailed usage guide |
| `prometheus.yml` | (Optional) Prometheus configuration example |

## Architecture

### Multi-Stage Docker Build

The Dockerfile uses multi-stage builds to optimize image size:

```
Stage 1 (Builder):
  ├─ Clone java-tron source
  ├─ Compile with Gradle
  └─ Extract FullNode.jar (~100 MB)

Stage 2 (Runtime):
  ├─ Lightweight JRE base image
  ├─ Copy only FullNode.jar
  └─ Final image: ~300 MB (vs ~1 GB with full source)
```

### Network Architecture

```
Docker Network: tron-dev-network
├─ tron-witness1 (SR)
│  ├─ Produces blocks
│  ├─ Connects to: (none, initial producer)
│  └─ Ports: Internal only (optional expose)
│
└─ tron-fullnode (Fullnode)
   ├─ Syncs blocks from SR
   ├─ Connects to: tron-witness1:18888
   └─ Ports: 8090, 8091, 50051, 9527 (exposed)
```

## Build Time

Expected build times (first build):

| Step | Time | Notes |
|------|------|-------|
| Clone repository | 30s - 1min | Depends on network speed |
| Gradle build | 5-10min | Depends on CPU |
| Docker image | 1-2min | Multi-stage optimization |
| Node startup | 1-2min | Genesis block initialization |
| **Total** | **8-15min** | Subsequent builds are faster with cache |

**Note**: Rebuilds with Docker cache are much faster (~2-3 minutes) if only configuration changes.

## Troubleshooting

### Build fails with "not found" error

**Problem**: JDK image not found

**Solution**: The Dockerfile automatically selects the JDK version based on your architecture. Ensure you have internet access to pull images.

### Containers fail to start

**Problem**: Insufficient memory

**Solution**: Allocate at least 8GB per node in Docker Desktop settings.

### Fullnode cannot connect to SR

**Problem**: Hostname mismatch

**Solution**: Ensure the SR hostname is `tron-witness1` in `docker-compose-build.yml` and the seed node configuration in `private_net_config_others.conf` matches.

### Port already in use

**Problem**: Port 8090 or others are occupied

**Solution**: Modify port mappings in `.env` file:

```bash
FULLNODE_HTTP_PORT=9090
FULLNODE_SOLIDITY_PORT=9091
```

### Slow build on ARM (Apple Silicon)

**Problem**: First build takes longer on M1/M2/M3

**Solution**: This is normal. The Dockerfile automatically uses JDK 17 for ARM64, which is optimized for Apple Silicon. Subsequent builds will be faster with cache.

## Advanced Usage

### Custom genesis block

To use a custom genesis block configuration:

1. Edit configuration files in `../../conf/`:
   - `private_net_config_witness1.conf` (SR configuration)
   - `private_net_config_others.conf` (Fullnode configuration)

2. Modify the `genesis.block` section with your custom parameters

3. Rebuild and restart:
   ```bash
   docker-compose -f docker-compose-build.yml down -v
   docker-compose -f docker-compose-build.yml up --build -d
   ```

### Adding more nodes

To add additional fullnodes, edit `docker-compose-build.yml`:

```yaml
  tron-fullnode-2:
    # Copy the tron-fullnode service definition
    # Change container_name and ports
    ports:
      - "8094:8090"  # Different host port
```

### CI/CD Integration

For automated builds in CI/CD pipelines:

```bash
# Example GitLab CI / GitHub Actions
docker-compose -f docker-compose-build.yml build --no-cache
docker-compose -f docker-compose-build.yml up -d

# Wait for nodes to be ready
sleep 120

# Run tests
curl http://localhost:8090/wallet/getnodeinfo

# Cleanup
docker-compose -f docker-compose-build.yml down -v
```

## Documentation

For more detailed information, refer to:

- **[DOCKER-COMPOSE-GUIDE.md](./DOCKER-COMPOSE-GUIDE.md)** - Comprehensive usage guide
- **[Metric Monitoring](../../metric_monitor/README.md)** - Prometheus metrics documentation
- **[TRON Developers](https://developers.tron.network/)** - Official TRON documentation
- **[java-tron GitHub](https://github.com/tronprotocol/java-tron)** - Source code repository

## License

This project follows the same license as [java-tron](https://github.com/tronprotocol/java-tron).
