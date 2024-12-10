package config

import (
	"log"
	"os"
	"path/filepath"

	configurator "github.com/OpenCHAMI/configurator/pkg"
	"github.com/OpenCHAMI/configurator/pkg/client"
	"gopkg.in/yaml.v2"
)

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
	Version     string                         `yaml:"version,omitempty"`
	Server      Server                         `yaml:"server,omitempty"`
	SmdClient   client.SmdClient               `yaml:"smd,omitempty"`
	AccessToken string                         `yaml:"access-token,omitempty"`
	Targets     map[string]configurator.Target `yaml:"targets,omitempty"`
	PluginDirs  []string                       `yaml:"plugins,omitempty"`
	CertPath    string                         `yaml:"cacert,omitempty"`
}

// Creates a new config with default parameters.
func New() Config {
	return Config{
		Version:   "",
		SmdClient: client.SmdClient{Host: "http://127.0.0.1:27779"},
		Targets: map[string]configurator.Target{
			"dnsmasq": configurator.Target{
				Plugin:        "",
				TemplatePaths: []string{},
			},
			"conman": configurator.Target{
				Plugin:        "",
				TemplatePaths: []string{},
			},
			"warewulf": configurator.Target{
				Plugin: "",
				TemplatePaths: []string{
					"templates/warewulf/defaults/node.jinja",
					"templates/warewulf/defaults/provision.jinja",
				},
			},
		},

		PluginDirs: []string{},
		Server: Server{
			Host: "127.0.0.1:3334",
			Jwks: Jwks{
				Uri:     "",
				Retries: 5,
			},
		},
	}
}

func Load(path string) Config {
	var c Config = New()
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

func (config *Config) Save(path string) {
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

func SaveDefault(path string) {
	path = filepath.Clean(path)
	if path == "" || path == "." {
		path = "config.yaml"
	}
	var c = New()
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
