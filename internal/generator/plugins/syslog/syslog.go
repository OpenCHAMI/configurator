package main

import (
	"fmt"

	configurator "github.com/OpenCHAMI/configurator/internal"
	"github.com/OpenCHAMI/configurator/internal/generator"
	"github.com/OpenCHAMI/configurator/internal/util"
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

func (g *Syslog) Generate(config *configurator.Config, opts ...util.Option) (generator.Files, error) {
	return nil, fmt.Errorf("plugin does not implement generation function")
}

var Generator Syslog
