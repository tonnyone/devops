# iptables 防火墙详解

## 概述

iptables是Linux系统中的包过滤防火墙，基于netfilter框架实现。它通过在内核网络协议栈的关键位置设置钩子(hook)来检查、修改或丢弃网络数据包。

## 核心概念：三表五链

### 五链 (Five Chains)

iptables的链对应netfilter在内核网络协议栈中的五个钩子点：

```
数据包流向图：

                    PREROUTING
                        ↓
                   路由判断
                   /        \
              本机接收      转发
                /              \
           INPUT              FORWARD
             ↓                   ↓
        本地进程              POSTROUTING
             ↓                   ↓
           OUTPUT              网卡输出
             ↓                   
        POSTROUTING             
             ↓                   
         网卡输出                
```

#### 1. PREROUTING链
- **位置**：数据包刚进入网络协议栈，路由判断之前
- **作用**：DNAT (目标地址转换)、端口映射
- **典型用途**：
  ```bash
  # 端口转发：将80端口转发到8080
  iptables -t nat -A PREROUTING -p tcp --dport 80 -j REDIRECT --to-port 8080
  
  # DNAT：将访问公网IP的请求转发到内网服务器
  iptables -t nat -A PREROUTING -d 公网IP -p tcp --dport 80 -j DNAT --to 192.168.1.100:80
  ```

#### 2. INPUT链
- **位置**：数据包经过路由判断，确定是发给本机的
- **作用**：控制进入本机的数据包
- **典型用途**：
  ```bash
  # 允许SSH连接
  iptables -A INPUT -p tcp --dport 22 -j ACCEPT
  
  # 拒绝特定IP访问
  iptables -A INPUT -s 192.168.1.100 -j DROP
  
  # 允许已建立的连接
  iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
  ```

#### 3. FORWARD链
- **位置**：数据包经过路由判断，确定需要转发
- **作用**：控制通过本机转发的数据包
- **典型用途**：
  ```bash
  # 允许内网访问外网
  iptables -A FORWARD -s 192.168.1.0/24 -j ACCEPT
  
  # 禁止特定网段互访
  iptables -A FORWARD -s 192.168.1.0/24 -d 192.168.2.0/24 -j DROP
  
  # NAT网关转发
  iptables -A FORWARD -i eth1 -o eth0 -j ACCEPT
  ```

#### 4. OUTPUT链
- **位置**：本机产生的数据包，发出前
- **作用**：控制本机发出的数据包
- **典型用途**：
  ```bash
  # 禁止本机访问特定网站
  iptables -A OUTPUT -d 某恶意网站IP -j DROP
  
  # 限制本机只能使用特定DNS
  iptables -A OUTPUT -p udp --dport 53 ! -d 8.8.8.8 -j DROP
  
  # 防止数据泄露
  iptables -A OUTPUT -p tcp --dport 443 -m owner --uid-owner apache -j DROP
  ```

#### 5. POSTROUTING链
- **位置**：数据包离开本机前的最后一个点
- **作用**：SNAT (源地址转换)、伪装
- **典型用途**：
  ```bash
  # SNAT：内网访问外网时改变源地址
  iptables -t nat -A POSTROUTING -s 192.168.1.0/24 -o eth0 -j SNAT --to 公网IP
  
  # MASQUERADE：动态SNAT (适用于动态IP)
  iptables -t nat -A POSTROUTING -s 192.168.1.0/24 -o ppp0 -j MASQUERADE
  
  # 源端口转换
  iptables -t nat -A POSTROUTING -p tcp --sport 80 -j SNAT --to :8080
  ```

### 三表 (Three Tables)

iptables的表定义了规则的类型和作用：

#### 1. filter表 (默认表)
- **功能**：包过滤，决定是否允许数据包通过
- **包含链**：INPUT, OUTPUT, FORWARD
- **默认策略**：通常设置为DROP或ACCEPT
- **使用场景**：
  ```bash
  # 基本防火墙规则
  iptables -A INPUT -p tcp --dport 80 -j ACCEPT      # 允许HTTP
  iptables -A INPUT -p tcp --dport 443 -j ACCEPT     # 允许HTTPS
  iptables -A INPUT -p tcp --dport 22 -j ACCEPT      # 允许SSH
  iptables -A INPUT -j DROP                          # 默认拒绝
  
  # 状态化防火墙
  iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
  iptables -A OUTPUT -m state --state NEW,ESTABLISHED,RELATED -j ACCEPT
  ```

