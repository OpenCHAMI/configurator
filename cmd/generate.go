//go:build client || all
// +build client all

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/OpenCHAMI/configurator/internal/generator"
	"github.com/OpenCHAMI/configurator/internal/util"
	"github.com/spf13/cobra"
)

var (
	tokenFetchRetries int
	pluginPaths       []string
	cacertPath        string
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

		// use cert path from cobra if empty
		// TODO: this needs to be checked for the correct desired behavior
		if config.CertPath == "" {
			config.CertPath = cacertPath
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
			outputBytes, err := generator.Generate(&config, params)
			if err != nil {
				fmt.Printf("failed to generate config: %v\n", err)
				os.Exit(1)
			}

			outputMap := util.ConvertMapOutput(outputBytes)
			if outputPath == "" {
				// write only to stdout by default
				if len(outputMap) == 1 {
					for _, contents := range outputMap {
						fmt.Printf("%s\n", string(contents))
					}
				} else {
					for path, contents := range outputMap {
						fmt.Printf("-- file: %s, size: %d\n%s\n", path, len(contents), string(contents))
					}
				}
			} else if outputPath != "" && targetCount == 1 && len(outputMap) == 1 {
				// write just a single file using provided name
				for _, contents := range outputBytes {
					// FIXME: fix output paths to not overwrite each other with multiple templates
					err := os.WriteFile(outputPath, contents, 0o644)
					if err != nil {
						fmt.Printf("failed to write config to file: %v", err)
						os.Exit(1)
					}
					fmt.Printf("wrote file to '%s'\n", outputPath)
				}
			} else if outputPath != "" && targetCount > 1 || len(outputMap) > 1 {
				// write multiple files in directory using template name
				for _, contents := range outputBytes {
					cleanPath := fmt.Sprintf("%s/%s.%s", filepath.Clean(outputPath), target, "conf")
					err := os.WriteFile(cleanPath, contents, 0o644)
					if err != nil {
						fmt.Printf("failed to write config to file: %v", err)
						os.Exit(1)
					}
					fmt.Printf("wrote file to '%s'\n", cleanPath)
				}
			}
		}

	},
}

func init() {
	generateCmd.Flags().StringSliceVar(&targets, "target", []string{}, "set the target configs to make")
	generateCmd.Flags().StringSliceVar(&pluginPaths, "plugins", []string{}, "set the generator plugins directory path")
	generateCmd.Flags().StringVarP(&outputPath, "output", "o", "", "set the output path for config targets")
	generateCmd.Flags().StringVar(&cacertPath, "ca-cert", "", "path to CA cert. (defaults to system CAs)")
	generateCmd.Flags().IntVar(&tokenFetchRetries, "fetch-retries", 5, "set the number of retries to fetch an access token")

	rootCmd.AddCommand(generateCmd)
}
