package generator

import (
	"fmt"
	"maps"
	"strings"

	configurator "github.com/OpenCHAMI/configurator/pkg"
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

func (g *Warewulf) Generate(config *configurator.Config, opts ...util.Option) (FileMap, error) {
	var (
		params    = GetParams(opts...)
		client    = GetClient(params)
		targetKey = params["target"].(string)
		target    = config.Targets[targetKey]
		outputs   = make(FileMap, len(target.FilePaths)+len(target.TemplatePaths))
	)

	// check if our client is included and is valid
	if client == nil {
		return nil, fmt.Errorf("invalid client (client is nil)")
	}

	// if we have a client, try making the request for the ethernet interfaces
	eths, err := client.FetchEthernetInterfaces(opts...)
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

	// print message if verbose param found
	if verbose, ok := params["verbose"].(bool); ok {
		if verbose {
			fmt.Printf("template: \n%s\n ethernet interfaces found: %v\n", strings.Join(target.TemplatePaths, "\n\t"), len(eths))
		}
	}

	// fetch redfish endpoints and handle errors
	eps, err := client.FetchRedfishEndpoints(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch redfish endpoints: %v", err)
	}
	if len(eps) <= 0 {
		return nil, fmt.Errorf("no redfish endpoints found")
	}

	// format output for template substitution
	nodeEntries := ""

	// load files and templates and copy to outputs
	files, err := LoadFiles(target.FilePaths...)
	if err != nil {
		return nil, fmt.Errorf("failed to load files: %v", err)
	}
	templates, err := ApplyTemplateFromFiles(Mappings{
		"node_entries": nodeEntries,
	}, target.TemplatePaths...)
	if err != nil {
		return nil, fmt.Errorf("failed to load templates: %v", err)
	}

	maps.Copy(outputs, files)
	maps.Copy(outputs, templates)

	// print message if verbose param is found
	if verbose, ok := params["verbose"].(bool); ok {
		if verbose {
			fmt.Printf("templates and files loaded: \n")
			for path, _ := range outputs {
				fmt.Printf("\t%s", path)
			}
		}
	}

	return outputs, err
}
