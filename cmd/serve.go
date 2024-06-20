//go:build server || all
// +build server all

package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/OpenCHAMI/configurator/internal/generator"
	"github.com/OpenCHAMI/configurator/internal/server"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start configurator as a server and listen for requests",
	Run: func(cmd *cobra.Command, args []string) {
		// make sure that we have a token present before trying to make request
		if config.AccessToken == "" {
			// TODO: make request to check if request will need token

			// check if OCHAMI_ACCESS_TOKEN env var is set if no access token is provided and use that instead
			accessToken := os.Getenv("ACCESS_TOKEN")
			if accessToken != "" {
				config.AccessToken = accessToken
			} else {
				// TODO: try and fetch token first if it is needed
				if verbose {
					fmt.Printf("No token found. Attempting to generate config without one...\n")
				}
			}
		}

		// use config plugins if none supplied via CLI
		if len(pluginPaths) <= 0 {
			pluginPaths = append(pluginPaths, config.PluginDirs...)
		}

		// show config as JSON and generators if verbose
		if verbose {
			b, err := json.MarshalIndent(config, "", "  ")
			if err != nil {
				fmt.Printf("failed to marshal config: %v\n", err)
			}
			fmt.Printf("%v\n", string(b))
		}

		// set up the routes and start the server
		server := server.Server{
			Server: &http.Server{
				Addr: fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port),
			},
			Jwks: server.Jwks{
				Uri:     config.Server.Jwks.Uri,
				Retries: config.Server.Jwks.Retries,
			},
			GeneratorParams: generator.Params{
				Args:        args,
				PluginPaths: pluginPaths,
				// Target: target,  // NOTE: targets are set via HTTP requests (ex: curl http://configurator:3334/generate?target=dnsmasq)
				Verbose: verbose,
			},
		}
		err := server.Serve(&config)
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("Server closed.")
		} else if err != nil {
			fmt.Errorf("failed to start server: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	serveCmd.Flags().StringVar(&config.Server.Host, "host", config.Server.Host, "set the server host")
	serveCmd.Flags().IntVar(&config.Server.Port, "port", config.Server.Port, "set the server port")
	serveCmd.Flags().StringSliceVar(&pluginPaths, "plugins", nil, "set the generator plugins directory path")
	serveCmd.Flags().StringVar(&config.Server.Jwks.Uri, "jwks-uri", config.Server.Jwks.Uri, "set the JWKS url to fetch public key")
	serveCmd.Flags().IntVar(&config.Server.Jwks.Retries, "jwks-fetch-retries", config.Server.Jwks.Retries, "set the JWKS fetch retry count")
	rootCmd.AddCommand(serveCmd)
}
