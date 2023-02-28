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

## 参考

- [tcpdmp filter 详细说明 ](http://alumni.cs.ucr.edu/~marios/ethereal-tcpdump.pdf)
- https://www.tcpdump.org/manpages/tcpdump.1.html
- https://opensource.com/article/18/10/introduction-tcpdump
- https://superuser.com/questions/497098/understanding-the-tcp-header
- https://juejin.cn/post/6844904084168769549