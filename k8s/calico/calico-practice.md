# Calico 网络解决方案学习指南

## 📖 Calico 简介

Calico 是一个开源的网络和网络安全解决方案，专为容器、虚拟机和基于主机的工作负载而设计。它是 Kubernetes 生态系统中最受欢迎的网络插件之一。

### 🎯 核心特性

- **高性能网络**: 基于 BGP 路由协议，提供高性能的网络连接
- **网络策略**: 提供细粒度的网络安全策略控制
- **多平台支持**: 支持 Kubernetes、OpenShift、Docker、OpenStack 等平台
- **IP 管理**: 自动分配和管理 IP 地址
- **网络隔离**: 支持多租户网络隔离
- **可扩展性**: 支持大规模集群部署

### 🏗️ 架构组件

#### 1. Calico Node
- 运行在每个节点上的 DaemonSet
- 负责路由编程和网络策略执行
- 包含 Felix、BIRD、confd 组件

#### 2. Felix
- Calico 的主要代理程序
- 负责编程路由表和 iptables 规则
- 监控端点和网络策略变化

#### 3. BIRD
- BGP 客户端，负责分发路由信息
- 与其他节点的 BIRD 实例建立 BGP 会话
- 确保整个集群的路由一致性

#### 4. Calico CNI Plugin
- 容器网络接口插件
- 负责为新创建的 Pod 分配 IP 地址
- 设置网络接口和路由

#### 5. etcd/Kubernetes API Server
- 存储 Calico 的配置和状态信息
- 在 Kubernetes 环境中直接使用 Kubernetes API

## � 组件状态查看（Felix / BIRD / confd 速查）

> 提示：若集群使用 VXLAN-only 且未启用 BGP，或使用 eBPF 数据面且未配置 BGP，BIRD/Confd 可能不运行或无会话，这是正常现象。

### 1) 定位 calico-node 与镜像
```bash
kubectl -n calico-system get pods -l k8s-app=calico-node -o wide
kubectl -n calico-system get pod <calico-node-pod> \
  -o jsonpath='{.spec.containers[?(@.name=="calico-node")].image}{"\n"}'
```

### 2) Felix（数据面代理）
```bash
# 日志
kubectl -n calico-system logs -f ds/calico-node --tail=200 | grep -i felix
kubectl -n calico-system logs <calico-node-pod> --tail=200 | grep -i felix

# 指标（默认 9091）
kubectl -n calico-system port-forward <calico-node-pod> 9091:9091
# 浏览 http://127.0.0.1:9091/metrics，关注 felix_ 前缀

# 节点状态（含 BGP 概览，如启用）
kubectl -n calico-system exec -it <calico-node-pod> -- calicoctl node status
```

### 3) BIRD（BGP 路由）
```bash
# 协议与邻居（不同版本命令可能是 birdc 或 birdcl）
kubectl -n calico-system exec -it <calico-node-pod> -- sh -lc 'birdc show protocols || birdcl show protocols'
kubectl -n calico-system exec -it <calico-node-pod> -- sh -lc 'birdc show status || birdcl show status'

# 路由表
kubectl -n calico-system exec -it <calico-node-pod> -- sh -lc 'birdc show route || birdcl show route'
kubectl -n calico-system exec -it <calico-node-pod> -- sh -lc 'birdc show route for 10.244.0.0/16 || birdcl show route for 10.244.0.0/16'

# 进程存在性
kubectl -n calico-system exec -it <calico-node-pod> -- pgrep -a bird || true
```

### 4) confd（生成 BIRD 配置）
```bash
# 日志与进程
kubectl -n calico-system logs <calico-node-pod> --tail=300 | grep -i confd || true
kubectl -n calico-system exec -it <calico-node-pod> -- pgrep -a confd || true

# 生成配置位置（随版本可能不同）
kubectl -n calico-system exec -it <calico-node-pod> -- sh -lc 'ls -l /etc/calico/bird* || ls -l /etc/calico/confd || true'
kubectl -n calico-system exec -it <calico-node-pod> -- sh -lc 'sed -n "1,120p" /etc/calico/bird.cfg 2>/dev/null || true'

# 触发 BIRD 重新加载配置
kubectl -n calico-system exec -it <calico-node-pod> -- sh -lc 'birdc configure || birdcl configure || true'
```

