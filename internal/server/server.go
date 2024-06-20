//go:build server || all
// +build server all

package server

import (
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

func (s *Server) Serve(config *configurator.Config) error {
	// create client just for the server to use to fetch data from SMD
	_ = &configurator.SmdClient{
		Host: config.SmdClient.Host,
		Port: config.SmdClient.Port,
	}

	// set the server address with config values
	s.Server.Addr = fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)

	// fetch JWKS public key from authorization server
	if config.Server.Jwks.Uri != "" && tokenAuth == nil {
		for i := 0; i < config.Server.Jwks.Retries; i++ {
			var err error
			tokenAuth, err = configurator.FetchPublicKeyFromURL(config.Server.Jwks.Uri)
			if err != nil {
				logrus.Errorf("failed to fetch JWKS: %w", err)
				continue
			}
			break
		}
	}

	var WriteError = func(w http.ResponseWriter, format string, a ...any) {
		errmsg := fmt.Sprintf(format, a...)
		fmt.Printf(errmsg)
		w.Write([]byte(errmsg))
	}

	// create new go-chi router with its routes
	router := chi.NewRouter()
	router.Use(middleware.RedirectSlashes)
	router.Use(middleware.Timeout(60 * time.Second))
	router.Group(func(r chi.Router) {
		if config.Server.Jwks.Uri != "" {
			r.Use(
				jwtauth.Verifier(tokenAuth),
				jwtauth.Authenticator(tokenAuth),
			)
		}
		r.HandleFunc("/generate", func(w http.ResponseWriter, r *http.Request) {
			s.GeneratorParams.Target = r.URL.Query().Get("target")
			output, err := generator.Generate(config, s.GeneratorParams)
			if err != nil {
				WriteError(w, "failed to generate config: %v\n", err)
				return
			}

			_, err = w.Write(output)
			if err != nil {
				WriteError(w, "failed to write response: %v", err)
				return
			}

			// NOTE: we probably don't want to hardcode the types, but should do for now
			// if _type == "dhcp" {
			// 	// fetch eths from SMD
			// 	eths, err := client.FetchEthernetInterfaces()
			// 	if err != nil {
			// 		logrus.Errorf("failed to fetch DHCP metadata: %v\n", err)
			// 		w.Write([]byte("An error has occurred"))
			// 		return
			// 	}
			// 	if len(eths) <= 0 {
			// 		logrus.Warnf("no ethernet interfaces found")
			// 		w.Write([]byte("no ethernet interfaces found"))
			// 		return
			// 	}
			// 	// generate a new config from that data

			// 	// b, err := g.GenerateDHCP(config, eths)
			// 	if err != nil {
			// 		logrus.Errorf("failed to generate DHCP: %v", err)
			// 		w.Write([]byte("An error has occurred."))
			// 		return
			// 	}
			// 	w.Write(b)
			// }
		})
		r.HandleFunc("/templates", func(w http.ResponseWriter, r *http.Request) {
			// TODO: handle GET request
			// TODO: handle POST request

		})
	})
	s.Handler = router
	return s.ListenAndServe()
}
