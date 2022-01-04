package urlx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func mockHTTPServer(h http.Handler) (string, func()) {
	listen, _ := net.Listen("tcp", "127.0.0.1:0")
	s := &http.Server{Handler: h}
	go func() { _ = s.Serve(listen) }()
	return "http:" + "//" + listen.Addr().String(), func() { _ = s.Shutdown(context.TODO()) }
}

func eq(t *testing.T, data [][2]any) {
	for _, kv := range data {
		if kv[0] != kv[1] {
			t.Fatalf("%s != %s", kv[:]...)
		}
	}
}

func TestRequest(t *testing.T) {
	addr, closer := mockHTTPServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set(HeaderContentType, "application/json")
		_ = json.NewEncoder(rw).Encode(map[string]string{
			"referer": r.Header.Get("Referer"),
			"accept":  r.Header.Get(HeaderAccept),
			"f1":      r.FormValue("f1"),
			"f2":      r.FormValue("f2"),
			"q1":      r.URL.Query().Get("q1"),
			"q2":      r.URL.Query().Get("q2"),
		})
	}))
	defer closer()

	data, err := Default(context.TODO()).UseClient(&http.Client{}).With(CookieEnabled(false)).
		Method(MethodPost).
		Url(addr).
		Query("q1=1&q2=two").
		SendForm(url.Values{"f1": {"a1"}, "f2": {"中文"}}).
		HeaderWith(AcceptJSON).
		// HeaderSet("Referer", "some_referer1").
		// HeaderDel("Referer").
		Bytes()
	if err != nil {
		t.Fatal(err)
	}

	var rMap map[string]string
	if err := json.Unmarshal(data, &rMap); err != nil {
		t.Fatal(err)
	}
	eq(t, [][2]any{
		{rMap["referer"], ""},
		{rMap["accept"], "application/json"},
		{rMap["f1"], "a1"},
		{rMap["f2"], "中文"},
		{rMap["q1"], "1"},
		{rMap["q2"], "two"},
	})
}

func TestForm(t *testing.T) {
	form1 := url.Values{
		"field1": {"value1"},
		"中文2":    {"value2"},
		"中文名3":   {"中文值3", "en3"},
	}

	addr, closer := mockHTTPServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		_, _ = rw.Write([]byte(r.Form.Encode()))
	}))
	defer closer()

	data, err := Default(nil).Url(addr + "?q1=1").SendForm(form1).Query("q3=4").Method(MethodPost).With(UseClient(&http.Client{})).Bytes()
	if err != nil {
		t.Fatal(err)
	}

	form1.Set("q1", "1")
	form1.Set("q3", "4")
	if !bytes.Equal(data, []byte(form1.Encode())) {
		t.Fatalf("FormNE:\ndata: %s\nform1:%s", data, form1.Encode())
	}
}

func TestHeader(t *testing.T) {
	headers := http.Header{
		http.CanonicalHeaderKey(HeaderAccept):         {"value1"},
		http.CanonicalHeaderKey(HeaderAcceptLanguage): {"value2"},
		http.CanonicalHeaderKey(HeaderCacheControl):   {"中文值3"},
	}

	addr, closer := mockHTTPServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		v := http.Header{}
		for k := range headers {
			v.Set(k, strings.Join(r.Header[k], ","))
		}
		_, _ = rw.Write([]byte(url.Values(v).Encode()))
	}))
	defer closer()

	r := Default(nil).Url(addr)
	for n, v := range headers {
		r.HeaderWith(HeaderSet(n, v...))
	}
	data, err := r.Bytes()
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(data, []byte(url.Values(headers).Encode())) {
		t.Fatalf("FormNE:\ndata: %s\nform1:%s", data, url.Values(headers).Encode())
	}
}

func TestTry(t *testing.T) {
	addr, closer := mockHTTPServer(nil)
	defer closer()
	ms := time.Millisecond
	errProxy := func(*http.Request) (*url.URL, error) { return url.Parse("https://localhsot:12345") }
	err := Default(nil).Url(addr).
		UseClient(&http.Client{Transport: &http.Transport{Proxy: errProxy}}).
		TryAt(50*ms, 50*ms, 500*ms).
		Process(nil)
	for err != nil {
		t.Logf("%T: %v", err, err)
		err = errors.Unwrap(err)
	}
}

func TestDirectError(t *testing.T) {
	addr, closer := mockHTTPServer(nil)
	defer closer()
	errBody := errors.New("ERRBODY")
	err := Default(nil).Url(addr).SendBody(func() (contentType string, body io.Reader, err error) { return "", nil, errBody }).Process(nil)
	eq(t, [][2]any{{err, errBody}})
}

func TestDownload(t *testing.T) {
	addr, closer := mockHTTPServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) { _, _ = rw.Write([]byte("1234")) }))
	defer closer()
	fn := "testdata/some"
	defer func() { _ = os.Remove(fn) }()
	if err := Default(nil).Url(addr).Download(fn); err != nil {
		t.Fatal(err)
	}
}