### 5) 常见判别
- VXLAN-only/未启用 BGP：看不到 BIRD/Confd 会话或进程，属正常。
- eBPF 数据面：仅在启用 BGP 分发路由时才会使用 BIRD。
- 防火墙阻断 TCP/179：BIRD 进程在，但会话处于 Idle/Connect；用 birdc 可见状态。
- 本机无 calicoctl：通过 kubectl exec 进入 calico-node Pod 内调用。

## �🚀 工作原理

### 1. 网络连接
```
Pod A (192.168.1.10) → Node A → BGP Router → Node B → Pod B (192.168.2.20)
```

### 2. 网络策略执行
- 基于 iptables 规则实现网络策略
- 支持 Kubernetes NetworkPolicy API
- 提供更丰富的 Calico NetworkPolicy

### 3. IP 地址管理 (IPAM)
- 自动分配 Pod IP 地址
- 支持多个 IP 池配置
- IP 地址回收和重用

## 🎓 推荐学习资源

### 📚 官方文档
1. **Calico 官方文档**: https://docs.projectcalico.org/
   - 最权威的学习资源
   - 包含安装、配置、故障排除指南

2. **Calico GitHub**: https://github.com/projectcalico/calico
   - 源代码和示例
   - Issue 讨论和最新更新

### 🎥 视频教程
1. **Calico 官方 YouTube 频道**
   - Project Calico 技术演讲
   - 实战演示和最佳实践

2. **CNCF 相关视频**
   - KubeCon 上的 Calico 技术分享
   - 云原生网络安全主题

### 📖 书籍推荐
1. **《Kubernetes 网络权威指南》**
   - 详细介绍 Kubernetes 网络模型
   - 包含 Calico 实战案例

2. **《云原生网络》**
   - 云原生环境下的网络解决方案
   - 网络安全和策略管理

### 🛠️ 实践教程
1. **官方 Getting Started 教程**
   ```bash
   # 快速开始教程
   https://docs.projectcalico.org/getting-started/
   ```

2. **Katacoda 交互式教程**
   - 在线实验环境
   - 无需本地环境搭建

## 🔧 快速开始

### 安装 Calico (Kubernetes)

#### 1. 使用 Operator 安装
```bash
# 安装 Tigera Operator
kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.25.0/manifests/tigera-operator.yaml

# 安装 Calico
kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.25.0/manifests/custom-resources.yaml
```

#### 2. 使用 Manifest 安装
```bash
# 直接安装 Calico
kubectl apply -f https://raw.githubusercontent.com/projectcalico/calico/v3.25.0/manifests/calico.yaml
```

### 验证安装
```bash
# 检查 Pod 状态
kubectl get pods -n calico-system

# 检查节点状态
kubectl get nodes -o wide

# 验证网络连接
kubectl exec -it <pod-name> -- ping <other-pod-ip>
```

## 🛡️ 网络策略示例

### 默认拒绝策略
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-all
  namespace: production
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
```

### 允许特定服务访问
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-frontend-to-backend
  namespace: production
spec:
  podSelector:
    matchLabels:
      app: backend
  policyTypes:
  - Ingress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: frontend
    ports:
    - protocol: TCP
      port: 8080
```

## � NetworkPolicy 查看与验证（速查）

### 快速查看与筛选
```bash
# 全集群/指定命名空间
kubectl get networkpolicy -A
kubectl get netpol -A -o wide
kubectl get netpol -n <ns>

# 查看详情
kubectl describe netpol <policy-name> -n <ns>
kubectl get netpol <policy-name> -n <ns> -o yaml

# 自定义列简表
kubectl get netpol -A -o custom-columns=NS:.metadata.namespace,NAME:.metadata.name,POLICY-TYPES:.spec.policyTypes

# 字段说明与用法
kubectl explain networkpolicy
kubectl explain networkpolicy.spec
kubectl explain networkpolicy.spec.{ingress,egress}

# 标签筛选与实时观察
kubectl get netpol -A -l 'team=payments'
kubectl get netpol -A -w
```

