//go:build client || all
// +build client all

package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/OpenCHAMI/configurator/pkg/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch a config file from a remote instance of configurator",
	Long:  "This command is simplified to make a HTTP request to the a configurator service.",
	Run: func(cmd *cobra.Command, args []string) {
		// make sure a host is set
		if remoteHost == "" {
			log.Error().Msg("no '--host' argument set")
			return
		}

		// check to see if an access token is available from env
		if conf.AccessToken == "" {
			// check if OCHAMI_ACCESS_TOKEN env var is set if no access token is provided and use that instead
			accessToken := os.Getenv("ACCESS_TOKEN")
			if accessToken != "" {
				conf.AccessToken = accessToken
			} else {
				// TODO: try and fetch token first if it is needed
				if verbose {
					log.Warn().Msg("No token found. Attempting to generate config without one...")
				}
			}
		}

		// add the "Authorization" header if an access token is supplied
		headers := map[string]string{}
		if accessToken != "" {
			headers["Authorization"] = "Bearer " + accessToken
		}

		for _, target := range targets {
			// make a request for each target
			url := fmt.Sprintf("%s/generate?target=%s", remoteHost, target)
			res, body, err := util.MakeRequest(url, http.MethodGet, nil, headers)
			if err != nil {
				log.Error().Err(err).Msg("failed to make request")
				return
			}
			// handle getting other error codes other than a 200
			if res != nil {
				if res.StatusCode == http.StatusOK {
					log.Info().Msgf("%s\n", string(body))
				}
			}
		}
	},
}

func init() {
	fetchCmd.Flags().StringVar(&remoteHost, "host", "", "set the remote configurator host and port")
	fetchCmd.Flags().StringSliceVar(&targets, "target", nil, "set the target configs to make")
	fetchCmd.Flags().StringVarP(&outputPath, "output", "o", "", "set the output path for config targets")
	fetchCmd.Flags().StringVar(&accessToken, "access-token", "o", "set the output path for config targets")

	rootCmd.AddCommand(fetchCmd)
}
