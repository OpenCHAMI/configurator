package generator

import (
	"bytes"
	"fmt"
	"maps"
	"os"
	"plugin"

	configurator "github.com/OpenCHAMI/configurator/internal"
	"github.com/OpenCHAMI/configurator/internal/util"
	"github.com/nikolalohinski/gonja/v2"
	"github.com/nikolalohinski/gonja/v2/exec"
	"github.com/sirupsen/logrus"
)

type Mappings = map[string]any
type Generator interface {
	GetName() string
	GetGroups() []string
	Generate(config *configurator.Config, opts ...util.Option) ([]byte, error)
}

type Params struct {
	Args        []string
	PluginPaths []string
	Target      string
	Verbose     bool
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
	for _, item := range items {
		if item.IsDir() {
			subitems, _ := os.ReadDir(item.Name())
			for _, subitem := range subitems {
				if !subitem.IsDir() {
					gen, err := LoadPlugin(subitem.Name())
					if err != nil {
						fmt.Printf("failed to load generator in directory '%s': %v\n", item.Name(), err)
						continue
					}
					if verbose, ok := params["verbose"].(bool); ok {
						if verbose {
							fmt.Printf("-- found plugin '%s'\n", item.Name())
						}
					}
					gens[gen.GetName()] = gen
				}
			}
		} else {
			gen, err := LoadPlugin(dirpath + item.Name())
			if err != nil {
				fmt.Printf("failed to load generator: %v\n", err)
				continue
			}
			if verbose, ok := params["verbose"].(bool); ok {
				if verbose {
					fmt.Printf("-- found plugin '%s'\n", dirpath+item.Name())
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

func WithOption(key string, value any) util.Option {
	return func(p util.Params) {
		p[key] = value
	}
}

// Helper function to get client in generator plugins.
func GetClient(params util.Params) *configurator.SmdClient {
	return util.Get[configurator.SmdClient](params, "client")
}

func GetParams(opts ...util.Option) util.Params {
	params := util.Params{}
	for _, opt := range opts {
		opt(params)
	}
	return params
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

func Generate(config *configurator.Config, params Params) ([]byte, error) {
	// load generator plugins to generate configs or to print
	var (
		generators = make(map[string]Generator)
		client     = configurator.SmdClient{
			Host:        config.SmdClient.Host,
			Port:        config.SmdClient.Port,
			AccessToken: config.AccessToken,
		}
	)

	// load all plugins from params
	for _, path := range params.PluginPaths {
		if params.Verbose {
			fmt.Printf("loading plugins from '%s'\n", path)
		}
		gens, err := LoadPlugins(path)
		if err != nil {
			fmt.Printf("failed to load plugins: %v\n", err)
			err = nil
			continue
		}

		// add loaded generator plugins to set
		maps.Copy(generators, gens)
	}

	// show available targets then exit
	if len(params.Args) == 0 && params.Target == "" {
		for g := range generators {
			fmt.Printf("-- found generator plugin \"%s\"\n", g)
		}
		return nil, nil
	}

	if params.Target == "" {
		logrus.Errorf("no target supplied (--target name)")
	} else {
		// run the generator plugin from target passed
		gen := generators[params.Target]
		if gen == nil {
			return nil, fmt.Errorf("invalid generator target (%s)", params.Target)
		}
		return gen.Generate(
			config,
			WithTemplate(gen.GetName()),
			WithClient(client),
		)
	}
	return nil, fmt.Errorf("an unknown error has occurred")
}
