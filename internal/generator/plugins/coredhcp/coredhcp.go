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

func (g *CoreDhcp) Generate(config *configurator.Config, opts ...util.Option) (map[string][]byte, error) {
	return nil, fmt.Errorf("plugin does not implement generation function")
}

var Generator CoreDhcp
