package urlx

import (
	"context"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"sync"
)

type MultipartBody struct {
	getFile func() (field, filename string, fileBody io.ReadCloser, err error)
	params  url.Values

	w    *io.PipeWriter
	r    *io.PipeReader
	mw   *multipart.Writer
	done chan error
	once sync.Once
}

func Multipart() *MultipartBody {
	return &MultipartBody{}
}

func (m *MultipartBody) Params(params url.Values) *MultipartBody {
	m.params = params
	return m
}

func (m *MultipartBody) File(getFile func() (field, filename string, fileBody io.ReadCloser, err error)) *MultipartBody {
	m.getFile = getFile
	return m
}

func (m *MultipartBody) LocalFile(field string, filename string) *MultipartBody {
	m.getFile = func() (field string, filename string, fileBody io.ReadCloser, err error) {
		name := filepath.Base(filename)
		f, err := os.Open(filename)
		if err != nil {
			return field, name, nil, err
		}
		return field, name, f, nil
	}
	return m
}

func (m *MultipartBody) Body() (contentType string, body io.Reader, err error) {
	m.once.Do(m.init)
	return m.mw.FormDataContentType(), io.NopCloser(m.r), nil
}

func (m *MultipartBody) WaitEnd(ctx context.Context) error {
	m.once.Do(m.init)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-m.done:
		return err
	}
}

func (m *MultipartBody) init() {
	m.r, m.w = io.Pipe()
	m.mw = multipart.NewWriter(m.w)
	m.done = make(chan error, 1)
	go func() {
		defer close(m.done)
		defer m.w.Close()
		defer m.r.Close()
		m.done <- m.readFile()
	}()
}

func (m *MultipartBody) readFile() error {
	for key, values := range m.params {
		for _, value := range values {
			if err := m.mw.WriteField(key, value); err != nil {
				return err
			}
		}
	}

	field, filename, fileBody, err := m.getFile()
	if err != nil {
		return err
	}
	defer func() {
		if closer, ok := fileBody.(io.Closer); ok {
			closer.Close()
		}
	}()

	formFile, err := m.mw.CreateFormFile(field, filepath.Base(filename))
	if err != nil {
		return err
	}

	if _, err := io.Copy(formFile, fileBody); err != nil {
		return err
	}

	return m.mw.Close()
}
