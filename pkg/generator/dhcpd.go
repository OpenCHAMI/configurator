package generator

import (
	"fmt"

	configurator "github.com/OpenCHAMI/configurator/pkg"
	"github.com/OpenCHAMI/configurator/pkg/util"
)

type DHCPd struct{}

func (g *DHCPd) GetName() string {
	return "dhcpd"
}

func (g *DHCPd) GetVersion() string {
	return util.GitCommit()
}

func (g *DHCPd) GetDescription() string {
	return fmt.Sprintf("Configurator generator plugin for '%s'.", g.GetName())
}

func (g *DHCPd) Generate(config *configurator.Config, opts ...util.Option) (FileMap, error) {
	var (
		params                                         = GetParams(opts...)
		client                                         = GetClient(params)
		targetKey                                      = params["target"].(string)
		target                                         = config.Targets[targetKey]
		compute_nodes                                  = ""
		eths          []configurator.EthernetInterface = nil
		err           error                            = nil
	)

	//
	if client != nil {
		eths, err = client.FetchEthernetInterfaces(opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch ethernet interfaces with client: %w", err)
		}
	}

	// check if we have the required params first
	if eths == nil {
		return nil, fmt.Errorf("invalid ethernet interfaces (variable is nil)")
	}
	if len(eths) <= 0 {
		return nil, fmt.Errorf("no ethernet interfaces found")
	}

	// format output to write to config file
	compute_nodes = "# ========== DYNAMICALLY GENERATED BY OPENCHAMI CONFIGURATOR ==========\n"
	for _, eth := range eths {
		if len(eth.IpAddresses) == 0 {
			continue
		}
		compute_nodes += fmt.Sprintf("host %s { hardware ethernet %s; fixed-address %s} ", eth.ComponentId, eth.MacAddress, eth.IpAddresses[0])
	}
	compute_nodes += "# ====================================================================="

	if verbose, ok := params["verbose"].(bool); ok {
		if verbose {
			fmt.Printf("")
		}
	}
	return ApplyTemplateFromFiles(Mappings{
		"plugin_name":        g.GetName(),
		"plugin_version":     g.GetVersion(),
		"plugin_description": g.GetDescription(),
		"compute_nodes":      compute_nodes,
		"node_entries":       "",
	}, target.TemplatePaths...)
}