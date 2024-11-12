package generator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"plugin"

	configurator "github.com/OpenCHAMI/configurator/pkg"
	"github.com/OpenCHAMI/configurator/pkg/util"
)

type Mappings map[string]any
type FileMap map[string][]byte
type FileList [][]byte
type Template []byte

// Generator interface used to define how files are created. Plugins can
// be created entirely independent of the main driver program.
type Generator interface {
	GetName() string
	GetVersion() string
	GetDescription() string
	Generate(config *configurator.Config, opts ...util.Option) (FileMap, error)
}

// Params defined and used by the "generate" subcommand.
// TODO: It may make more sense to just pass this to 'Generate()' instead of using
// the functional options pattern.
type Params struct {
	Args          []string
	Host          string
	Port          int
	Generators    map[string]Generator
	TemplatePaths []string
	PluginPath    string
	PluginArgs    map[string]string
	Target        string
	Verbose       bool
}

// Converts the file outputs from map[string][]byte to map[string]string.
func ConvertContentsToString(f FileMap) map[string]string {
	n := make(map[string]string, len(f))
	for k, v := range f {
		n[k] = string(v)
	}
	return n
}

// Loads files without applying any Jinja 2 templating.
func LoadFiles(paths ...string) (FileMap, error) {
	var outputs = FileMap{}
	for _, path := range paths {
		expandedPaths, err := filepath.Glob(path)
		if err != nil {
			return nil, fmt.Errorf("failed to glob path: %w", err)
		}
		for _, expandedPath := range expandedPaths {
			info, err := os.Stat(expandedPath)
			if err != nil {
				fmt.Println(err)
				return nil, fmt.Errorf("failed to stat file or directory: %w", err)
			}
			// skip any directories found
			if info.IsDir() {
				continue
			}
			b, err := os.ReadFile(expandedPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read file: %w", err)
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
		return nil, fmt.Errorf("failed to test if plugin path is directory: %w", err)
	}

	// try and open the plugin
	p, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin: %w", err)
	}

	// load the "Generator" symbol from plugin
	symbol, err := p.Lookup("Generator")
	if err != nil {
		return nil, fmt.Errorf("failed to look up symbol at path '%s': %w", path, err)
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
		generators = make(map[string]Generator)
		params     = util.ToDict(opts...)
	)

	//
	err := filepath.Walk(dirpath, func(path string, info fs.FileInfo, err error) error {
		// skip trying to load generator plugin if directory or error
		if info.IsDir() || err != nil {
			return nil
		}

		// load the generator plugin from current path
		gen, err := LoadPlugin(path)
		if err != nil {
			return fmt.Errorf("failed to load generator in directory '%s': %w", path, err)
		}

		// show the plugins found if verbose flag is set
		if util.GetVerbose(params) {
			fmt.Printf("-- found plugin '%s'\n", gen.GetName())
		}

		// map each generator plugin by name for lookup
		generators[gen.GetName()] = gen
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return generators, nil
}

func With(key string, value any) util.Option {
	return func(p util.Params) {
		if p != nil {
			p[key] = value
		}
	}
}

// Option to specify "target" in parameter map. This is used to set which generator
// to use to generate a config file.
func WithTarget(target string) util.Option {
	return With("target", target)
}

func WithArgs(args []string) util.Option {
	return With("args", args)
}

func WithPluginArgs(pluginArgs map[string]string) util.Option {
	return With("plugin-args", pluginArgs)
}

// Option to specify "type" in parameter map. This is not currently used.
func WithType(_type string) util.Option {
	return With("type", _type)
}

// Option to the plugin to load
func WithPlugin(path string) util.Option {
	return func(p util.Params) {
		if p != nil {
			plugin, err := LoadPlugin(path)
			if err != nil {
				return
			}
			p["plugin"] = plugin
		}
	}
}

func WithTemplates(paths []string) util.Option {
	return func(p util.Params) {
		if p != nil {
			templates, err := LoadTemplates(paths)
			if err != nil {

			}
			p["templates"] = templates
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

func GetArgs(params util.Params) []string {
	args := util.Get[[]string](params, "args")
	if args != nil {
		return *args
	}
	return nil
}

func GetPluginArgs(params util.Params) map[string]string {
	pluginArgs := util.Get[map[string]string](params, "plugin-args")
	if pluginArgs != nil {
		return *pluginArgs
	}
	return nil
}

// Generate() is the main function to generate a collection of files and returns them as a map.
// This function only expects a path to a plugin and paths to a collection of templates to
// be used. This function will only load the plugin on-demand and fetch resources as needed.
func Generate(config *configurator.Config, params Params) (FileMap, error) {
	var (
		gen    Generator
		client = configurator.NewSmdClient()
	)

	return gen.Generate(
		config,
		WithPlugin(params.PluginPath),
		WithTemplates(params.TemplatePaths),
		WithClient(client),
	)
}

// Main function to generate a collection of files as a map with the path as the key and
// the contents of the file as the value. This function currently expects a list of plugin
// paths to load all plugins within a directory. Then, each plugin's generator.GenerateWithTarget()
// function is called for each target specified.
//
// This function is the corresponding implementation for the "generate" CLI subcommand.
// It is also call when running the configurator as a service with the "/generate" route.
//
// TODO: Separate loading plugins so we can load them once when running as a service.
// TODO: Handle host/port arguments from CLI flags correctly
func GenerateWithTarget(config *configurator.Config, params Params) (FileMap, error) {
	// load generator plugins to generate configs or to print
	var (
		client = configurator.NewSmdClient(
			configurator.WithClientHost(config.SmdClient.Host),
			configurator.WithClientPort(config.SmdClient.Port),
			configurator.WithClientAccessToken(config.AccessToken),
			configurator.WithCertPoolFile(config.CertPath),
		)
	)

	// check if a target is supplied
	if len(params.Args) == 0 && params.Target == "" {
		return nil, fmt.Errorf("must specify a target")
	}

	// load target information from config
	target, ok := config.Targets[params.Target]
	if !ok {
		return nil, fmt.Errorf("target not found in config")
	}

	// if plugin path specified from CLI, use that instead
	if params.PluginPath != "" {
		target.PluginPath = params.PluginPath
	}

	// only load the plugin needed for this target
	generator, err := LoadPlugin(target.PluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load plugin: %w", err)
	}

	// run the generator plugin from target passed
	return generator.Generate(
		config,
		WithTarget(generator.GetName()),
		WithClient(client),
		WithPluginArgs(params.PluginArgs),
		WithArgs(params.Args),
	)
}
