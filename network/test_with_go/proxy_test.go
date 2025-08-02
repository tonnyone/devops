package proxy_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"golang.org/x/net/proxy"
)

const (
	// 代理服务器配置
	proxyHost = "remote:8080"
	proxyUser = "admin"
	proxyPass = "123456"

	// 测试目标
	testHTTPURL  = "http://www.baidu.com"
	testHTTPSURL = "https://www.baidu.com"
)

// TestHTTPProxy 测试 HTTP 代理
func TestHTTPProxy(t *testing.T) {
	fmt.Println("=== 测试 HTTP 代理 ===")

	// 构建代理URL
	proxyURL, err := url.Parse(fmt.Sprintf("http://%s:%s@%s", proxyUser, proxyPass, proxyHost))
	if err != nil {
		t.Fatalf("解析代理URL失败: %v", err)
	}

	// 创建HTTP客户端
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: 30 * time.Second,
	}

	// 测试HTTP请求
	t.Run("HTTP请求", func(t *testing.T) {
		resp, err := client.Get(testHTTPURL)
		if err != nil {
			t.Fatalf("HTTP请求失败: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("读取响应失败: %v", err)
		}

		fmt.Printf("HTTP响应状态: %s\n", resp.Status)
		fmt.Printf("HTTP响应内容: %s\n", string(body))

		if resp.StatusCode != 200 {
			t.Errorf("期望状态码200，实际得到: %d", resp.StatusCode)
		}
	})

	// 测试HTTPS请求
	t.Run("HTTPS请求", func(t *testing.T) {
		resp, err := client.Get(testHTTPSURL)
		if err != nil {
			t.Fatalf("HTTPS请求失败: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("读取响应失败: %v", err)
		}

		fmt.Printf("HTTPS响应状态: %s\n", resp.Status)
		fmt.Printf("HTTPS响应内容: %s\n", string(body))

		if resp.StatusCode != 200 {
			t.Errorf("期望状态码200，实际得到: %d", resp.StatusCode)
		}
	})
}

// TestSOCKS5Proxy 测试 SOCKS5 代理
func TestSOCKS5Proxy(t *testing.T) {
	fmt.Println("\n=== 测试 SOCKS5 代理 ===")

	t.Run("方法1-Transport.Proxy", func(t *testing.T) {
		// 方法1：使用 Transport.Proxy
		// 注意：这种方式实际上是告诉HTTP客户端使用SOCKS5代理
		proxyURL, err := url.Parse(fmt.Sprintf("socks5://%s:%s@%s", proxyUser, proxyPass, proxyHost))
		if err != nil {
			t.Fatalf("解析SOCKS5代理URL失败: %v", err)
		}
		client := &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
			Timeout: 30 * time.Second,
		}

		resp, err := client.Get(testHTTPSURL)
		if err != nil {
			t.Fatalf("SOCKS5请求失败: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("读取响应失败: %v", err)
		}

		fmt.Printf("SOCKS5方法1响应状态: %s\n", resp.Status)
		fmt.Printf("SOCKS5方法1响应内容长度: %d 字节\n", len(body))
	})

	t.Run("方法2-SOCKS5Dialer", func(t *testing.T) {
		// 方法2：使用专用的SOCKS5 Dialer
		// 这种方式更直接地控制网络连接层面
		fmt.Println("\n--- SOCKS5 Dialer 方式 ---")
		auth := &proxy.Auth{
			User:     proxyUser,
			Password: proxyPass,
		}

		dialer, err := proxy.SOCKS5("tcp", proxyHost, auth, proxy.Direct)
		if err != nil {
			t.Fatalf("创建SOCKS5 dialer失败: %v", err)
		}
		client := &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
					return dialer.Dial(network, address)
				},
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
			Timeout: 30 * time.Second,
		}

		resp, err := client.Get(testHTTPSURL)
		if err != nil {
			t.Fatalf("SOCKS5请求失败: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("读取响应失败: %v", err)
		}

		fmt.Printf("SOCKS5方法2响应状态: %s\n", resp.Status)
		fmt.Printf("SOCKS5方法2响应内容长度: %d 字节\n", len(body))
		fmt.Println("\n=== 方法2说明 ===")
		fmt.Println("这也是显式代理配置")
		fmt.Println("- 通过替换DialContext来使用SOCKS5代理")
		fmt.Println("- 更底层的网络连接控制")
		fmt.Println("- 可以精确控制连接建立过程")
	})
}
