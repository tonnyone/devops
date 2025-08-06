# 路由器与路由表详解

## 📋 路由表基础

### 什么是路由表？

路由表（Routing Table）是网络设备（路由器、主机）用来存储网络路径信息的数据结构。它告诉设备如何将数据包转发到目标网络。

```
┌─────────────────────────────────────────────────────────┐
│                    路由表的作用                         │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  数据包到达 → 查询路由表 → 确定出接口 → 转发数据包      │
│                    ↓                                    │
│              路由表条目：                               │
│              目标网络 → 下一跳 → 出接口 → 度量值        │
│                                                         │
│  就像GPS导航：                                          │
│  "要到北京" → "走京沪高速" → "从2号出口" → "500公里"     │
└─────────────────────────────────────────────────────────┘
```

### 路由表的基本结构

```bash
# Linux路由表示例
$ ip route show
default via 192.168.1.1 dev eth0 
10.244.0.0/16 dev docker0 proto kernel scope link src 10.244.0.1 
10.244.1.0/24 via 192.168.1.101 dev eth0 proto bird 
172.17.0.0/16 dev docker0 proto kernel scope link src 172.17.0.1 
192.168.1.0/24 dev eth0 proto kernel scope link src 192.168.1.100 

# 路由表字段解释
┌──────────────┬─────────────────┬──────────────┬─────────────┐
│   目标网络   │     下一跳      │   出接口     │   协议来源  │
├──────────────┼─────────────────┼──────────────┼─────────────┤
│ default      │ 192.168.1.1     │ eth0         │ 默认路由    │
│ 10.244.1.0/24│ 192.168.1.101   │ eth0         │ BGP学习     │
│ 192.168.1.0/24│ 直连           │ eth0         │ 内核生成    │
└──────────────┴─────────────────┴──────────────┴─────────────┘
```

### 路由表条目详解

每个路由条目包含以下关键信息：

#### 1. 目标网络（Destination）
```bash
# 不同类型的目标网络
0.0.0.0/0        # 默认路由（所有未匹配的流量）
192.168.1.0/24   # 特定网段
192.168.1.100/32 # 主机路由（单个IP）
10.244.0.0/16    # 容器网络段
```

#### 2. 下一跳（Next Hop/Gateway）
```bash
# 下一跳的不同形式
192.168.1.1      # 指定网关IP
0.0.0.0          # 直连网络（本地交付）
*                # 直连网络的另一种表示
```

#### 3. 出接口（Interface）
```bash
# 常见接口类型
eth0             # 以太网接口
wlan0            # WiFi接口
lo               # 回环接口
docker0          # Docker桥接口
tunl0            # IPIP隧道接口
```

#### 4. 路由度量（Metric）
```bash
# 度量值表示路由优先级
metric 0         # 最高优先级
metric 100       # 中等优先级
metric 600       # 较低优先级（如WiFi）
```

#### 5. 协议来源（Protocol）
```bash
kernel          # 内核自动生成
static          # 静态配置
dhcp            # DHCP获得
bird            # BGP协议学习
```

## 🔍 路由匹配原理

### 最长前缀匹配

路由表使用**最长前缀匹配**（Longest Prefix Match）算法选择最佳路由：

```
例子：数据包目标IP为 192.168.1.100

路由表中的候选路由：
1. 0.0.0.0/0          (默认路由，匹配长度0)
2. 192.168.0.0/16     (匹配长度16)
3. 192.168.1.0/24     (匹配长度24) ← 最长匹配，选择此条
4. 192.168.1.100/32   (如果存在，匹配长度32，最优)

选择原则：匹配长度越长，路由越具体，优先级越高
```

### 路由匹配流程图

```
┌─────────────────────────────────────────────────────────┐
│                  路由匹配决策流程                       │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  数据包到达                                             │
│       ↓                                                 │
│  提取目标IP地址                                         │
│       ↓                                                 │
│  遍历路由表条目                                         │
│       ↓                                                 │
│  找到所有匹配的路由 ──┐                                 │
│       ↓              │                                  │
│  按前缀长度排序 ←────┘                                  │
│       ↓                                                 │
│  选择最长匹配 ──→ 相同长度时按度量值选择                │
│       ↓                                                 │
│  确定下一跳和出接口                                     │
│       ↓                                                 │
│  转发数据包                                             │
└─────────────────────────────────────────────────────────┘
```

