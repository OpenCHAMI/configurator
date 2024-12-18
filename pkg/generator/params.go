package generator

import (
	configurator "github.com/OpenCHAMI/configurator/pkg"
	"github.com/OpenCHAMI/configurator/pkg/client"
	"github.com/OpenCHAMI/configurator/pkg/config"
)

type (
	// Params used by the generator
	Params struct {
		Templates  map[string]Template
		Files      map[string][]byte
		ClientOpts []client.Option
		Verbose    bool
	}
	Option func(Params)
)

func ToParams(opts ...Option) Params {
	params := Params{}
	for _, opt := range opts {
		opt(params)
	}
	return params
}

func WithClientOpts(opts ...client.Option) Option {
	return func(p Params) {
		p.ClientOpts = opts
	}
}

func WithTemplates(templates map[string]Template) Option {
	return func(p Params) {
		p.Templates = templates
	}
}

// Helper function to get the target in generator.Generate() plugin implementations.
func GetTarget(config *config.Config, key string) configurator.Target {
	return config.Targets[key]
}
