#────────────────────────────────────────────────────────────
# ogspy – minimal cross-compile Makefile
#────────────────────────────────────────────────────────────
BINARY      ?= ogspy
BUILD_DIR   ?= build
VERSION     ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "0.0.0")

GOFLAGS     := -trimpath -ldflags="-s -w -X 'main.__version__=$(VERSION)'"
export CGO_ENABLED := 0

# Cross-compile matrix --------------------------------------------------------
TARGETS = \
  darwin-amd64 \
  darwin-arm64 \
  linux-amd64 \
  linux-arm64

all: $(TARGETS) ## Build all target binaries

$(TARGETS):
	@echo ">> building $@ (version $(VERSION))"
	@mkdir -p $(BUILD_DIR)
	@OS=$(word 1,$(subst -, ,$@)) && \
	  ARCH=$(word 2,$(subst -, ,$@)) && \
	  GOOS=$$OS GOARCH=$$ARCH go build $(GOFLAGS) \
	    -o $(BUILD_DIR)/$(BINARY)-$(VERSION)-$$OS-$$ARCH

# Convenience wrappers --------------------------------------------------------
install: host = $(shell uname -s | tr A-Z a-z)-$(shell uname -m)
install: ## Install the host binary into /usr/local/bin
	@echo ">> installing $(host)"
	@sudo install -m 0755 $(BUILD_DIR)/$(BINARY)-$(VERSION)-$(host) /usr/local/bin/$(BINARY)

clean:  ## Remove build artefacts
	@rm -rf $(BUILD_DIR)

sha:    ## Print SHA-256 for all artefacts (for checksums / Homebrew)
	@sha256sum $(BUILD_DIR)/* | sed 's|$(BUILD_DIR)/||'

# Quick & dirty Homebrew formula scaffold (stdout) ----------------------------
formula: ## Generate a draft Homebrew formula for darwin (stdout)
	@CPU_ARCHS="amd64 arm64"; \
	echo "class Ogspy < Formula"; \
	echo "  desc \"CLI to inspect, validate and monitor Open Graph metadata\""; \
	echo "  homepage \"https://github.com/vincenzomaritato/$(BINARY)\""; \
	echo "  version \"$(VERSION)\""; \
	for ARCH in $$CPU_ARCHS; do \
	  FNAME=$(BINARY)-$(VERSION)-darwin-$$ARCH; \
	  SHA=$$(shasum -a 256 $(BUILD_DIR)/$$FNAME | awk '{print $$1}'); \
	  if [ "$$ARCH" = "arm64" ]; then \
	    echo "  if Hardware::CPU.arm?"; \
	  else \
	    echo "  else"; \
	  fi; \
	  echo "    url \"https://github.com/vincenzomaritato/$(BINARY)/releases/download/$(VERSION)/$$FNAME\""; \
	  echo "    sha256 \"$$SHA\""; \
	done; \
	echo "  end"; \
	echo ""; \
	echo "  def install"; \
	echo "    bin.install Dir[\"ogspy*\"][0] => \"ogspy\""; \
	echo "  end"; \
	echo "end"

# Help target -----------------------------------------------------------------
.PHONY: $(TARGETS) all install clean sha formula help
help:  ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?##' $(MAKEFILE_LIST) | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'