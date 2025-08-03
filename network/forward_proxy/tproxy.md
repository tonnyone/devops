# TPROXY透明代理详解

## 什么是TPROXY？

TPROXY（Transparent Proxy）是Linux内核提供的透明代理技术，相比于传统的REDIRECT方式，TPROXY提供了更高性能和更强功能的透明代理解决方案。

### TPROXY vs REDIRECT 核心区别

| 特性 | REDIRECT | TPROXY |
|------|----------|---------|
| **数据包处理** | 修改目标地址 | 不修改数据包 |
| **原始目标获取** | SO_ORIGINAL_DST系统调用 | getsockname()直接获取 |
| **协议支持** | 仅TCP | TCP + UDP |
| **性能开销** | 中等（需要NAT和conntrack） | 低（直接路由劫持） |
| **配置复杂度** | 简单 | 复杂（需要策略路由） |
| **内核版本要求** | 低（2.4+） | 高（2.6.28+） |

## TPROXY工作原理深度解析

### 核心技术栈

```bash
# TPROXY技术栈组成
应用程序 (透明代理服务器)
    ↓
IP_TRANSPARENT socket选项
    ↓  
iptables TPROXY目标 + fwmark标记
    ↓
策略路由 (Policy Routing)
    ↓
local路由类型
    ↓
内核netfilter框架
```

### 1. iptables TPROXY目标详解

#### TPROXY规则语法
```bash
iptables -t mangle -A PREROUTING -p tcp --dport 80,443 \
    -j TPROXY --tproxy-mark 1 --on-port 8080

# 参数解析：
# -t mangle: 使用mangle表（用于修改数据包标记）
# PREROUTING: 在路由决策前处理数据包
# --tproxy-mark 1: 给匹配的数据包打上fwmark=1的标记
# --on-port 8080: 指定TPROXY监听端口
```

#### TPROXY的内核处理流程
```bash
# 数据包在内核中的处理路径

1. 【数据包到达】
   客户端发起: curl http://example.com:80
   数据包: [客户端IP:随机端口] -> [example.com:80]

2. 【mangle表PREROUTING链】
   - 数据包进入netfilter PREROUTING钩子
   - 匹配TPROXY规则（目标端口80）
   - 执行TPROXY目标：
     * 给数据包设置 skb->mark = 1
     * 记录TPROXY端口信息（8080）
     * 数据包继续处理，目标地址保持不变

3. 【路由决策阶段】
   - 内核查询路由表，发现数据包有fwmark=1
   - 根据策略路由规则，使用专门的tproxy路由表
   - 查询tproxy表，匹配到 "local default" 路由
   - "local"类型路由的特殊效果：
     * 将example.com:80当作本地地址处理
     * 数据包不会被转发，而是投递给本地

4. 【本地投递】
   - 内核寻找监听在端口8080的socket
   - 要求该socket设置了IP_TRANSPARENT选项
   - 找到TPROXY代理程序的监听socket
   - 将连接投递给代理程序

5. 【代理程序处理】
   - 代理程序接受连接
   - 调用getsockname()获取"本地地址"
   - 由于数据包未被修改，getsockname()返回原始目标example.com:80
   - 代理程序知道真实目标，建立上游连接
```

### 2. 策略路由（Policy Routing）机制

#### Linux路由系统概述
```bash
# Linux的路由决策是多层次的

1. 【路由规则（Rules）】- 决定使用哪个路由表
2. 【路由表（Tables）】- 包含具体的路由条目
3. 【路由条目（Routes）】- 指定如何到达目标

# 查看路由规则
ip rule list
# 默认输出：
# 0:    from all lookup local      # 本地路由表（最高优先级）
# 32766: from all lookup main      # 主路由表
# 32767: from all lookup default   # 默认路由表
```

