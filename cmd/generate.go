//go:build client || all
// +build client all

package cmd

import (
	"fmt"
	"os"
	"strings"

	configurator "github.com/OpenCHAMI/configurator/internal"
	"github.com/OpenCHAMI/configurator/internal/generator"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	targets           []string
	tokenFetchRetries int
)
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Create a config file from current system state",
	Run: func(cmd *cobra.Command, args []string) {
		client := configurator.SmdClient{
			Host:        config.SmdHost,
			Port:        config.SmdPort,
			AccessToken: config.AccessToken,
		}

		// make sure that we have a token present before trying to make request
		if config.AccessToken == "" {
			// check if OCHAMI_ACCESS_TOKEN env var is set if no access token is provided and use that instead

			accessToken := os.Getenv("OCHAMI_ACCESS_TOKEN")
			if accessToken != "" {
				config.AccessToken = accessToken
			} else {
				fmt.Printf("No token found. Attempting to generate config without one...\n")
			}
		}

		if targets == nil {
			logrus.Errorf("no target supplied (--target type:template)")
		} else {
			for _, target := range targets {
				// split the target and type
				tmp := strings.Split(target, ":")
				g := generator.Generator{
					Type:     tmp[0],
					Template: tmp[1],
				}

				// NOTE: we probably don't want to hardcode the types, but should do for now
				if g.Type == "dhcp" {
					// fetch eths from SMD
					eths, err := client.FetchEthernetInterfaces()
					if err != nil {
						logrus.Errorf("failed to fetch DHCP metadata: %v\n", err)
					}
					if len(eths) <= 0 {
						break
					}
					// generate a new config from that data

					g.GenerateDHCP(&config, eths)
				} else if g.Type == "dns" {
					// TODO: fetch from SMD
					// TODO: generate config from pulled info

				} else if g.Type == "syslog" {

				} else if g.Type == "ansible" {

				} else if g.Type == "warewulf" {

				}

			}
		}

	},
}

func init() {
	generateCmd.Flags().StringSliceVar(&targets, "target", nil, "set the target configs to make")
	generateCmd.Flags().IntVar(&tokenFetchRetries, "fetch-retries", 5, "set the number of retries to fetch an access token")

	rootCmd.AddCommand(generateCmd)
}
