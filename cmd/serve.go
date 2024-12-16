//go:build server || all
// +build server all

package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/OpenCHAMI/configurator/pkg/server"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start configurator as a server and listen for requests",
	Run: func(cmd *cobra.Command, args []string) {
		// make sure that we have a token present before trying to make request
		if conf.AccessToken == "" {
			// check if ACCESS_TOKEN env var is set if no access token is provided and use that instead
			accessToken := os.Getenv("ACCESS_TOKEN")
			if accessToken != "" {
				conf.AccessToken = accessToken
			} else {
				if verbose {
					log.Warn().Msg("No token found. Continuing without one...\n")
				}
			}
		}

		// show config as JSON and generators if verbose
		if verbose {
			b, err := json.MarshalIndent(conf, "", "\t")
			if err != nil {
				log.Error().Err(err).Msg("failed to marshal config")
			}
			fmt.Printf("%v\n", string(b))
		}

		// start listening with the server
		var (
			s   *server.Server = server.New(&conf)
			err error          = s.Serve()
		)
		if errors.Is(err, http.ErrServerClosed) {
			if verbose {
				log.Info().Msg("server closed")
			}
		} else if err != nil {
			log.Error().Err(err).Msg("failed to start server")
			os.Exit(1)
		}
	},
}

func init() {
	serveCmd.Flags().StringVar(&conf.Server.Host, "host", conf.Server.Host, "set the server host and port")
	// serveCmd.Flags().StringVar(&pluginPath, "plugin", "", "set the generator plugins directory path")
	serveCmd.Flags().StringVar(&conf.Server.Jwks.Uri, "jwks-uri", conf.Server.Jwks.Uri, "set the JWKS url to fetch public key")
	serveCmd.Flags().IntVar(&conf.Server.Jwks.Retries, "jwks-fetch-retries", conf.Server.Jwks.Retries, "set the JWKS fetch retry count")
	rootCmd.AddCommand(serveCmd)
}
