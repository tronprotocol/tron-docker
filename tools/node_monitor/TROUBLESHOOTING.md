# 故障排查指南

## Prometheus 无法连接 node_monitor

### 错误信息
```
Error scraping target: Get "http://node-monitor:9091/metrics": dial tcp: lookup node-monitor on 127.0.0.11:53: no such host
```

### 原因
`node-monitor` 没有在 Docker Compose 中定义，Prometheus 容器无法通过 DNS 解析这个主机名。

### 解决方案

#### 方案 1：node_monitor 在主机上运行（推荐用于开发测试）

如果 `node_monitor` 在主机上运行，需要修改 `prometheus.yml` 使用特殊的主机名来访问主机：

**macOS/Windows:**
```yaml
- targets:
    - host.docker.internal:9091
```

**Linux:**
```yaml
- targets:
    - 172.17.0.1:9091  # Docker 默认网关 IP
```

或者使用 `host.docker.internal`（Docker 20.10+ 支持）：
```yaml
- targets:
    - host.docker.internal:9091
```

**验证主机 IP（Linux）:**
```bash
docker network inspect bridge | grep Gateway
```

#### 方案 2：node_monitor 在 Docker 容器中运行（推荐用于生产）

将 `node_monitor` 添加到 `docker-compose.yml`：

```yaml
services:
  node-monitor:
    build:
      context: ../tools/node_monitor
      dockerfile: Dockerfile
    container_name: node-monitor
    networks:
      - tron_network
    ports:
      - "9091:9091"
    environment:
      - NODE_URL=http://tron-node1:8090
    command:
      - "-node-url"
      - "http://tron-node1:8090"
      - "-metrics-addr"
      - "0.0.0.0:9091"
      - "-interval"
      - "10s"
```

然后在 `prometheus.yml` 中使用容器名：
```yaml
- targets:
    - node-monitor:9091
```

### 验证步骤

1. **检查 node_monitor 是否运行：**
   ```bash
   # 如果主机运行
   curl http://localhost:9091/metrics
   
   # 如果容器运行
   docker exec node-monitor curl http://localhost:9091/metrics
   ```

2. **从 Prometheus 容器测试连接：**
   ```bash
   # 测试 host.docker.internal（macOS/Windows）
   docker exec prometheus wget -qO- http://host.docker.internal:9091/metrics
   
   # 测试 Docker 网关（Linux）
   docker exec prometheus wget -qO- http://172.17.0.1:9091/metrics
   
   # 测试容器名（如果 node_monitor 在 Docker 中）
   docker exec prometheus wget -qO- http://node-monitor:9091/metrics
   ```

3. **检查 Prometheus 目标状态：**
   - 访问 `http://localhost:9090/targets`
   - 查看 `tron-node-monitor` 的状态
   - 应该显示 "UP"

### 常见问题

**Q: Linux 上 `host.docker.internal` 不可用？**
A: 使用 Docker 网关 IP `172.17.0.1`，或添加 `extra_hosts` 到 Prometheus 服务：
```yaml
prometheus:
  extra_hosts:
    - "host.docker.internal:host-gateway"
```

**Q: 仍然无法连接？**
A: 检查：
1. node_monitor 是否监听 `0.0.0.0:9091`（不是 `127.0.0.1:9091`）
2. 防火墙是否阻止了 9091 端口
3. Docker 网络配置是否正确
