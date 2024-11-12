package main

import (
	"fmt"

	configurator "github.com/OpenCHAMI/configurator/pkg"
	"github.com/OpenCHAMI/configurator/pkg/generator"
	"github.com/OpenCHAMI/configurator/pkg/util"
	"github.com/mitchellh/mapstructure"
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
		params                                       = generator.GetParams(opts...)
		client                                       = generator.GetClient(params)
		target                                       = generator.GetTarget(config, params["target"].(string))
		ethHost     string                           = configurator.GetIPAddress(params)
		external    []configurator.EthernetInterface = nil
		internal    []configurator.EthernetInterface = nil
		ethExternal []map[string]any                 = nil
		ethInternal []map[string]any                 = nil
		// hostEth  *configurator.EthernetInterface  = nil
		err error = nil
	)

	if ethHost != "" {
		client.FetchEthernetInterfaces(
			configurator.WithIPAddress(ethHost),
		)
		_ = ethHost
	}

	// if we have a client, try making the request for the ethernet interfaces
	if client != nil {
		fmt.Printf("host: %s:%d\n", client.Host, client.Port)
		// try and get all external ethernet interfaces
		external, err = client.FetchEthernetInterfaces(
			configurator.WithNetwork("External"),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch external ethernet interfaces with client: %v", err)
		}
		// try and get all internal ethernet interfaces
		internal, err = client.FetchEthernetInterfaces(
			configurator.WithNetwork("Internal"),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch internal ethernet interfaces with client: %v", err)
		}
		// TODO: try and get host ethernet interface using hostname

	}

	// we can *maybe* get the {{ host.*_netmask }} from the IP address + CIDR
	// otherwise, assume /24

	err = mapstructure.Decode(internal, &ethExternal)
	if err != nil {
		log.Error().Err(err).Msg("failed to decode internal interfaces")
	}

	err = mapstructure.Decode(external, &ethInternal)
	if err != nil {
		log.Error().Err(err).Msg("failed to decode external interfaces")
	}

	// try to decode subnet mask from IP address if in CIDR notation

	// create a shim layer
	return generator.ApplyTemplateFromFiles(generator.Mappings{
		"host.name":               "", // ??
		"host.internal_interface": "", // get from cloud-init??
		"host.internal_subnet":    "", // get from cloud-init??
		"host.internal_netmask":   "", // ??
		"host.internal_ip":        "", // ??
		"host.external_interface": "", // get from cloud-init??
		"host.external_subnet":    "", // get from cloud-init??
		"host.external_netmask":   "", // ??
		"host.external_ip":        "", // ??
		"host":                    "",
		"ext_ethernet_interfaces": ethExternal,
		"int_ethernet_interfaces": ethInternal,
	}, target.TemplatePaths...)
}

var Generator Nodes
