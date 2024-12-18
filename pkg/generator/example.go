//go:build example || plugins
// +build example plugins

package generator

import (
	"fmt"

	configurator "github.com/OpenCHAMI/configurator/pkg"
	"github.com/OpenCHAMI/configurator/pkg/generator"
	"github.com/OpenCHAMI/configurator/pkg/util"
)

type Example struct {
	Message string
}

func (g *Example) GetName() string {
	return "example"
}

func (g *Example) GetVersion() string {
	return util.GitCommit()
}

func (g *Example) GetDescription() string {
	return fmt.Sprintf("Configurator generator plugin for '%s'.", g.GetName())
}

func (g *Example) Generate(config *configurator.Config, opts ...util.Option) (generator.FileMap, error) {
	g.Message = `
	This is an example generator plugin. See the file in 'internal/generator/plugins/example/example.go' on
	information about constructing plugins and plugin requirements.`
	return generator.FileMap{"example": []byte(g.Message)}, nil
}

var Generator Example
