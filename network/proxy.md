# 代理测试

使用 [gost](https://github.com/ginuerzh/gost) 搭建代理，通过 curl 和 go 代码作为client测试
dns解析是发生效本机还是代理服务器

## 启动 proxy server

启动 http 代理 server
```shell
gost -L=http://admin:123456@0.0.0.0:8080
```

启动 sock5 代理 server
```shell
gost -L=socks5://admin:123456@0.0.0.0:8080
```

测试方法： 分别禁用本地的dns解析和代理所在服务器上的dns解析，测试代理是否生效

### CURL 命令

curl 通过 http proxy 访问互联网资源
```shell
curl -v -x 192.168.1.100:8080 -U admin:123456  https://www.baidu.com/
```
实验结果:
curl 会使用 **代理** 的 dns server 来访问远程 http 网址

curl 模拟 socks5 代理:
```shell
curl --socks5 192.168.1.100:8080 --proxy-user admin:123456 https://www.baidu.com/
```
实验结果:
curl 会使用 **本地** 的 dns server 访问百度(**这里需要注意一下**),
如果要使用代理的dns server 请使用下面的命令

```shell
curl -v --socks5-hostname 192.168.1.100:8080 --proxy-user admin:123456 https://www.baidu.com/
```

### go http client

http 代理使用方法
```code
url, _ := url.Parse("http://admin:123456@192.168.1.100:8080")
client := http.Client{
    Transport: &http.Transport{
        Proxy: http.ProxyURL(url),
    },
}
result, err := client.Get("https://www.baidu.com")
```
实验结果:
go http client 会使用 **代理** 的 dns server 来访问远程 http 网址

socks5 第一种代理使用方法
```code
url, _ := url.Parse("socks5://admin:123456@192.168.1.100:8080")
client := http.Client{
    Transport: &http.Transport{
        Proxy: http.ProxyURL(url),
    },
}
result, err := client.Get("https://www.baidu.com")
```
实验结果:
go http client 会使用 **代理** 的 dns server 来访问远程 http 网址

socks5 第二种代理使用方法
```code
dialer, err := proxy.SOCKS5("tcp", "192.168.1.100:8080", &proxy.Auth{
    User:     "admin",
    Password: "123456",
}, proxy.Direct)
if err != nil {
    fmt.Fprintln(os.Stderr, "can't connect to the proxy:", err)
}
client := http.Client{
    Transport: &http.Transport{DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
        return dialer.Dial(network, address)
    }},
}
result, err := client.Get("https://www.baidu.com")
```
实验结果:
go http client 会使用 **代理** 的 dns server 来访问远程 http 网址
