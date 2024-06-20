//go:build client || all
// +build client all

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/OpenCHAMI/configurator/internal/generator"
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
		// if we have more than one target and output is set, create configs in directory
		targetCount := len(targets)
		if outputPath != "" && targetCount > 1 {
			err := os.MkdirAll(outputPath, 0o755)
			if err != nil {
				fmt.Printf("failed to make output directory: %v", err)
				os.Exit(1)
			}
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

		// generate config with each supplied target
		for _, target := range targets {
			params := generator.Params{
				Args:        args,
				PluginPaths: pluginPaths,
				Target:      target,
				Verbose:     verbose,
			}
			output, err := generator.Generate(&config, params)
			if err != nil {
				fmt.Printf("failed to generate config: %v\n", err)
				os.Exit(1)
			}

			// write config output if no specific targetPath is set
			if outputPath == "" {
				// write only to stdout
				fmt.Printf("%s\n", string(output))
			} else if outputPath != "" && targetCount == 1 {
				// write just a single file using template name
				err := os.WriteFile(outputPath, output, 0o644)
				if err != nil {
					fmt.Printf("failed to write config to file: %v", err)
					os.Exit(1)
				}
			} else if outputPath != "" && targetCount > 1 {
				// write multiple files in directory using template name
				err := os.WriteFile(fmt.Sprintf("%s/%s.%s", filepath.Clean(outputPath), target, ".conf"), output, 0o644)
				if err != nil {
					fmt.Printf("failed to write config to file: %v", err)
					os.Exit(1)
				}
			}
		}

	},
}

func init() {
	generateCmd.Flags().StringSliceVar(&targets, "target", []string{}, "set the target configs to make")
	generateCmd.Flags().StringSliceVar(&pluginPaths, "plugins", []string{}, "set the generator plugins directory path")
	generateCmd.Flags().StringVarP(&outputPath, "output", "o", "", "set the output path for config targets")
	generateCmd.Flags().IntVar(&tokenFetchRetries, "fetch-retries", 5, "set the number of retries to fetch an access token")

	rootCmd.AddCommand(generateCmd)
}