## 📊 路由表类型

### 1. 主机路由表

```bash
# 查看主机路由表
ip route show table main

# 输出示例
default via 192.168.1.1 dev eth0 proto dhcp metric 100 
169.254.0.0/16 dev eth0 scope link metric 1000 
192.168.1.0/24 dev eth0 proto dhcp scope link src 192.168.1.100 metric 100
```

### 2. 策略路由表

Linux支持多个路由表，用于实现复杂的路由策略：

```bash
# 查看所有路由表
ip route show table all

# 常见系统路由表
ip route show table local    # 本地路由表
ip route show table main     # 主路由表
ip route show table default  # 默认路由表

# 自定义路由表
echo "200 custom_table" >> /etc/iproute2/rt_tables
ip route add 10.0.0.0/8 via 192.168.1.1 table custom_table
```

### 3. 路由器路由表

```bash
# 企业路由器路由表示例（Cisco格式）
Router# show ip route
Gateway of last resort is 203.0.113.1 to network 0.0.0.0

C    192.168.1.0/24 is directly connected, GigabitEthernet0/1
C    192.168.2.0/24 is directly connected, GigabitEthernet0/2
S*   0.0.0.0/0 [1/0] via 203.0.113.1
O    10.0.1.0/24 [110/2] via 192.168.1.2, 00:05:32, GigabitEthernet0/1
B    172.16.0.0/16 [20/0] via 192.168.2.100, 00:10:15

# 路由代码说明
C - Connected (直连)
S - Static (静态)
O - OSPF
B - BGP
* - 候选默认路由
```

## ⚙️ 路由表操作

### 查看路由表

```bash
# 基本查看命令
ip route                    # 显示主路由表
ip route show table main   # 显示主路由表
ip route show table local  # 显示本地路由表
route -n                   # 传统命令（数字格式）
netstat -rn                # 另一种查看方式

# 查看特定路由
ip route get 8.8.8.8       # 查看到达8.8.8.8的路由
ip route show 192.168.1.0/24  # 查看特定网段路由
```

### 添加路由

```bash
# 添加静态路由
ip route add 10.0.0.0/8 via 192.168.1.1 dev eth0

# 添加主机路由
ip route add 192.168.2.100/32 via 192.168.1.1

# 添加默认路由
ip route add default via 192.168.1.1 dev eth0

# 添加带度量值的路由
ip route add 10.0.0.0/8 via 192.168.1.1 metric 100

# 永久保存路由（Ubuntu/Debian）
echo "10.0.0.0/8 via 192.168.1.1 dev eth0" >> /etc/network/interfaces

# 永久保存路由（CentOS/RHEL）
echo "10.0.0.0/8 via 192.168.1.1 dev eth0" >> /etc/sysconfig/static-routes
```

### 删除路由

```bash
# 删除特定路由
ip route del 10.0.0.0/8 via 192.168.1.1

# 删除默认路由
ip route del default via 192.168.1.1

# 删除所有路由（危险操作）
ip route flush table main
```

### 修改路由

```bash
# 替换路由
ip route replace 10.0.0.0/8 via 192.168.1.2 dev eth0

# 修改默认路由
ip route replace default via 192.168.1.2 dev eth0
```

## 🔧 高级路由表功能

### 1. 负载均衡路由

```bash
# 配置等价多路径（ECMP）
ip route add 10.0.0.0/8 \
  nexthop via 192.168.1.1 weight 1 \
  nexthop via 192.168.1.2 weight 1

# 查看负载均衡路由
ip route show 10.0.0.0/8
# 输出：10.0.0.0/8 
#         nexthop via 192.168.1.1 dev eth0 weight 1 
#         nexthop via 192.168.1.2 dev eth0 weight 1
```

### 2. 源路由（Policy Routing）

