# Calico 初级学习指南 (1-2周)

> 本文档专为 Calico 初学者设计，涵盖基础概念到实践操作的完整学习路径

## 🎯 学习目标

通过1-2周的学习，你将掌握：
- Kubernetes 网络模型基础
- CNI (Container Network Interface) 基本概念
- Calico 的安装和基本配置
- Pod 间网络通信的实践操作

## 📚 第一周：理论基础

### Day 1-2: Kubernetes 网络模型

#### 📖 核心概念
Kubernetes 网络模型遵循以下原则：
- **扁平网络**: 所有 Pod 都在一个扁平的网络空间中
- **无 NAT 通信**: Pod 之间可以直接通信，无需 NAT
- **唯一 IP**: 每个 Pod 都有唯一的集群内 IP 地址

#### 🔍 深入理解：为什么 Kubernetes 可以实现无 NAT 通信？

**无 NAT 通信**是 Kubernetes 网络模型的关键特性，这意味着 Pod 之间的通信不需要网络地址转换。让我们深入了解其技术原理：

##### 1. **什么是 NAT？为什么要避免它？**

**NAT (Network Address Translation) 的问题：**
```
传统 NAT 通信流程：
Pod-A (私有IP: 192.168.1.10:8080) 
    ↓
NAT 网关 (转换为公网IP: 203.0.113.1:30001)
    ↓ (通过互联网)
NAT 网关 (转换回私有IP: 192.168.2.1:8080)
    ↓
Pod-B (私有IP: 192.168.2.20:8080)

问题：
- IP 地址被转换，应用无法获得真实的源 IP
- 端口映射复杂，需要额外的端口管理
- 连接状态跟踪开销大
- 某些协议（如 FTP、SIP）不兼容 NAT
```

##### 2. **Kubernetes 如何实现无 NAT 通信？**

**核心原理：全局路由表 + 直接路由**

```
Pod-A (10.244.1.10:8080) ←直接路由→ Pod-B (10.244.2.20:8080)

无需地址转换！源 IP 和目标 IP 都保持不变
```

**技术实现步骤：**

**A. 集群级 IP 地址分配**
```bash
# 每个节点分配一个子网段
节点 A: 10.244.1.0/24  (可分配 254 个 Pod IP)
节点 B: 10.244.2.0/24  (可分配 254 个 Pod IP)
节点 C: 10.244.3.0/24  (可分配 254 个 Pod IP)

# Pod 获得集群内全局唯一的 IP
Pod-A: 10.244.1.10  (在节点 A)
Pod-B: 10.244.2.20  (在节点 B)
```

**B. 自动路由分发机制**
```bash
# 当 Pod 创建时，Calico 自动更新所有节点的路由表

# 节点 A 的路由表
10.244.1.0/24 dev cali123abc scope link  # 本地 Pod
10.244.2.0/24 via 192.168.100.101       # 到节点 B 的路由
10.244.3.0/24 via 192.168.100.102       # 到节点 C 的路由

# 节点 B 的路由表  
10.244.2.0/24 dev cali456def scope link  # 本地 Pod
10.244.1.0/24 via 192.168.100.100       # 到节点 A 的路由
10.244.3.0/24 via 192.168.100.102       # 到节点 C 的路由
```

**C. BGP 协议自动同步**
```
节点 A ←→ BGP 会话 ←→ 节点 B
   ↓                    ↓
"我有 10.244.1.0/24"   "我有 10.244.2.0/24"
   ↓                    ↓
自动学习并更新路由表
```

##### 3. **数据包传输的完整流程**

```
发送端 (Pod-A: 10.244.1.10)
    ↓
1. 应用发送数据：src=10.244.1.10:8080, dst=10.244.2.20:8080
    ↓
2. 内核查路由表：10.244.2.0/24 via 192.168.100.101
    ↓
3. 封装以太网帧：dst_mac=节点B的MAC地址
    ↓
4. 物理网络传输到节点 B
    ↓
5. 节点 B 接收并查路由：10.244.2.20 dev cali456def
    ↓
6. 直接转发到 Pod-B
    ↓
接收端 (Pod-B: 10.244.2.20) 收到原始 IP 包
```

**关键点：整个过程中源 IP 和目标 IP 从未改变！**

##### 4. **与 NAT 方式的对比**

