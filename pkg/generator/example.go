package generator

import (
	"fmt"

	"github.com/OpenCHAMI/configurator/pkg/config"
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

func (g *Example) Generate(config *config.Config, params Params) (FileMap, error) {
	g.Message = `
	This is an example generator plugin. See the file in 'internal/generator/plugins/example/example.go' on
	information about constructing plugins and plugin requirements.`
	return FileMap{"example": []byte(g.Message)}, nil
}
