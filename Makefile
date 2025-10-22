BINARY := webcrawler

Q := @
ifdef VERBOSE
Q :=
endif

GOFLAGS := -v

INSTALL_DIR := $(shell go env GOPATH)/bin

.PHONY: all install build run test clean

all: build

build:
	$(Q)echo "Building $(BINARY)..."
	$(Q)go build $(GOFLAGS) -o $(BINARY) .

install: build 
	$(Q)echo "Installing $(BINARY) to $(INSTALL_DIR)"
	$(Q)go install $(GOFLAGS) .

run: build
	$(Q)echo "Running $(BINARY)"
	$(Q)./$(BINARY)

test: install
	$(Q)echo "Running tests..." 
	$(Q)$(BINARY)

clean:
	$(Q)echo "Cleaning up..."
	$(Q)rm -f $(BINARY)