| 特性 | Kubernetes 无 NAT | 传统 NAT 网络 |
|------|------------------|---------------|
| **源 IP 保持** | ✅ 完全保持 | ❌ 被转换 |
| **端口冲突** | ✅ 无冲突 | ❌ 需要端口映射 |
| **协议支持** | ✅ 所有协议 | ❌ 某些协议受限 |
| **性能开销** | ✅ 直接路由，低延迟 | ❌ NAT 转换开销 |
| **连接跟踪** | ✅ 无需状态表 | ❌ 需要维护连接状态 |
| **双向通信** | ✅ 对等通信 | ❌ 需要端口转发配置 |

##### 5. **实际验证无 NAT 通信**

**验证命令：**
```bash
# 1. 创建两个测试 Pod
kubectl run pod-a --image=nicolaka/netshoot --command -- sleep 3600
kubectl run pod-b --image=nicolaka/netshoot --command -- sleep 3600

# 2. 获取 Pod IP
POD_A_IP=$(kubectl get pod pod-a -o jsonpath='{.status.podIP}')
POD_B_IP=$(kubectl get pod pod-b -o jsonpath='{.status.podIP}')

# 3. 从 Pod-A 访问 Pod-B，查看网络包
kubectl exec pod-a -- tcpdump -i eth0 -n host $POD_B_IP &
kubectl exec pod-a -- ping -c 1 $POD_B_IP

# 4. 验证源 IP 保持不变
kubectl exec pod-b -- netstat -an | grep $POD_A_IP
```

**期望结果：**
- Pod-B 看到的连接源 IP 就是 Pod-A 的真实 IP
- 没有任何地址转换痕迹

##### 6. **无 NAT 的业务价值**

**A. 安全审计**
```bash
# 应用可以获得真实的客户端 IP
kubectl logs web-pod | grep "Client IP: 10.244.1.10"
# 而不是看到网关的 IP
```

**B. 服务发现简化**
```yaml
# 服务可以直接使用 Pod IP 通信
apiVersion: v1
kind: Service
metadata:
  name: backend-service
spec:
  clusterIP: None  # Headless Service
  selector:
    app: backend
```

**C. 协议兼容性**
```bash
# 支持需要端到端连接的协议
# 如 gRPC 双向流、WebSocket、数据库连接池等
```

##### 7. **Calico 的无 NAT 实现细节**

**veth pair + 直接路由：**
```
Pod 网络命名空间
    ↓ (veth pair)
主机网络命名空间
    ↓ (直接路由，无 bridge)
物理网卡
    ↓ (BGP 路由)
目标节点
```

**与其他 CNI 的对比：**
- **Flannel (VXLAN 模式)**: 使用隧道封装，有额外开销
- **Flannel (host-gw 模式)**: 类似 Calico，直接路由
- **Weave**: 可选择路由或隧道模式
- **Calico**: 默认直接路由，性能最优

这就是 Kubernetes 实现无 NAT 通信的完整原理——通过全局 IP 分配 + 自动路由分发 + BGP 协议同步，实现了真正的"扁平网络"！

#### 🔍 深入理解：为什么说 Kubernetes 是扁平网络？

**扁平网络 (Flat Network)** 是 Kubernetes 网络模型的核心特征，这里的"扁平"有着特定的技术含义：

##### 1. **网络拓扑层面的扁平**
```
传统多层网络架构：
┌─────────────────────────────────────┐
│         应用层 (Layer 7)            │
├─────────────────────────────────────┤
│         会话层 (Layer 5)            │
├─────────────────────────────────────┤
│         网络层 (Layer 3)            │
├─────────────────────────────────────┤
│         数据链路层 (Layer 2)        │
└─────────────────────────────────────┘

Kubernetes 扁平网络：
┌─────────────────────────────────────┐
│    所有 Pod 在同一个 IP 网段        │
│    Pod-A(10.1.1.10) ←→ Pod-B(10.1.2.20) │
│         直接三层路由通信            │
└─────────────────────────────────────┘
```

##### 2. **地址空间的扁平化**
- **统一地址空间**: 整个集群使用同一个大的 IP 地址空间（如 10.244.0.0/16）
- **无子网隔离**: Pod 不需要通过网关或 NAT 就能直接访问其他 Pod
- **路由透明**: 每个 Pod 都能"看到"集群内所有其他 Pod 的真实 IP

##### 3. **与传统网络架构的对比**

