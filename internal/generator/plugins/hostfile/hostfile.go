package main

import (
	"fmt"

	configurator "github.com/OpenCHAMI/configurator/internal"
	"github.com/OpenCHAMI/configurator/internal/generator"
	"github.com/OpenCHAMI/configurator/internal/util"
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

func (g *Hostfile) Generate(config *configurator.Config, opts ...util.Option) (generator.Files, error) {
	return nil, fmt.Errorf("plugin does not implement generation function")
}

var Generator Hostfile
