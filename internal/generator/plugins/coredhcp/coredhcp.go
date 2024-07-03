package main

import (
	"fmt"

	configurator "github.com/OpenCHAMI/configurator/internal"
	"github.com/OpenCHAMI/configurator/internal/util"
)

type CoreDhcp struct{}

func (g *CoreDhcp) GetName() string {
	return "coredhcp"
}

func (g *CoreDhcp) GetVersion() string {
	return util.GitCommit()
}

func (g *CoreDhcp) GetDescription() string {
	return fmt.Sprintf("Configurator generator plugin for '%s' to generate config files. This plugin is not complete and still a WIP.", g.GetName())
}

func (g *CoreDhcp) Generate(config *configurator.Config, opts ...util.Option) (map[string][]byte, error) {
	return nil, fmt.Errorf("plugin does not implement generation function")
}

var Generator CoreDhcp
