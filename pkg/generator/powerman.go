package generator

import (
	"fmt"

	"github.com/OpenCHAMI/configurator/pkg/config"
	"github.com/OpenCHAMI/configurator/pkg/util"
)

type Powerman struct{}

func (g *Powerman) GetName() string {
	return "powerman"
}

func (g *Powerman) GetVersion() string {
	return util.GitCommit()
}

func (g *Powerman) GetDescription() string {
	return fmt.Sprintf("Configurator generator plugin for '%s'.", g.GetName())
}

func (g *Powerman) Generate(config *config.Config, opts ...Option) (FileMap, error) {
	return nil, fmt.Errorf("plugin does not implement generation function")
}
