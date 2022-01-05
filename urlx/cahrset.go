package urlx

import (
	"io"
	"log"
	"mime"
	"net/http"
	"strings"

	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func CharsetDecode(next Process) Process {
	return func(resp *http.Response, body io.ReadCloser) error {
		defer body.Close()
		var r io.Reader = body
		mimeType, params, _ := mime.ParseMediaType(resp.Header.Get(HeaderContentType))
		if strings.HasPrefix(mimeType, "text/") && len(params) > 0 {
			if charset := strings.TrimSpace(params["charset"]); charset != "" {
				codec, err := htmlindex.Get(charset)
				if err != nil {
					log.Printf("not support charset: %s", charset)
				} else if codec != unicode.UTF8 {
					r = transform.NewReader(r, codec.NewDecoder())
					resp.Header.Set(HeaderContentType, mimeType)
				}
			}
		}
		return next(resp, io.NopCloser(r))
	}
}
