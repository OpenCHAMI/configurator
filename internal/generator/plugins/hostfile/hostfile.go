package main

import (
	"fmt"

	configurator "github.com/OpenCHAMI/configurator/internal"
	"github.com/OpenCHAMI/configurator/internal/util"
)

type Hostfile struct{}

func (g *Hostfile) GetName() string {
	return "hostfile"
}

func (g *Hostfile) Generate(config *configurator.Config, opts ...util.Option) (map[string][]byte, error) {
	return nil, fmt.Errorf("plugin does not implement generation function")
}

var Generator Hostfile
