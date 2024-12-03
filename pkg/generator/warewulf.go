package generator

import (
	"fmt"
	"maps"

	"github.com/OpenCHAMI/configurator/pkg/client"
	"github.com/OpenCHAMI/configurator/pkg/config"
	"github.com/OpenCHAMI/configurator/pkg/util"
)

type Warewulf struct{}

func (g *Warewulf) GetName() string {
	return "warewulf"
}

func (g *Warewulf) GetVersion() string {
	return util.GitCommit()
}

func (g *Warewulf) GetDescription() string {
	return "Configurator generator plugin for 'warewulf' config files."
}

func (g *Warewulf) Generate(config *config.Config, params Params) (FileMap, error) {
	var (
		smdClient   = client.NewSmdClient(params.ClientOpts...)
		outputs     = make(FileMap, len(params.Templates))
		nodeEntries = ""
	)

	// if we have a client, try making the request for the ethernet interfaces
	eths, err := smdClient.FetchEthernetInterfaces(params.Verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ethernet interfaces with client: %v", err)
	}

	// check if we have the required params first
	if eths == nil {
		return nil, fmt.Errorf("invalid ethernet interfaces (variable is nil)")
	}
	if len(eths) <= 0 {
		return nil, fmt.Errorf("no ethernet interfaces found")
	}

	// fetch redfish endpoints and handle errors
	eps, err := smdClient.FetchRedfishEndpoints(params.Verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch redfish endpoints: %v", err)
	}
	if len(eps) <= 0 {
		return nil, fmt.Errorf("no redfish endpoints found")
	}

	templates, err := ApplyTemplates(Mappings{
		"node_entries": nodeEntries,
	}, params.Templates)
	if err != nil {
		return nil, fmt.Errorf("failed to load templates: %v", err)
	}

	maps.Copy(outputs, params.Files)
	maps.Copy(outputs, templates)

	// print message if verbose param is found
	if params.Verbose {
		fmt.Printf("templates and files loaded: \n")
		for path, _ := range outputs {
			fmt.Printf("\t%s", path)
		}
	}

	return outputs, err
}
