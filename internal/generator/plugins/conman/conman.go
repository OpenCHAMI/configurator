package main

import (
	"bytes"
	"fmt"

	configurator "github.com/OpenCHAMI/configurator/internal"
	"github.com/OpenCHAMI/configurator/internal/generator"
	"github.com/OpenCHAMI/configurator/internal/util"
	"github.com/nikolalohinski/gonja/v2"
	"github.com/nikolalohinski/gonja/v2/exec"
)

type Conman struct{}

func (g *Conman) GetName() string {
	return "conman"
}

func (g *Conman) GetGroups() []string {
	return []string{"conman"}
}

func (g *Conman) Generate(config *configurator.Config, opts ...util.Option) ([]byte, error) {
	params := generator.GetParams(opts...)
	var (
		template = params["template"].(string)
		path     = config.TemplatePaths[template]
	)
	data := exec.NewContext(map[string]any{})

	t, err := gonja.FromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read template from file: %v", err)
	}
	output := "# ========== GENERATED BY OCHAMI CONFIGURATOR ==========\n"
	output += "# ======================================================"
	b := bytes.Buffer{}
	if err = t.Execute(&b, data); err != nil {
		return nil, fmt.Errorf("failed to execute: %v", err)
	}

	return b.Bytes(), nil
}

var Generator Conman
