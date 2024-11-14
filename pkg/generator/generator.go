package generator

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"plugin"

	configurator "github.com/OpenCHAMI/configurator/pkg"
	"github.com/OpenCHAMI/configurator/pkg/util"
	"github.com/nikolalohinski/gonja/v2"
	"github.com/nikolalohinski/gonja/v2/exec"
	"github.com/rs/zerolog/log"
)

type (
	Mappings map[string]any
	FileMap  map[string][]byte
	FileList [][]byte
	Template []byte

	// Generator interface used to define how files are created. Plugins can
	// be created entirely independent of the main driver program.
	Generator interface {
		GetName() string
		GetVersion() string
		GetDescription() string
		Generate(config *configurator.Config, opts ...util.Option) (FileMap, error)
	}

	// Params defined and used by the "generate" subcommand.
	Params struct {
		Args          []string
		TemplatePaths []string
		PluginPath    string
		Target        string
		Verbose       bool
	}
)

var DefaultGenerators = createDefaultGenerators()

func createDefaultGenerators() map[string]Generator {
	var (
		generatorMap = map[string]Generator{}
		generators   = []Generator{
			&Conman{}, &DHCPd{}, &DNSMasq{}, &Hostfile{},
			&Powerman{}, &Syslog{}, &Warewulf{},
		}
	)
	for _, g := range generators {
		generatorMap[g.GetName()] = g
	}
	return generatorMap
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
		if params.GetVerbose() {
			fmt.Printf("-- found plugin '%s'\n", gen.GetName())
		}

		// map each generator plugin by name for lookup
		generators[gen.GetName()] = gen
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	// items, _ := os.ReadDir(dirpath)
	// for _, item := range items {
	// 	if item.IsDir() {
	// 		subitems, _ := os.ReadDir(item.Name())
	// 		for _, subitem := range subitems {
	// 			if !subitem.IsDir() {
	// 				gen, err := LoadPlugin(subitem.Name())
	// 				if err != nil {
	// 					fmt.Printf("failed to load generator in directory '%s': %v\n", item.Name(), err)
	// 					continue
	// 				}
	// 				if verbose, ok := params["verbose"].(bool); ok {
	// 					if verbose {
	// 						fmt.Printf("-- found plugin '%s'\n", item.Name())
	// 					}
	// 				}
	// 				gens[gen.GetName()] = gen
	// 			}
	// 		}
	// 	} else {
	// 		gen, err := LoadPlugin(dirpath + item.Name())
	// 		if err != nil {
	// 			fmt.Printf("failed to load plugin: %v\n", err)
	// 			continue
	// 		}
	// 		if verbose, ok := params["verbose"].(bool); ok {
	// 			if verbose {
	// 				fmt.Printf("-- found plugin '%s'\n", dirpath+item.Name())
	// 			}
	// 		}
	// 		gens[gen.GetName()] = gen
	// 	}
	// }

	return generators, nil
}

func LoadTemplate(path string) (Template, error) {
	// skip loading template if path is a directory with no error
	if isDir, err := util.IsDirectory(path); err == nil && isDir {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to test if template path is directory: %w", err)
	}

	// try and read the contents of the file
	// NOTE: we don't care if this is actually a Jinja template
	// or not...at least for now.
	return os.ReadFile(path)
}

func LoadTemplates(paths []string, opts ...util.Option) (map[string]Template, error) {
	var (
		templates = make(map[string]Template)
		params    = util.ToDict(opts...)
	)

	for _, path := range paths {
		err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
			// skip trying to load generator plugin if directory or error
			if info.IsDir() || err != nil {
				return nil
			}

			// load the contents of the template
			template, err := LoadTemplate(path)
			if err != nil {
				return fmt.Errorf("failed to load generator in directory '%s': %w", path, err)
			}

			// show the templates loaded if verbose flag is set
			if params.GetVerbose() {
				fmt.Printf("-- loaded tempalte '%s'\n", path)
			}

			// map each template by the path it was loaded from
			templates[path] = template
			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("failed to walk directory: %w", err)
		}
	}

	return templates, nil
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

// Wrapper function to slightly abstract away some of the nuances with using gonja
// into a single function call. This function is *mostly* for convenience and
// simplication. If no paths are supplied, then no templates will be applied and
// there will be no output.
//
// The "FileList" returns a slice of byte arrays in the same order as the argument
// list supplied, but with the Jinja templating applied.
func ApplyTemplates(mappings Mappings, contents ...[]byte) (FileList, error) {
	var (
		data    = exec.NewContext(mappings)
		outputs = FileList{}
	)

	for _, b := range contents {
		// load jinja template from file
		t, err := gonja.FromBytes(b)
		if err != nil {
			return nil, fmt.Errorf("failed to read template from file: %w", err)
		}

		// execute/render jinja template
		b := bytes.Buffer{}
		if err = t.Execute(&b, data); err != nil {
			return nil, fmt.Errorf("failed to execute: %w", err)
		}
		outputs = append(outputs, b.Bytes())
	}

	return outputs, nil
}

// Wrapper function similiar to "ApplyTemplates" but takes file paths as arguments.
// This function will load templates from a file instead of using file contents.
func ApplyTemplateFromFiles(mappings Mappings, paths ...string) (FileMap, error) {
	var (
		data    = exec.NewContext(mappings)
		outputs = FileMap{}
	)

	for _, path := range paths {
		// load jinja template from file
		t, err := gonja.FromFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read template from file: %w", err)
		}

		// execute/render jinja template
		b := bytes.Buffer{}
		if err = t.Execute(&b, data); err != nil {
			return nil, fmt.Errorf("failed to execute: %w", err)
		}
		outputs[path] = b.Bytes()
	}

	return outputs, nil
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
func GenerateWithTarget(config *configurator.Config, params Params) (FileMap, error) {
	// load generator plugins to generate configs or to print
	var (
		client = configurator.NewSmdClient(
			configurator.WithHost(config.SmdClient.Host),
			configurator.WithPort(config.SmdClient.Port),
			configurator.WithAccessToken(config.AccessToken),
			configurator.WithCertPoolFile(config.CertPath),
		)
		target    configurator.Target
		generator Generator
		err       error
		ok        bool
	)

	// check if a target is supplied
	if len(params.Args) == 0 && params.Target == "" {
		return nil, fmt.Errorf("must specify a target")
	}

	// load target information from config
	target, ok = config.Targets[params.Target]
	if !ok {
		return nil, fmt.Errorf("target not found in config")
	}

	// if plugin path specified from CLI, use that instead
	if params.PluginPath != "" {
		target.PluginPath = params.PluginPath
	}

	// check if generator is built-in first before loading
	generator, ok = DefaultGenerators[params.Target]
	if !ok {
		// only load the plugin needed for this target if we don't find default
		log.Error().Msg("did not find target in default generators")
		generator, err = LoadPlugin(target.PluginPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load plugin: %w", err)
		}
	}

	// run the generator plugin from target passed
	return generator.Generate(
		config,
		WithTarget(generator.GetName()),
		WithClient(client),
	)
}
