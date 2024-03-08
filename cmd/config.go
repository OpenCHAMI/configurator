package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	configurator "github.com/OpenCHAMI/configurator/internal"
	"github.com/OpenCHAMI/configurator/internal/util"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Create a new default config file",
	Run: func(cmd *cobra.Command, args []string) {
		// create a new config at all args (paths)
		for _, path := range args {
			// check and make sure something doesn't exist first
			if exists, err := util.PathExists(path); exists || err != nil {
				fmt.Printf("file or directory exists\n")
				continue
			}
			configurator.SaveDefaultConfig(path)
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
