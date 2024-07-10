package cmd

import (
	"fmt"
	"os"

	configurator "github.com/OpenCHAMI/configurator/pkg"
	"github.com/OpenCHAMI/configurator/pkg/util"
	"github.com/spf13/cobra"
)

var (
	configPath string
	config     configurator.Config
	verbose    bool
	targets    []string
	outputPath string
)

var rootCmd = &cobra.Command{
	Use:   "configurator",
	Short: "Tool for building common config files",
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
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "./config.yaml", "set the config path")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "set to enable verbose output")
}

func initConfig() {
	if configPath != "" {
		exists, err := util.PathExists(configPath)
		if err != nil {
			fmt.Printf("failed to load config")
			os.Exit(1)
		} else if exists {
			config = configurator.LoadConfig(configPath)
		} else {
			config = configurator.NewConfig()
		}
	} else {
		config = configurator.NewConfig()
	}

	//
	// set environment variables to override config values
	//

	// set the JWKS url if we find the CONFIGURATOR_JWKS_URL environment variable
	jwksUrl := os.Getenv("CONFIGURATOR_JWKS_URL")
	if jwksUrl != "" {
		config.Server.Jwks.Uri = jwksUrl
	}
}
