
# build everything at once
all: plugins exe test

# build the main executable to make configs
main: exe
driver: exe
binaries: exe
exe:
	go build --tags=all -o configurator


docker: binaries plugins
	docker build . --build-arg REGISTRY_HOST=${REGISTRY_HOST} --no-cache --pull --tag '${NAME}:${VERSION}'

# build all of the generators into plugins
plugins:
	mkdir -p lib
	go build -buildmode=plugin -o lib/conman.so pkg/generator/plugins/conman/conman.go
	go build -buildmode=plugin -o lib/coredhcp.so pkg/generator/plugins/coredhcp/coredhcp.go
	go build -buildmode=plugin -o lib/dhcpd.so pkg/generator/plugins/dhcpd/dhcpd.go
	go build -buildmode=plugin -o lib/dnsmasq.so pkg/generator/plugins/dnsmasq/dnsmasq.go
	go build -buildmode=plugin -o lib/example.so pkg/generator/plugins/example/example.go
	go build -buildmode=plugin -o lib/hostfile.so pkg/generator/plugins/hostfile/hostfile.go
	go build -buildmode=plugin -o lib/powerman.so pkg/generator/plugins/powerman/powerman.go
	go build -buildmode=plugin -o lib/syslog.so pkg/generator/plugins/syslog/syslog.go
	go build -buildmode=plugin -o lib/warewulf.so pkg/generator/plugins/warewulf/warewulf.go

# remove executable and all built plugins
clean:
	rm configurator
	rm lib/*

# run all of the unit tests
test:
	go test ./tests/generate_test.go --tags=all
