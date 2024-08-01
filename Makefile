# Unless set otherwise, the container runtime is Docker
DOCKER ?= docker

prog ?= configurator
git_tag := $(shell git describe --abbrev=0 --tags)
sources := main.go $(wildcard cmd/*.go)
plugin_source_prefix := pkg/generator/plugins
plugin_sources := $(filter-out %_test.go,$(wildcard $(plugin_source_prefix)/*/*.go))
plugin_binaries := $(addprefix lib/,$(patsubst %.go,%.so,$(notdir $(plugin_sources))))

# build everything at once
.PHONY: all
all: plugins exe test

# build the main executable to make configs
.PHONY: main driver binaries exe
main: exe
driver: exe
binaries: exe
exe: $(prog)

# build named executable from go sources
$(prog): $(sources)
	go build --tags=all -o $(prog)

.PHONY: container
container: binaries plugins
	$(DOCKER) build . --build-arg --no-cache --pull --tag '$(prog):$(git_tag)-dirty'

.PHONY: container-testing
container-testing: binaries plugins
	$(DOCKER) build . --tag $(prog):testing

# build all of the generators into plugins
.PHONY: plugins
plugins: $(plugin_binaries)

# how to make each plugin
lib/%.so: pkg/generator/plugins/%/*.go
	mkdir -p lib
	go build -buildmode=plugin -o $@ $<

# remove executable and all built plugins
.PHONY: clean
clean:
	rm -f configurator
	rm -f lib/*

# run all of the unit tests
.PHONY: test
test: $(prog) $(plugin_binaries)
	go test ./tests/generate_test.go --tags=all