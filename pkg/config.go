package configurator

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Options struct{}

type Target struct {
	PluginPath    string   `yaml:"plugin,omitempty"`
	TemplatePaths []string `yaml:"templates,omitempty"`
	FilePaths     []string `yaml:"files,omitempty"`
	RunTargets    []string `yaml:"targets,omitempty"`
}

type Jwks struct {
	Uri     string `yaml:"uri"`
	Retries int    `yaml:"retries,omitempty"`
}

type Server struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Jwks Jwks   `yaml:"jwks,omitempty"`
}

type Config struct {
	Version     string            `yaml:"version,omitempty"`
	Server      Server            `yaml:"server,omitempty"`
	SmdClient   SmdClient         `yaml:"smd,omitempty"`
	AccessToken string            `yaml:"access-token,omitempty"`
	Targets     map[string]Target `yaml:"targets,omitempty"`
	PluginDirs  []string          `yaml:"plugins,omitempty"`
	CertPath    string            `yaml:"cacert,omitempty"`
	Options     Options           `yaml:"options,omitempty"`
}

// Creates a new config with default parameters.
func NewConfig() Config {
	return Config{
		Version: "",
		SmdClient: SmdClient{
			Host: "http://127.0.0.1",
			Port: 27779,
		},
		Targets: map[string]Target{
			"dnsmasq": Target{
				PluginPath:    "",
				TemplatePaths: []string{},
			},
			"conman": Target{
				PluginPath:    "",
				TemplatePaths: []string{},
			},
			"warewulf": Target{
				PluginPath: "",
				TemplatePaths: []string{
					"templates/warewulf/defaults/node.jinja",
					"templates/warewulf/defaults/provision.jinja",
				},
			},
		},

		PluginDirs: []string{},
		Server: Server{
			Host: "127.0.0.1",
			Port: 3334,
			Jwks: Jwks{
				Uri:     "",
				Retries: 5,
			},
		},
		Options: Options{},
	}
}

func LoadConfig(path string) Config {
	var c Config = NewConfig()
	file, err := os.ReadFile(path)
	if err != nil {
		log.Printf("failed to read config file: %v\n", err)
		return c
	}
	err = yaml.Unmarshal(file, &c)
	if err != nil {
		log.Fatalf("failed to unmarshal config: %v\n", err)
		return c
	}
	return c
}

func (config *Config) SaveConfig(path string) {
	path = filepath.Clean(path)
	if path == "" || path == "." {
		path = "config.yaml"
	}
	data, err := yaml.Marshal(config)
	if err != nil {
		log.Printf("failed to marshal config: %v\n", err)
		return
	}
	err = os.WriteFile(path, data, os.ModePerm)
	if err != nil {
		log.Printf("failed to write default config file: %v\n", err)
		return
	}
}

func SaveDefaultConfig(path string) {
	path = filepath.Clean(path)
	if path == "" || path == "." {
		path = "config.yaml"
	}
	var c = NewConfig()
	data, err := yaml.Marshal(c)
	if err != nil {
		log.Printf("failed to marshal config: %v\n", err)
		return
	}
	err = os.WriteFile(path, data, os.ModePerm)
	if err != nil {
		log.Printf("failed to write default config file: %v\n", err)
		return
	}
}
