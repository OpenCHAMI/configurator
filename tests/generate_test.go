package tests

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	configurator "github.com/OpenCHAMI/configurator/pkg"
	"github.com/OpenCHAMI/configurator/pkg/config"
	"github.com/OpenCHAMI/configurator/pkg/generator"
	"github.com/OpenCHAMI/configurator/pkg/server"
	"github.com/OpenCHAMI/configurator/pkg/util"
)

var (
	workDir    string
	replaceDir string
	err        error
)

// A valid test generator that implements the `Generator` interface.
type TestGenerator struct{}

func (g *TestGenerator) GetName() string    { return "test" }
func (g *TestGenerator) GetVersion() string { return "v1.0.0" }
func (g *TestGenerator) GetDescription() string {
	return "This is a plugin created for running tests."
}
func (g *TestGenerator) Generate(config *config.Config, params generator.Params) (generator.FileMap, error) {
	// Jinja 2 template file
	files := map[string]generator.Template{
		"test1": generator.Template{
			Contents: []byte(`
Name:        {{plugin_name}}
Version:     {{plugin_version}}
Description: {{plugin_description}}

This is the first test template file.
			`),
		},
		"test2": generator.Template{
			Contents: []byte(`
This is another testing Jinja 2 template file using {{plugin_name}}.
			`),
		},
	}

	// apply Jinja templates to file
	fileMap, err := generator.ApplyTemplates(generator.Mappings{
		"plugin_name":        g.GetName(),
		"plugin_version":     g.GetVersion(),
		"plugin_description": g.GetDescription(),
	}, files)
	if err != nil {
		return nil, fmt.Errorf("failed to apply templates: %v", err)
	}

	// make sure we have a valid config we can access
	if config == nil {
		return nil, fmt.Errorf("invalid config (config is nil)")
	}

	// TODO: make sure we can get a target

	// make sure we have the same number of files in file list
	var (
		fileInputCount  = len(files)
		fileOutputCount = len(fileMap)
	)
	if fileInputCount != fileOutputCount {
		return nil, fmt.Errorf("file output count (%d) is not the same as the input (%d)", fileOutputCount, fileInputCount)
	}

	return fileMap, nil
}

func init() {
	workDir, err = os.Getwd()
	if err != nil {
		log.Fatalf("failed to get working directory: %v", err)
	}
	replaceDir = fmt.Sprintf("%s", filepath.Dir(workDir))
}

