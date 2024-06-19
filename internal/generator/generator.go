package generator

import (
	"bytes"
	"fmt"
	"os"
	"plugin"

	configurator "github.com/OpenCHAMI/configurator/internal"
	"github.com/OpenCHAMI/configurator/internal/util"
	"github.com/nikolalohinski/gonja/v2"
	"github.com/nikolalohinski/gonja/v2/exec"
)

type Mappings = map[string]any
type Generator interface {
	GetName() string
	GetGroups() []string
	Generate(config *configurator.Config, opts ...util.Option) ([]byte, error)
}

func LoadPlugin(path string) (Generator, error) {
	p, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load plugin: %v", err)
	}

	symbol, err := p.Lookup("Generator")
	if err != nil {
		return nil, fmt.Errorf("failed to look up symbol: %v", err)
	}

	gen, ok := symbol.(Generator)
	if !ok {
		return nil, fmt.Errorf("failed to load the correct symbol type")
	}
	return gen, nil
}

func LoadPlugins(dirpath string, opts ...util.Option) (map[string]Generator, error) {
	// check if verbose option is supplied
	var (
		gens   = make(map[string]Generator)
		params = util.GetParams(opts...)
	)

	items, _ := os.ReadDir(dirpath)
	var LoadGenerator = func(path string) (Generator, error) {
		// load each generator plugin
		p, err := plugin.Open(path)
		if err != nil {
			return nil, fmt.Errorf("failed to load plugin: %v", err)
		}

		// lookup symbol in plugin
		symbol, err := p.Lookup("Generator")
		if err != nil {
			return nil, fmt.Errorf("failed to look up symbol: %v", err)
		}

		// assert that the loaded symbol is the correct type
		gen, ok := symbol.(Generator)
		if !ok {
			return nil, fmt.Errorf("failed to load the correct symbol type")
		}
		return gen, nil
	}
	for _, item := range items {
		if item.IsDir() {
			subitems, _ := os.ReadDir(item.Name())
			for _, subitem := range subitems {
				if !subitem.IsDir() {
					gen, err := LoadGenerator(subitem.Name())
					if err != nil {
						fmt.Printf("failed to load generator in directory '%s': %v\n", item.Name(), err)
						continue
					}
					if verbose, ok := params["verbose"].(bool); ok {
						if verbose {
							fmt.Printf("found plugin '%s'\n", item.Name())
						}
					}
					gens[gen.GetName()] = gen
				}
			}
		} else {
			gen, err := LoadGenerator(dirpath + item.Name())
			if err != nil {
				fmt.Printf("failed to load generator: %v\n", err)
				continue
			}
			if verbose, ok := params["verbose"].(bool); ok {
				if verbose {
					fmt.Printf("found plugin '%s'\n", dirpath+item.Name())
				}
			}
			gens[gen.GetName()] = gen
		}
	}

	return gens, nil
}

func WithTemplate(_template string) util.Option {
	return func(p util.Params) {
		if p != nil {
			p["template"] = _template
		}
	}
}

func WithType(_type string) util.Option {
	return func(p util.Params) {
		if p != nil {
			p["type"] = _type
		}
	}
}

func WithClient(client configurator.SmdClient) util.Option {
	return func(p util.Params) {
		p["client"] = client
	}
}

// Syntactic sugar generic function to get parameter from util.Params.
func Get[T any](params util.Params, key string) *T {
	if v, ok := params[key].(T); ok {
		return &v
	}
	return nil
}

// Helper function to get client in generator plugins.
func GetClient(params util.Params) *configurator.SmdClient {
	return Get[configurator.SmdClient](params, "client")
}

func GetParams(opts ...util.Option) util.Params {
	params := util.Params{}
	for _, opt := range opts {
		opt(params)
	}
	return params
}

func Generate(g Generator, config *configurator.Config, opts ...util.Option) {
	g.Generate(config, opts...)
}

func ApplyTemplate(path string, mappings map[string]any) ([]byte, error) {
	data := exec.NewContext(mappings)

	// load jinja template from file
	t, err := gonja.FromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read template from file: %v", err)
	}

	// execute/render jinja template
	b := bytes.Buffer{}
	if err = t.Execute(&b, data); err != nil {
		return nil, fmt.Errorf("failed to execute: %v", err)
	}

	return b.Bytes(), nil
}
