# OpenCHAMI Configurator

The `configurator` (portmanteau of config + generator) is an extensible tool that fetchs data from an instance of [SMD](https://github.com/OpenCHAMI/smd) to generate commonly used config files based on Jinja 2 template files. The tool and generator plugins are written in Go and plugins can be written by following the ["Creating Generator Plugins"](#creating-generator-plugins) section of this README.

## Building and Usage

The `configurator` is built using standard `go` build tools. The project separates the client and server with build tags. To get started, clone the project, download the dependencies, and build the project:

```bash
git clone https://github.com/OpenCHAMI/configurator.git
go mod tidy
go build --tags all # equivalent to `go build --tags client,server``

## ...or just run `make` in project directory
```

This will build the main driver program, but also requires generator plugins to define how new config files are generated. The default plugins can be built using the following build command:

```bash
go build -buildmode=plugin -o lib/conman.so internal/generator/plugins/conman/conman.go
go build -buildmode=plugin -o lib/coredhcp.so internal/generator/plugins/coredhcp/coredhcp.go
go build -buildmode=plugin -o lib/dnsmasq.so internal/generator/plugins/dnsmasq/dnsmasq.go
go build -buildmode=plugin -o lib/powerman.so internal/generator/plugins/powerman/powerman.go
go build -buildmode=plugin -o lib/syslog.so internal/generator/plugins/syslog/syslog.go
```

**NOTE: Not all of the plugins have completed generation implementations and are WIP.**

These commands will build the default plugins and store them in the "lib" directory by default. Alternatively, the plugins can be built using `make plugins` if GNU make is installed and available. After you build the plugins, run the following to use the tool:

```bash
./configurator generate --config config.yaml --target dnsmasq -o dnsmasq.conf
```

This will generate a new `dnsmasq` config file based on the Jinja 2 template specified in the config file for "dnsmasq". The `--target` flag specifies the type of config file to generate by its name (see the [`Creating Generator Plugins`](#creating-generator-plugins) section for details). The `configurator` tool requires a valid access token when making requests to an instance of SMD that has protected routes.

The tool can also run as a service to generate files for clients:

```bash
./configurator serve --config config.yaml
```

Once the server is up and listening for HTTP requests, you can try making a request to it with `curl` or `configurator fetch`. Both commands below are essentially equivalent:

```bash
curl http://127.0.0.1:3334/generate?target=dnsmasq -H "Authorization: Bearer $ACCESS_TOKEN"
# ...or...
./configurator fetch --target dnsmasq --host http://127.0.0.1 --port 3334 
```

This will do the same thing as the `generate` subcommand, but remotely. The access token is only required if the `CONFIGURATOR_JWKS_URL` environment variable is set. The `ACCESS_TOKEN` environment variable passed to `curl` and it's corresponding CLI argument both expects a token as a JWT.

### Creating Generator Plugins

The `configurator` uses generator plugins to define how config files are generated using a `Generator` interface.  The interface is defined like so:

```go
type Files = map[string][]byte
type Generator interface {
  GetName() string
  GetVersion() string
  GetDescription() string
  Generate(config *configurator.Config, opts ...util.Option) (Files, error)
}
```

A new plugin can be created by implementing the methods from interface and exporting a symbol with `Generator` as the name and the plugin struct as the type. The `GetName()` function returns the name that is used for looking up the corresponding template set in your config file. It can also be included in the templated files with the default plugins using the `{{ plugin_name }}` in your template. The `GetVersion()` and `GetDescription()` functions returns the version and description of the plugin which can be included in the templated files using `{{ plugin_version }}` and `{{ plugin_description }}` respectively with the default plugins. The `Generate` function is where the magic happens to build the config file from a template.

```go
package main

type MyGenerator struct {}

func (g *MyGenerator) GetName() string {
  // just an example...this can be done however you want
  pluginInfo := LoadFromFile("path/to/plugin/info.json")
  return pluginInfo["name"]
}

func (g *MyGenerator) GetVersion() string {
  return "v1.0.0"
}

func (g *MyGenerator) GetDescription() string {
  return "This is an example plugin."
}

func (g *MyGenerator) Generate(config *configurator.Config, opts ...util.Option) (generator.Files, error) {
  // do config generation stuff here...
  var (
    params = generator.GetParams(opts...)
    client = generator.GetClient(params)
    output = ""
  )
  if client {
    eths, err := client.FetchEthernetInterfaces(opts...)
    // ... blah, blah, blah, check error, format output, and so on...
  }

  // apply the substitutions to Jinja template and return output as byte array
  return generator.ApplyTemplate(path, generator.Mappings{
    "plugin_name":        g.GetName(),
    "plugin_version":     g.GetVersion(),
    "plugin_description": g.GetDescription(),
    "output": output,
  })
}

// this MUST be named "Generator" for symbol lookup in main driver
var Generator MyGenerator
```

Finally, build the plugin and put it somewhere specified by `plugins` in your config. Make sure that the package is `main` before building.

```bash
go build -buildmode=plugin -o lib/mygenerator.so path/to/mygenerator.go
```

Now your plugin should be available to use with the `configurator` main driver. If you get an error about not loading the correct symbol type, make sure that you generator function definitions match the `Generator` interface exactly.

## Configuration

Here is an example config file to start using configurator:

```yaml
server:         # Server-related parameters when using as service
  host: 127.0.0.1
  port: 3334
  jwks:         # Set the JWKS uri to protect /generate route
    uri: ""
    retries: 5
smd: .          # SMD-related parameters
  host: http://127.0.0.1
  port: 27779
plugins:        # path to plugin directories
  - "lib/"
targets:        # targets to call with --target flag
  dnsmasq:
    templates:
      - templates/dnsmasq.jinja
  warewulf:
    templates:  # files using Jinja templating
      - templates/warewulf/vnfs/dhcpd-template.jinja
      - templates/warewulf/vnfs/dnsmasq-template.jinja
    files:      # files to be copied without templating
      - templates/warewulf/defaults/provision.jinja
      - templates/warewulf/defaults/node.jinja
      - templates/warewulf/filesystem/examples/*
      - templates/warewulf/vnfs/*
      - templates/warewulf/bootstrap.jinja
      - templates/warewulf/database.jinja
    targets:    # additional targets to run 
      - dnsmasq
```

The `server` section sets the properties for running the `configurator` tool as a service and is not required if you're only using the CLI. Also note that the `jwks-uri` parameter is only needs for protecting endpoints. If it is not set, then the API is entirely public. The `smd` section tells the `configurator` tool where to find SMD to pull state management data used by the internal client. The `templates` section is where the paths are mapped to each generator plugin by its name (see the [`Creating Generator Plugins`](#creating-generator-plugins) section for details). The `plugins` is a list of paths to load generator plugins.

## Running the Tests

The `configurator` project includes a collection of tests focused on verifying plugin behavior and generating files. The tests do not currently test fetching information from SMD (or whatever remote source). The tests can be ran with either of the following commands:

```bash
go test ./tests/generate_test.go --tags=all
# ...or alternatively with GNU make...
make tests
```


## Known Issues

- Adds a new `OAuthClient` with every token request
- Plugins are being loaded each time a file is generated

## TODO

- Add group functionality to create by files by groups
- Extend SMD client functionality (or make extensible?)
- Handle authentication with `OAuthClient`'s correctly
