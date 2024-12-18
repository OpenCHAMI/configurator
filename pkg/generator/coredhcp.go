package generator

import (
	"fmt"

	"github.com/OpenCHAMI/configurator/pkg/config"
	"github.com/OpenCHAMI/configurator/pkg/util"
)

type CoreDhcp struct{}

func (g *CoreDhcp) GetName() string {
	return "coredhcp"
}

func (g *CoreDhcp) GetVersion() string {
	return util.GitCommit()
}

func (g *CoreDhcp) GetDescription() string {
	return fmt.Sprintf("Configurator generator plugin for '%s' to generate config files. (WIP)", g.GetName())
}

func (g *CoreDhcp) Generate(config *config.Config, params Params) (FileMap, error) {
	return nil, fmt.Errorf("plugin does not implement generation function")
}
