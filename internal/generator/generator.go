package generator

import (
	"os"

	"fmt"

	configurator "github.com/OpenCHAMI/configurator/internal"
	"github.com/nikolalohinski/gonja/v2"
	"github.com/nikolalohinski/gonja/v2/exec"
)

type Generator struct {
}

func New() *Generator {
	return &Generator{}
}

func (g *Generator) GenerateDNS(config *configurator.Config) {
	// generate file using jinja template
	// TODO: load template file for DNS
	// TODO: substitute DNS data fetched from SMD
	// TODO: print generated config file to STDOUT
}

func (g *Generator) GenerateDHCP(config *configurator.Config, target string, eths []configurator.EthernetInterface) error {
	// generate file using gonja template
	// TODO: load template file for DHCP
	path := config.TemplatePaths[target]
	fmt.Printf("path: %s\neth count: %v\n", path, len(eths))
	t, err := gonja.FromFile(config.TemplatePaths[target])
	if err != nil {
		return fmt.Errorf("failed to read template from file: %v", err)
	}
	template := "# ========== GENERATED BY OCHAMI CONFIGURATOR ==========\n"
	for _, eth := range eths {
		if eth.Type == "NodeBMC" {
			template += "dhcp-host=" + eth.MacAddress + "," + eth.ComponentId + "," + eth.IpAddresses[0].IpAddress + "\n"
		} else {
			template += "dhcp-host=" + eth.MacAddress + "," + eth.ComponentId + "," + eth.IpAddresses[0].IpAddress + "\n"
		}
	}
	template += "# ======================================================"
	data := exec.NewContext(map[string]any{
		"hosts": template,
	})
	if err = t.Execute(os.Stdout, data); err != nil {
		return fmt.Errorf("failed to execute: %v", err)
	}

	return nil
}
