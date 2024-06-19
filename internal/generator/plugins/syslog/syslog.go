package main

import (
	configurator "github.com/OpenCHAMI/configurator/internal"
	"github.com/OpenCHAMI/configurator/internal/util"
)

type Syslog struct{}

func (g *Syslog) GetName() string {
	return "syslog"
}

func (g *Syslog) GetGroups() []string {
	return []string{"log"}
}

func (g *Syslog) Generate(config *configurator.Config, opts ...util.Option) ([]byte, error) {
	return nil, nil
}

var Generator Syslog
