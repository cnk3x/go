package urlx

import (
	"net/http"
	"strings"
)

const (
	HeaderAccept         = "Accept"
	HeaderAcceptLanguage = "Accept-Language"
	HeaderAcceptEncoding = "Accept-Encoding"
	HeaderUserAgent      = "User-Agent"
	HeaderContentType    = "Content-Type"
	HeaderReferer        = "Referer"
	HeaderCacheControl   = "Cache-Control" // no-cache
	HeaderPragma         = "Pragma"        // no-cache
)

// HeaderWith 设置请求头
func (c *Request) HeaderWith(options ...HeaderOption) *Request {
	c.headers = append(c.headers, options...)
	return c
}

// HeaderSet 设置请求头
func HeaderSet(key string, values ...string) HeaderOption {
	return func(headers http.Header) {
		headers.Set(key, strings.Join(values, ","))
	}
}

// HeaderDel 删除请求头
func HeaderDel(keys ...string) HeaderOption {
	return func(headers http.Header) {
		for _, key := range keys {
			headers.Del(key)
		}
	}
}

// AcceptLanguage 接受语言
func AcceptLanguage(acceptLanguage string) HeaderOption {
	return HeaderSet(HeaderAcceptLanguage, acceptLanguage)
}

// Accept 接受格式
func Accept(accept string) HeaderOption {
	return HeaderSet(HeaderAccept, accept)
}

// UserAgent 浏览器代理字符串
func UserAgent(userAgent string) HeaderOption {
	return HeaderSet(HeaderUserAgent, userAgent)
}

// Referer 引用地址
func Referer(referer string) HeaderOption {
	return HeaderSet(HeaderReferer, referer)
}

var (
	// NoCache 无缓存
	NoCache = HeaderOption(func(headers http.Header) {
		headers.Set(HeaderCacheControl, "no-cache")
		headers.Set(HeaderPragma, "no-cache")
	})

	// AcceptChinese 接受中文
	AcceptChinese = AcceptLanguage("zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6,zh-TW;q=0.5")

	// AcceptHTML 接受网页浏览器格式
	AcceptHTML = Accept("text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")

	// AcceptJSON 接受JSON格式
	AcceptJSON = Accept("application/json")

	// AcceptXML 接受XML格式
	AcceptXML = Accept("application/xml,text/xml")

	// MacEdge Mac Edge 浏览器的 UserAgent
	MacEdge = UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.55 Safari/537.36 Edg/96.0.1054.43")

	// WindowsEdge Windows Edge 浏览器的 UserAgent
	WindowsEdge = UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.55 Safari/537.36 Edg/96.0.1054.43")
)
