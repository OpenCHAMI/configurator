package main

import (
	configurator "github.com/OpenCHAMI/configurator/internal"
	"github.com/OpenCHAMI/configurator/internal/util"
)

type Powerman struct{}

func (g *Powerman) GetName() string {
	return "powerman"
}

func (g *Powerman) GetGroups() []string {
	return []string{"powerman"}
}

func (g *Powerman) Generate(config *configurator.Config, opts ...util.Option) ([]byte, error) {
	return nil, nil
}

var Generator Powerman
