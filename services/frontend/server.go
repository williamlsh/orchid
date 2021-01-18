package frontend

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/net/netutil"
)

// Server represents an HTTP server.
type Server struct {
	*http.Server

	// MaxConnections limits the number of accepted simultaneous connections.
	// Defaults to 0, indicating no limit.
	MaxConnections int
}

// NewServer creates a new HTTP server.
func NewServer(addr string, handler http.Handler) *Server {
	return &Server{
		Server: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
			TLSConfig: &tls.Config{
				NextProtos:       []string{"h2", "http/1.1"},
				MinVersion:       tls.VersionTLS12,
				CurvePreferences: []tls.CurveID{tls.CurveP256, tls.X25519},
				CipherSuites: []uint16{
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				},
				PreferServerCipherSuites: true,
			},
		},
	}
}

// IsTLS checks wether TLS is enabled.
func (srv *Server) IsTLS() bool {
	return len(srv.TLSConfig.Certificates) > 0 || srv.TLSConfig.GetCertificate != nil
}

// Start starts either a HTTP server or HTTPS server.
func (srv *Server) Start() error {
	ln, err := srv.Listen()
	if err != nil {
		return err
	}
	if srv.IsTLS() {
		ln = tls.NewListener(ln, srv.TLSConfig)
	}
	return srv.Serve(ln)
}

// Listen returns a TCP listener, is MaxConnections is larger than 0, it limits the listener's max connections.
func (srv *Server) Listen() (net.Listener, error) {
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return nil, err
	}

	if srv.MaxConnections > 0 {
		ln = netutil.LimitListener(ln, srv.MaxConnections)
	}

	return ln, nil
}

// GetCertificate injects auto cert from Let's Encrypt.
// It also starts an HTTP server to serve Let's Encrypt's ACME challenge.
func (srv *Server) GetCertificate(hosts ...string) {
	certManager := autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache("cert-cache"),
		// Put your domain here:
		HostPolicy: autocert.HostWhitelist(hosts...),
	}
	srv.TLSConfig.GetCertificate = certManager.GetCertificate

	go http.ListenAndServe(":8081", certManager.HTTPHandler(nil))
}