// Test building and loading plugins
func TestPlugin(t *testing.T) {
	var (
		testPluginDir        = t.TempDir()
		testPluginPath       = fmt.Sprintf("%s/test-plugin.so", testPluginDir)
		testPluginSourcePath = fmt.Sprintf("%s/test-plugin.go", testPluginDir)
		testPluginSource     = []byte(
			`package main

import (
	"github.com/OpenCHAMI/configurator/pkg/config"
	"github.com/OpenCHAMI/configurator/pkg/generator"
)

type TestGenerator struct{}

func (g *TestGenerator) GetName() string    { return "test" }
func (g *TestGenerator) GetVersion() string { return "v1.0.0" }
func (g *TestGenerator) GetDescription() string {
	return "This is a plugin creating for running tests."
}
func (g *TestGenerator) Generate(config *config.Config, params generator.Params) (generator.FileMap, error) {
	return generator.FileMap{"test": []byte("test")}, nil
}

var Generator TestGenerator`)
	)

	// get directory to replace remote pkg with local
	// _, filename, _, _ := runtime.Caller(0)
	// replaceDir := fmt.Sprintf("%s", filepath.Dir(workDir))

	// show all paths to make sure we're using the correct ones
	fmt.Printf("(TestPlugin) working directory:     %v\n", workDir)
	fmt.Printf("(TestPlugin) plugin directory:      %v\n", testPluginDir)
	fmt.Printf("(TestPlugin) plugin path:           %v\n", testPluginPath)
	fmt.Printf("(TestPlugin) plugin source path:    %v\n", testPluginSourcePath)

	// make temporary directory to test plugin
	err = os.MkdirAll(testPluginDir, 0o777)
	if err != nil {
		t.Fatalf("failed to make temporary directory: %v", err)
	}

	// dump the plugin source code to a file
	err = os.WriteFile(testPluginSourcePath, testPluginSource, os.ModePerm)
	if err != nil {
		t.Fatalf("failed to write test plugin file: %v", err)
	}

	// make sure the source file was actually written
	fileInfo, err := os.Stat(testPluginSourcePath)
	if err != nil {
		t.Fatalf("failed to stat path: %v", err)
	}
	if fileInfo.IsDir() {
		t.Fatalf("expected file but found directory")
	}

	// change to testing directory to run command
	err = os.Chdir(testPluginDir)
	if err != nil {
		t.Fatalf("failed to 'cd' to temporary directory: %v", err)
	}

	// initialize the plugin directory as a Go project
	cmd := exec.Command("bash", "-c", "go mod init github.com/OpenCHAMI/configurator-test-plugin")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to execute command: %v\n%s", err, string(output))
	}

	// use the local `pkg` instead of the release one
	cmd = exec.Command("bash", "-c", fmt.Sprintf("go mod edit -replace=github.com/OpenCHAMI/configurator=%s", replaceDir))
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to execute command: %v\n%s", err, string(output))
	}

	// run `go mod tidy` for dependencies
	cmd = exec.Command("bash", "-c", "go mod tidy")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to execute command: %v\n%s", err, string(output))
	}

	// execute command to build the plugin
	cmd = exec.Command("bash", "-c", "go build -buildmode=plugin -o=test-plugin.so test-plugin.go")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to execute command: %v\n%s", err, string(output))
	}

	// stat the file to confirm that the plugin was built
	fileInfo, err = os.Stat(testPluginPath)
	if err != nil {
		t.Fatalf("failed to stat plugin file: %v", err)
	}
	if fileInfo.IsDir() {
		t.Fatalf("directory file but a file was expected")
	}
	if fileInfo.Size() <= 0 {
		t.Fatal("found an empty file or file with size of 0 bytes")
	}

	// test loading plugins both individually and in a dir
	gen, err := generator.LoadPlugin("test-plugin.so")
	if err != nil {
		t.Fatalf("failed to load the test plugin: %v", err)
	}

	// test that we have all expected methods with type assertions
	if _, ok := gen.(interface {
		GetName() string
		GetVersion() string
		GetDescription() string
		Generate(*config.Config, generator.Params) (generator.FileMap, error)
	}); !ok {
		t.Error("plugin does not implement all of the generator interface")
	}

	// test loading plugins from a directory (should just load a single one)
	gens, err := generator.LoadPlugins(testPluginDir)
	if err != nil {
		t.Fatalf("failed to load plugins in '%s': %v", testPluginDir, err)
	}

	// test all of the plugins loaded from a directory (should expect same result as above)
	for _, gen := range gens {
		if _, ok := gen.(interface {
			GetName() string
			GetVersion() string
			GetDescription() string
			Generate(*config.Config, generator.Params) (generator.FileMap, error)
		}); !ok {
			t.Error("plugin does not implement all of the generator interface")
		}
	}

}

