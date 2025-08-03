## nc command 

```
全称 netcat ，非常有用的网路调试工具示例如下： 
```

## server 端监听一个端口 并在客户端建链接，一旦链接建立客户端和服务端任意一端发送的数据，另一端都回收到打印出来

```shell
nc -l 1234
```

```shell
nc 127.0.0.1 1234
```

### 数据传输

server端启动

```shell
nc -l 1234 > b.txt
```

client端发送

```shell
nc 127.0.0.1 1234 < /tmp/a.txt
```

### 发起简单的http请求

```shell
printf "GET / HTTP/1.1\r\n\r\n" | nc baidu.com 80
```

### 构造更复杂的请求(通过SMTP发送邮件)

```shell
nc [-C] localhost 25 << EOF
HELO host.example.com
MAIL FROM:<user@host.example.com>
RCPT TO:<user2@host.example.com>
DATA
Body of e-mail.
.
QUIT
EOF
```

## 端口扫描

```-z``` flag 会返回是目标打开的端口，而不是初始化一个连接，不添加-z 的标记会通过tcp连接到目标端口

```shell
# 查看当前主机监听的端口号
nc -zv 127.0.0.1 1-1000
```

### 下面的命令会显示出协议的banner,如果打开的话

```shell
echo "QUIT" | nc 127.0.0.1 1-1000
```

### 示例

```shell
# 通过源端口的31337去连接远程的443端口，5s后超时
nc -p 31337 -w 5 host.example.com 42
```

```shell
# 通过udp协议连接53端口
nc -u 8.8.8.8 53
```

```shell
# 使用本地192.168.2.102作为连接的ip去连接
printf "GET / HTTP/1.1\r\n\r\n" | nc -s 192.168.2.102 baidu.com 80
```

```shell
# 使用 gost 启动代理(https://github.com/ginuerzh/gost)
gost -L=:8080
# nc 通过 http connect代理 连接 http服务
nc -x10.253.1.65:9070 -Xconnect -x127.0.0.1:8080 www.baidu.com 80
```

## 参考

- https://www.computerhope.com/unix/nc.htm
- https://en.wikipedia.org/wiki/Netcat#Proxying
- https://v2.gost.run/sni/ 