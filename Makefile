GOBIN = $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN = $(shell go env GOPATH)/bin
endif
ifeq ($(PREFIX),)
    PREFIX := /usr/local
endif
ifeq ($(SERVICE_PREFIX),)
    SERVICE_PREFIX := /etc/systemd/system
endif
ifeq ($(XDG_CONFIG_HOME),)
    XDG_CONFIG_HOME := /etc
endif
ifeq ($(OS),)
	OS := linux
endif
ifeq ($(ARCH),)
	ARCH := amd64
endif
ifeq ($(CMD),)
	CMD := themer
endif

.PHONY: help
help:
	@echo "Usage: [variables] make <target>"
	@echo
	@echo "This Makefile makes use of dependency lists."
	@echo "The artifacts are compiled only if any of their dependencies are newer than them."
	@echo
	@echo "Commands:"
	@echo "\tbuild            \tBuilds the binary."
	@echo "\tinstall          \tInstall the binary and a basic config file."
	@echo "\t                 \tSet the value of the PREFIX (default: /usr/local) to change the installation location."
	@echo "\t                 \tSet the value of the XDG_CONFIG_HOME to change the config location."
	@echo "\tinstall-autostart\tInstall the binary, a basic config file and a systemd service file."
	@echo "\t                 \tSet the value of the PREFIX (default: /usr/local) to change the installation location."
	@echo "\t                 \tSet the value of the SERVICE_PREFIX (default: /etc/systemd/system) to change the installation location."
	@echo "\t                 \tSet the value of the XDG_CONFIG_HOME to change the config location."
	@echo
	@echo "Utilities:"
	@echo "\tclear         \tRemoves all build artifacts."

.PHONY: install
install:
	$(MAKE) compile-and-install

.PHONY: install-autostart
install-autostart:
	$(MAKE) compile-and-install
	$(MAKE) install-service

.PHONY: compile-and-install
compile-and-install:
	$(MAKE) compile
	install -d $(DESTDIR)$(PREFIX)/bin
	install -m 755 build/$(CMD)/$(CMD)_$(OS)_$(ARCH) $(DESTDIR)$(PREFIX)/bin/$(CMD)
	install -d $(XDG_CONFIG_HOME)/themer
	install -b -S .old templates/config.json.template $(XDG_CONFIG_HOME)/themer/config.json
	install -d $(PREFIX)/share/applications

.PHONY: install-service
install-service:
	install -m 764 templates/themer.service.template $(SERVICE_PREFIX)/themer.service 
	sed -i'' "s#ExecStart\=themer#ExecStart\=\"$$(realpath $(DESTDIR)$(PREFIX)/bin/$(CMD))\"#" $(SERVICE_PREFIX)/themer.service 
	echo "Remember to enable the service within systemd!"
	echo
	echo "If you plan to use it as a user service, make sure to replace the dependency from \"multi-user.target\" to \"default.target\"."

.PHONY: build
build:
	$(MAKE) compile

.PHONY: compile
compile: build/$(CMD)/$(CMD)_$(OS)_$(ARCH)

build/$(CMD)/$(CMD)_$(OS)_$(ARCH): $(shell find . -type f -name '*.go' -print)
	mkdir -p $(shell dirname $@)
	rm -f $@
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build -o $@ ./cmd/$(CMD)
	chmod 0700 build/$(CMD)/*

clear:
	rm -rf build/*