// Test that expects to fail with a specific error using a partially
// implemented generator. The purpose of this test is to make sure we're
// seeing the correct error that we would expect in these situations.
// The errors should be something like:
//   - no symbol:      "failed to look up symbol at path"
//   - invalid symbol: "failed to load the correct symbol type at path"
func TestPluginWithInvalidOrNoSymbol(t *testing.T) {
	var (
		testPluginDir        = t.TempDir()
		testPluginPath       = fmt.Sprintf("%s/invalid-plugin.so", testPluginDir)
		testPluginSourcePath = fmt.Sprintf("%s/invalid-plugin.go", testPluginDir)
		testPluginSource     = []byte(`
package main

// An invalid generator that does not or partially implements
// the "Generator" interface.
type InvalidGenerator struct{}
func (g *InvalidGenerator) GetName() string { return "invalid" }
var Generator InvalidGenerator
		`)
	)

	// show all paths to make sure we're using the correct ones
	fmt.Printf("(TestPluginWithInvalidOrNoSymbol) working directory:     %v\n", workDir)
	fmt.Printf("(TestPluginWithInvalidOrNoSymbol) plugin directory:      %v\n", testPluginDir)
	fmt.Printf("(TestPluginWithInvalidOrNoSymbol) plugin path:           %v\n", testPluginPath)
	fmt.Printf("(TestPluginWithInvalidOrNoSymbol) plugin source path:    %v\n", testPluginSourcePath)

	// get directory to replace remote pkg with local
	// _, filename, _, _ := runtime.Caller(0)

	// make temporary directory to test plugin
	err = os.MkdirAll(testPluginDir, os.ModeDir)
	if err != nil {
		t.Fatalf("failed to make temporary directory: %v", err)
	}

	// dump the plugin source code to a file
	err = os.WriteFile(testPluginSourcePath, testPluginSource, os.ModePerm)
	if err != nil {
		t.Fatalf("failed to write test plugin file: %v", err)
	}

	// make sure the source file was actually written
	fileInfo, err := os.Stat(testPluginSourcePath)
	if err != nil {
		t.Fatalf("failed to stat path: %v", err)
	}
	if fileInfo.IsDir() {
		t.Fatalf("expected file but found directory")
	}

	// change to testing directory to run command
	err = os.Chdir(testPluginDir)
	if err != nil {
		t.Fatalf("failed to 'cd' to temporary directory: %v", err)
	}

	// initialize the plugin directory as a Go project
	cmd := exec.Command("bash", "-c", "go mod init github.com/OpenCHAMI/configurator-test-plugin")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to execute command: %v\n%s", err, string(output))
	}

	// use the local `pkg` instead of the release one
	cmd = exec.Command("bash", "-c", fmt.Sprintf("go mod edit -replace=github.com/OpenCHAMI/configurator=%s", replaceDir))
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to execute command: %v\n%s", err, string(output))
	}

	// run `go mod tidy` for dependencies
	cmd = exec.Command("bash", "-c", "go mod tidy")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to execute command: %v\n%s", err, string(output))
	}

	// execute command to build the plugin
	cmd = exec.Command("bash", "-c", "go build -buildmode=plugin -o=invalid-plugin.so invalid-plugin.go")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to execute command: %v\n%s", err, string(output))
	}

	// stat the file to confirm that the plugin was built
	fileInfo, err = os.Stat(testPluginPath)
	if err != nil {
		t.Fatalf("failed to stat plugin file: %v", err)
	}
	if fileInfo.IsDir() {
		t.Fatalf("directory file but a file was expected")
	}
	if fileInfo.Size() <= 0 {
		t.Fatal("found an empty file or file with size of 0 bytes")
	}

	// try and load plugin, but expect specific error
	_, err = generator.LoadPlugin(testPluginSourcePath)
	if err == nil {
		t.Fatalf("expected an error, but returned nil")
	}
}

// Test that expects to successfully "generate" a file using the built-in
// example plugin with no fetching.
//
// NOTE: Normally we would dynamically load a generator from a plugin, but
// we're not doing it here since that's not what is being tested.
func TestGenerateExample(t *testing.T) {
	var (
		conf = config.New()
		gen  = TestGenerator{}
	)

	// make sure our generator returns expected strings
	t.Run("properties", func(t *testing.T) {
		if gen.GetName() != "test" {
			t.Error("test generator return unexpected name")
		}
		if gen.GetVersion() != "v1.0.0" {
			t.Error("test generator return unexpected version")
		}
		if gen.GetDescription() != "This is a plugin created for running tests." {
			t.Error("test generator return unexpected description")
		}
	})

	// try to generate a file with templating applied
	fileMap, err := gen.Generate(&conf, generator.Params{})
	if err != nil {
		t.Fatalf("failed to generate file: %v", err)
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
// NOTE: This test uses the default server settings to run. Also, no need to
// try and load the plugin from a lib here either.
func TestGenerateExampleWithServer(t *testing.T) {
	var (
		conf    = config.New()
		gen     = TestGenerator{}
		headers = make(map[string]string, 0)
	)

	// NOTE: Currently, the server needs a config to know where to get load plugins,
	// and how to handle targets/templates. This will be simplified in the future to
	// decoupled the server from required a config altogether.
	conf.Targets["test"] = configurator.Target{
		TemplatePaths: []string{},
		FilePaths:     []string{},
	}

	// create new server, add test generator, and start in background
	server := server.New(&conf)
	generator.DefaultGenerators["test"] = &gen
	go server.Serve()

	// make request to server to generate a file
	res, b, err := util.MakeRequest("http://127.0.0.1:3334/generate?target=test", http.MethodGet, nil, headers)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expect status code 200 from response but received %d instead", res.StatusCode)
	}

	// test for specific output from request
	//
	// NOTE: we don't actually use the config in this plugin implementation,
	// but we do check that a valid config was passed.
	fileMap, err := gen.Generate(&conf, generator.Params{})
	if err != nil {
		t.Fatalf("failed to generate file: %v", err)
	}
	for path, contents := range fileMap {
		tmp := make(map[string]string, 1)
		err := json.Unmarshal(b, &tmp)
		if err != nil {
			t.Errorf("failed to unmarshal response: %v", err)
			continue
		}
		if string(contents) != string(tmp[path]) {
			t.Fatalf("response does not match expected output...\nexpected:%s\noutput:%s", string(contents), string(tmp[path]))
		}
	}
}