```bash
# 基于源IP的路由策略
ip rule add from 192.168.1.0/24 table 100
ip route add default via 10.0.0.1 table 100

# 基于标记的路由策略
iptables -t mangle -A OUTPUT -d 8.8.8.8 -j MARK --set-mark 1
ip rule add fwmark 1 table 200
ip route add default via 172.16.0.1 table 200

# 查看路由规则
ip rule show
```

### 3. 路由缓存

```bash
# 查看路由缓存（较新内核已移除）
ip route show cached

# 清除路由缓存
ip route flush cache

# 查看邻居缓存（ARP表）
ip neighbor show
```

## 🐛 路由表故障排除

### 常见问题诊断

#### 1. 网络不通排查流程

```bash
# 1. 检查路由表
ip route show

# 2. 测试连通性
ping -c 3 目标IP

# 3. 跟踪路由路径
traceroute 目标IP
# 或者
mtr 目标IP

# 4. 检查ARP表
ip neighbor show

# 5. 检查网络接口
ip addr show
```

#### 2. 默认路由问题

```bash
# 检查默认路由
ip route | grep default

# 没有默认路由时添加
ip route add default via 192.168.1.1 dev eth0

# 多个默认路由冲突
ip route del default  # 删除所有默认路由
ip route add default via 正确网关IP dev eth0
```

#### 3. 路由优先级问题

```bash
# 查看详细路由信息
ip route show table all

# 检查路由规则
ip rule show

# 测试特定路由
ip route get 目标IP via 指定网关
```

### 调试工具

```bash
# 1. 路由跟踪
traceroute -n 8.8.8.8        # 显示数字IP
mtr --report-cycles 10 8.8.8.8  # 持续监控

# 2. 网络诊断
ss -tuln                      # 查看监听端口
netstat -i                    # 接口统计
ethtool eth0                  # 网卡状态

# 3. 数据包捕获
tcpdump -i eth0 icmp          # 抓取ICMP包
tcpdump -i any host 8.8.8.8   # 抓取特定主机流量
```

## 📈 路由表性能优化

### 1. 路由聚合

```bash
# 将多个细分路由聚合成一个
# 替换前：
ip route add 192.168.1.0/24 via 10.0.0.1
ip route add 192.168.2.0/24 via 10.0.0.1
ip route add 192.168.3.0/24 via 10.0.0.1

# 聚合后：
ip route add 192.168.0.0/22 via 10.0.0.1  # 覆盖上述三个网段
```

### 2. 静态路由优化

```bash
# 使用更具体的路由减少查找时间
# 推荐：
ip route add 192.168.1.100/32 via 10.0.0.1  # 主机路由

# 不推荐：
ip route add 192.168.0.0/16 via 10.0.0.1    # 过于宽泛
```

### 3. 接口绑定

```bash
# 直接指定接口，减少ARP查询
ip route add 10.0.0.0/8 dev eth1  # 直连接口
# 而不是
ip route add 10.0.0.0/8 via 网关IP  # 需要ARP解析
```

## 💡 最佳实践

### 1. 路由表设计原则

```bash
# 1. 遵循层次化设计
核心层    ← 默认路由指向上级
汇聚层    ← 聚合下级网段
接入层    ← 具体主机路由

# 2. 避免路由环路
- 确保路由对称性
- 使用路由度量值
- 定期检查路由路径
```

### 2. 监控和维护

```bash
# 定期备份路由表
ip route show > /backup/route_$(date +%Y%m%d).txt

# 监控路由变化
# 使用脚本定期检查关键路由
#!/bin/bash
CRITICAL_ROUTES=("8.8.8.8" "192.168.1.1")
for route in "${CRITICAL_ROUTES[@]}"; do
    if ! ping -c 1 -W 1 "$route" &>/dev/null; then
        echo "警告：无法访问 $route"
    fi
done
```

### 3. 安全考虑

```bash
# 1. 防止路由劫持
- 使用静态路由代替动态路由
- 限制BGP邻居
- 启用路由认证

# 2. 访问控制
- 使用防火墙规则配合路由
- 实施入口过滤
- 监控异常路由更新
```

---

*路由表是网络通信的基础，理解其工作原理对于网络管理和故障排除至关重要。*
