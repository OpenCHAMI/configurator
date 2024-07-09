package tests

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"testing"

	"command-line-arguments/Users/allend/Desktop/projects/ochami/configurator/internal/server/server.go"

	configurator "github.com/OpenCHAMI/configurator/internal"
	"github.com/OpenCHAMI/configurator/internal/generator"
	"github.com/OpenCHAMI/configurator/internal/util"
	"github.com/docker/docker/api/server"
)

// A valid test generator that implements the `Generator` interface.
type TestGenerator struct{}

func (g *TestGenerator) GetName() string    { return "test" }
func (g *TestGenerator) GetVersion() string { return "v1.0.0" }
func (g *TestGenerator) GetDescription() string {
	return "This is a plugin creating for running tests."
}
func (g *TestGenerator) Generate(config *configurator.Config, opts ...util.Option) (generator.FileMap, error) {
	// Jinja 2 template file
	files := [][]byte{
		[]byte(`
Name:        {{plugin_name}}
Version:     {{plugin_version}}
Description: {{plugin_description}}

This is the first test template file.
		`),
		[]byte(`
This is another testing Jinja 2 template file using {{plugin_name}}.
		`),
	}

	// apply Jinja templates to file
	fileList, err := generator.ApplyTemplates(generator.Mappings{
		"plugin_name":        g.GetName(),
		"plugin_version":     g.GetVersion(),
		"plugin_description": g.GetDescription(),
	}, files...)
	if err != nil {
		return nil, fmt.Errorf("failed to apply templates: %v", err)
	}

	// make sure we're able to receive certain arguments when passed
	params := generator.GetParams(opts...)
	if len(params) <= 0 {
		return nil, fmt.Errorf("expect at least one params, but found none")
	}

	// TODO: make sure we can get the target

	// make sure we're able to get the client as well
	client := generator.GetClient(params)
	if client == nil {
		return nil, fmt.Errorf("invalid client (client is nil)")
	}

	// make sure we have the same number of files in file list
	if len(files) != len(fileList) {
		return nil, fmt.Errorf("file list output count is not the same as the input")
	}

	// convert file list to file map
	fileMap := make(generator.FileMap, len(fileList))
	for i, contents := range fileList {
		fileMap[fmt.Sprintf("t-%d.txt", i)] = contents
	}

	return fileMap, nil
}

// An invalid generator that does not or partially implements
// the `Generator` interface.
type InvalidGenerator struct{}

// Test building and loading plugins
func TestPlugin(t *testing.T) {
	var (
		tempDir    = t.TempDir()
		testPlugin = []byte(`
package main

type TestGenerator struct{}

func (g *TestGenerator) GetName() string        { return "test" }
func (g *TestGenerator) GetVersion() string     { return "v1.0.0" }
func (g *TestGenerator) GetDescription() string { return "This is a plugin creating for running tests." }
func (g *TestGenerator) Generate(config *configurator.Config, opts ...util.Option) (generator.FileMap, error) {
	return generator.FileMap{"test": []byte("test")}, nil
}
var Generator TestGenerator
		`)
		testPluginPath = fmt.Sprintf("%s/testplugin.go")
	)

	// build a test plugin
	t.Run("build", func(t *testing.T) {
		// dump the plugin source code to a file
		err := os.WriteFile(testPluginPath, testPlugin, os.ModePerm)
		if err != nil {
			t.Fatalf("failed to write test plugin file: %v", err)
		}

		// execute command to build the plugin
		cmd := exec.Command("go", "build", "-buildmode=plugin", fmt.Sprintf("-o=%s/test.so", tempDir))
		if cmd.Err != nil {
			t.Fatalf("failed to build plugin: %v", cmd.Err)
		}

		// stat the file to confirm that it was built
		fileInfo, err := os.Stat(testPluginPath)
		if err != nil {
			t.Fatalf("failed to stat plugin file: %v", err)
		}
		if fileInfo.IsDir() {
			t.Fatalf("directory file but a file was expected")
		}
		if fileInfo.Size() <= 0 {
			t.Fatal("found an empty file or file with size of 0 bytes")
		}

	})

	// test loading plugins both individually and in a dir
	t.Run("load", func(t *testing.T) {
		gen, err := generator.LoadPlugin(testPluginPath)
		if err != nil {
			t.Fatalf("failed to load the test plugin: %v", err)
		}

		// test that we have all expected methods with type assertions
		if _, ok := gen.(interface {
			GetName() string
			GetVersion() string
			GetDescription() string
			Generate(*configurator.Config, ...util.Option) (generator.FileMap, error)
		}); !ok {
			t.Error("plugin does not implement all of the generator interface")
		}

		// test loading plugins from a directory (should just load a single one)
		gens, err := generator.LoadPlugins(tempDir)
		if err != nil {
			t.Fatalf("failed to load plugins in '%s': %v", tempDir, err)
		}

		// test all of the plugins loaded from a directory (should expect same result as above)
		for _, gen := range gens {
			if _, ok := gen.(interface {
				GetName() string
				GetVersion() string
				GetDescription() string
				Generate(*configurator.Config, ...util.Option) (generator.FileMap, error)
			}); !ok {
				t.Error("plugin does not implement all of the generator interface")
			}
		}
	})

}