#### TPROXY专用路由配置
```bash
# 第1步：创建专用路由表
echo "100 tproxy" >> /etc/iproute2/rt_tables
# 作用：定义路由表ID=100，名称=tproxy

# 第2步：添加策略路由规则  
ip rule add fwmark 1 table tproxy
# 作用：fwmark=1的数据包使用tproxy路由表查询

# 第3步：在tproxy表中添加local路由
ip route add local default dev lo table tproxy
# 作用：将所有目标地址（default=0.0.0.0/0）当作本地地址处理

# 验证配置
ip rule list
# 应该看到：
# 32765: from all fwmark 0x1 lookup tproxy

ip route list table tproxy  
# 应该看到：
# local default dev lo scope host
```

#### "local"路由类型的特殊机制
```bash
# local路由类型的作用机制

# 普通路由（转发类型）
ip route add 8.8.8.8/32 via 192.168.1.1 dev eth0
# 效果：发往8.8.8.8的数据包通过eth0转发给192.168.1.1

# local路由（本地类型）
ip route add local 8.8.8.8/32 dev lo  
# 效果：发往8.8.8.8的数据包当作发往本机127.0.0.1处理

# TPROXY使用local default的神奇效果：
ip route add local default dev lo table tproxy
# 效果：任何目标地址都被当作本地地址
# 这使得监听程序可以接收任意目标地址的连接
```

### 3. IP_TRANSPARENT socket选项详解

#### IP_TRANSPARENT的作用
```c
// 普通socket的限制示例
int fd = socket(AF_INET, SOCK_STREAM, 0);
struct sockaddr_in addr = {
    .sin_family = AF_INET,
    .sin_addr.s_addr = inet_addr("8.8.8.8"),  // 非本机IP
    .sin_port = htons(80)
};

// 普通socket无法绑定非本机IP
int result = bind(fd, (struct sockaddr*)&addr, sizeof(addr));
// result = -1, errno = EADDRNOTAVAIL (Cannot assign requested address)
```

```c
// 设置IP_TRANSPARENT后的效果
int transparent = 1;
int result = setsockopt(fd, SOL_IP, IP_TRANSPARENT, &transparent, sizeof(transparent));
// 现在可以成功绑定任意IP地址

result = bind(fd, (struct sockaddr*)&addr, sizeof(addr));
// result = 0 (成功!)
```

#### IP_TRANSPARENT的内核实现原理
```bash
# IP_TRANSPARENT在内核中的检查逻辑

1. 【bind()系统调用】
   应用程序调用bind()绑定非本机地址

2. 【内核地址检查】
   inet_bind() -> inet_addr_type() 检查地址类型
   
3. 【IP_TRANSPARENT检查】
   if (地址不是本机 && !socket->transparent) {
       return -EADDRNOTAVAIL;  // 拒绝绑定
   }
   
4. 【允许绑定】
   如果设置了IP_TRANSPARENT，跳过地址检查
   允许绑定任意地址

5. 【连接接收】
   当数据包到达时，内核会查找匹配的socket
   IP_TRANSPARENT socket可以接收任意目标地址的连接
```

## TPROXY完整实现示例

