package cmd

import (
	"fmt"
	"os"

	"github.com/OpenCHAMI/configurator/pkg/config"
	"github.com/OpenCHAMI/configurator/pkg/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	conf        config.Config
	configPath  string
	cacertPath  string
	verbose     bool
	targets     []string
	outputPath  string
	accessToken string
	remoteHost  string
)

var rootCmd = &cobra.Command{
	Use:   "configurator",
	Short: "Dynamically generate files defined by generators",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "set the config path")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "set to enable verbose output")
	rootCmd.PersistentFlags().StringVar(&cacertPath, "cacert", "", "path to CA cert. (defaults to system CAs)")
}

func initConfig() {
	// empty from not being set
	if configPath != "" {
		exists, err := util.PathExists(configPath)
		if err != nil {
			log.Error().Err(err).Str("path", configPath).Msg("failed to load config")
			os.Exit(1)
		} else if exists {
			conf = config.Load(configPath)
		} else {
			// show error and exit since a path was specified
			log.Error().Str("path", configPath).Msg("config file not found")
			os.Exit(1)
		}
	} else {
		// set to the default value and create a new one
		configPath = "./config.yaml"
		conf = config.New()
	}

	//
	// set environment variables to override config values
	//

	// set the JWKS url if we find the CONFIGURATOR_JWKS_URL environment variable
	jwksUrl := os.Getenv("CONFIGURATOR_JWKS_URL")
	if jwksUrl != "" {
		conf.Server.Jwks.Uri = jwksUrl
	}
}