### 关联 Pod（被策略选中）
```bash
kubectl get pod -n <ns> --show-labels
# 若策略使用 matchLabels: { app: backend }
kubectl get pod -n <ns> -l app=backend -o wide
```
提示：复杂的 matchExpressions 需对照 YAML 人工核对后用等价选择器组合查询。

### Calico 扩展策略（如启用 Calico）
```bash
# 列出 Calico 策略与全局策略
calicoctl get networkpolicy -A
calicoctl get globalnetworkpolicy -A

# 查看具体定义
calicoctl get networkpolicy <ns>.<name> -o yaml
calicoctl get globalnetworkpolicy <name> -o yaml

# 判断是否安装了 Calico CRD
kubectl get crd | grep -i projectcalico
```

### 快速验证策略是否生效
```bash
# 启动临时测试 Pod（建议位于同一命名空间）
kubectl run -it --rm np-tester -n <ns> --image=alpine:3.20 --restart=Never -- sh
# 容器内：
apk add --no-cache curl busybox-extras
curl -v http://<pod-ip-or-svc>:<port> --max-time 3   # 测 TCP 入/出站
nc -u -vz <pod-ip> <udp-port>                         # 测 UDP 入/出站

# 或直接从业务 Pod 测试
kubectl exec -it <workload-pod> -n <ns> -- sh -c 'apk add --no-cache curl; curl -v http://<dst>:<port> --max-time 3'

# 若使用 Calico，辅助排障
kubectl -n calico-system logs -f ds/calico-node --tail=300
calicoctl node status
```

## �🔍 常用命令

### calicoctl 工具
```bash
# 安装 calicoctl
curl -L https://github.com/projectcalico/calico/releases/latest/download/calicoctl-linux-amd64 -o calicoctl
chmod +x calicoctl

# 查看节点状态
calicoctl node status

# 查看 IP 池
calicoctl get ippool -o wide

# 查看网络策略
calicoctl get networkpolicy

# 查看端点信息
calicoctl get workloadendpoint
```

### 故障排除命令
```bash
# 检查 BGP 连接状态
calicoctl node status

# 查看路由表
ip route show

# 检查 iptables 规则
iptables -L -n

# 查看 Calico 日志
kubectl logs -n calico-system <calico-pod-name>
```

## 🚨 常见问题排查

### 1. Pod 无法通信
- 检查 BGP 连接状态
- 验证路由表配置
- 检查网络策略规则

### 2. IP 地址分配失败
- 检查 IP 池配置
- 验证 IPAM 配置
- 查看 CNI 日志

### 3. 网络策略不生效
- 验证策略语法
- 检查标签选择器
- 查看 iptables 规则

## 📈 性能优化

### 1. BGP 优化
```yaml
apiVersion: projectcalico.org/v3
kind: BGPConfiguration
metadata:
  name: default
spec:
  logSeverityScreen: Info
  nodeToNodeMeshEnabled: false  # 禁用全网格模式
  asNumber: 64512
```

### 2. 路由反射器配置
```yaml
apiVersion: projectcalico.org/v3
kind: BGPPeer
metadata:
  name: route-reflector
spec:
  peerIP: 10.0.0.1
  asNumber: 64512
```

## 🔗 相关链接