#### 2. nat表
- **功能**：网络地址转换 (NAT)
- **包含链**：PREROUTING, POSTROUTING, OUTPUT
- **使用场景**：
  ```bash
  # 典型的NAT网关配置
  # 1. 开启转发
  echo 1 > /proc/sys/net/ipv4/ip_forward
  
  # 2. 配置SNAT让内网访问外网
  iptables -t nat -A POSTROUTING -s 192.168.1.0/24 -o eth0 -j MASQUERADE
  
  # 3. 配置DNAT让外网访问内网服务
  iptables -t nat -A PREROUTING -p tcp --dport 80 -j DNAT --to 192.168.1.100:80
  iptables -t nat -A PREROUTING -p tcp --dport 443 -j DNAT --to 192.168.1.100:443
  
  # 4. 配置端口映射
  iptables -t nat -A PREROUTING -p tcp --dport 2222 -j DNAT --to 192.168.1.100:22
  ```

#### 3. mangle表
- **功能**：修改数据包头部信息
- **包含链**：所有五条链都包含
- **使用场景**：
  ```bash
  # 修改TOS (Type of Service) 字段
  iptables -t mangle -A OUTPUT -p tcp --sport 80 -j TOS --set-tos 0x10
  
  # 修改TTL值
  iptables -t mangle -A OUTPUT -j TTL --ttl-set 64
  
  # 标记数据包 (配合tc进行流量控制)
  iptables -t mangle -A OUTPUT -p tcp --dport 80 -j MARK --set-mark 1
  
  # 修改DSCP字段 (QoS)
  iptables -t mangle -A OUTPUT -p tcp --dport 22 -j DSCP --set-dscp 46
  ```

### nat表 vs mangle表：修改功能的区别

虽然nat表和mangle表都可以修改数据包，但它们有本质区别：

#### nat表的修改特点
- **目的**：实现网络地址转换 (NAT)
- **修改内容**：IP地址和端口号
- **自动特性**：
  - 自动建立连接跟踪 (conntrack)
  - 自动处理返回数据包的反向转换
  - 只对连接的第一个数据包生效，后续数据包自动应用相同转换

```bash
# nat表示例：DNAT转换
iptables -t nat -A PREROUTING -d 192.168.1.1 -p tcp --dport 80 -j DNAT --to 10.0.0.100:8080

# 连接跟踪自动处理：
# 1. 第一个包：192.168.1.1:80 → 10.0.0.100:8080 (应用DNAT规则)
# 2. 后续包：192.168.1.1:80 → 10.0.0.100:8080 (conntrack自动转换)
# 3. 返回包：10.0.0.100:8080 → 192.168.1.1:80 (conntrack自动反向转换)
```

#### mangle表的修改特点
- **目的**：包标记、QoS、特殊处理
- **修改内容**：IP头部各种字段 (TOS、TTL、DSCP等)
- **手动特性**：
  - 每个数据包都需要单独处理
  - 不会自动建立连接跟踪
  - 需要手动处理双向流量

```bash
# mangle表示例：修改TTL
iptables -t mangle -A OUTPUT -j TTL --ttl-set 64

# 每个数据包都会被处理：
# 1. 第一个包：TTL修改为64
# 2. 第二个包：TTL修改为64
# 3. 每个包：TTL修改为64 (需要逐个处理)
```

#### 具体对比分析

| 特性 | nat表 | mangle表 |
|------|-------|----------|
| **主要用途** | 网络地址转换 | 包标记和QoS |
| **修改范围** | IP地址、端口 | 所有IP头部字段 |
| **连接跟踪** | 自动建立 | 不自动建立 |
| **反向处理** | 自动处理 | 需手动配置 |
| **处理频率** | 只处理首包 | 处理每个包 |
| **典型场景** | NAT网关、端口映射 | 流量标记、QoS |

#### 实际应用示例对比

**场景1：端口映射 (nat表最佳)**
```bash
# nat表实现 - 推荐方式
iptables -t nat -A PREROUTING -p tcp --dport 80 -j DNAT --to 192.168.1.100:8080
# 自动处理：
# - 入站：外部:80 → 192.168.1.100:8080
# - 出站：192.168.1.100:8080 → 外部:80 (自动反向转换)

# mangle表实现 - 不推荐，复杂且不完整
iptables -t mangle -A PREROUTING -p tcp --dport 80 -j DADDR --set-daddr 192.168.1.100
iptables -t mangle -A PREROUTING -p tcp --dport 80 -j DPORT --set-dport 8080
# 问题：
# - 需要手动配置返回路径
# - 每个包都要处理，性能较差
# - 不会自动维护连接状态
```

