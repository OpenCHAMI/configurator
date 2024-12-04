//go:build client || all
// +build client all

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/OpenCHAMI/configurator/pkg/config"
	"github.com/OpenCHAMI/configurator/pkg/generator"
	"github.com/OpenCHAMI/configurator/pkg/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	tokenFetchRetries int
	templatePaths     []string
	pluginPath        string
	useCompression    bool
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a config file from state management",
	Run: func(cmd *cobra.Command, args []string) {
		// make sure that we have a token present before trying to make request
		if conf.AccessToken == "" {
			// check if OCHAMI_ACCESS_TOKEN env var is set if no access token is provided and use that instead
			accessToken := os.Getenv("ACCESS_TOKEN")
			if accessToken != "" {
				conf.AccessToken = accessToken
			} else {
				// TODO: try and fetch token first if it is needed
				if verbose {
					log.Warn().Msg("No token found. Attempting to generate conf without one...\n")
				}
			}
		}

		// use cert path from cobra if empty
		if conf.CertPath == "" {
			conf.CertPath = cacertPath
		}

		// show conf as JSON and generators if verbose
		if verbose {
			b, err := json.MarshalIndent(conf, "", "  ")
			if err != nil {
				log.Error().Err(err).Printf("failed to marshal config")
			}
			// print the config file as JSON
			fmt.Printf("%v\n", string(b))
		}

		// run all of the target recursively until completion if provided
		if len(targets) > 0 {
			RunTargets(&conf, args, targets...)
		} else {
			if pluginPath == "" {
				log.Error().Msg("no plugin path specified")
				return
			}

			// load the templates to use
			templates := map[string]generator.Template{}
			for _, path := range templatePaths {
				template := generator.Template{}
				template.LoadFromFile(path)
				if !template.IsEmpty() {
					templates[path] = template
				}
			}

			params := generator.Params{
				Templates: templates,
			}

			// run generator.Generate() with just plugin path and templates provided
			outputBytes, err := generator.Generate(pluginPath, params)
			if err != nil {
				log.Error().Err(err).Msg("failed to generate files")
			}

			// if we have more than one target and output is set, create configs in directory
			outputMap := generator.ConvertContentsToString(outputBytes)
			writeOutput(outputBytes, len(targets), len(outputMap))
		}
	},
}

// Generate files by supplying a list of targets as string values. Currently,
// targets are defined statically in a config file. Targets are ran recursively
// if more targets are nested in a defined target, but will not run additional
// child targets if it is the same as the parent.
//
// NOTE: This may be changed in the future how this is done.
func RunTargets(conf *config.Config, args []string, targets ...string) {
	// generate config with each supplied target
	for _, target := range targets {
		outputBytes, err := generator.GenerateWithTarget(conf, target)
		if err != nil {
			log.Error().Err(err).Msg("failed to generate config")
			os.Exit(1)
		}

		// if we have more than one target and output is set, create configs in directory
		outputMap := generator.ConvertContentsToString(outputBytes)
		writeOutput(outputBytes, len(targets), len(outputMap))

		// remove any targets that are the same as current to prevent infinite loop
		nextTargets := util.CopyIf(conf.Targets[target].RunTargets, func(nextTarget string) bool {
			return nextTarget != target
		})

		// ...then, run any other targets that the current target has
		RunTargets(conf, args, nextTargets...)
	}
}

func writeOutput(outputBytes generator.FileMap, targetCount int, templateCount int) {
	outputMap := generator.ConvertContentsToString(outputBytes)
	if outputPath == "" {
		// write only to stdout by default
		if len(outputMap) == 1 {
			for _, contents := range outputMap {
				fmt.Printf("%s\n", string(contents))
			}
		} else {
			for path, contents := range outputMap {
				fmt.Printf("-- file: %s, size: %d B\n%s\n", path, len(contents), string(contents))
			}
		}
	} else if outputPath != "" && targetCount == 1 && templateCount == 1 {
		// write just a single file using provided name
		for _, contents := range outputBytes {
			err := os.WriteFile(outputPath, contents, 0o644)
			if err != nil {
				log.Error().Err(err).Msg("failed to write conf to file")
				os.Exit(1)
			}
			log.Info().Msgf("wrote file to '%s'\n", outputPath)
		}
	} else if outputPath != "" && targetCount > 1 && useCompression {
		// write multiple files to archive, compress, then save to output path
		out, err := os.Create(fmt.Sprintf("%s.tar.gz", outputPath))
		if err != nil {
			log.Error().Err(err).Msg("failed to write archive")
			os.Exit(1)
		}
		files := make([]string, len(outputBytes))
		i := 0
		for path := range outputBytes {
			files[i] = path
			i++
		}
		err = util.CreateArchive(files, out)
		if err != nil {
			log.Error().Err(err).Msg("failed to create archive")
			os.Exit(1)
		}

	} else if outputPath != "" && targetCount > 1 || templateCount > 1 {
		// write multiple files in directory using template name
		err := os.MkdirAll(filepath.Clean(outputPath), 0o755)
		if err != nil {
			log.Error().Err(err).Msg("failed to make output directory")
			os.Exit(1)
		}
		for path, contents := range outputBytes {
			filename := filepath.Base(path)
			cleanPath := fmt.Sprintf("%s/%s", filepath.Clean(outputPath), filename)
			err := os.WriteFile(cleanPath, contents, 0o755)
			if err != nil {
				log.Error().Err(err).Msg("failed to write conf to file")
				os.Exit(1)
			}
			log.Info().Msgf("wrote file to '%s'\n", cleanPath)
		}
	}
}

func init() {
	generateCmd.Flags().StringSliceVar(&targets, "target", []string{}, "set the targets to run pre-defined conf")
	generateCmd.Flags().StringSliceVar(&templatePaths, "template", []string{}, "set the paths for the Jinja 2 templates to use")
	generateCmd.Flags().StringVar(&pluginPath, "plugin", "", "set the generator plugin path")
	generateCmd.Flags().StringVarP(&outputPath, "output", "o", "", "set the output path for conf targets")
	generateCmd.Flags().IntVar(&tokenFetchRetries, "fetch-retries", 5, "set the number of retries to fetch an access token")
	generateCmd.Flags().StringVar(&remoteHost, "host", "http://localhost", "set the remote host")
	generateCmd.Flags().BoolVar(&useCompression, "compress", false, "set whether to archive and compress multiple file outputs")

	// requires either 'target' by itself or 'plugin' and 'templates' together
	// generateCmd.MarkFlagsOneRequired("target", "plugin")
	generateCmd.MarkFlagsMutuallyExclusive("target", "plugin")
	generateCmd.MarkFlagsMutuallyExclusive("target", "template")
	generateCmd.MarkFlagsRequiredTogether("plugin", "template")

	rootCmd.AddCommand(generateCmd)
}
