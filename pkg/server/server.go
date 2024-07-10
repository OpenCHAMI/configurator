//go:build server || all
// +build server all

package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	configurator "github.com/OpenCHAMI/configurator/pkg"
	"github.com/OpenCHAMI/configurator/pkg/generator"
	"github.com/OpenCHAMI/jwtauth/v5"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
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
func (s *Server) Serve() error {
	// create client just for the server to use to fetch data from SMD
	_ = &configurator.SmdClient{
		Host: s.Config.SmdClient.Host,
		Port: s.Config.SmdClient.Port,
	}

	// set the server address with config values
	s.Server.Addr = fmt.Sprintf("%s:%d", s.Config.Server.Host, s.Config.Server.Port)

	// fetch JWKS public key from authorization server
	if s.Config.Server.Jwks.Uri != "" && tokenAuth == nil {
		for i := 0; i < s.Config.Server.Jwks.Retries; i++ {
			var err error
			tokenAuth, err = configurator.FetchPublicKeyFromURL(s.Config.Server.Jwks.Uri)
			if err != nil {
				logrus.Errorf("failed to fetch JWKS: %w", err)
				continue
			}
			break
		}
	}

	// create new go-chi router with its routes
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.StripSlashes)
	router.Use(middleware.Timeout(60 * time.Second))
	if s.Config.Server.Jwks.Uri != "" {
		router.Group(func(r chi.Router) {
			r.Use(
				jwtauth.Verifier(tokenAuth),
				jwtauth.Authenticator(tokenAuth),
			)

			// protected routes if using auth
			r.HandleFunc("/generate", s.Generate)
			r.HandleFunc("/templates", s.ManageTemplates)
		})
	} else {
		// public routes without auth
		router.HandleFunc("/generate", s.Generate)
		router.HandleFunc("/templates", s.ManageTemplates)
	}

	// always available public routes go here (none at the moment)

	s.Handler = router
	return s.ListenAndServe()
}

func (s *Server) Close() {

}

// This is the corresponding service function to generate templated files, that
// works similarly to the CLI variant. This function takes similiar arguments as
// query parameters that are included in the HTTP request URL.
func (s *Server) Generate(w http.ResponseWriter, r *http.Request) {
	// get all of the expect query URL params and validate
	s.GeneratorParams.Target = r.URL.Query().Get("target")
	if s.GeneratorParams.Target == "" {
		writeError(w, "no targets supplied")
		return
	}

	// generate a new config file from supplied params
	outputs, err := generator.Generate(s.Config, s.GeneratorParams)
	if err != nil {
		writeError(w, "failed to generate config: %v", err)
		return
	}

	// marshal output to JSON then send response to client
	tmp := generator.ConvertContentsToString(outputs)
	b, err := json.Marshal(tmp)
	if err != nil {
		writeError(w, "failed to marshal output: %v", err)
		return
	}
	_, err = w.Write(b)
	if err != nil {
		writeError(w, "failed to write response: %v", err)
		return
	}
}

// Incomplete WIP function for managing templates remotely. There is currently no
// internal API to do this yet.
//
// TODO: need to implement template managing API first in "internal/generator/templates" or something
func (s *Server) ManageTemplates(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("this is not implemented yet"))
	if err != nil {
		writeError(w, "failed to write response: %v", err)
		return
	}
}

// Wrapper function to simplify writting error message responses. This function
// is only intended to be used with the service and nothing else.
func writeError(w http.ResponseWriter, format string, a ...any) {
	errmsg := fmt.Sprintf(format, a...)
	fmt.Printf(errmsg)
	w.Write([]byte(errmsg))
}
