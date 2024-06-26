package cmd

import (
	"fmt"
	"maps"
	"strings"

	"github.com/OpenCHAMI/configurator/internal/generator"
	"github.com/spf13/cobra"
)

var (
	pluginDirs []string
	generators map[string]generator.Generator
)

var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect generator plugin information",
	Run: func(cmd *cobra.Command, args []string) {
		// load specific plugins from positional args
		generators = make(map[string]generator.Generator)
		for _, path := range args {
			gen, err := generator.LoadPlugin(path)
			if err != nil {
				fmt.Printf("failed to load plugin at path '%s': %v\n", path, err)
				continue
			}
			generators[path] = gen
		}

		// load plugins and print all plugin details
		if len(pluginDirs) > 0 {

		} else {
			for _, pluginDir := range config.PluginDirs {
				gens, err := generator.LoadPlugins(pluginDir)
				if err != nil {
					fmt.Printf("failed to load plugin: %v\n", err)
					continue
				}
				maps.Copy(generators, gens)
			}
		}

		// print all generator information
		const WIDTH = 40
		if len(generators) > 0 {
			o := ""
			for _, g := range generators {
				o += fmt.Sprintf("- Name:         %s\n", g.GetName())
				o += fmt.Sprintf("  Version:      %s\n", g.GetVersion())
				o += fmt.Sprintf("  Description:  %s\n", g.GetDescription())
				o += "\n"
			}
			o = strings.TrimRight(o, "\n")
			fmt.Printf("%s", o)
		}
	},
}

func init() {
	rootCmd.AddCommand(inspectCmd)
}
