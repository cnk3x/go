package urlx

import (
	"context"
	"io"
	"net/http"
	"time"
)

// New 以一些选项开始初始化请求器
func New(ctx context.Context, options ...Option) *Request {
	return (&Request{ctx: ctx}).With(options...)
}

// 一些特定方法的定义
type (
	Option       = func(*Request) error                                   // 请求选项
	Body         = func() (contentType string, body io.Reader, err error) // 请求提交内容构造方法
	HeaderOption = func(headers http.Header)                              // 请求头处理
)

// Request 请求构造
type Request struct {
	ctx     context.Context        // Context
	options []func(*Request) error // options

	// request fields
	method    string         // 接口请求方法
	url       string         // 请求地址
	query     string         // 请求链接参数
	buildBody Body           // 请求内容
	headers   []HeaderOption // 请求头处理

	// response fields
	beforeMw []ProcessMw // 中间件

	// client fields
	tryTimes []time.Duration // 重试时间和时机
	client   *http.Client    // client
}

/*请求公共设置*/

// With 增加选项
func (c *Request) With(options ...Option) *Request {
	c.options = append(c.options, options...)
	return c
}

// Url 设置请求链接
func (c *Request) Url(url string) *Request {
	c.url = url
	return c
}

// Query 设置请求Query参数
func (c *Request) Query(query string) *Request {
	c.query = query
	return c
}
