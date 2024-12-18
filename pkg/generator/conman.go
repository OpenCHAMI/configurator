package generator

import (
	"fmt"

	configurator "github.com/OpenCHAMI/configurator/pkg"
	"github.com/OpenCHAMI/configurator/pkg/util"
)

type Conman struct{}

func (g *Conman) GetName() string {
	return "conman"
}

func (g *Conman) GetVersion() string {
	return util.GitCommit()
}

func (g *Conman) GetDescription() string {
	return fmt.Sprintf("Configurator generator plugin for '%s'.", g.GetName())
}

func (g *Conman) Generate(config *configurator.Config, opts ...util.Option) (FileMap, error) {
	var (
		params                                   = GetParams(opts...)
		client                                   = GetClient(params)
		targetKey                                = params["target"].(string) // required param
		target                                   = config.Targets[targetKey]
		eps       []configurator.RedfishEndpoint = nil
		err       error                          = nil
		// serverOpts = ""
		// globalOpts = ""
		consoles = ""
	)

	// fetch required data from SMD to create config
	if client != nil {
		eps, err = client.FetchRedfishEndpoints(opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch redfish endpoints with client: %v", err)
		}
	}

	// add any additional conman or server opts
	// if extraOpts, ok := params["opts"].(map[string]any); ok {

	// }

	// format output to write to config file
	consoles = "# ========== DYNAMICALLY GENERATED BY OPENCHAMI CONFIGURATOR ==========\n"
	for _, ep := range eps {
		consoles += fmt.Sprintf("CONSOLE name=%s dev=ipmi:%s-bmc ipmiopts=U:%s,P:%s,W:solpayloadsize\n", ep.Name, ep.Name, ep.User, ep.Password)
	}
	consoles += "# ====================================================================="

	// apply template substitutions and return output as byte array
	return ApplyTemplateFromFiles(Mappings{
		"plugin_name":        g.GetName(),
		"plugin_version":     g.GetVersion(),
		"plugin_description": g.GetDescription(),
		"server_opts":        "",
		"global_opts":        "",
		"consoles":           consoles,
	}, target.TemplatePaths...)
}
