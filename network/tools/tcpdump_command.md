## tcpdump command 

```
tcpdump ，非常有用的网络调试工具 
```

## 列出所有网络接口

```shell
sudo tcpdump --list-interfaces
sudo tcpdump -D 
```

```shell
# 所有接口抓取
sudo tcpdump --interface any
```

```shell
# 只抓取5个包
sudo tcpdump -i eth0 -c 5 
```

```shell
# tcpdump 会默认吧ip和端口使用名称展示，这里使用ip和端口展示
sudo tcpdump -i any -c5 -nn
```

```shell
# 指定ip和端口过滤和目标ip以及端口相关的包(无论包是源还是目的)
sudo tcpdump -i any -c5 -nn host 10.255.31.110 and port 30000
```

```shell
# 指定源ip和端口
sudo tcpdump -i any -c5 -nn src 192.168.122.98 and port 80
```

```shell
# 指定协议ICMP协议包
sudo tcpdump -i any -c5 -nn icmp
```

```shell
# 复杂的过滤方式
sudo tcpdump -i any -c5 -nn "port 80 and (src 192.168.122.98 or src 54.204.39.132)"
```

```shell
# -X 打印包的内容通过hex形式，-A 使用ASCII的格式打印
sudo tcpdump -i any -c10 -nn -A port 80

# -x hex打印 
sudo tcpdump -i any -nn -x -S "host 10.255.31.110 and port 3000 or host 10.248.162.60 and port 30000"
```

```shell
# 抓取到的包保存到文件中-w
sudo tcpdump -i any -c10 -nn -A port 80 -w aa.pcap
# 读取 pcap 包的内容
tcpdump -nn -r webserver.pcap
```

### 包解析结果分析

**TCP 三次握手**
```shell
00:54:45.223572 IP 10.255.31.110.30000 > 10.248.162.60.443: Flags [S], seq 1798586549, win 65535, options [mss 1372,nop,wscale 6,nop,nop,TS val 3113795521 ecr 0,sackOK,eol], length 0
        0x0000:  4520 0040 0000 4000 3506 6df7 0aff 1f6e
        0x0010:  0af8 a23c 7530 01bb 6b34 40b5 0000 0000
        0x0020:  b002 ffff c383 0000 0204 055c 0103 0306
        0x0030:  0101 080a b998 bfc1 0000 0000 0402 0000
00:54:45.223966 IP 10.248.162.60.443 > 10.255.31.110.30000: Flags [S.], seq 34566580, ack 1798586550, win 28960, options [mss 1460,sackOK,TS val 89927522 ecr 3113795521,nop,wscale 10], length 0
        0x0000:  4500 003c 0000 4000 3f06 641b 0af8 a23c
        0x0010:  0aff 1f6e 01bb 7530 020f 71b4 6b34 40b6
        0x0020:  a012 7120 d7cf 0000 0204 05b4 0402 080a
        0x0030:  055c 2f62 b998 bfc1 0103 030a
00:54:45.234917 IP 10.255.31.110.30000 > 10.248.162.60.443: Flags [.], ack 34566581, win 2061, options [nop,nop,TS val 3113795533 ecr 89927522], length 0
        0x0000:  4520 0034 0000 4000 3506 6e03 0aff 1f6e
        0x0010:  0af8 a23c 7530 01bb 6b34 40b6 020f 71b5
        0x0020:  8010 080d 5250 0000 0101 080a b998 bfcd
        0x0030:  055c 2f62
```

#### 输出字段解析

1. 时钟
2. 网络协议
3. 源IP和端口
4. 目的IP和端口
5. Flags

| Value  | FlagType | Description       |
|--------|----------|-------------------|
| S      | SYN      | Connection Start  |
| F      | FIN      | Connection Finish |
| P      | PUSH     | Data push         |
| R      | RST      | Connection reset  |
| R      | ACK      | Acknowledgment    |

#### 第一个网络包内容说明

