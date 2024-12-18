//go:build server || all
// +build server all

package server

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	configurator "github.com/OpenCHAMI/configurator/pkg"
	"github.com/OpenCHAMI/configurator/pkg/generator"
	"github.com/OpenCHAMI/jwtauth/v5"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/sirupsen/logrus"

	openchami_authenticator "github.com/openchami/chi-middleware/auth"
	openchami_logger "github.com/openchami/chi-middleware/log"
	"github.com/rs/zerolog/log"
)

var (
	tokenAuth *jwtauth.JWTAuth = nil
)

type Jwks struct {
	Uri     string
	Retries int
}
type Server struct {
	*http.Server
	Config          *configurator.Config
	Jwks            Jwks `yaml:"jwks"`
	GeneratorParams generator.Params
	TokenAuth       *jwtauth.JWTAuth
}

// Constructor to make a new server instance with an optional config.
func New(config *configurator.Config) *Server {
	// create default config if none supplied
	if config == nil {
		c := configurator.NewConfig()
		config = &c
	}
	// return based on config values
	return &Server{
		Config: config,
		Server: &http.Server{
			Addr: fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port),
		},
		Jwks: Jwks{
			Uri:     config.Server.Jwks.Uri,
			Retries: config.Server.Jwks.Retries,
		},
	}
}

// Main function to start up configurator as a service.
func (s *Server) Serve(cacertPath string) error {
	// create client just for the server to use to fetch data from SMD
	client := &configurator.SmdClient{
		Host: s.Config.SmdClient.Host,
		Port: s.Config.SmdClient.Port,
	}

	// add cert to client if `--cacert` flag is passed
	if cacertPath != "" {
		cacert, _ := os.ReadFile(cacertPath)
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(cacert)
		client.Transport = &http.Transport{
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

	// set the server address with config values
	s.Server.Addr = fmt.Sprintf("%s:%d", s.Config.Server.Host, s.Config.Server.Port)

	// fetch JWKS public key from authorization server
	if s.Config.Server.Jwks.Uri != "" && tokenAuth == nil {
		for i := 0; i < s.Config.Server.Jwks.Retries; i++ {
			var err error
			tokenAuth, err = configurator.FetchPublicKeyFromURL(s.Config.Server.Jwks.Uri)
			if err != nil {
				logrus.Errorf("failed to fetch JWKS: %v", err)
				continue
			}
			break
		}
	}

	// Setup logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// create new go-chi router with its routes
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.StripSlashes)
	router.Use(middleware.Timeout(60 * time.Second))
	router.Use(openchami_logger.OpenCHAMILogger(logger))
	if s.Config.Server.Jwks.Uri != "" {
		router.Group(func(r chi.Router) {
			r.Use(
				jwtauth.Verifier(tokenAuth),
				openchami_authenticator.AuthenticatorWithRequiredClaims(tokenAuth, []string{"sub", "iss", "aud"}),
			)

			// protected routes if using auth
			r.HandleFunc("/generate", s.Generate(client))
			r.HandleFunc("/templates", s.ManageTemplates)
		})
	} else {
		// public routes without auth
		router.HandleFunc("/generate", s.Generate(client))
		router.HandleFunc("/templates", s.ManageTemplates)
	}

	// always available public routes go here (none at the moment)

	s.Handler = router
	return s.ListenAndServe()
}

// TODO: implement a way to shut the server down
func (s *Server) Close() {

}

// This is the corresponding service function to generate templated files, that
// works similarly to the CLI variant. This function takes similiar arguments as
// query parameters that are included in the HTTP request URL.
func (s *Server) Generate(client *configurator.SmdClient) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		// get all of the expect query URL params and validate
		s.GeneratorParams.Target = r.URL.Query().Get("target")
		s.GeneratorParams.Client = client
		if s.GeneratorParams.Target == "" {
			writeErrorResponse(w, "must specify a target")
			return
		}

		// generate a new config file from supplied params
		outputs, err := generator.GenerateWithTarget(s.Config, s.GeneratorParams)
		if err != nil {
			writeErrorResponse(w, "failed to generate file: %v", err)
			return
		}

		// marshal output to JSON then send response to client
		tmp := generator.ConvertContentsToString(outputs)
		b, err := json.Marshal(tmp)
		if err != nil {
			writeErrorResponse(w, "failed to marshal output: %v", err)
			return
		}
		_, err = w.Write(b)
		if err != nil {
			writeErrorResponse(w, "failed to write response: %v", err)
			return
		}
	}
}

// Incomplete WIP function for managing templates remotely. There is currently no
// internal API to do this yet.
//
// TODO: need to implement template managing API first in "internal/generator/templates" or something
func (s *Server) ManageTemplates(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("this is not implemented yet"))
	if err != nil {
		writeErrorResponse(w, "failed to write response: %v", err)
		return
	}
}

// Wrapper function to simplify writting error message responses. This function
// is only intended to be used with the service and nothing else.
func writeErrorResponse(w http.ResponseWriter, format string, a ...any) error {
	errmsg := fmt.Sprintf(format, a...)
	w.Write([]byte(errmsg))
	return fmt.Errorf(errmsg)
}
