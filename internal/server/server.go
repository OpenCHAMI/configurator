//go:build server || all
// +build server all

package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	configurator "github.com/OpenCHAMI/configurator/internal"
	"github.com/OpenCHAMI/configurator/internal/generator"
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

func New() *Server {
	return &Server{
		Server: &http.Server{
			Addr: "localhost:3334",
		},
		Jwks: Jwks{
			Uri:     "",
			Retries: 5,
		},
	}
}

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

	// always public routes go here (none at the moment)

	s.Handler = router
	return s.ListenAndServe()
}

func WriteError(w http.ResponseWriter, format string, a ...any) {
	errmsg := fmt.Sprintf(format, a...)
	fmt.Printf(errmsg)
	w.Write([]byte(errmsg))
}

func (s *Server) Generate(w http.ResponseWriter, r *http.Request) {
	s.GeneratorParams.Target = r.URL.Query().Get("target")
	outputs, err := generator.Generate(s.Config, s.GeneratorParams)
	if err != nil {
		WriteError(w, "failed to generate config: %v", err)
		return
	}

	// convert byte arrays to string
	tmp := map[string]string{}
	for path, output := range outputs {
		tmp[path] = string(output)
	}

	// marshal output to JSON then send
	b, err := json.Marshal(tmp)
	if err != nil {
		WriteError(w, "failed to marshal output: %v", err)
		return
	}
	_, err = w.Write(b)
	if err != nil {
		WriteError(w, "failed to write response: %v", err)
		return
	}
}

func (s *Server) ManageTemplates(w http.ResponseWriter, r *http.Request) {
	// TODO: need to implement template managing API first in "internal/generator/templates" or something
	_, err := w.Write([]byte("this is not implemented yet"))
	if err != nil {
		WriteError(w, "failed to write response: %v", err)
		return
	}
}