**传统虚拟化网络（非扁平）:**
```
VM-A (192.168.1.10) 
    ↓ (通过虚拟交换机)
虚拟网关 (192.168.1.1)
    ↓ (NAT 转换)
物理网络 (10.0.0.0/8)
    ↓ (路由)
虚拟网关 (192.168.2.1)
    ↓ (通过虚拟交换机)  
VM-B (192.168.2.20)
```

**Kubernetes 扁平网络:**
```
Pod-A (10.244.1.10) ←直接路由→ Pod-B (10.244.2.20)
```

##### 4. **技术实现原理**

扁平网络通过以下技术实现：

**A. 集群级别的路由表**
```bash
# 节点 A 的路由表
10.244.1.0/24 dev cali123abc scope link  # 本节点 Pod
10.244.2.0/24 via 192.168.1.101        # 节点 B 的 Pod

# 节点 B 的路由表  
10.244.2.0/24 dev cali456def scope link  # 本节点 Pod
10.244.1.0/24 via 192.168.1.100        # 节点 A 的 Pod
```

**B. BGP 路由协议自动同步**
```
节点 A ←→ BGP 会话 ←→ 节点 B
   ↓                    ↓
自动学习对方的Pod路由信息
```

##### 5. **扁平网络的优势**

1. **简化网络模型**: 
   - 开发者不需要考虑复杂的网络拓扑
   - 服务发现更加直观

2. **性能优势**:
   - 减少网络跳数，降低延迟
   - 避免 NAT 转换的性能开销

3. **故障排查简单**:
   - 网络路径清晰可见
   - 问题定位更容易

4. **支持原生协议**:
   - 支持多播、广播等特殊协议
   - 保持网络协议的完整性

##### 6. **实际验证扁平网络**

你可以通过以下命令验证扁平网络特性：

```bash
# 1. 查看所有 Pod IP (它们在同一个地址空间)
kubectl get pods -A -o wide

# 2. 从任意 Pod ping 另一个 Pod (直接使用 IP)
kubectl exec pod-a -- ping 10.244.2.20

# 3. 查看路由路径 (应该是直接路由，无 NAT)
kubectl exec pod-a -- traceroute 10.244.2.20

# 4. 验证无端口映射 (Pod 端口就是实际端口)
kubectl exec pod-a -- netstat -tlnp
```

##### 7. **扁平网络 vs 其他网络模式**

| 网络模式 | 地址转换 | 网络性能 | 复杂度 | 协议支持 |
|---------|----------|----------|---------|----------|
| 扁平网络 | 无 | 高 | 低 | 完整 |
| NAT 网络 | 有 | 中等 | 中等 | 受限 |
| 隧道网络 | 封装 | 低 | 高 | 完整 |

这就是为什么 Kubernetes 选择扁平网络模型的原因——它提供了最接近"裸机"网络的体验，同时保持了容器化的便利性。

#### 🔍 网络层次结构
```
┌─────────────────────────────────────┐
│              Service                │  ← 4. 服务发现和负载均衡
├─────────────────────────────────────┤
│               Pod                   │  ← 3. 应用容器
├─────────────────────────────────────┤
│              Node                   │  ← 2. 物理/虚拟机节点
├─────────────────────────────────────┤
│             Cluster                 │  ← 1. 集群网络
└─────────────────────────────────────┘
```

