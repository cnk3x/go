package urlx

import (
	"io"
	"net/http"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/flate"
	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/s2"
	"github.com/klauspost/compress/snappy"
	"github.com/klauspost/compress/zstd"
)

// DecompressionBody 解压Body
func DecompressionBody(next Process) Process {
	return func(resp *http.Response, body io.ReadCloser) (err error) {
		defer body.Close()
		contentEncoding := resp.Header.Get(HeaderContentEncoding)
		if contentEncoding != "" {
			decoded := true
			switch contentEncoding {
			case "br":
				body = io.NopCloser(brotli.NewReader(body))
			case "deflate":
				body = flate.NewReader(body)
			case "gzip":
				body, err = gzip.NewReader(body)
			case "s2":
				body = io.NopCloser(s2.NewReader(body))
			case "snappy":
				body = io.NopCloser(snappy.NewReader(body))
			case "zstd":
				b, er := zstd.NewReader(body)
				if er != nil {
					return er
				}
				body = b.IOReadCloser()
			default:
				decoded = false
			}
			if err != nil {
				return
			}
			if decoded {
				resp.Header.Del(HeaderContentEncoding)
			}
		}
		return next(resp, io.NopCloser(body))
	}
}
