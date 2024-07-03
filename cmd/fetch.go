//go:build client || all
// +build client all

package cmd

import (
	"fmt"
	"net/http"

	"github.com/OpenCHAMI/configurator/internal/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	accessToken string
	remoteHost  string
	remotePort  int
)

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch a config file from a remote instance of configurator",
	Run: func(cmd *cobra.Command, args []string) {
		// make sure a host is set
		if remoteHost == "" {
			logrus.Errorf("no '--host' argument set")
			return
		}

		headers := map[string]string{}
		if accessToken != "" {
			headers["Authorization"] = "Bearer " + accessToken
		}

		for _, target := range targets {
			// make a request for each target
			url := fmt.Sprintf("%s:%d/generate?target=%s", remoteHost, remotePort, target)
			res, body, err := util.MakeRequest(url, http.MethodGet, nil, headers)
			if err != nil {
				logrus.Errorf("failed to make request: %v", err)
				return
			}
			if res != nil {
				if res.StatusCode == http.StatusOK {
					fmt.Printf("%s\n", string(body))
				}
			}
		}
	},
}

func init() {
	fetchCmd.Flags().StringVar(&remoteHost, "host", "", "set the remote configurator host")
	fetchCmd.Flags().IntVar(&remotePort, "port", 3334, "set the remote configurator port")
	fetchCmd.Flags().StringSliceVar(&targets, "target", nil, "set the target configs to make")
	fetchCmd.Flags().StringVarP(&outputPath, "output", "o", "", "set the output path for config targets")
	fetchCmd.Flags().StringVar(&accessToken, "access-token", "o", "", "set the output path for config targets")

	rootCmd.AddCommand(fetchCmd)
}
