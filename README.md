# OpenCHAMI Configurator

Configurator is a tool that fetchs data from an instance of [SMD](https://github.com/OpenCHAMI/smd) to generate commonly used config files.

## Building and Usage

Configurator is built using Go:

```bash
go mod tidy && go build
```

To use the tool, run the following:

```bash
./configurator generate --config config.yaml --target dhcp:dnsmasq
```

This will generate a new DHCP `dnsmasq` config file based on the Jinja 2 template specified in the config file for "dnsmasq". The `--target` flag is set by passing an argument in the form of "type:template" to specify the type of config file being generate and the template file to use respectively. The configurator requires valid access token when making requests to an instance of SMD that has protected routes.

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
  ansible: templates/ansible
  powerman: templates/powerman
  conman: templates/conman
```


## Known Issues

- Adds a new `OAuthClient` with every token request

## TODO