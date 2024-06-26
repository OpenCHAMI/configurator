package main

import (
	"fmt"
	"maps"

	configurator "github.com/OpenCHAMI/configurator/internal"
	"github.com/OpenCHAMI/configurator/internal/generator"
	"github.com/OpenCHAMI/configurator/internal/util"
)

type Warewulf struct{}

func (g *Warewulf) GetName() string {
	return "warewulf"
}

func (g *Warewulf) GetGroups() []string {
	return []string{"warewulf"}
}

func (g *Warewulf) Generate(config *configurator.Config, opts ...util.Option) (generator.Files, error) {
	var (
		params    = generator.GetParams(opts...)
		client    = generator.GetClient(params)
		targetKey = params["target"].(string)
		target    = config.Targets[targetKey]
		outputs   = make(generator.Files, len(target.FilePaths)+len(target.Templates))
	)

	// check if our client is included and is valid
	if client == nil {
		return nil, fmt.Errorf("invalid client (client is nil)")
	}

	// fetch redfish endpoints and handle errors
	eps, err := client.FetchRedfishEndpoints(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch redfish endpoints: %v", err)
	}
	if len(eps) <= 0 {
		return nil, fmt.Errorf("no redfish endpoints found")
	}

	// load files and templates and copy to outputs
	files, err := generator.LoadFiles(target.FilePaths...)
	if err != nil {
		return nil, fmt.Errorf("failed to load files: %v", err)
	}
	templates, err := generator.ApplyTemplates(generator.Mappings{}, target.Templates...)
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

var Generator Warewulf
