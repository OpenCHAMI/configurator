package configurator

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Options struct{}

type Jwks struct {
	Uri     string `yaml:"uri"`
	Retries int    `yaml:"retries"`
}

type Server struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Jwks Jwks   `yaml:"jwks"`
}

type Config struct {
	Version       string            `yaml:"version"`
	Server        Server            `yaml:"server"`
	SmdClient     SmdClient         `yaml:"smd"`
	AccessToken   string            `yaml:"access-token"`
	TemplatePaths map[string]string `yaml:"templates"`
	Plugins       []string          `yaml:"plugins"`
	Options       Options           `yaml:"options"`
}

func NewConfig() Config {
	return Config{
		Version: "",
		SmdClient: SmdClient{
			Host: "http://127.0.0.1",
			Port: 27779,
		},
		TemplatePaths: map[string]string{
			"dnsmasq":  "templates/dnsmasq.jinja",
			"syslog":   "templates/syslog.jinja",
			"ansible":  "templates/ansible.jinja",
			"powerman": "templates/powerman.jinja",
			"conman":   "templates/conman.jinja",
		},
		Plugins: []string{},
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