**场景2：QoS流量标记 (mangle表专用)**
```bash
# mangle表实现 - 正确方式
iptables -t mangle -A OUTPUT -p tcp --dport 22 -j DSCP --set-dscp 46    # SSH高优先级
iptables -t mangle -A OUTPUT -p tcp --dport 80 -j DSCP --set-dscp 26    # HTTP中优先级
iptables -t mangle -A OUTPUT -p tcp --dport 443 -j DSCP --set-dscp 26   # HTTPS中优先级

# nat表无法实现 - 没有相应的target
# nat表只有：SNAT, DNAT, MASQUERADE, REDIRECT, NETMAP等NAT相关target
```

**场景3：TTL修改 (只能用mangle表)**
```bash
# mangle表实现 - 唯一方式
iptables -t mangle -A OUTPUT -j TTL --ttl-set 64

# nat表无法实现 - 没有TTL相关的target
```

#### 内核处理机制差异

```c
// nat表的连接跟踪机制 (简化)
struct nf_conn *ct = nf_ct_get(skb);
if (ct && ct->status & IPS_NAT_DONE_MASK) {
    // 已有NAT转换记录，直接应用
    return nf_nat_apply_transform(skb, ct);
} else {
    // 新连接，执行NAT规则并记录
    result = nf_nat_rule_find(skb);
    nf_ct_set_nat_info(ct, result);
    return result;
}

// mangle表的直接处理机制 (简化)
for (each_packet) {
    // 每个包都执行完整的规则匹配
    result = nf_mangle_rule_find(skb);
    nf_mangle_apply_transform(skb, result);
}
```

#### 选择原则

**使用nat表的情况：**
- 需要修改IP地址或端口进行NAT转换
- 希望自动处理双向流量
- 实现负载均衡、端口映射
- 需要连接跟踪功能

**使用mangle表的情况：**
- 需要修改除IP/端口外的其他头部字段
- 实现QoS和流量分类
- 包标记用于后续处理
- 修改TTL、TOS、DSCP等字段

## NAT在不同链中的应用详解

### DNAT：PREROUTING vs OUTPUT

#### PREROUTING链中的DNAT（经典用法）
```bash
# 用途：处理从外部进入的连接
iptables -t nat -A PREROUTING -d 公网IP -p tcp --dport 80 -j DNAT --to 192.168.1.100:8080

# 数据包流向：
外部客户端 → 网络接口 → PREROUTING(DNAT) → 路由判断 → FORWARD链 → 内网服务器
```

**应用场景：**
- 公网服务器端口映射到内网
- 负载均衡器分发外部请求
- DMZ区域服务暴露

#### OUTPUT链中的DNAT（特殊用法）
```bash
# 用途：处理本机发出的连接
iptables -t nat -A OUTPUT -d 8.8.8.8 -p tcp --dport 53 -j DNAT --to 192.168.1.1:53

# 数据包流向：
本机进程 → OUTPUT(DNAT) → 路由判断 → POSTROUTING → 网络接口 → 目标服务器
```

**应用场景：**
- 本机DNS重定向（将8.8.8.8重定向到内网DNS）
- 本机应用代理（将特定服务重定向到代理服务器）
- 开发环境流量劫持

**OUTPUT链DNAT的典型案例：**
```bash
#!/bin/bash
# 案例：将本机所有DNS查询重定向到内网DNS服务器

# 原始情况：本机进程查询8.8.8.8
# dig @8.8.8.8 google.com

# 配置OUTPUT DNAT
iptables -t nat -A OUTPUT -d 8.8.8.8 -p udp --dport 53 -j DNAT --to 192.168.1.1:53
iptables -t nat -A OUTPUT -d 8.8.8.8 -p tcp --dport 53 -j DNAT --to 192.168.1.1:53

# 效果：本机进程以为在查询8.8.8.8，实际查询的是192.168.1.1
# 应用程序无感知，管理员可以控制DNS查询行为
```

### SNAT：POSTROUTING vs OUTPUT

#### POSTROUTING链中的SNAT（标准用法）
```bash
# 用途：处理即将离开本机的所有数据包（本机产生+转发）
iptables -t nat -A POSTROUTING -s 192.168.1.0/24 -o eth0 -j SNAT --to 公网IP

# 覆盖场景：
# 1. 本机产生的数据包：本机进程 → OUTPUT → POSTROUTING(SNAT) → 网络接口
# 2. 转发的数据包：    网络接口 → FORWARD → POSTROUTING(SNAT) → 网络接口
```

**应用场景：**
- NAT网关（内网访问外网）
- VPN服务器出口
- 多网卡服务器的源地址统一

#### OUTPUT链中的SNAT（罕见用法）
```bash
# 用途：仅处理本机进程发出的数据包
iptables -t nat -A OUTPUT -s 192.168.1.10 -j SNAT --to 192.168.1.20

# 数据包流向：
本机进程 → OUTPUT(SNAT) → 路由判断 → POSTROUTING → 网络接口
```