前面 20个字节 是网络层ip包的首部: 
```
4520 0040 0000 4000 3506 6df7 0aff 1f6e
0af8 a23c

4520 0040: (0100) ip协议版本V4, 5(0101): 首部长度(单位,以32比特为单位): 20 bytes ,20(区分服务:DSCP), 0040(总长度,标识IP头部加上上层数据的数据包大小) 64字节
0000 4000: 0000(标识符,用于ip分片重组)，4(标志)，000(片偏移)
3506 6df7: 35(生存时间TTl),06(协议号:标识IP协议上层应用。当上层协议为ICMP时，协议号为1，TCP协议号为6，UDP的协议号为17)，6df7(首部校验和)
0aff 1f6e: 0aff1f6e(源ip地址)
0af8 a23c: (目的ip地址)
```

后面是IP包的包体，TCP的首部(从7530开始):
```
7530 01bb: 7530 是源端口(30000),01bb目的端口(443),
6b34 40b5: 序号: 6b3440b5(1798586549),  
0000 0000: 确认号: 00000000,
b002 ffff: (10110000 00000010 11111111 11111111) : 1011(首部长度:11,以32比特为单位，这里首部长44个字节),0000(保留字段)，00(CWR,ECE) 000010(标记字段)，11111111 11111111 (接收窗口 65535)
c383 0000：c383(因特网校验和) 0000(紧急数据指针）
0204 055c:（可选与变长选项字段）
0103 0306:（可选与变长选项字段）
0101 080a:（可选与变长选项字段）
b998 bfc1:（可选与变长选项字段）
0000 0000:（可选与变长选项字段）
0402 0000:（可选与变长选项字段）
```

### TCP 四次挥手

```
20:48:32.277781 IP 10.248.162.60.443 > 10.255.31.110.30000: Flags [F.], seq 1, ack 1, win 29, options [nop,nop,TS val 75154618 ecr 1504977382], length 0
20:48:32.369280 IP 10.255.31.110.30000 > 10.248.162.60.443: Flags [.], ack 2, win 2061, options [nop,nop,TS val 1505007476 ecr 75154618], length 0
20:48:32.369831 IP 10.255.31.110.30000 > 10.248.162.60.443: Flags [F.], seq 1, ack 2, win 2061, options [nop,nop,TS val 1505007476 ecr 75154618], length 0
20:48:32.369886 IP 10.248.162.60.443 > 10.255.31.110.30000: Flags [.], ack 2, win 29, options [nop,nop,TS val 75154710 ecr 1505007476], length 0
```

> When you use -xx, tcpdump outputs the link layer header of all packets, so the first 4 bytes of the output aren't TCP – they are part of the Ethernet frame.
> Even with plain -x, tcpdump would print the IP header before TCP/UDP.
> If you want to see the packet structure, use Wireshark instead – it will display every packet as a tree, and highlight the specific bytes for every value.


### tcpdump 抓包的原理

#### 1. 底层机制：libpcap

tcpdump基于libpcap库实现数据包捕获，工作在Linux内核的网络协议栈中：

```
应用层程序(tcpdump)
        ↑
    libpcap库
        ↑
   Packet Socket (AF_PACKET)
        ↑
    内核网络协议栈
        ↑
    网络设备驱动
        ↑
    物理网络接口
```

#### 2. 抓包位置

tcpdump在网络协议栈的不同层次都可以抓取数据包：

```
┌─────────────────┐
│   应用层数据    │ ← tcpdump无法直接抓取应用层原始数据
├─────────────────┤
│   传输层(TCP)   │ ← tcpdump -i any port 80
├─────────────────┤
│   网络层(IP)    │ ← tcpdump -i any host 192.168.1.1
├─────────────────┤
│  数据链路层     │ ← tcpdump -i eth0 (包含以太网头部)
├─────────────────┤
│   物理层        │ ← 网卡接收的原始比特流
└─────────────────┘
```

#### 3. 抓包流程

**第1步：网卡接收数据包**
```
网络数据包 → 网卡 → DMA传输到内存 → 触发中断
```

**第2步：内核处理**
```c
// 简化的内核处理流程
网卡中断处理程序() {
    从网卡缓冲区读取数据包;
    分配sk_buff结构体;
    将数据包放入接收队列;
    唤醒网络软中断;
}

网络软中断处理() {
    for (每个待处理的数据包) {
        调用协议栈处理函数;
        如果有AF_PACKET套接字监听 {
            复制数据包到套接字缓冲区;  // tcpdump在这里获取数据
        }
        继续协议栈正常处理;
    }
}
```

