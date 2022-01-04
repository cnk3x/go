package urlx

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type ProgressFunc = func(total, cur, speed float64)

// Progress 下载进度插件
func Progress(progress ProgressFunc) ProcessMw {
	return func(next Process) Process {
		return func(resp *http.Response, body io.ReadCloser) error {
			defer body.Close()
			return next(resp, body)
		}
	}
}

// Download 下载到文件
func (c *Request) Download(fn string) (err error) {
	return c.Process(func(resp *http.Response, body io.ReadCloser) (err error) {
		defer body.Close()
		tempFn := fn + ".urlx_dl_temp"
		if err = os.MkdirAll(filepath.Dir(tempFn), 0755); err != nil {
			return
		}

		err = func() error {
			f, err := os.Create(tempFn)
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = io.Copy(f, resp.Body)
			return err
		}()

		return os.Rename(tempFn, fn)
	})
}
