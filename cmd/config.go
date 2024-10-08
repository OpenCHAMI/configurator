package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	configurator "github.com/OpenCHAMI/configurator/pkg"
	"github.com/OpenCHAMI/configurator/pkg/util"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Create a new default config file",
	Run: func(cmd *cobra.Command, args []string) {
		// create a new config at all args (paths)
		//
		// TODO: change this to only take a single arg since more
		// than one arg is *maybe* a mistake
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
