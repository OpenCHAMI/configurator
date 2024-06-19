package main

import (
	configurator "github.com/OpenCHAMI/configurator/internal"
	"github.com/OpenCHAMI/configurator/internal/util"
)

type CoreDhcp struct{}

func (g *CoreDhcp) GetName() string {
	return "coredhcp"
}

func (g *CoreDhcp) GetGroups() []string {
	return []string{"coredhcp"}
}

func (g *CoreDhcp) Generate(config *configurator.Config, opts ...util.Option) ([]byte, error) {
	return nil, nil
}

var Generator CoreDhcp