**应用场景：**
- 进程级别的源地址伪装
- 特定应用的源IP控制
- 多IP服务器的出口IP管理

**OUTPUT链SNAT的实际案例：**
```bash
#!/bin/bash
# 案例：多IP服务器的应用级源IP控制

# 服务器有多个IP：192.168.1.10, 192.168.1.11, 192.168.1.12
# 需求：不同应用使用不同的源IP发出请求

# Web服务使用.10发出请求
iptables -t nat -A OUTPUT -m owner --uid-owner www-data -j SNAT --to 192.168.1.10

# 数据库服务使用.11发出请求  
iptables -t nat -A OUTPUT -m owner --uid-owner mysql -j SNAT --to 192.168.1.11

# 备份服务使用.12发出请求
iptables -t nat -A OUTPUT -m owner --uid-owner backup -j SNAT --to 192.168.1.12

# 效果：同一台服务器上的不同服务使用不同源IP对外通信
```

### NAT链使用的完整矩阵

| NAT类型 | PREROUTING | INPUT | FORWARD | OUTPUT | POSTROUTING |
|---------|------------|-------|---------|--------|-------------|
| **DNAT** | ✅ 外部到内部 | ❌ | ❌ | ✅ 本机重定向 | ❌ |
| **SNAT** | ❌ | ❌ | ❌ | ✅ 本机源控制 | ✅ 全局源控制 |
| **REDIRECT** | ✅ 透明代理 | ❌ | ❌ | ✅ 本机代理 | ❌ |
| **MASQUERADE** | ❌ | ❌ | ❌ | ✅ 动态IP | ✅ 动态IP |

### 数据包流向与NAT时机

```bash
# 完整的数据包流向图（包含NAT时机）

【外部到内部的数据包】
网络接口 → PREROUTING(DNAT可能) → 路由判断 → INPUT/FORWARD → POSTROUTING(SNAT可能) → 网络接口/本地进程

【本机发出的数据包】  
本地进程 → OUTPUT(DNAT/SNAT可能) → 路由判断 → POSTROUTING(SNAT可能) → 网络接口

【转发的数据包】
网络接口 → PREROUTING(DNAT可能) → 路由判断 → FORWARD → POSTROUTING(SNAT可能) → 网络接口
```

### 实际应用场景对比

#### 场景1：企业NAT网关
```bash
# 标准配置：PREROUTING DNAT + POSTROUTING SNAT
# 外网访问内网服务
iptables -t nat -A PREROUTING -d 公网IP -p tcp --dport 80 -j DNAT --to 192.168.1.100:80

# 内网访问外网
iptables -t nat -A POSTROUTING -s 192.168.1.0/24 -o eth0 -j MASQUERADE
```

#### 场景2：开发环境流量控制
```bash
# 特殊配置：OUTPUT DNAT + OUTPUT SNAT
# 开发机访问测试环境：将生产环境API重定向到测试环境
iptables -t nat -A OUTPUT -d 生产API_IP -p tcp --dport 443 -j DNAT --to 测试API_IP:443

# 同时修改源IP，模拟特定来源
iptables -t nat -A OUTPUT -d 测试API_IP -p tcp --dport 443 -j SNAT --to 模拟源IP
```

#### 场景3：容器网络
```bash
# Docker式配置：多级NAT
# 容器访问外网：容器IP → 宿主机IP → 公网IP
# POSTROUTING处理两次SNAT

# 第一次：容器网段 → 宿主机IP
iptables -t nat -A POSTROUTING -s 172.17.0.0/16 -o docker0 -j MASQUERADE

# 第二次：宿主机 → 公网
iptables -t nat -A POSTROUTING -s 172.17.0.0/16 -o eth0 -j MASQUERADE
```

### 表和链的关系矩阵

| 表\链 | PREROUTING | INPUT | FORWARD | OUTPUT | POSTROUTING |
|-------|------------|-------|---------|--------|-------------|
| **filter** | ❌ | ✅ | ✅ | ✅ | ❌ |
| **nat** | ✅ | ❌ | ❌ | ✅ | ✅ |
| **mangle** | ✅ | ✅ | ✅ | ✅ | ✅ |

### 数据包处理优先级

当数据包经过iptables时，会按照以下优先级顺序处理：

```
1. mangle表的PREROUTING链
2. nat表的PREROUTING链  
3. mangle表的INPUT/FORWARD链
4. filter表的INPUT/FORWARD链
5. mangle表的OUTPUT链
6. nat表的OUTPUT链
7. filter表的OUTPUT链
8. mangle表的POSTROUTING链
9. nat表的POSTROUTING链
```

