package generator

import (
	"fmt"

	"github.com/OpenCHAMI/configurator/pkg/config"
	"github.com/OpenCHAMI/configurator/pkg/util"
)

type Hostfile struct{}

func (g *Hostfile) GetName() string {
	return "hostfile"
}

func (g *Hostfile) GetVersion() string {
	return util.GitCommit()
}

func (g *Hostfile) GetDescription() string {
	return fmt.Sprintf("Configurator generator plugin for '%s'.", g.GetName())
}

func (g *Hostfile) Generate(config *config.Config, opts ...Option) (FileMap, error) {
	return nil, fmt.Errorf("plugin does not implement generation function")
}
