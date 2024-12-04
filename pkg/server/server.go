//go:build server || all
// +build server all

package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	configurator "github.com/OpenCHAMI/configurator/pkg"
	"github.com/OpenCHAMI/configurator/pkg/client"
	"github.com/OpenCHAMI/configurator/pkg/config"
	"github.com/OpenCHAMI/configurator/pkg/generator"
	"github.com/OpenCHAMI/jwtauth/v5"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"

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
	Config          *config.Config
	Jwks            Jwks `yaml:"jwks"`
	GeneratorParams generator.Params
	TokenAuth       *jwtauth.JWTAuth
	Targets         map[string]Target
}

type Target struct {
	Name       string               `json:"name"`
	PluginPath string               `json:"plugin"`
	Templates  []generator.Template `json:"templates"`
}

// Constructor to make a new server instance with an optional config.
func New(conf *config.Config) *Server {
	// create default config if none supplied
	if conf == nil {
		c := config.New()
		conf = &c
	}
	// return based on config values
	return &Server{
		Config: conf,
		Server: &http.Server{
			Addr: fmt.Sprintf("%s:%d", conf.Server.Host, conf.Server.Port),
		},
		Jwks: Jwks{
			Uri:     conf.Server.Jwks.Uri,
			Retries: conf.Server.Jwks.Retries,
		},
	}
}

// Main function to start up configurator as a service.
func (s *Server) Serve() error {
	// Setup logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// set the server address with config values
	s.Server.Addr = s.Config.Server.Host

	// fetch JWKS public key from authorization server
	if s.Config.Server.Jwks.Uri != "" && tokenAuth == nil {
		for i := 0; i < s.Config.Server.Jwks.Retries; i++ {
			var err error
			tokenAuth, err = configurator.FetchPublicKeyFromURL(s.Config.Server.Jwks.Uri)
			if err != nil {
				log.Error().Err(err).Msgf("failed to fetch JWKS")
				continue
			}
			break
		}
	}

	// create client with opts to use to fetch data from SMD
	opts := []client.Option{
		client.WithAccessToken(s.Config.AccessToken),
		client.WithCertPoolFile(s.Config.CertPath),
	}

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
			r.HandleFunc("/generate", s.Generate(opts...))
			r.Post("/targets", s.createTarget)
		})
	} else {
		// public routes without auth
		router.HandleFunc("/generate", s.Generate(opts...))
		router.Post("/targets", s.createTarget)
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
func (s *Server) Generate(opts ...client.Option) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get all of the expect query URL params and validate
		var (
			target string = r.URL.Query().Get("target")
		)
		s.GeneratorParams = parseGeneratorParams(r, opts...)
		if target == "" {
			writeErrorResponse(w, "must specify a target")
			return
		}

		// try to generate with target supplied by client first
		var (
			t       *Target = s.getTarget(target)
			outputs generator.FileMap
			err     error
		)

		if t != nil {
			outputs, err = generator.Generate(t.PluginPath, s.GeneratorParams)
			if err != nil {

			}
		} else {
			// try and generate a new config file from supplied params
			outputs, err = generator.GenerateWithTarget(s.Config, target)
			if err != nil {
				writeErrorResponse(w, "failed to generate file: %v", err)
				return
			}
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

// Create a new target with name, generator, templates, and files.
//
// Example:
//
//	curl -X POST /target?name=test&plugin=dnsmasq
//
// TODO: need to implement template managing API first in "internal/generator/templates" or something
func (s *Server) createTarget(w http.ResponseWriter, r *http.Request) {
	var (
		target = Target{}
		bytes  []byte
		err    error
	)
	if r == nil {
		writeErrorResponse(w, "request is invalid")
		return
	}

	bytes, err = io.ReadAll(r.Body)
	if err != nil {
		writeErrorResponse(w, "failed to read response body: %v", err)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(bytes, &target)
	if err != nil {
		writeErrorResponse(w, "failed to unmarshal target: %v", err)
		return
	}

	// make sure a plugin and at least one template is supplied
	if target.Name == "" {
		writeErrorResponse(w, "target name is required")
		return
	}
	if target.PluginPath == "" {
		writeErrorResponse(w, "must supply a generator name")
		return
	}
	if len(target.Templates) <= 0 {
		writeErrorResponse(w, "must provided at least one template")
		return
	}

	s.Targets[target.Name] = target

}

func (s *Server) getTarget(target string) *Target {
	t, ok := s.Targets[target]
	if ok {
		return &t
	}
	return nil
}

// Wrapper function to simplify writting error message responses. This function
// is only intended to be used with the service and nothing else.
func writeErrorResponse(w http.ResponseWriter, format string, a ...any) error {
	errmsg := fmt.Sprintf(format, a...)
	log.Error().Msg(errmsg)
	http.Error(w, errmsg, http.StatusInternalServerError)
	return fmt.Errorf(errmsg)
}

func parseGeneratorParams(r *http.Request, opts ...client.Option) generator.Params {
	var params = generator.Params{
		ClientOpts: opts,
	}
	return params
}
