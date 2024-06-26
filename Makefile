
# build everything at once
all: plugins exe

# build the main executable to make configs
main: exe
exe:
	go build --tags=all -o configurator

# build all of the generators into plugins
plugins:
	mkdir -p lib
	go build -buildmode=plugin -o lib/conman.so internal/generator/plugins/conman/conman.go
	go build -buildmode=plugin -o lib/coredhcp.so internal/generator/plugins/coredhcp/coredhcp.go
	go build -buildmode=plugin -o lib/dnsmasq.so internal/generator/plugins/dnsmasq/dnsmasq.go
	go build -buildmode=plugin -o lib/powerman.so internal/generator/plugins/powerman/powerman.go
	go build -buildmode=plugin -o lib/syslog.so internal/generator/plugins/syslog/syslog.go
