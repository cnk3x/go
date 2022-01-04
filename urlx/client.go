package urlx

import (
	"net/http"
	"net/http/cookiejar"
	"time"
)

/* 设置客户端 */

// TryAt 失败重试，等待休眠时间
func (c *Request) TryAt(times ...time.Duration) *Request {
	c.tryTimes = times
	return c
}

// UseClient 使用的客户端定义
func (c *Request) UseClient(client *http.Client) *Request {
	c.client = client
	return c
}

// CookieEnabled 开关 Cookie
func CookieEnabled(enabled ...bool) Option {
	if len(enabled) == 0 || enabled[0] {
		jar, _ := cookiejar.New(nil)
		return Jar(jar)
	}
	return Jar(nil)
}

// Jar 设置Cookie容器
func Jar(jar http.CookieJar) Option {
	return func(c *Request) error {
		c.client.Jar = jar
		return nil
	}
}

// UseClient 使用自定义的HTTP客户端
func UseClient(client *http.Client) Option {
	return func(r *Request) error {
		r.client = client
		return nil
	}
}
