package urlx

import (
	"bytes"
	"encoding/xml"
	"io"
	"net/url"
	"strings"

	"github.com/goccy/go-json"
	"github.com/google/go-querystring/query"
)

// SendBody 设置请求提交内容
func (c *Request) SendBody(body Body) *Request {
	c.buildBody = body
	return c
}

// SendForm 提交表单
func (c *Request) SendForm(v any) *Request {
	return c.SendBody(func() (contentType string, body io.Reader, err error) {
		contentType = "application/x-www-form-urlencoded; charset=utf-8"
		switch o := v.(type) {
		case io.Reader:
			body = o
		case []byte:
			body = bytes.NewReader(o)
		case string:
			body = strings.NewReader(o)
		case bytes.Buffer:
			body = &o
		case *bytes.Buffer:
			body = o
		case url.Values:
			body = strings.NewReader(o.Encode())
		case *url.Values:
			body = strings.NewReader(o.Encode())
		case map[string]string:
			values := url.Values{}
			for k, v := range o {
				values.Set(k, v)
			}
			body = strings.NewReader(values.Encode())
		default:
			if r, ok := o.(io.Reader); ok {
				body = r
			} else {
				var values url.Values
				if values, err = query.Values(v); err == nil {
					body = strings.NewReader(values.Encode())
				}
			}
		}
		return
	})
}

// SendJSON 提交JSON
func (c *Request) SendJSON(v any) *Request {
	return c.SendBody(func() (contentType string, body io.Reader, err error) {
		contentType = "application/json; charset=utf-8"
		switch o := v.(type) {
		case io.Reader:
			body = o
		case []byte:
			body = bytes.NewReader(o)
		case string:
			body = strings.NewReader(o)
		case bytes.Buffer:
			body = bytes.NewReader(o.Bytes())
		case *bytes.Buffer:
			body = bytes.NewReader(o.Bytes())
		default:
			var data []byte
			if data, err = json.Marshal(v); err == nil {
				body = bytes.NewReader(data)
			}
		}
		return
	})
}

// SendXML 提交XML
func (c *Request) SendXML(v any) *Request {
	return c.SendBody(func() (contentType string, body io.Reader, err error) {
		contentType = "application/xml; charset=utf-8"
		switch o := v.(type) {
		case io.Reader:
			body = o
		case []byte:
			body = bytes.NewReader(o)
		case string:
			body = strings.NewReader(o)
		case bytes.Buffer:
			body = bytes.NewReader(o.Bytes())
		case *bytes.Buffer:
			body = bytes.NewReader(o.Bytes())
		default:
			var data []byte
			if data, err = xml.Marshal(v); err == nil {
				body = bytes.NewReader(data)
			}
		}
		return
	})
}

// // FormEncode 编码为QueryString
// func FormEncode(v any) (string, error) {
// 	switch o := v.(type) {
// 	case url.Values:
// 		return o.Encode(), nil
// 	case *url.Values:
// 		return o.Encode(), nil
// 	case map[string]string:
// 		u := url.Values{}
// 		for k, v := range o {
// 			u.Set(k, v)
// 		}
// 		return u.Encode(), nil
// 	case string:
// 		return o, nil
// 	case []byte:
// 		return string(o), nil
// 	case *bytes.Buffer:
// 		return o.String(), nil
// 	}
//
// 	if r, ok := v.(io.Reader); ok {
// 		data, _ := io.ReadAll(r)
// 		return string(data), nil
// 	}
//
// 	values, err := query.Values(v)
// 	if err != nil {
// 		return "", err
// 	}
// 	return values.Encode(), nil
// }
