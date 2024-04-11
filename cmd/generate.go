//go:build client || all
// +build client all

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	configurator "github.com/OpenCHAMI/configurator/internal"
	"github.com/OpenCHAMI/configurator/internal/generator"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	tokenFetchRetries int
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a config file from system state",
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
			// if we have more than one target and output is set, create configs in directory
			targetCount := len(targets)
			if outputPath != "" && targetCount > 1 {
				err := os.MkdirAll(outputPath, 0o755)
				if err != nil {
					logrus.Errorf("failed to make output directory: %v", err)
					return
				}
			}

			for _, target := range targets {
				// split the target and type
				tmp := strings.Split(target, ":")

				// make sure each target has at least two args
				if len(tmp) < 2 {
					message := "target"
					if len(tmp) == 1 {
						message += fmt.Sprintf(" '%s'", tmp[1])
					}
					message += " does not provide enough arguments (args: \"type:template\")"
					logrus.Errorf(message)
					continue
				}
				g := generator.Generator{
					Type:     tmp[0],
					Template: tmp[1],
				}

				// check if another param is specified
				targetPath := ""
				if len(tmp) > 2 {
					targetPath = tmp[2]
				}

				// NOTE: we probably don't want to hardcode the types, but should do for now
				ext := ""
				contents := []byte{}
				if g.Type == "dhcp" {
					// fetch eths from SMD
					eths, err := client.FetchEthernetInterfaces()
					if err != nil {
						logrus.Errorf("failed to fetch DHCP metadata: %v\n", err)
						continue
					}
					if len(eths) <= 0 {
						continue
					}
					// generate a new config from that data
					contents, err = g.GenerateDHCP(&config, eths)
					if err != nil {
						logrus.Errorf("failed to generate DHCP config file: %v\n", err)
						continue
					}
					ext = "conf"
				} else if g.Type == "dns" {
					// TODO: fetch from SMD
					// TODO: generate config from pulled info

				} else if g.Type == "syslog" {

				} else if g.Type == "ansible" {

				} else if g.Type == "warewulf" {

				}

				// write config output if no specific targetPath is set
				if targetPath == "" {
					if outputPath == "" {
						// write only to stdout
						fmt.Printf("%s\n", "")
					} else if outputPath != "" && targetCount == 1 {
						// write just a single file using template name
						err := os.WriteFile(outputPath, contents, 0o644)
						if err != nil {
							logrus.Errorf("failed to write config to file: %v", err)
							continue
						}
					} else if outputPath != "" && targetCount > 1 {
						// write multiple files in directory using template name
						err := os.WriteFile(fmt.Sprintf("%s/%s.%s", filepath.Clean(outputPath), g.Template, ext), contents, 0o644)
						if err != nil {
							logrus.Errorf("failed to write config to file: %v", err)
							continue
						}
					}
				}
			} // for targets
		}
	},
}

func init() {
	generateCmd.Flags().StringSliceVar(&targets, "target", nil, "set the target configs to make")
	generateCmd.Flags().StringVarP(&outputPath, "output", "o", "", "set the output path for config targets")
	generateCmd.Flags().IntVar(&tokenFetchRetries, "fetch-retries", 5, "set the number of retries to fetch an access token")

	rootCmd.AddCommand(generateCmd)
}
