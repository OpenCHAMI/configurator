package generator

import (
	"bytes"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"plugin"

	configurator "github.com/OpenCHAMI/configurator/internal"
	"github.com/OpenCHAMI/configurator/internal/util"
	"github.com/nikolalohinski/gonja/v2"
	"github.com/nikolalohinski/gonja/v2/exec"
	"github.com/sirupsen/logrus"
)

type Mappings map[string]any
type Files map[string][]byte

// Generator interface used to define how files are created. Plugins can
// be created entirely independent of the main driver program.
type Generator interface {
	GetName() string
	GetVersion() string
	GetDescription() string
	Generate(config *configurator.Config, opts ...util.Option) (Files, error)
}

// Params defined and used by the "generate" subcommand.
type Params struct {
	Args        []string
	PluginPaths []string
	Target      string
	Verbose     bool
}

func ConvertContentsToString(f Files) map[string]string {
	n := make(map[string]string, len(f))
	for k, v := range f {
		n[k] = string(v)
	}
	return n
}

// Loads files without applying any Jinja 2 templating.
func LoadFiles(paths ...string) (Files, error) {
	var outputs = Files{}
	for _, path := range paths {
		expandedPaths, err := filepath.Glob(path)
		if err != nil {
			return nil, fmt.Errorf("failed to glob path: %v", err)
		}
		for _, expandedPath := range expandedPaths {
			info, err := os.Stat(expandedPath)
			if err != nil {
				fmt.Println(err)
				return nil, fmt.Errorf("failed to stat file or directory: %v", err)
			}
			// skip any directories found
			if info.IsDir() {
				continue
			}
			b, err := os.ReadFile(expandedPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read file: %v", err)
			}

			outputs[expandedPath] = b
		}
	}

	return outputs, nil
}

// Loads a single generator plugin given a single file path.
func LoadPlugin(path string) (Generator, error) {
	// skip loading plugin if path is a directory with no error
	if isDir, err := util.IsDirectory(path); err == nil && isDir {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to test if path is directory: %v", err)
	}

	// try and open the plugin
	p, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin: %v", err)
	}

	// load the "Generator" symbol from plugin
	symbol, err := p.Lookup("Generator")
	if err != nil {
		return nil, fmt.Errorf("failed to look up symbol at path '%s': %v", path, err)
	}

	// assert that the plugin loaded has a valid generator
	gen, ok := symbol.(Generator)
	if !ok {
		return nil, fmt.Errorf("failed to load the correct symbol type at path '%s'", path)
	}
	return gen, nil
}

// Loads all generator plugins in a given directory.
//
// Returns a map of generators. Each generator can be accessed by the name
// returned by the generator.GetName() implemented.
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
				fmt.Printf("failed to load plugin: %v\n", err)
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

// Option to specify "target" in parameter map. This is used to set which generator
// to use to generate a config file.
func WithTarget(target string) util.Option {
	return func(p util.Params) {
		if p != nil {
			p["target"] = target
		}
	}
}

// Option to specify "type" in parameter map. This is not currently used.
func WithType(_type string) util.Option {
	return func(p util.Params) {
		if p != nil {
			p["type"] = _type
		}
	}
}

// Option to a specific client to include in implementing plugin generator.Generate().
//
// NOTE: This may be changed to pass some kind of client interface as an argument in
// the future instead.
func WithClient(client configurator.SmdClient) util.Option {
	return func(p util.Params) {
		p["client"] = client
	}
}

// Helper function to get client in generator.Generate() plugin implementations.
func GetClient(params util.Params) *configurator.SmdClient {
	return util.Get[configurator.SmdClient](params, "client")
}

// Helper function to get the target in generator.Generate() plugin implementations.
func GetTarget(config *configurator.Config, key string) configurator.Target {
	return config.Targets[key]
}

// Helper function to load all options set with With*() into parameter map.
func GetParams(opts ...util.Option) util.Params {
	params := util.Params{}
	for _, opt := range opts {
		opt(params)
	}
	return params
}

// Wrapper function to slightly abstract away some of the nuances with using gonja
// into a single function call. This function is *mostly* for convenience and
// simplication.
func ApplyTemplates(mappings map[string]any, paths ...string) (Files, error) {
	var (
		data    = exec.NewContext(mappings)
		outputs = Files{}
	)

	for _, path := range paths {
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
		outputs[path] = b.Bytes()
	}

	return outputs, nil
}

// Main function to generate a collection of files as a map with the path as the key and
// the contents of the file as the value. This function currently expects a list of plugin
// paths to load all plugins within a directory. Then, each plugin's generator.Generate()
// function is called for each target specified.
//
// This function is the corresponding implementation for the "generate" CLI subcommand.
// It is also call when running the configurator as a service with the "/generate" route.
//
// TODO: Separate loading plugins so we can load them once when running as a service.
func Generate(config *configurator.Config, params Params) (Files, error) {
	// load generator plugins to generate configs or to print
	var (
		generators = make(map[string]Generator)
		client     = configurator.NewSmdClient(
			configurator.WithHost(config.SmdClient.Host),
			configurator.WithPort(config.SmdClient.Port),
			configurator.WithAccessToken(config.AccessToken),
			configurator.WithSecureTLS(config.CertPath),
		)
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
			WithTarget(gen.GetName()),
			WithClient(client),
		)
	}
	return nil, fmt.Errorf("an unknown error has occurred")
}
