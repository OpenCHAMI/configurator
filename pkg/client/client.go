package client

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"os"
	"time"
)

type Option func(*Params)
type Params struct {
	Host        string `yaml:"host"`
	AccessToken string `yaml:"access-token"`
	Transport   *http.Transport
}

func ToParams(opts ...Option) *Params {
	params := &Params{}
	for _, opt := range opts {
		opt(params)
	}
	return params
}

func WithHost(host string) Option {
	return func(c *Params) {
		c.Host = host
	}
}

func WithAccessToken(token string) Option {
	return func(c *Params) {
		c.AccessToken = token
	}
}

func WithCertPool(certPool *x509.CertPool) Option {
	return func(c *Params) {
		c.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            certPool,
				InsecureSkipVerify: true,
			},
			DisableKeepAlives: true,
			Dial: (&net.Dialer{
				Timeout:   120 * time.Second,
				KeepAlive: 120 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   120 * time.Second,
			ResponseHeaderTimeout: 120 * time.Second,
		}
	}
}

// FIXME: Need to check for errors when reading from a file
func WithCertPoolFile(certPath string) Option {
	if certPath == "" {
		return func(sc *Params) {}
	}
	cacert, _ := os.ReadFile(certPath)
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(cacert)
	return WithCertPool(certPool)
}
