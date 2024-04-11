# OpenCHAMI Configurator

Configurator is a tool that fetchs data from an instance of [SMD](https://github.com/OpenCHAMI/smd) to generate commonly used config files.

## Building and Usage

Configurator is built using standard `go` build tools. The project separates the client and server with build tags. To get started, clone the project, download the dependencies, and build the project:

```bash
git clone https://github.com/OpenCHAMI/configurator.git
go mod tidy
go build --tags all # equivalent to `go build --tags client,server``
```

To use the tool, run the following:

```bash
./configurator generate --config config.yaml --target dhcp:dnsmasq
```

This will generate a new DHCP `dnsmasq` config file based on the Jinja 2 template specified in the config file for "dnsmasq". The `--target` flag is set by passing an argument in the form of "type:template" to specify the type of config file being generate and the template file to use respectively. The configurator requires valid access token when making requests to an instance of SMD that has protected routes.

The tool can also be ran as a microservice:

```bash
./configurator serve --config config.yaml
```

Once the server is up and listening for HTTP requests, you can try making a request to it with curl:

```bash
curl http://127.0.0.1:3334/target?type=dhcp&template=dnsmasq
```

This will do the same thing as the `generate` subcommand, but remotely.

## Configuration

Here is an example config file to start using configurator:

```yaml
version: ""
smd-host: http://127.0.0.1
smd-port: 27779
access-token: 
templates:
  coredhcp: templates/dhcp/coredhcp.config.jinja
  dnsmasq: templates/dhcp/dnsmasq.conf.jinja
  syslog: templates/syslog.jinja
  ansible: templates/ansible.j2
  powerman: templates/powerman.jinja
  conman: templates/conman.jinja
```


## Known Issues

- Adds a new `OAuthClient` with every token request

## TODO