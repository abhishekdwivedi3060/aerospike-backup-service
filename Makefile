# Go parameters
WORKSPACE = $(shell git rev-parse --show-toplevel)
GOCMD = go
UNAME = $(shell uname -sm | tr ' ' '-')
CGO_CFLAGS=-I $(WORKSPACE)/modules/aerospike-tools-backup/modules/c-client/target/$(UNAME)/include \
 -I $(WORKSPACE)/modules/aerospike-tools-backup/include
GOBUILD = CGO_CFLAGS="$(CGO_CFLAGS)" CGO_ENABLED=1 $(GOCMD) build
GOTEST = $(GOCMD) test
GOCLEAN = $(GOCMD) clean
GO_VERSION = 1.21.4
GOBIN_VERSION = $(shell $(GO) version 2>/dev/null)
OS = $(shell uname | tr '[:upper:]' '[:lower:]')
ARCH = $(shell uname -m)
ifeq ($(ARCH),x86_64)
	ARCH = amd64
else ifeq ($(ARCH),aarch64)
	ARCH = arm64
endif

LSB_EXISTS := $(shell which lsb_release 2> /dev/null)
ifeq ($(LSB_EXISTS),)
	DISTRO_FULL := $(shell . /etc/os-release 2> /dev/null; echo $$NAME | tr ' ' '_')
	DISTRO_VERSION := $(shell . /etc/os-release 2> /dev/null; echo $$VERSION_ID | tr ' ' '_')
else
	DISTRO_FULL := $(shell lsb_release -i | cut -f2- | tr ' ' '_')
	DISTRO_VERSION := $(shell lsb_release -r | cut -f2- | tr ' ' '_')
endif

ifeq ($(DISTRO_FULL),Debian)
	DISTRO_SHORT = debian
else ifeq ($(DISTRO_FULL),Ubuntu)
	DISTRO_SHORT = ubuntu
else ifeq ($(DISTRO_FULL),Amazon_Linux)
	DISTRO_SHORT = amzn
else ifeq ($(DISTRO_FULL),CentOS_Linux)
	DISTRO_SHORT = el
else ifeq ($(DISTRO_FULL),Red_Hat_Enterprise_Linux)
	DISTRO_SHORT = el
else ifeq ($(DISTRO_FULL),Rocky_Linux)
	DISTRO_SHORT = rocky
endif

BINARY_NAME = aerospike-backup-service
GIT_TAG = $(shell git describe --tags)

