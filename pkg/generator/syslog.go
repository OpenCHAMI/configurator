package generator

import (
	"fmt"

	configurator "github.com/OpenCHAMI/configurator/pkg"
	"github.com/OpenCHAMI/configurator/pkg/util"
)

type Syslog struct{}

func (g *Syslog) GetName() string {
	return "syslog"
}

func (g *Syslog) GetVersion() string {
	return util.GitCommit()
}

func (g *Syslog) GetDescription() string {
	return fmt.Sprintf("Configurator generator plugin for '%s'.", g.GetName())
}

func (g *Syslog) Generate(config *configurator.Config, opts ...util.Option) (FileMap, error) {
	return nil, fmt.Errorf("plugin does not implement generation function")
}
