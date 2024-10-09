package main

import (
	"encoding/json"
	"fmt"

	configurator "github.com/OpenCHAMI/configurator/pkg"
	"github.com/OpenCHAMI/configurator/pkg/generator"
	"github.com/OpenCHAMI/configurator/pkg/util"
	"github.com/rs/zerolog/log"
)

type Nodes struct{}

func (g *Nodes) GetName() string {
	return "nodes"
}

func (g *Nodes) GetVersion() string {
	return util.GitCommit()
}

func (g *Nodes) GetDescription() string {
	return "Plugin for generating 'ospfd.conf' and 'zebra.conf' with hostnames."
}

func (g *Nodes) Generate(config *configurator.Config, opts ...util.Option) (generator.FileMap, error) {
	var (
		params = generator.GetParams(opts...)
		// args                                       = generator.GetArgs(params)
		client    = generator.GetClient(params)
		targetKey = params["target"].(string) // required param
		target    = config.Targets[targetKey]
		// eps       []configurator.RedfishEndpoint   = nil
		external []configurator.EthernetInterface = nil
		internal []configurator.EthernetInterface = nil
		err      error                            = nil
	)

	// if we have a client, try making the request for the ethernet interfaces
	if client != nil {
		fmt.Printf("host: %s:%d\n", client.Host, client.Port)
		external, err = client.FetchEthernetInterfaces(
			configurator.WithNetwork("External"),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch external ethernet interfaces with client: %v", err)
		}
	}

	for _, eth := range external {
		b, err := json.MarshalIndent(eth, "", "  ")
		if err != nil {
			log.Error().Err(err).Msg("failed to unmarshal ethernet interface")
			continue
		}
		fmt.Printf("%s\n", string(b))
	}

	if client != nil {
		internal, err = client.FetchEthernetInterfaces(
			configurator.WithNetwork("Internal"),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch internal ethernet interfaces with client: %v", err)
		}
	}

	for _, eth := range internal {
		b, err := json.MarshalIndent(eth, "", "  ")
		if err != nil {
			log.Error().Err(err).Msg("failed to unmarshal ethernet interface")
			continue
		}
		fmt.Printf("%s\n", string(b))
	}

	// we should only expect a single ex

	// we can *maybe* get the {{ host.*_netmask }} from the IP address + CIDR
	// otherwise, assume /24

	// what is the {{ host.*_subnet }} value?

	_ = external
	_ = internal

	// create a shim layer
	return generator.ApplyTemplateFromFiles(generator.Mappings{
		"host.name":               "", // ??
		"host.internal_interface": "", // get from SMD components ethernet interfaces
		"host.internal_subnet":    "", // get from cloud-init
		"host.internal_netmask":   "", // ??
		"host.internal_ip":        "", // ??
		"host.external_interface": "", // get from cloud-init
		"host.external_subnet":    "", // get from cloud-init
		"host.external_netmask":   "", // ??
		"host.external_ip":        "", // ??
	}, target.TemplatePaths...)
}

var Generator Nodes