- [Calico 官网](https://www.projectcalico.org/)
- [Kubernetes 网络模型](https://kubernetes.io/docs/concepts/cluster-administration/networking/)
- [CNI 规范](https://github.com/containernetworking/cni)
- [BGP 协议基础](https://tools.ietf.org/html/rfc4271)

## 📝 学习路径建议

### 初级阶段 (1-2周)
1. 理解 Kubernetes 网络模型
2. 学习 CNI 基本概念
3. 安装和配置 Calico
4. 实践基本的 Pod 间通信

### 中级阶段 (2-4周)
1. 深入理解 Calico 架构
2. 学习网络策略配置
3. 掌握 BGP 路由原理
4. 学习故障排除方法

### 高级阶段 (1-2月)
1. 大规模集群网络设计
2. 性能调优和监控
3. 自定义网络策略开发
4. 与其他网络解决方案对比

---

*最后更新: 2025年7月*

## 🧰 VXLAN 模式运维与常用命令

本节聚焦 Calico 在 VXLAN 模式（或 VXLANCrossSubnet）下的日常运维、排障与变更，提供可直接执行的检查与修复步骤。

### 1) 模式启用与状态确认（Day 0/Day 1）

- 确认 IPPool 封装策略与 NAT：
  ```bash
  calicoctl get ippool -o wide
  calicoctl get ippool -o yaml | sed -n '/^---/p;/^kind: IPPool/,$p' | grep -E 'name:|cidr:|vxlanMode:|ipipMode:|natOutgoing:' -n
  ```
- 确认 Felix VXLAN 开关与参数：
  ```bash
  kubectl get felixconfiguration default -o yaml | grep -E 'vxlanEnabled|vxlanPort|vxlanVNI|bpfEnabled|wireguardEnabled'
  ```
- 节点侧 VXLAN 设备与内核视图：
  ```bash
  ip -d link show vxlan.calico     # 查看 VNI/端口/本地地址
  ip addr show vxlan.calico        # IP/MAC/MTU
  ss -lun | grep 4789              # UDP 4789 监听/收发
  bridge fdb show dev vxlan.calico # 远端 VTEP 学习表
  ip neigh show dev vxlan.calico   # 邻居条目
  ```

启用方式（二选一）：
- 通过 IPPool：将 `vxlanMode: Always|CrossSubnet`，并将 `ipipMode: Never`。
- 通过 Operator（Installation CR）：`calicoNetwork.encapsulation: VXLAN|VXLANCrossSubnet`（推荐在变更窗口完成）。

MTU 提示：VXLAN 额外头部约 50B，常见路径 MTU 1500 时，Pod veth MTU 建议 1450。若路径 MTU 更小，需要进一步下调以避免分片。

### 2) 连通性自检 Runbook（Day 2）

按由易到难的顺序快速定位问题：
1. Pod ➜ Pod（同节点/跨节点）
   ```bash
   # 同节点
   kubectl exec -it <pod-a> -- ping -c3 <pod-b-ip>
   # 跨节点
   kubectl exec -it <pod-a-on-node1> -- ping -c3 <pod-b-on-node2-ip>
   ```
2. 节点 ➜ 节点 UDP 端口可达
   ```bash
   # 防火墙/安全组需放通 UDP/4789
   nc -u -vz <node2-internal-ip> 4789 || true
   ```
3. 抓包验证 VXLAN 封装是否发生
   ```bash
   # 在发送节点物理网卡
   sudo tcpdump -ni <underlay-if> udp port 4789 -vv
   # 在解封节点 VXLAN 设备（可见内层以太帧/ARP/IP）
   sudo tcpdump -ni vxlan.calico -vv
   ```
4. 核对 VTEP/FDB 学习
   ```bash
   bridge fdb show dev vxlan.calico | grep -E "self|dst"
   # 正常应看到远端 VTEP（dst <node-ip>）与本地 self 项
   ```
5. 路径 MTU 快检（IPv4）
   ```bash
   # 1472 = 1500 - 20(IP) - 8(ICMP); VXLAN 需再预留约50B
   ping -M do -s 1400 <remote-node-ip>   # 依据实际链路酌情调整
   ```

### 3) 常见问题与修复

- UDP/4789 被阻断或丢包
  ```bash
  # 检查本机防火墙
  iptables -S | grep 4789 || true
  # 云安全组/边界防火墙需放通 UDP 4789（双向）
  ```

- MTU 不匹配导致丢包/分片
  ```bash
  ip addr show vxlan.calico | grep mtu
  # 观察 dmesg 与 tcpdump，若出现 fragmentation needed 或重传，需下调 veth/VXLAN MTU
  ```
  处置：统一将 Pod veth MTU 设为 1450（或更低以契合实际链路 MTU），并在维护窗口滚动重建工作负载以生效。

- VTEP/FDB 条目异常（黑洞/错误下一跳）
  ```bash
  bridge fdb show dev vxlan.calico
  # 谨慎：刷新后 Felix 会重新编程
  sudo bridge fdb flush dev vxlan.calico
  ```

- 邻居项陈旧（ARP/NDP）
  ```bash
  ip neigh show dev vxlan.calico | grep FAILED || true
  sudo ip neigh flush dev vxlan.calico
  ```

- 反向路由过滤（rp_filter）导致回包被丢弃
  ```bash
  sysctl net.ipv4.conf.all.rp_filter
  sysctl net.ipv4.conf.default.rp_filter
  # 建议为 0(关闭) 或 2(宽松) 以适配 overlay
  ```

- 网卡 VXLAN 相关 offload 优化
  ```bash
  ethtool -k <underlay-if> | grep -E 'udp_tnl|gro|gso|tso'
  # 若硬件支持，开启 udp_tnl_segmentation/gro 可降低 CPU 占用
  ```

### 4) 模式迁移与回滚（BGP/IPIP → VXLAN）

建议在低峰期按以下顺序进行，并小批量灰度：
1. 预检查：
   - 放通节点间 UDP/4789。
   - 评估链路 MTU 并规划 Pod MTU（如 1450）。
2. 打开 VXLAN 能力：
   ```bash
   # Operator 安装：调整 Installation CR（encapsulation: VXLAN/VXLANCrossSubnet）
   # Manifest 安装：确保 felixconfiguration 中 vxlanEnabled: true
   ```
3. 调整 IPPool：
   ```bash
   # 将现有/新建 IPPool 设置为 vxlanMode: Always|CrossSubnet, ipipMode: Never
   calicoctl get ippool -o yaml > ippool.yaml
   # 编辑后回写
   calicoctl apply -f ippool.yaml
   ```
4. 同步 MTU 与滚动：
   - 统一 MTU，必要时滚动重建工作负载以应用新 veth MTU。
   - 观察跨节点流量是否已走 VXLAN（tcpdump 验证）。
5. 验证与收尾：
   - 跨节点 Pod 连通、业务回归、指标稳定后，清理旧的 IPIP 相关配置（如 `tunl0` 未再使用）。

回滚思路：反向将 IPPool 改回 `ipipMode: CrossSubnet/Always`、关闭 VXLAN，按变更前状态恢复，期间保持 UDP/4789 放通直至回滚完成。

### 5) 监控与诊断

- calicoctl/节点侧：
  ```bash
  calicoctl node status
  calicoctl get ippool -o wide
  calicoctl node diags        # 采集诊断包
  kubectl -n calico-system logs -f ds/calico-node --tail=500
  ```

- 指标（Prometheus）：
  ```bash
  # Felix 指标通常暴露在 calico-node Pod 的 9091 端口
  kubectl -n calico-system port-forward <calico-node-pod> 9091:9091
  # 浏览 /metrics，关注端点数、策略安装、若开启 eBPF 关注 bpf 相关指标
  ```

### 6) VXLAN 速查命令清单（可收藏）

```bash
# 模式/配置
calicoctl get ippool -o wide
kubectl get felixconfiguration default -o yaml | grep -E 'vxlan|bpf|wireguard'

# 设备/转发表
ip -d link show vxlan.calico
bridge fdb show dev vxlan.calico
ip neigh show dev vxlan.calico
ss -lun | grep 4789

# 抓包
sudo tcpdump -ni <underlay-if> udp port 4789 -vv
sudo tcpdump -ni vxlan.calico -vv

# MTU/链路
ip addr show vxlan.calico | grep mtu
ping -M do -s 1400 <remote-node-ip>

# 故障辅助
sudo bridge fdb flush dev vxlan.calico      # 谨慎
sudo ip neigh flush dev vxlan.calico        # 谨慎
calicoctl node diags
```

> 更多 VXLAN 原理与对比可参考同目录的《calico-network-modes.md》中的 VXLAN 章节。