### 实际应用示例

#### 场景1：简单的Web服务器防火墙

```bash
#!/bin/bash
# 清空现有规则
iptables -F
iptables -t nat -F
iptables -t mangle -F

# 设置默认策略
iptables -P INPUT DROP
iptables -P FORWARD DROP
iptables -P OUTPUT ACCEPT

# 允许本地回环
iptables -A INPUT -i lo -j ACCEPT

# 允许已建立的连接
iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT

# 允许SSH (限制IP范围)
iptables -A INPUT -p tcp --dport 22 -s 管理员IP/24 -j ACCEPT

# 允许Web服务
iptables -A INPUT -p tcp --dport 80 -j ACCEPT
iptables -A INPUT -p tcp --dport 443 -j ACCEPT

# 允许DNS查询响应
iptables -A INPUT -p udp --sport 53 -j ACCEPT

# 记录被拒绝的连接
iptables -A INPUT -j LOG --log-prefix "DROPPED: "
iptables -A INPUT -j DROP
```

#### 场景2：NAT网关配置

```bash
#!/bin/bash
# 开启IP转发
echo 1 > /proc/sys/net/ipv4/ip_forward

# 清空规则
iptables -F
iptables -t nat -F

# 设置filter表规则
iptables -P FORWARD DROP
iptables -A FORWARD -i eth1 -o eth0 -s 192.168.1.0/24 -j ACCEPT
iptables -A FORWARD -i eth0 -o eth1 -m state --state ESTABLISHED,RELATED -j ACCEPT

# 设置NAT规则
iptables -t nat -A POSTROUTING -s 192.168.1.0/24 -o eth0 -j MASQUERADE

# 端口映射 (内网Web服务器)
iptables -t nat -A PREROUTING -i eth0 -p tcp --dport 80 -j DNAT --to 192.168.1.100:80
iptables -t nat -A PREROUTING -i eth0 -p tcp --dport 443 -j DNAT --to 192.168.1.100:443

# 对应的filter规则
iptables -A FORWARD -i eth0 -o eth1 -p tcp --dport 80 -d 192.168.1.100 -j ACCEPT
iptables -A FORWARD -i eth0 -o eth1 -p tcp --dport 443 -d 192.168.1.100 -j ACCEPT
```

#### 场景3：高级流量控制 (配合mangle表)

```bash
#!/bin/bash
# 标记不同类型的流量
iptables -t mangle -A OUTPUT -p tcp --dport 22 -j MARK --set-mark 1    # SSH流量
iptables -t mangle -A OUTPUT -p tcp --dport 80 -j MARK --set-mark 2    # HTTP流量
iptables -t mangle -A OUTPUT -p tcp --dport 443 -j MARK --set-mark 2   # HTTPS流量
iptables -t mangle -A OUTPUT -p udp --dport 53 -j MARK --set-mark 3    # DNS流量

# 设置QoS (需要配合tc命令)
# tc qdisc add dev eth0 root handle 1: htb default 30
# tc class add dev eth0 parent 1: classid 1:1 htb rate 1mbit
# tc class add dev eth0 parent 1:1 classid 1:10 htb rate 500kbit ceil 800kbit # SSH优先
# tc class add dev eth0 parent 1:1 classid 1:20 htb rate 300kbit ceil 600kbit # WEB服务
# tc class add dev eth0 parent 1:1 classid 1:30 htb rate 200kbit ceil 400kbit # 其他
# 
# tc filter add dev eth0 protocol ip parent 1:0 prio 1 handle 1 fw classid 1:10
# tc filter add dev eth0 protocol ip parent 1:0 prio 2 handle 2 fw classid 1:20
# tc filter add dev eth0 protocol ip parent 1:0 prio 3 handle 3 fw classid 1:30
```

## 总结

**三表五链**是iptables的核心架构：

### 五链的作用总结
1. **PREROUTING**：数据包进入时的预处理 (DNAT、端口映射)
2. **INPUT**：控制进入本机的流量
3. **FORWARD**：控制转发流量 (路由器、网关功能)
4. **OUTPUT**：控制本机发出的流量
5. **POSTROUTING**：数据包发出前的后处理 (SNAT、伪装)

### 三表的功能总结
1. **filter表**：包过滤防火墙 (最常用)
2. **nat表**：网络地址转换 (NAT网关必须)
3. **mangle表**：包修改 (QoS、流量控制)

理解了三表五链的概念，就掌握了iptables的基础架构，可以根据实际需求灵活配置防火墙规则。
