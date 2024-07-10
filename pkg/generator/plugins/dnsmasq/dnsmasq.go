package main

import (
	"fmt"
	"strings"

	configurator "github.com/OpenCHAMI/configurator/pkg"
	"github.com/OpenCHAMI/configurator/pkg/generator"
	"github.com/OpenCHAMI/configurator/pkg/util"
)

type DnsMasq struct{}

func (g *DnsMasq) GetName() string {
	return "dnsmasq"
}

func (g *DnsMasq) GetVersion() string {
	return util.GitCommit()
}

func (g *DnsMasq) GetDescription() string {
	return fmt.Sprintf("Configurator generator plugin for '%s'.", g.GetName())
}

func (g *DnsMasq) Generate(config *configurator.Config, opts ...util.Option) (generator.FileMap, error) {
	// make sure we have a valid config first
	if config == nil {
		return nil, fmt.Errorf("invalid config (config is nil)")
	}

	// set all the defaults for variables
	var (
		params                                     = generator.GetParams(opts...)
		client                                     = generator.GetClient(params)
		targetKey                                  = params["target"].(string) // required param
		target                                     = config.Targets[targetKey]
		eths      []configurator.EthernetInterface = nil
		err       error                            = nil
	)

	// if we have a client, try making the request for the ethernet interfaces
	if client != nil {
		eths, err = client.FetchEthernetInterfaces(opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch ethernet interfaces with client: %v", err)
		}
	}

	// check if we have the required params first
	if eths == nil {
		return nil, fmt.Errorf("invalid ethernet interfaces (variable is nil)")
	}
	if len(eths) <= 0 {
		return nil, fmt.Errorf("no ethernet interfaces found")
	}

	// print message if verbose param found
	if verbose, ok := params["verbose"].(bool); ok {
		if verbose {
			fmt.Printf("template: \n%s\nethernet interfaces found: %v\n", strings.Join(target.TemplatePaths, "\n\t"), len(eths))
		}
	}

	// format output to write to config file
	output := "# ========== DYNAMICALLY GENERATED BY OPENCHAMI CONFIGURATOR ==========\n"
	for _, eth := range eths {
		if eth.Type == "NodeBMC" {
			output += "dhcp-host=" + eth.MacAddress + "," + eth.ComponentId + "," + eth.IpAddresses[0].IpAddress + "\n"
		} else {
			output += "dhcp-host=" + eth.MacAddress + "," + eth.ComponentId + "," + eth.IpAddresses[0].IpAddress + "\n"
		}
	}
	output += "# ====================================================================="

	// apply template substitutions and return output as byte array
	return generator.ApplyTemplateFromFiles(generator.Mappings{
		"plugin_name":        g.GetName(),
		"plugin_version":     g.GetVersion(),
		"plugin_description": g.GetDescription(),
		"dhcp-hosts":         output,
	}, target.TemplatePaths...)
}

var Generator DnsMasq