#### 📝 学习任务
1. **理解文档**: 阅读 [Kubernetes 网络模型官方文档](https://kubernetes.io/docs/concepts/cluster-administration/networking/)

2. **实践练习**: 创建简单的 Pod，观察其网络配置
   ```bash
   # 创建测试 Pod
   kubectl run test-pod --image=nginx --rm -it -- /bin/bash
   
   # 查看网络接口
   ip addr show
   
   # 查看路由表
   ip route show
   ```

### Day 3-4: CNI 基本概念

#### 📖 什么是 CNI
CNI (Container Network Interface) 是容器网络的标准化接口，定义了：
- 容器运行时如何调用网络插件
- 网络插件如何配置容器网络
- 统一的网络管理标准

#### 🏗️ CNI 工作流程
```
容器创建 → 调用 CNI 插件 → 分配 IP → 配置网络接口 → 设置路由
```

#### 📝 学习任务

1. **查看 CNI 配置**:
   ```bash
   # 查看 CNI 配置文件
   sudo ls /etc/cni/net.d/
   sudo cat /etc/cni/net.d/10-calico.conflist
   ```

2. **理解 CNI 规范**: 阅读 [CNI 官方规范](https://github.com/containernetworking/cni)

### Day 5-7: Calico 架构基础

#### 📖 Calico 简介复习
Calico 是一个基于 BGP 的网络解决方案，提供：
- **纯三层路由**: 不使用 overlay 网络
- **高性能**: 接近原生网络性能
- **网络策略**: 细粒度的安全控制

#### 🏗️ 核心组件详解

##### 1. Felix (核心代理)

```bash
# 查看 Felix 状态
kubectl logs -n calico-system -l k8s-app=calico-node | grep felix
```

**主要功能**:
- 监听 Kubernetes API 变化
- 编程 Linux 内核路由表
- 配置 iptables 规则
- 报告节点健康状态

##### 2. BIRD (BGP 客户端)
```bash
# 查看 BGP 状态
kubectl exec -n calico-system <calico-node-pod> -- bird -s
```

**主要功能**:
- 与其他节点交换路由信息
- 维护 BGP 会话
- 确保路由一致性

##### 3. CNI 插件
**主要功能**:
- Pod 创建时分配 IP 地址
- 配置 veth pair
- 设置默认路由

#### 📝 学习任务
1. **查看组件状态**:
   ```bash
   # 查看 Calico 相关 Pod
   kubectl get pods -n calico-system
   
   # 查看详细信息
   kubectl describe pod -n calico-system <calico-node-pod>
   ```

2. **理解数据流向**:
   ```
   Pod A → veth pair → 主机路由表 → BGP → 目标主机 → veth pair → Pod B
   ```

## 🛠️ 第二周：实践操作

### Day 8-10: Calico 安装配置

#### 🚀 环境准备

##### 前置要求
- Kubernetes 集群 (版本 1.20+)
- 节点间网络互通
- 关闭 swap

##### 安装步骤

**方法一: 使用 Operator (推荐)**
```bash
# 1. 安装 Tigera Operator
kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.25.0/manifests/tigera-operator.yaml

# 2. 等待 Operator 就绪
kubectl wait --for=condition=Ready pod -l name=tigera-operator -n tigera-operator --timeout=300s

# 3. 安装 Calico
kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.25.0/manifests/custom-resources.yaml
```

**方法二: 直接安装**
```bash
# 直接安装 Calico
kubectl apply -f https://raw.githubusercontent.com/projectcalico/calico/v3.25.0/manifests/calico.yaml
```

#### ✅ 验证安装

```bash
# 1. 检查 Pod 状态
kubectl get pods -n calico-system

# 2. 检查节点状态
kubectl get nodes -o wide

# 3. 验证所有节点就绪
kubectl get nodes --no-headers | awk '{print $2}' | grep -v Ready || echo "所有节点就绪"
```

#### 📝 学习任务
1. **完整安装流程**: 在测试环境完成 Calico 安装
2. **故障排查**: 如果安装失败，学会查看日志排查问题
   ```bash
   # 查看 Operator 日志
   kubectl logs -n tigera-operator deployment/tigera-operator
   
   # 查看 Calico 节点日志
   kubectl logs -n calico-system -l k8s-app=calico-node
   ```

### Day 11-12: 基本工具使用

#### 🔧 安装 calicoctl

```bash
# 下载 calicoctl (Linux)
curl -L https://github.com/projectcalico/calico/releases/latest/download/calicoctl-linux-amd64 -o calicoctl

# 设置执行权限
chmod +x calicoctl

# 移动到系统路径
sudo mv calicoctl /usr/local/bin/

# 验证安装
calicoctl version
```

#### 📊 基本命令练习

```bash
# 1. 查看节点状态
calicoctl node status

# 2. 查看 IP 池
calicoctl get ippool -o wide

# 3. 查看工作负载端点
calicoctl get workloadendpoint

# 4. 查看 BGP 配置
calicoctl get bgpconfig default -o yaml
```

#### 📝 学习任务
1. **熟练使用工具**: 练习所有基本命令
2. **理解输出**: 理解每个命令输出的含义

### Day 13-14: Pod 间通信实践

#### 🧪 创建测试环境

```bash
# 1. 创建测试命名空间
kubectl create namespace calico-test

# 2. 创建测试 Pod
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: pod-a
  namespace: calico-test
  labels:
    app: test-app-a
spec:
  containers:
  - name: container-a
    image: nicolaka/netshoot
    command: ["/bin/bash", "-c", "sleep 3600"]
---
apiVersion: v1
kind: Pod
metadata:
  name: pod-b
  namespace: calico-test
  labels:
    app: test-app-b
spec:
  containers:
  - name: container-b
    image: nicolaka/netshoot
    command: ["/bin/bash", "-c", "sleep 3600"]
EOF
```

#### 🔍 网络连通性测试

```bash
# 1. 获取 Pod IP
kubectl get pods -n calico-test -o wide

# 2. 测试 Pod 间连通性
POD_B_IP=$(kubectl get pod pod-b -n calico-test -o jsonpath='{.status.podIP}')
kubectl exec -n calico-test pod-a -- ping -c 3 $POD_B_IP

# 3. 测试跨节点通信 (如果 Pod 在不同节点)
kubectl exec -n calico-test pod-a -- traceroute $POD_B_IP
```

#### 🔬 深入分析网络路径

```bash
# 1. 查看路由表
kubectl exec -n calico-test pod-a -- ip route show

# 2. 查看网络接口
kubectl exec -n calico-test pod-a -- ip addr show

# 3. 在主机上查看路由
# (在 Pod 所在的节点执行)
ip route show | grep <pod-ip>
```

#### 📝 学习任务
1. **网络测试**: 完成所有连通性测试
2. **路径分析**: 理解数据包的传输路径
3. **问题排查**: 如果遇到连通性问题，学会分析和解决

## 🎯 初级阶段总结

### ✅ 应该掌握的技能

#### 理论知识
- [x] Kubernetes 网络模型基本原理
- [x] CNI 的作用和工作机制
- [x] Calico 架构和核心组件
- [x] BGP 路由的基本概念

#### 实践技能
- [x] 在 Kubernetes 集群中安装 Calico
- [x] 使用 kubectl 和 calicoctl 基本命令
- [x] 验证 Pod 间网络连通性
- [x] 进行基本的网络故障排查

### 🔍 自我检测题

1. **理论题**:
   - Kubernetes 网络模型的三个基本要求是什么？
   - CNI 在容器网络中扮演什么角色？
   - Calico 的 Felix 组件主要负责什么功能？

2. **实践题**:
   - 如何检查 Calico 安装是否成功？
   - 如何查看某个 Pod 的网络配置？
   - 如何测试两个 Pod 之间的网络连通性？

### 🚀 进入中级阶段的准备

完成初级阶段后，你应该：
1. **巩固基础**: 确保理论知识扎实
2. **多做实验**: 在不同环境中练习安装和配置
3. **阅读文档**: 开始关注 Calico 的高级特性
4. **准备进阶**: 开始学习网络策略和BGP配置

## 📚 推荐阅读

### 必读文档
- [Kubernetes 网络概念](https://kubernetes.io/docs/concepts/cluster-administration/networking/)
- [Calico 快速开始](https://docs.projectcalico.org/getting-started/)
- [CNI 规范](https://github.com/containernetworking/cni/blob/master/SPEC.md)

### 参考资料
- [Kubernetes 网络故障排查](https://kubernetes.io/docs/tasks/debug-application-cluster/debug-service/)
- [Calico 故障排查指南](https://docs.projectcalico.org/maintenance/troubleshoot/)

## 🆘 常见问题

### Q1: 安装后 Pod 无法获取 IP 地址
**A**: 检查以下项目：
```bash
# 检查 CNI 配置
ls /etc/cni/net.d/

# 检查 Calico 节点状态
kubectl get pods -n calico-system

# 查看节点日志
kubectl logs -n calico-system -l k8s-app=calico-node
```

### Q2: Pod 间无法通信
**A**: 逐步排查：
```bash
# 1. 检查 Pod IP 分配
kubectl get pods -o wide

# 2. 检查路由表
calicoctl node status

# 3. 测试网络连通性
kubectl exec <pod-a> -- ping <pod-b-ip>
```

### Q3: 如何重置 Calico 配置
**A**: 清理和重新安装：
```bash
# 删除 Calico 资源
kubectl delete -f https://raw.githubusercontent.com/projectcalico/calico/v3.25.0/manifests/calico.yaml

# 清理节点配置 (在每个节点执行)
sudo rm -rf /etc/cni/net.d/10-calico.conflist
sudo rm -rf /var/lib/calico/

# 重新安装
kubectl apply -f https://raw.githubusercontent.com/projectcalico/calico/v3.25.0/manifests/calico.yaml
```

---

*完成初级阶段后，建议继续学习 [Calico 中级指南](./intermediate-guide.md)*

*最后更新: 2025年7月*