CMD_DIR = cmd/backup
TARGET_DIR = target
PKG_DIR = build/package
PREP_DIR = $(TARGET_DIR)/pkg_install
CONFIG_FILES = $(wildcard config/*)
POST_INSTALL_SCRIPT = $(PKG_DIR)/post-install.sh
TOOLS_DIR = modules/aerospike-tools-backup

MAINTAINER = "Aerospike"
DESCRIPTION = "Aerospike Backup Service"
URL = "https://www.aerospike.com"
VENDOR = "Aerospike"
LICENSE = "Apache License 2.0"

FPM_COMMON_ARGS = \
	--force \
	--input-type dir \
	--name $(BINARY_NAME) \
	--version $(GIT_TAG) \
	--chdir $(PREP_DIR) \
	--maintainer $(MAINTAINER) \
	--description $(DESCRIPTION) \
	--url $(URL) \
	--vendor $(VENDOR) \
	--license $(LICENSE) \
	--after-install $(POST_INSTALL_SCRIPT)

.PHONY: install-aws-sdk-cpp
install-aws-sdk-cpp:
	git clone --recurse-submodules https://github.com/aws/aws-sdk-cpp
	cd $(WORKSPACE)/aws-sdk-cpp && \
	cmake -S . -B build \
	-DCMAKE_BUILD_TYPE=Release \
	-DBUILD_ONLY="s3" \
	-DBUILD_SHARED_LIBS=OFF \
	-DENABLE_TESTING=OFF \
	-DCMAKE_INSTALL_PREFIX=/usr/local \
	-DCMAKE_INSTALL_LIBDIR=lib
	cd $(WORKSPACE)/aws-sdk-cpp/ && sudo make -C build
	cd $(WORKSPACE)/aws-sdk-cpp/build && sudo make install

.PHONY: install-go
install-go:
ifdef GOBIN_VERSION
else
	curl -L "https://go.dev/dl/go$(GO_VERSION).$(OS)-$(ARCH).tar.gz" > "go$(GO_VERSION).$(OS)-$(ARCH).tar.gz"; \
	sudo tar -C /usr/local -xzf "go$(GO_VERSION).$(OS)-$(ARCH).tar.gz" && rm "go$(GO_VERSION).$(OS)-$(ARCH).tar.gz"; \
	echo export PATH="\$PATH:/usr/local/go/bin" >> $(HOME)/.profile
endif

.PHONY: install-deb-build-deps
install-deb-build-deps:
	sudo apt-get update
	sudo sh -c 'curl -OL https://go.dev/dl/go1.21.4.linux-amd64.tar.gz && tar -C /usr/local -xzf go1.21.4.linux-amd64.tar.gz && rm go1.21.4.linux-amd64.tar.gz && echo "export PATH=\$PATH:/usr/local/go/bin" >> /etc/profile.d/go.sh'
	sudo apt-get install -y \
	build-essential \
	libssl-dev \
	libuv1-dev \
	libcurl4-openssl-dev \
	libzstd-dev \
	cmake \
	pkg-config \
	zlib1g-dev \
	debhelper \
	lintian \
	devscripts

.PHONY: build-submodules
build-submodules install-aws-sdk-cpp:
	cd $(TOOLS_DIR) && git submodule update --init --recursive
	$(MAKE) -C $(TOOLS_DIR) shared EVENT_LIB=libuv AWS_SDK_STATIC_PATH=/usr/local/lib
	./scripts/copy_shared.sh

.PHONY: clean-submodules
clean-submodules:
	$(MAKE) -C $(TOOLS_DIR) clean

.PHONY: build
build:
	mkdir -p $(TARGET_DIR)
	$(GOBUILD) -o $(TARGET_DIR)/$(BINARY_NAME) ./$(CMD_DIR)

.PHONY: test
test:
	$(GOTEST) -v ./...

.PHONY: package
package: rpm deb tar

.PHONY: rpm
rpm: build prep
	$(eval ARCH := $(shell uname -m))
	$(eval DISTRO_VERSION := $(shell echo $(DISTRO_VERSION) | cut -d'.' -f1)) # Only major version for RPM
	fpm $(FPM_COMMON_ARGS) \
		--output-type rpm \
		--package $(TARGET_DIR)/$(BINARY_NAME)-$(GIT_TAG)-1.$(DISTRO_SHORT)$(DISTRO_VERSION).$(ARCH).rpm

.PHONY: deb
deb:
	cd $(WORKSPACE)/package && dpkg-buildpackage
	mv $(WORKSPACE)/$(BINARY_NAME)_* $(WORKSPACE)/target

.PHONY: tar
tar: build prep
	fpm $(FPM_COMMON_ARGS) \
		--output-type tar \
		--package $(TARGET_DIR)/$(BINARY_NAME)_$(GIT_TAG)_$(DISTRO_SHORT)$(DISTRO_VERSION)_$(ARCH).tgz

.PHONY: prep
prep:
ifndef DISTRO_FULL
	$(error Distro not found)
endif

ifndef DISTRO_VERSION
	$(error Distro Version not found)
endif

	@echo "Distro: $(DISTRO_FULL)"
	@echo "Distro Version: $(DISTRO_VERSION)"

	@which git > /dev/null || (echo "Git is not installed"; exit 1)
	@which fpm > /dev/null || (echo "FPM is not installed"; exit 1)

	install -d $(PREP_DIR)
	install -d $(PREP_DIR)/usr/local/bin
	install -d $(PREP_DIR)/var/log/aerospike
	install -d $(PREP_DIR)/etc/$(BINARY_NAME)
	install -d $(PREP_DIR)/etc/systemd/system
	install -m 755 $(TARGET_DIR)/$(BINARY_NAME) $(PREP_DIR)/usr/local/bin/$(BINARY_NAME)
	install -m 644 $(CONFIG_FILES) $(PREP_DIR)/etc/$(BINARY_NAME)/
	install -m 644 $(PKG_DIR)/$(BINARY_NAME).service $(PREP_DIR)/etc/systemd/system/$(BINARY_NAME).service

.PHONY: clean
clean:
	$(GOCLEAN)
	rm -rf $(TARGET_DIR)

.PHONY: all
all: build test package