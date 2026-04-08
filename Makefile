# Variables
BINARY   = ip2proxyliteconvert
PREFIX  ?= /usr/local
BINDIR   = $(PREFIX)/bin
INSTALL_PATH = $(BINDIR)/$(BINARY)

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
SRC_PATH = ./cmd/ip2proxyliteconvert
GEN_PATH  = ./cmd/gen-exceptions

Q = $(if $(V),,@)

.PHONY: all build install uninstall clean help tidy gen test

all: tidy build

tidy:
	@echo "--- Tidying Dependencies ---"
	$(Q)go mod tidy

gen:
	@echo "--- Generating Exception Test Data ---"
	$(Q)go run $(GEN_PATH)

build:
	@echo "--- Building $(BINARY) ($(VERSION)) ---"
	$(Q)go build -ldflags="-X main.version=$(VERSION)" -o $(BINARY) $(SRC_PATH)

test: gen
	@echo "--- Running Tests ---"
	$(Q)go test -v ./...

install: build
	@echo "--- Installing to $(DESTDIR)$(INSTALL_PATH) ---"
	install -m 755 $(BINARY) $(DESTDIR)$(INSTALL_PATH)

uninstall:
	@echo "--- Removing $(BINARY) ---"
	rm -f $(DESTDIR)$(INSTALL_PATH)

clean:
	@echo "--- Cleaning Workspace ---"
	rm -f $(BINARY)

help:
	@echo "Usage: make [V=1] [target]"
	@echo "Targets:"
	@echo "  all       - Tidy and build"
	@echo "  gen       - Generate exception test CSVs"
	@echo "  test      - Generate data and run all tests"
	@echo "  install   - Install binary to $(INSTALL_PATH)"
	@echo "  clean     - Remove binary and test samples"