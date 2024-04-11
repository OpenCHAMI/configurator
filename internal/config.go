package configurator

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Options struct {
	JwksUri     string `yaml:"jwks-uri"`
	JwksRetries int    `yaml:"jwks-retries"`
}

type Server struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}
type Config struct {
	Version       string            `yaml:"version"`
	SmdHost       string            `yaml:"smd-host"`
	SmdPort       int               `yaml:"smd-port"`
	AccessToken   string            `yaml:"access-token"`
	TemplatePaths map[string]string `yaml:"templates"`
	Server        Server            `yaml:"server"`
	Options       Options           `yaml:"options"`
}

func NewConfig() Config {
	return Config{
		Version: "",
		SmdHost: "http://127.0.0.1",
		SmdPort: 27779,
		TemplatePaths: map[string]string{
			"dnsmasq":  "templates/dhcp/dnsmasq.conf",
			"syslog":   "templates/syslog/",
			"ansible":  "templates/ansible",
			"powerman": "templates/powerman",
			"conman":   "templates/conman",
		},
		Server: Server{
			Host: "127.0.0.1",
			Port: 3334,
		},
		Options: Options{
			JwksUri:     "",
			JwksRetries: 5,
		},
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
