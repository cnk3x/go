package urlx

import (
	"context"
	"encoding/xml"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
)

/*处理响应*/

const HeaderContentEncoding = "Content-Encoding"

type (
	Process   = func(resp *http.Response, body io.ReadCloser) error // 响应处理器
	ProcessMw = func(next Process) Process                          // 响应预处理器
)

var ProcessNil = func(resp *http.Response, body io.ReadCloser) error { return nil }

// ProcessWith 在处理之前的预处理
func (c *Request) ProcessWith(mws ...ProcessMw) *Request {
	c.beforeMw = append(c.beforeMw, mws...)
	return c
}

// Process 处理响应
func (c *Request) Process(process Process) error {
	if c.client == nil {
		c.client = &http.Client{}
	}

	for _, apply := range c.options {
		if err := apply(c); err != nil {
			return err
		}
	}

	if c.ctx == nil {
		c.ctx = context.Background()
	}

	if c.method == "" {
		c.method = http.MethodGet
	}

	requestUrl := c.url
	if c.query != "" {
		if strings.Contains(requestUrl, "?") {
			requestUrl += "&" + c.query
		} else {
			requestUrl += "?" + c.query
		}
	}

	if c.buildBody == nil {
		c.buildBody = func() (contentType string, body io.Reader, err error) { return "", nil, nil }
	}

	var resp *http.Response
	for i := 0; i < len(c.tryTimes)+1; i++ {
		contentType, body, err := c.buildBody()
		if err != nil {
			return err
		}

		req, err := http.NewRequestWithContext(c.ctx, c.method, requestUrl, body)
		if err != nil {
			return err
		}

		if contentType != "" {
			req.Header.Set(HeaderContentType, contentType)
		}

		for _, headerOption := range c.headers {
			headerOption(req.Header)
		}

		if resp, err = c.client.Do(req); err != nil {
			var ne net.Error
			if i < len(c.tryTimes) && errors.As(err, &ne) {
				log.Printf("第%d次出错: %v, %s后重试", i+1, err, c.tryTimes[i])
				select {
				case <-c.ctx.Done():
					return err
				case <-time.After(c.tryTimes[i]):
					continue
				}
			}
			log.Printf("第%d次出错: %v, 返回错误", i+1, err)
			return err
		}
		break
	}

	defer resp.Body.Close()
	if process == nil {
		process = ProcessNil
	}
	for _, before := range c.beforeMw {
		process = before(process)
	}

	return process(resp, io.NopCloser(resp.Body))
}

// Bytes 处理响应字节
func (c *Request) Bytes() (data []byte, err error) {
	err = c.Process(func(resp *http.Response, body io.ReadCloser) (err error) {
		data, err = io.ReadAll(resp.Body)
		return
	})
	return
}

// JSON 处理响应
func (c *Request) JSON(out any) (data []byte, err error) {
	if data, err = c.Bytes(); err == nil {
		err = json.Unmarshal(data, out)
	}
	return
}

// XML 处理响应
func (c *Request) XML(out any) (data []byte, err error) {
	if data, err = c.Bytes(); err == nil {
		err = xml.Unmarshal(data, out)
	}
	return
}

// YAML 处理响应
func (c *Request) YAML(out any) (data []byte, err error) {
	if data, err = c.Bytes(); err == nil {
		err = yaml.Unmarshal(data, out)
	}
	return
}

// Status .
func Status(processStatus func(status int) Process) ProcessMw {
	return func(next Process) Process {
		return func(resp *http.Response, body io.ReadCloser) error {
			if process := processStatus(resp.StatusCode); process != nil {
				return process(resp, body)
			}
			return next(resp, body)
		}
	}
}

// JSON 处理响应
func JSON(out any) Process {
	return func(resp *http.Response, body io.ReadCloser) error {
		defer body.Close()
		return json.NewDecoder(body).Decode(out)
	}
}

// XML 处理响应
func XML(out any) Process {
	return func(resp *http.Response, body io.ReadCloser) error {
		defer body.Close()
		return xml.NewDecoder(body).Decode(out)
	}
}

// YAML 处理响应
func YAML(out any) Process {
	return func(resp *http.Response, body io.ReadCloser) error {
		defer body.Close()
		return yaml.NewDecoder(body).Decode(out)
	}
}