### 系统配置脚本
```bash
#!/bin/bash
# TPROXY透明代理完整配置脚本

set -e

# 配置参数
TPROXY_PORT=8080
TPROXY_MARK=1
TPROXY_TABLE=100

echo "🚀 开始配置TPROXY透明代理..."

# 1. 创建专用路由表
echo "${TPROXY_TABLE} tproxy" >> /etc/iproute2/rt_tables 2>/dev/null || true

# 2. 清理现有规则（避免重复）
iptables -t mangle -F 2>/dev/null || true
ip rule del fwmark ${TPROXY_MARK} table tproxy 2>/dev/null || true
ip route flush table tproxy 2>/dev/null || true

# 3. 配置iptables TPROXY规则
echo "📋 配置iptables规则..."

# 创建自定义链
iptables -t mangle -N TPROXY_TRANSPARENT 2>/dev/null || true

# 排除本地流量
iptables -t mangle -A TPROXY_TRANSPARENT -d 127.0.0.0/8 -j RETURN
iptables -t mangle -A TPROXY_TRANSPARENT -d 192.168.0.0/16 -j RETURN
iptables -t mangle -A TPROXY_TRANSPARENT -d 10.0.0.0/8 -j RETURN

# TPROXY规则
iptables -t mangle -A TPROXY_TRANSPARENT -p tcp --dport 80,443 \
    -j TPROXY --tproxy-mark ${TPROXY_MARK} --on-port ${TPROXY_PORT}

# 应用到PREROUTING链
iptables -t mangle -A PREROUTING -j TPROXY_TRANSPARENT

# 4. 配置策略路由
echo "🗺️  配置策略路由..."
ip rule add fwmark ${TPROXY_MARK} table tproxy
ip route add local default dev lo table tproxy

# 5. 启用IP转发
echo "🔀 启用IP转发..."
echo 1 > /proc/sys/net/ipv4/ip_forward
echo 1 > /proc/sys/net/ipv4/conf/all/route_localnet

# 6. 验证配置
echo "✅ 验证配置..."
echo "路由规则:"
ip rule list | grep tproxy

echo "路由表内容:"
ip route list table tproxy

echo "iptables规则:"
iptables -t mangle -L TPROXY_TRANSPARENT -n

echo "🎉 TPROXY配置完成!"
```