**第3步：tcpdump获取数据包**
```c
// tcpdump的工作原理（简化）
int main() {
    // 1. 创建AF_PACKET套接字
    int sock = socket(AF_PACKET, SOCK_RAW, htons(ETH_P_ALL));
    
    // 2. 绑定到指定网络接口
    bind(sock, (struct sockaddr*)&sockaddr_ll, sizeof(sockaddr_ll));
    
    // 3. 设置过滤器（BPF）
    setsockopt(sock, SOL_SOCKET, SO_ATTACH_FILTER, &filter, sizeof(filter));
    
    // 4. 循环接收数据包
    while (running) {
        recvfrom(sock, buffer, buffer_size, 0, NULL, NULL);
        解析并显示数据包内容();
    }
}
```

#### 4. BPF过滤机制

Berkeley Packet Filter (BPF) 在内核中高效过滤数据包：

```
原始数据包流 → BPF虚拟机 → 过滤后的数据包 → tcpdump
                  ↑
              过滤规则编译后的字节码
```

**BPF指令示例：**
```bash
# tcpdump编译过滤规则
tcpdump -d "host 192.168.1.1 and port 80"

# 输出BPF指令：
(000) ldh      [12]                    # 加载以太网类型字段
(001) jeq      #0x800            jt 2  jf 18   # 判断是否为IP包
(002) ld       [26]                    # 加载源IP地址
(003) jeq      #0xc0a80101       jt 4  jf 8    # 判断是否为192.168.1.1
...
```

#### 5. 性能考虑

**内核态vs用户态数据复制：**
```
每个数据包的处理路径：
网卡缓冲区 → 内核sk_buff → 用户态缓冲区 → tcpdump处理

优化机制：
1. 零拷贝：使用mmap减少数据复制
2. 批量处理：一次性处理多个数据包
3. 环形缓冲区：避免内存分配/释放开销
```

**抓包对性能的影响：**
```bash
# 高流量环境下的性能影响
echo "抓包会消耗CPU和内存资源："
echo "1. 数据包复制开销"
echo "2. BPF过滤计算开销" 
echo "3. 用户态处理开销"
echo "4. 存储I/O开销（如果保存到文件）"
```

#### 6. 特殊情况说明

**混杂模式（Promiscuous Mode）：**
```bash
# 启用混杂模式，抓取所有经过网卡的数据包
sudo tcpdump -i eth0 -p  # -p 禁用混杂模式

# 混杂模式下网卡行为：
正常模式: 只接收目标MAC地址为本机的数据包
混杂模式: 接收所有经过网卡的数据包（交换机环境下仅限同一冲突域）
```

**抓包位置限制：**
```
┌─────────────┐    数据包流向    ┌─────────────┐
│   客户端    │ ─────────────► │   服务器    │
└─────────────┘                └─────────────┘
       │                              │
       ▼                              ▼
  客户端网卡                      服务器网卡
  可以抓取：                      可以抓取：
  - 发出的数据包                  - 接收的数据包
  - 接收的数据包                  - 发出的数据包
  
  无法抓取：
  - 中间路由器的转发过程
  - 其他主机之间的通信（除非在同一冲突域）
```

#### 7. 实用调试技巧

```bash
# 抓包时避免自己的SSH连接干扰
sudo tcpdump -i any -nn not port 22

# 实时查看HTTP请求内容
sudo tcpdump -i any -A -s 0 'tcp port 80 and (((ip[2:2] - ((ip[0]&0xf)<<2)) - ((tcp[12]&0xf0)>>2)) != 0)'

# 抓取特定进程的网络流量（需要配合netstat等工具）
sudo netstat -anp | grep :80  # 找到进程PID
sudo tcpdump -i any -nn "port 80"
```

这就是tcpdump的抓包原理：通过AF_PACKET套接字在内核网络协议栈中获取数据包副本，使用BPF进行高效过滤，最终在用户态进行解析和显示。






## 参考

- [tcpdmp filter 详细说明 ](http://alumni.cs.ucr.edu/~marios/ethereal-tcpdump.pdf)
- https://www.tcpdump.org/manpages/tcpdump.1.html
- https://opensource.com/article/18/10/introduction-tcpdump
- https://superuser.com/questions/497098/understanding-the-tcp-header
- https://juejin.cn/post/6844904084168769549