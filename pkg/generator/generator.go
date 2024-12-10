package generator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"plugin"

	configurator "github.com/OpenCHAMI/configurator/pkg"
	"github.com/OpenCHAMI/configurator/pkg/client"
	"github.com/OpenCHAMI/configurator/pkg/config"
	"github.com/OpenCHAMI/configurator/pkg/util"
	"github.com/rs/zerolog/log"
)

type (
	Mappings map[string]any
	FileMap  map[string][]byte
	FileList [][]byte

	// Generator interface used to define how files are created. Plugins can
	// be created entirely independent of the main driver program.
	Generator interface {
		GetName() string
		GetVersion() string
		GetDescription() string
		Generate(config *config.Config, params Params) (FileMap, error)
	}
)

var DefaultGenerators = createDefaultGenerators()

func createDefaultGenerators() map[string]Generator {
	var (
		generatorMap = map[string]Generator{}
		generators   = []Generator{
			&Conman{}, &DHCPd{}, &DNSMasq{}, &Warewulf{}, &Example{},
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
func LoadPlugins(dirpath string, opts ...Option) (map[string]Generator, error) {
	// check if verbose option is supplied
	var (
		generators = make(map[string]Generator)
		params     = ToParams(opts...)
	)

	//
	err := filepath.Walk(dirpath, func(path string, info fs.FileInfo, err error) error {
		// skip trying to load generator plugin if directory or error
		if info.IsDir() || err != nil {
			return nil
		}

		// only try loading if file has .so extension
		if filepath.Ext(path) != ".so" {
			return nil
		}

		// load the generator plugin from current path
		gen, err := LoadPlugin(path)
		if err != nil {
			return fmt.Errorf("failed to load generator in directory '%s': %w", path, err)
		}

		// show the plugins found if verbose flag is set
		if params.Verbose {
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

// Generate() is the main function to generate a collection of files and returns them as a map.
// This function only expects a path to a plugin and paths to a collection of templates to
// be used. This function will only load the plugin on-demand and fetch resources as needed.
//
// This function requires that a target and plugin path be set at minimum.
func Generate(plugin string, params Params) (FileMap, error) {
	var (
		generator Generator
		ok        bool
		err       error
	)

	// check if generator is built-in first before loading external plugin
	generator, ok = DefaultGenerators[plugin]
	if !ok {
		// only load the plugin needed for this target if we don't find default
		log.Error().Msg("could not find target in default generators")
		generator, err = LoadPlugin(plugin)
		if err != nil {
			return nil, fmt.Errorf("failed to load plugin from file: %v", err)
		}
	}

	return generator.Generate(nil, params)
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
func GenerateWithTarget(config *config.Config, target string) (FileMap, error) {
	// load generator plugins to generate configs or to print
	var (
		opts       []client.Option
		targetInfo configurator.Target
		generator  Generator
		params     Params
		err        error
		ok         bool
	)

	// check if a target is supplied
	if target == "" {
		return nil, fmt.Errorf("must specify a target")
	}

	// load target information from config
	targetInfo, ok = config.Targets[target]
	if !ok {
		log.Warn().Msg("target not found in config")
	}

	// if no plugin supplied in config target, then using the target supplied
	if targetInfo.Plugin == "" {
		targetInfo.Plugin = target
	}

	// check if generator is built-in first before loading
	generator, ok = DefaultGenerators[target]
	if !ok {
		// only load the plugin needed for this target if we don't find default
		log.Error().Msg("could not find target in default generators")
		generator, err = LoadPlugin(targetInfo.Plugin)
		if err != nil {
			return nil, fmt.Errorf("failed to load plugin: %v", err)
		}
	}

	// prepare params to pass into generator
	params.Templates = map[string]Template{}
	for _, templatePath := range targetInfo.TemplatePaths {
		template := Template{}
		template.LoadFromFile(templatePath)
		params.Templates[templatePath] = template
	}

	// set the client options
	if config.AccessToken != "" {
		params.ClientOpts = append(opts, client.WithAccessToken(config.AccessToken))
	}
	if config.CertPath != "" {
		params.ClientOpts = append(opts, client.WithCertPoolFile(config.CertPath))
	}

	// load files that are not to be copied
	params.Files, err = LoadFiles(targetInfo.FilePaths...)
	if err != nil {
		return nil, fmt.Errorf("failed to load files to copy: %v", err)
	}

	// run the generator plugin from target passed
	return generator.Generate(config, params)
}