### Go语言TPROXY服务器实现
```go
package main

import (
    "fmt"
    "io"
    "log"
    "net"
    "os"
    "syscall"
    "unsafe"
)

const (
    // Linux特定常量
    IP_TRANSPARENT = 19
    SO_REUSEADDR   = 2
)

type TProxyServer struct {
    listenAddr string
    port       int
}

func NewTProxyServer(port int) *TProxyServer {
    return &TProxyServer{
        listenAddr: fmt.Sprintf(":%d", port),
        port:       port,
    }
}

func (t *TProxyServer) Start() error {
    log.Printf("🌐 启动TPROXY服务器在端口 %d", t.port)
    
    // 创建原始socket
    fd, err := t.createTProxySocket()
    if err != nil {
        return fmt.Errorf("创建TPROXY socket失败: %v", err)
    }
    defer syscall.Close(fd)
    
    // 绑定和监听
    if err := t.bindAndListen(fd); err != nil {
        return fmt.Errorf("绑定监听失败: %v", err)
    }
    
    log.Printf("✅ TPROXY服务器启动成功，等待连接...")
    
    // 接受连接循环
    for {
        clientFd, clientAddr, err := syscall.Accept(fd)
        if err != nil {
            log.Printf("❌ 接受连接失败: %v", err)
            continue
        }
        
        go t.handleConnection(clientFd, clientAddr)
    }
}

func (t *TProxyServer) createTProxySocket() (int, error) {
    // 创建TCP socket
    fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
    if err != nil {
        return -1, err
    }
    
    // 设置IP_TRANSPARENT选项（TPROXY的关键）
    if err := syscall.SetsockoptInt(fd, syscall.SOL_IP, IP_TRANSPARENT, 1); err != nil {
        syscall.Close(fd)
        return -1, fmt.Errorf("设置IP_TRANSPARENT失败: %v", err)
    }
    
    // 设置SO_REUSEADDR选项
    if err := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, SO_REUSEADDR, 1); err != nil {
        syscall.Close(fd)
        return -1, fmt.Errorf("设置SO_REUSEADDR失败: %v", err)
    }
    
    log.Printf("🔧 socket选项配置完成")
    return fd, nil
}

func (t *TProxyServer) bindAndListen(fd int) error {
    // 绑定到指定端口
    addr := &syscall.SockaddrInet4{
        Port: t.port,
        Addr: [4]byte{0, 0, 0, 0}, // 监听所有接口
    }
    
    if err := syscall.Bind(fd, addr); err != nil {
        return fmt.Errorf("bind失败: %v", err)
    }
    
    // 开始监听
    if err := syscall.Listen(fd, 128); err != nil {
        return fmt.Errorf("listen失败: %v", err)
    }
    
    return nil
}

func (t *TProxyServer) handleConnection(clientFd int, clientAddr syscall.Sockaddr) {
    defer syscall.Close(clientFd)
    
    // 获取客户端地址字符串
    clientAddrStr := t.sockaddrToString(clientAddr)
    
    // TPROXY的关键：获取原始目标地址
    // 在TPROXY模式下，getsockname返回的就是原始目标地址
    originalTarget, err := t.getOriginalTarget(clientFd)
    if err != nil {
        log.Printf("❌ 获取原始目标失败: %v", err)
        return
    }
    
    log.Printf("🎯 TPROXY连接: %s -> %s", clientAddrStr, originalTarget)
    
    // 连接到真实目标
    targetConn, err := net.Dial("tcp", originalTarget)
    if err != nil {
        log.Printf("❌ 连接目标失败 %s: %v", originalTarget, err)
        return
    }
    defer targetConn.Close()
    
    // 创建客户端连接包装器
    clientConn := &SocketConn{fd: clientFd}
    
    log.Printf("✅ TPROXY隧道建立: %s <-> %s", clientAddrStr, originalTarget)
    
    // 双向数据转发
    t.forwardData(clientConn, targetConn, originalTarget)
}

func (t *TProxyServer) getOriginalTarget(clientFd int) (string, error) {
    // TPROXY模式下，getsockname返回的就是原始目标地址
    addr, err := syscall.Getsockname(clientFd)
    if err != nil {
        return "", err
    }
    
    return t.sockaddrToString(addr), nil
}

func (t *TProxyServer) sockaddrToString(addr syscall.Sockaddr) string {
    switch a := addr.(type) {
    case *syscall.SockaddrInet4:
        ip := net.IPv4(a.Addr[0], a.Addr[1], a.Addr[2], a.Addr[3])
        return fmt.Sprintf("%s:%d", ip.String(), a.Port)
    case *syscall.SockaddrInet6:
        ip := net.IP(a.Addr[:])
        return fmt.Sprintf("[%s]:%d", ip.String(), a.Port)
    default:
        return "unknown"
    }
}

func (t *TProxyServer) forwardData(client, target net.Conn, targetAddr string) {
    // 启动双向数据转发
    done := make(chan struct{}, 2)
    
    // 客户端 -> 目标
    go func() {
        defer func() { done <- struct{}{} }()
        written, err := io.Copy(target, client)
        if err != nil {
            log.Printf("❌ 客户端->目标转发错误: %v", err)
        } else {
            log.Printf("→ 客户端->目标: %d 字节", written)
        }
        target.Close()
    }()
    
    // 目标 -> 客户端  
    go func() {
        defer func() { done <- struct{}{} }()
        written, err := io.Copy(client, target)
        if err != nil {
            log.Printf("❌ 目标->客户端转发错误: %v", err)
        } else {
            log.Printf("← 目标->客户端: %d 字节", written)
        }
        client.Close()
    }()
    
    // 等待任一方向转发完成
    <-done
    log.Printf("🔒 TPROXY隧道关闭: %s", targetAddr)
}

// Socket文件描述符包装器
type SocketConn struct {
    fd int
}

func (sc *SocketConn) Read(b []byte) (n int, err error) {
    return syscall.Read(sc.fd, b)
}

func (sc *SocketConn) Write(b []byte) (n int, err error) {
    return syscall.Write(sc.fd, b)
}

func (sc *SocketConn) Close() error {
    return syscall.Close(sc.fd)
}

func (sc *SocketConn) LocalAddr() net.Addr {
    // 实现net.Conn接口（可选）
    return nil
}

func (sc *SocketConn) RemoteAddr() net.Addr {
    // 实现net.Conn接口（可选）  
    return nil
}

func (sc *SocketConn) SetDeadline(t time.Time) error {
    // 实现net.Conn接口（可选）
    return nil
}

func (sc *SocketConn) SetReadDeadline(t time.Time) error {
    // 实现net.Conn接口（可选）
    return nil
}

func (sc *SocketConn) SetWriteDeadline(t time.Time) error {
    // 实现net.Conn接口（可选）
    return nil
}

func main() {
    // 检查root权限
    if os.Geteuid() != 0 {
        log.Fatal("❌ TPROXY需要root权限运行")
    }
    
    server := NewTProxyServer(8080)
    if err := server.Start(); err != nil {
        log.Fatal("❌ TPROXY服务器启动失败:", err)
    }
}
```