// Test that expects to successfully "generate" a file using the built-in
// example plugin with no fetching.
//
// NOTE: Normally we would dynamically load a generator from a plugin, but
// we're not doing it here since that's not what is being tested.
func TestGenerateExample(t *testing.T) {
	var (
		config = configurator.NewConfig()
		client = configurator.NewSmdClient()
		gen    = TestGenerator{}
	)

	// make sure our generator returns expected strings
	t.Run("properties", func(t *testing.T) {
		if gen.GetName() != "test" {
			t.Error("test generator return unexpected name")
		}
		if gen.GetVersion() != "v1.0.0" {
			t.Error("test generator return unexpected version")
		}
		if gen.GetDescription() != "test" {
			t.Error("test generator return unexpected description")
		}
	})

	// try to generate a file with templating applied
	fileMap, err := gen.Generate(
		&config,
		generator.WithTarget("test"),
		generator.WithClient(client),
	)
	if err != nil {
		t.Fatal("failed to generate file: %v", err)
	}

	// test for 2 expected files to be generated in the output (hint: check the
	// TestGenerator.Generate implementation)
	if len(fileMap) != 2 {
		t.Error("expected 2 files in generated output")
	}
}

// Test that expects to successfully "generate" a file using the built-in
// example plugin but by making a HTTP request to a service instance instead.
//
// NOTE: This test uses the default server settings to run.
func TestGenerateExampleWithServer(t *testing.T) {
	var (
		config  = configurator.NewConfig()
		gen     = TestGenerator{}
		headers = make(map[string]string, 0)
	)

	// add a test target to config
	config.Targets["test"] = configurator.Target{}

	// start up a new server in background
	server := server.New(config)
	go server.Server()

	// make request to server to generate a file
	res, b, err := util.MakeRequest("http://127.0.0.1:3334/generate?target=test", http.MethodGet, nil, headers)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expect status code 200 from response but received %d instead", res.StatusCode)
	}

	// test for specific output from request
	fileMap, err := gen.Generate(&config)
	if err != nil {
		t.Fatalf("failed to generate file: %v", err)
	}
	if string(fileMap["test"]) != string(b) {
		t.Fatal("response does not match expected output")
	}

}

// Test that expects to fail with a specific error: "failed to loook up
// symbol at path"
func TestGenerateWithNoSymbol(t *testing.T) {
}

// Test that expects to fail with a specific error: "failed to load the
// correct symbol type at path"
func TestGenerateWithInvalidGenerator() {
}
