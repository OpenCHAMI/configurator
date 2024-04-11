//go:build server || all
// +build server all

package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/OpenCHAMI/configurator/internal/server"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start configurator as a server and listen for requests",
	Run: func(cmd *cobra.Command, args []string) {
		// set up the routes and start the server
		server := server.New()
		err := server.Start(&config)
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("Server closed.")
		} else if err != nil {
			logrus.Errorf("failed to start server: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	serveCmd.Flags().StringVar(&config.Server.Host, "host", config.Server.Host, "set the server host")
	serveCmd.Flags().IntVar(&config.Server.Port, "port", config.Server.Port, "set the server port")
	serveCmd.Flags().StringVar(&config.Options.JwksUri, "jwks-uri", config.Options.JwksUri, "set the JWKS url to fetch public key")
	serveCmd.Flags().IntVar(&config.Options.JwksRetries, "jwks-fetch-retries", config.Options.JwksRetries, "set the JWKS fetch retry count")
	rootCmd.AddCommand(serveCmd)
}
