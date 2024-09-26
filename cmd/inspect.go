package cmd

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/OpenCHAMI/configurator/pkg/generator"
	"github.com/rodaine/table"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	byTarget bool
)

var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect generator plugin information",
	Long:  "The 'inspect' sub-command takes a list of directories and prints all found plugin information.",
	Run: func(cmd *cobra.Command, args []string) {
		// set up table formatter
		table.DefaultHeaderFormatter = func(format string, vals ...interface{}) string {
			return strings.ToUpper(fmt.Sprintf(format, vals...))
		}

		// TODO: remove duplicate args from CLI

		// load specific plugins from positional args
		var generators = make(map[string]generator.Generator)
		for _, path := range args {
			err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
				}
				gen, err := generator.LoadPlugin(path)
				if err != nil {
					return err
				}
				generators[gen.GetName()] = gen
				return nil
			})

			if err != nil {
				log.Error().Err(err).Msg("failed to walk directory")
				continue
			}
		}

		// print all generator plugin information found
		tbl := table.New("Name", "Version", "Description")
		for _, g := range generators {
			tbl.AddRow(g.GetName(), g.GetVersion(), g.GetDescription())
		}
		if len(generators) > 0 {
			tbl.Print()
		}
	},
}

func init() {
	inspectCmd.Flags().BoolVar(&byTarget, "by-target", false, "set whether to ")
	rootCmd.AddCommand(inspectCmd)
}
