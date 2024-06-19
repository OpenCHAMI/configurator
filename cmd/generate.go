//go:build client || all
// +build client all

package cmd

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"

	configurator "github.com/OpenCHAMI/configurator/internal"
	"github.com/OpenCHAMI/configurator/internal/generator"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	tokenFetchRetries int
	pluginPaths       []string
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a config file from state management",
	Run: func(cmd *cobra.Command, args []string) {
		// load generator plugins to generate configs or to print
		var (
			generators = make(map[string]generator.Generator)
			client     = configurator.SmdClient{
				Host:        config.SmdClient.Host,
				Port:        config.SmdClient.Port,
				AccessToken: config.AccessToken,
			}
		)
		for _, path := range pluginPaths {
			if verbose {
				fmt.Printf("loading plugins from '%s'\n", path)
			}
			gens, err := generator.LoadPlugins(path)
			if err != nil {
				fmt.Printf("failed to load plugins: %v\n", err)
				err = nil
				continue
			}

			// add loaded generator plugins to set
			maps.Copy(generators, gens)
		}

		// show config as JSON and generators if verbose
		if verbose {
			b, err := json.MarshalIndent(config, "", "  ")
			if err != nil {
				fmt.Printf("failed to marshal config: %v\n", err)
			}
			fmt.Printf("%v\n", string(b))
		}

		// show available targets then exit
		if len(args) == 0 && len(targets) == 0 {
			for g := range generators {
				fmt.Printf("\tplugin: %s, name:\n", g)
			}
			os.Exit(0)
		}

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
				// tmp := strings.Split(target, ":")

				// make sure each target has at least two args
				// if len(tmp) < 2 {
				// 	message := "target"
				// 	if len(tmp) == 1 {
				// 		message += fmt.Sprintf(" '%s'", tmp[0])
				// 	}
				// 	message += " does not provide enough arguments (args: \"type:template\")"
				// 	logrus.Errorf(message)
				// 	continue
				// }
				// var (
				// 	_type     = tmp[0]
				// 	_template = tmp[1]
				// )
				// g := generator.Generator{
				// 	Type:     tmp[0],
				// 	Template: tmp[1],
				// }

				// check if another param is specified
				// targetPath := ""
				// if len(tmp) > 2 {
				// 	targetPath = tmp[2]
				// }

				// run the generator plugin from target passed
				gen := generators[target]
				if gen == nil {
					fmt.Printf("invalid generator target (%s)\n", target)
					continue
				}
				output, err := gen.Generate(
					&config,
					generator.WithTemplate(gen.GetName()),
					generator.WithClient(client),
				)
				if err != nil {
					fmt.Printf("failed to generate config: %v\n", err)
					continue
				}

				// NOTE: we probably don't want to hardcode the types, but should do for now
				// ext := ""
				// contents := []byte{}
				// if _type == "dhcp" {
				// 	// fetch eths from SMD
				// 	eths, err := client.FetchEthernetInterfaces()
				// 	if err != nil {
				// 		logrus.Errorf("failed to fetch DHCP metadata: %v\n", err)
				// 		continue
				// 	}
				// 	if len(eths) <= 0 {
				// 		continue
				// 	}
				// 	// generate a new config from that data
				// 	contents, err = g.GenerateDHCP(&config, eths)
				// 	if err != nil {
				// 		logrus.Errorf("failed to generate DHCP config file: %v\n", err)
				// 		continue
				// 	}
				// 	ext = "conf"
				// } else if g.Type == "dns" {
				// 	// TODO: fetch from SMD
				// 	// TODO: generate config from pulled info

				// } else if g.Type == "syslog" {

				// } else if g.Type == "ansible" {

				// } else if g.Type == "warewulf" {

				// }

				// write config output if no specific targetPath is set
				// if targetPath == "" {
				if outputPath == "" {
					// write only to stdout
					fmt.Printf("%s\n", string(output))
				} else if outputPath != "" && targetCount == 1 {
					// write just a single file using template name
					err := os.WriteFile(outputPath, output, 0o644)
					if err != nil {
						logrus.Errorf("failed to write config to file: %v", err)
						continue
					}
				} else if outputPath != "" && targetCount > 1 {
					// write multiple files in directory using template name
					err := os.WriteFile(fmt.Sprintf("%s/%s.%s", filepath.Clean(outputPath), target, ".conf"), output, 0o644)
					if err != nil {
						logrus.Errorf("failed to write config to file: %v", err)
						continue
					}
				}
				// }
			} // for targets
		}
	},
}

func init() {
	generateCmd.Flags().StringSliceVar(&targets, "target", nil, "set the target configs to make")
	generateCmd.Flags().StringSliceVar(&pluginPaths, "plugin", nil, "set the generator plugins directory path to shared libraries")
	generateCmd.Flags().StringVarP(&outputPath, "output", "o", "", "set the output path for config targets")
	generateCmd.Flags().IntVar(&tokenFetchRetries, "fetch-retries", 5, "set the number of retries to fetch an access token")

	rootCmd.AddCommand(generateCmd)
}
