package httpx

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"os"
	"sync"

	"golang.org/x/crypto/acme/autocert"
)

type ServerOptions struct {
	Listen string
	HTTPS  struct {
		Enabled     bool
		Listen      string
		SSLRedirect bool
		HSTS        bool
		CertFile    string
		Auto        bool
		Domains     []string
		CertDIR     string
		Email       string
	}
}

type Server struct {
	options ServerOptions
	mu      sync.Mutex
	wait    Wait
	log     *log.Logger
	quit    chan struct{}
}

func New(options ServerOptions) *Server {
	return &Server{
		options: options,
		log:     log.Default(),
		quit:    make(chan struct{}, 1),
	}
}

// Logger 设置logger
func (s *Server) Logger(log *log.Logger) {
	s.log = log
}

// Serve 开始服务
func (s *Server) Serve(ctx context.Context, mux http.Handler) error {
	s.startServer(ctx, listen(s.options.Listen), mux)
	if s.options.HTTPS.Enabled {
		if s.options.HTTPS.Auto {
			s.startServer(ctx, s.autoHTTPS(), mux)
		} else {
			s.startServer(ctx, s.listenHTTPS(), mux)
		}
	}
	return nil
}

// Quit 主动退出
func (s *Server) Quit() {
	close(s.quit)
}

// WaitForExit 等待退出
func (s *Server) WaitForExit(ctx context.Context) {
	s.wait.Wait(ctx)
}

func (s *Server) startServer(ctx context.Context, listen net.Listener, mux http.Handler) {
	hs := &http.Server{
		Handler: mux,
		BaseContext: func(l net.Listener) context.Context {
			log.Printf("已启动: %s", l.Addr())
			return ctx
		},
	}

	hs.RegisterOnShutdown(func() {
		log.Printf("退出")
	})

	s.wait.Add(1)
	go func() {
		defer s.wait.Done()
		if err := hs.Serve(listen); err != nil && err != http.ErrServerClosed {
			log.Printf("已退出: %v", err)
			return
		}
		log.Printf("已退出")
	}()

	go func() {
		select {
		case <-ctx.Done():
		case <-s.quit:
		}
		log.Printf("退出")
		hs.Shutdown(context.TODO())
	}()
}

func (s *Server) autoHTTPS() net.Listener {
	acm := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache(s.options.HTTPS.CertDIR),
		HostPolicy: autocert.HostWhitelist(s.options.HTTPS.Domains...),
		Email:      s.options.HTTPS.Email,
	}
	return tls.NewListener(listen(s.options.HTTPS.Listen), acm.TLSConfig())
}

func (s *Server) listenHTTPS() net.Listener {
	data, err := os.ReadFile(s.options.HTTPS.CertFile)
	if err != nil {
		return listenErr(err)
	}
	cert, err := GetCertificate(data)
	if err != nil {
		return listenErr(err)
	}
	return tls.NewListener(listen(s.options.HTTPS.Listen), &tls.Config{Certificates: []tls.Certificate{*cert}})
}
