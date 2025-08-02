# http & socks5 代理测试
使用 [gost](https://github.com/ginuerzh/gost) 搭建代理，通过 curl 和 go 代码作为 client试

## 启动 proxy server
```shell
#http 
gost -L=http://admin:123456@0.0.0.0:8080
```

```shell
#socks
gost -L=socks5://admin:123456@0.0.0.0:8080
```
测试方法： 分别禁用本地的dns解析和代理所在服务器上的dns解析，测试代理是否生效

## CURL命令
curl 通过 http proxy 访问互联网资源 
```shell
# 注意这里的dns是使用的代理服务器的DNS
curl -v -x remote:8080 -U admin:123456  https://www.baidu.com/

```

curl 模拟 socks5 代理:
```shell
# 注意这里是使用的客户端本地的DNS
curl --socks5 remote:8080 --proxy-user admin:123456 https://www.baidu.com/

# 这种方式才是使用代理服务器的DNS
curl -v --socks5-hostname remote:8080 --proxy-user admin:123456 https://www.baidu.com/
```

### SOCKS5 RFC 1928 协议标准
SOCKS5协议**同时支持**两种DNS解析方式，通过ATYP (Address Type)字段区分：

| ATYP值 | 地址类型 | DNS解析方式 | 说明 |
|--------|----------|-------------|------|
| **0x01** | IPv4地址 | **客户端DNS** | 客户端解析域名，发送IP给代理 |
| **0x03** | 域名 | **代理DNS** | 客户端发送域名，代理负责解析 |
| **0x04** | IPv6地址 | **客户端DNS** | 客户端解析域名，发送IPv6给代理 |

### 协议报文格式
```
SOCKS5 CONNECT请求:
+----+-----+-------+------+----------+----------+
|VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
+----+-----+-------+------+----------+----------+
| 05 | 01  |  00   |  XX  | Variable |    2     |
+----+-----+-------+------+----------+----------+
```

**客户端DNS (ATYP=0x01)**:
```
05 01 00 01 [C0 A8 01 01] 01 BB
→ 连接到 192.168.1.1:443
```

**代理DNS (ATYP=0x03)**:
```
05 01 00 03 [0F] [www.example.com] 01 BB
→ 连接到 www.example.com:443 (代理解析)
```

### curl命令与协议对应关系
```shell
# 客户端DNS解析 → ATYP=0x01
curl --socks5 proxy:port URL
curl -x socks5://proxy:port URL

# 代理DNS解析 → ATYP=0x03  
curl --socks5-hostname proxy:port URL
```

### golang http client
```code
url, _ := url.Parse("http://admin:123456@192.168.1.100:8080")
client := http.Client{
    Transport: &http.Transport{
        Proxy: http.ProxyURL(url),
    },
}
result, err := client.Get("https://www.baidu.com")
```
### golang socks5 client 方式

SOCKS5在Go中的两种DNS解析实现：
#### 方式1: 客户端DNS解析 (ATYP=0x01)
```go
// 手动解析域名
ips, err := net.LookupIP("www.example.com")
targetIP := ips[0].String()

// 创建SOCKS5拨号器
dialer, err := proxy.SOCKS5("tcp", "192.168.1.100:8080", &proxy.Auth{
    User:     "admin",
    Password: "123456",
}, proxy.Direct)

client := http.Client{
    Transport: &http.Transport{
        DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
            // 发送IP给代理，不需要代理解析DNS
            return dialer.Dial(network, address)
        },
        TLSClientConfig: &tls.Config{
            ServerName: "www.example.com", // TLS握手需要原始域名
        },
    },
}

// 使用解析后的IP构造URL
result, err := client.Get(fmt.Sprintf("https://%s", targetIP))
```

#### 方式2: 代理DNS解析 (ATYP=0x03)
```go
// 标准方式，让代理解析DNS
dialer, err := proxy.SOCKS5("tcp", "192.168.1.100:8080", &proxy.Auth{
    User:     "admin",
    Password: "123456",
}, proxy.Direct)

client := http.Client{
    Transport: &http.Transport{
        DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
            // 直接发送域名给代理，代理负责DNS解析
            return dialer.Dial(network, address)
        },
    },
}

// 直接使用域名
result, err := client.Get("https://www.baidu.com")
```

#### Transport.Proxy方式 (简化版)
```go
// 使用Transport.Proxy，Go运行时自动处理SOCKS5协议
proxyURL, _ := url.Parse("socks5://admin:123456@192.168.1.100:8080")
client := http.Client{
    Transport: &http.Transport{
        Proxy: http.ProxyURL(proxyURL),
    },
}
result, err := client.Get("https://www.baidu.com")
```