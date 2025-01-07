# OpenCHAMI Configurator

The `configurator` is an extensible tool that is capable of dynamically generating files on the fly. The tool includes a built-in generator that fetchs data from an instance of [SMD](https://github.com/OpenCHAMI/smd) to generate files based on Jinja 2 template files. The tool and generator plugins are written in Go and plugins can be written by following the ["Creating Generator Plugins"](#creating-generator-plugins) section of this README.

## Building and Usage

The `configurator` is built using standard `go` build tools. The project separates the client, server, and generator components using build tags. To get started, clone the project, download the dependencies, and build the project:

```bash
git clone https://github.com/OpenCHAMI/configurator.git
go mod tidy
go build --tags all # equivalent to `go build --tags client,server``
```

This will build the main driver program with the default generators that are found in the `pkg/generators` directory.

> [!WARNING]
> Not all of the plugins have completed generation implementations and are a WIP.

### Running Configurator with CLI

After you build the program, run the following command to use the tool:

```bash
export ACCESS_TOKEN=eyJhbGciOiJIUzI1NiIs...
./configurator generate --config config.yaml --target coredhcp -o coredhcp.conf --cacert ochami.pem
```

This will generate a new `coredhcp` config file based on the Jinja 2 template specified in the config file for "coredhcp". The files will be written to `coredhcp.conf` as specified with the `-o/--output` flag. The `--target` flag specifies the type of config file to generate by its name (see the [`Creating Generator Plugins`](#creating-generator-plugins) section for details).

In other words, there should be an entry in the config file that looks like this:

```yaml
...
targets:
  coredhcp:
    plugin: "lib/coredhcp.so"  # optional, if we want to use an external plugin instead
    templates:
      - templates/coredhcp.j2
...

```

> [!NOTE]
> The `configurator` tool requires a valid access token when making requests to an instance of SMD that has protected routes.

### Running Configurator as a Service

The tool can also run as a service to generate files for clients:

```bash
export CONFIGURATOR_JWKS_URL="http://my.openchami.cluster:8443/key"
./configurator serve --config config.yaml
```

Once the server is up and listening for HTTP requests, you can try making a request to it with `curl` or `configurator fetch`. Both commands below are essentially equivalent:

```bash
export ACCESS_TOKEN=eyJhbGciOiJIUzI1NiIs...
curl http://127.0.0.1:3334/generate?target=coredhcp -X GET -H "Authorization: Bearer $ACCESS_TOKEN" --cacert ochami.pem
# ...or...
./configurator fetch --target coredhcp --host http://127.0.0.1:3334 --cacert ochami.pem
```

This will do the same thing as the `generate` subcommand, but through a GET request where the file contents is returned in the response. The access token is only required if the `CONFIGURATOR_JWKS_URL` environment variable is set when starting the server with `serve`. The `ACCESS_TOKEN` environment variable is passed to `curl` using the `Authorization` header and expects a token as a JWT.

### Docker

New images can be built and tested using the `Dockerfile` provided in the project. However, the binary executable and the generator plugins must first be built before building the image since the Docker build copies the binary over. Therefore, build all of the binaries first by following the first section of ["Building and Usage"](#building-and-usage). Running `make docker` from the Makefile will automate this process. Otherwise, run the `docker build` command after building the executable and libraries.

```bash
docker build -t configurator:testing path/to/configurator/Dockerfile
```

If you want to easily include your own external generator plugins, you can build it and copy the `lib.so` file to `lib/`. Make sure that the `Generator` interface is implemented correctly as described in the ["Creating Generator Plugins"](#creating-generator-plugins) or the plugin will not load (you should get an error that specifically says this). Additionally, the name string returned from the `GetName()` method is used for looking up the plugin with the `--target` flag by the main driver program.

Alternatively, pull the latest existing image/container from the GitHub container repository.

```bash
docker pull ghcr.io/openchami/configurator:latest
```

Then, run the Docker container similarly to running the binary.

```bash
export ACCESS_TOKEN=eyJhbGciOiJIUzI1NiIs...
docker run ghcr.io/openchami/configurator:latest configurator generate --config config.yaml --target coredhcp -o coredhcp.conf --cacert configurator.pem
```

### Creating Generator Plugins

The `configurator` uses built-in and user-defined generators that implement the `Generator` interface to describe how config files should be generated. The interface is defined like so:

```go
// maps the file path to its contents
type FileMap = map[string][]byte

// interface for generator plugins
type Generator interface {
  GetName() string
  GetVersion() string
  GetDescription() string
  Generate(config *configurator.Config, opts ...util.Option) (FileMap, error)
}
```

A new plugin can be created by implementing the methods from interface and exporting a symbol with `Generator` as the name and the plugin struct as the type. The `GetName()` function returns the name that is used for looking up the corresponding target set in your config file. It can also be included in the templated files with the default plugins using the `{{ plugin_name }}` in your template. The `GetVersion()` and `GetDescription()` functions returns the version and description of the plugin which can be included in the templated files using `{{ plugin_version }}` and `{{ plugin_description }}` respectively with the default plugins. The `Generate` function is where the magic happens to build the config file from a template.

```go
package main

type MyGenerator struct {
  PluginInfo map[string]any
}

var pluginInfo map[string]any

// this function is not a part of the `Generator` interface
func (g *MyGenerator) LoadFromFile() map[string]any{ /*...*/ }

func (g *MyGenerator) GetName() string {
  // just an example...this can be done however you want
  g.PluginInfo := LoadFromFile("path/to/plugin/info.json")
  return g.PluginInfo["name"]
}

func (g *MyGenerator) GetVersion() string {
  return g.PluginInfo["version"] // "v1.0.0"
}

func (g *MyGenerator) GetDescription() string {
  return g.PluginInfo["description"] // "This is an example plugin."
}

func (g *MyGenerator) Generate(config *configurator.Config, opts ...util.Option) (generator.FileMap, error) {
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

  // apply the substitutions to Jinja template and return output as FileMap (i.e. path and it's contents)
  return generator.ApplyTemplate(path, generator.Mappings{
    "plugin_name":        g.GetName(),
    "plugin_version":     g.GetVersion(),
    "plugin_description": g.GetDescription(),
    "output": output,
  })
}

> [!NOTE]
> The keys in `generator.ApplyTemplate` must not contain illegal characters such as a `-` or else the templates will not apply correctly.


// this MUST be named "Generator" for symbol lookup in main driver
var Generator MyGenerator
```

Finally, build the plugin and put it somewhere specified by `plugins` in your config. Make sure that the package is `main` before building.

```bash
go build -buildmode=plugin -o lib/mygenerator.so path/to/mygenerator.go
```

Now your plugin should be available to use with the `configurator` main driver program. If you get an error about not loading the correct symbol type, make sure that your generator function definitions match the `Generator` interface entirely and that you don't have a partially implemented interface.

## Configuration

Here is an example config file to start using configurator:

```yaml
server:         # Server-related parameters when using as service
  host: 127.0.0.1
  port: 3334
  jwks:         # Set the JWKS uri for protected routes
    uri: ""
    retries: 5
smd:            # SMD-related parameters
  host: http://127.0.0.1:27779
plugins:        # path to plugin directories
  - "lib/"
targets:        # targets to call with --target flag
  coredhcp:
    templates:
      - templates/coredhcp.j2
    files:      # files to be copied without templating
      - extra/nodes.conf
    targets:    # additional targets to run (does not run recursively)
      - dnsmasq
```

The `server` section sets the properties for running the `configurator` tool as a service and is not required if you're only using the CLI. Also note that the `jwks.uri` parameter is only needed for protecting endpoints. If it is not set, then all API routes are entirely public. The `smd` section tells the `configurator` tool where to find the SMD service to pull state management data used internally by the client's generator. The `templates` section is where the paths are mapped to each generator by its name (see the [`Creating Generator Plugins`](#creating-generator-plugins) section for details). The `plugins` is a list of paths to search for and load external generator plugins.

## Running the Tests

The `configurator` project includes a collection of tests focused on verifying plugin behavior and generating files. The tests do not include fetching information from any remote sources, can be ran with the following command:

```bash
go test ./tests/generate_test.go --tags=all
```

## Known Issues

- Adds a new `OAuthClient` with every token request
- Plugins are being loaded each time a file is generated

## TODO

- Add group functionality to create by files by groups
- Extend SMD client functionality (or make extensible?)
- Handle authentication with `OAuthClient`'s correctly
