package main

import (
	"fmt"

	configurator "github.com/OpenCHAMI/configurator/internal"
	"github.com/OpenCHAMI/configurator/internal/generator"
	"github.com/OpenCHAMI/configurator/internal/util"
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

func (g *Powerman) Generate(config *configurator.Config, opts ...util.Option) (generator.FileMap, error) {
	return nil, fmt.Errorf("plugin does not implement generation function")
}

var Generator Powerman