## TPROXY高级特性

### 1. UDP支持
```bash
# UDP TPROXY配置
iptables -t mangle -A PREROUTING -p udp --dport 53 \
    -j TPROXY --tproxy-mark 1 --on-port 8053

# UDP TPROXY的特殊处理：
# 1. 使用recvfrom()接收数据包
# 2. 通过IP_RECVORIGDSTADDR获取原始目标
# 3. 维护UDP会话状态
# 4. 处理NAT超时
```

### 2. IPv6支持
```bash
# IPv6 TPROXY配置
ip6tables -t mangle -A PREROUTING -p tcp --dport 80,443 \
    -j TPROXY --tproxy-mark 1 --on-port 8080

# IPv6路由配置
ip -6 rule add fwmark 1 table tproxy
ip -6 route add local default dev lo table tproxy
```

### 3. 性能优化
```bash
# 系统参数优化
echo 'net.core.rmem_max = 134217728' >> /etc/sysctl.conf
echo 'net.core.wmem_max = 134217728' >> /etc/sysctl.conf
echo 'net.ipv4.tcp_rmem = 4096 65536 134217728' >> /etc/sysctl.conf
echo 'net.ipv4.tcp_wmem = 4096 65536 134217728' >> /etc/sysctl.conf

# 禁用不必要的conntrack（TPROXY不依赖）
iptables -t raw -A PREROUTING -p tcp --dport 80,443 -j NOTRACK
iptables -t raw -A OUTPUT -p tcp --sport 80,443 -j NOTRACK
```

## 故障排除和调试

### 常见问题诊断
```bash
# 1. 检查内核支持
grep TPROXY /boot/config-$(uname -r)
# 应该看到: CONFIG_NETFILTER_TPROXY=m

# 2. 检查模块加载
lsmod | grep xt_TPROXY
# 如果没有，手动加载: modprobe xt_TPROXY

# 3. 检查策略路由
ip rule list | grep tproxy
ip route list table tproxy

# 4. 检查iptables规则
iptables -t mangle -L -n -v

# 5. 监控数据包标记
tcpdump -i any -n 'tcp and port 80' -v
# 查看是否有TPROXY标记

# 6. 检查socket状态
ss -tlpn | grep :8080
# 确认TPROXY程序正在监听
```

### 调试技巧
```bash
# 开启详细日志
echo 1 > /proc/sys/net/netfilter/nf_log_all_netns

# 跟踪数据包流向
iptables -t mangle -I PREROUTING -j LOG --log-prefix "MANGLE-PRE: "
iptables -t mangle -I POSTROUTING -j LOG --log-prefix "MANGLE-POST: "

# 查看日志
tail -f /var/log/kern.log | grep "MANGLE"
```

## 参考资料

> https://jimmysong.io/blog/what-is-tproxy/
> https://pkg.go.dev/github.com/LiamHaworth/go-tproxy#section-readme
> https://gsoc-blog.ecklm.com/iptables-redirect-vs.-dnat-vs.-tproxy/
> https://www.kernel.org/doc/Documentation/networking/tproxy.txt

## 前置知识