# From https://raw.githubusercontent.com/vincentbernat/hellogopher/1cf9ebeeaea41d2e159e196f7d176d6b12ff0eb2/Makefile
# Licensed under Creative Commons Zero v1.0 Universal (CC0)

MODULE   = $(shell env GO111MODULE=on $(GO) list -m)
DATE    ?= $(shell date +%FT%T%z)
VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || \
			cat .version 2> /dev/null || echo v0)
PKGS     = $(or $(PKG),$(shell env GO111MODULE=on $(GO) list ./...))
BIN      = bin

GO      = go
TIMEOUT = 15
V = 0
Q = $(if $(filter 1,$V),,@)
M = $(shell if [ "$$(tput colors 2> /dev/null || echo 0)" -ge 8 ]; then printf "\033[34;1m▶\033[0m"; else printf "▶"; fi)

export GO111MODULE=on

GENERATED = # List of generated files

.PHONY: all
all: check-license fmt lint $(GENERATED) | $(BIN) ; $(info $(M) building package…) @ ## Build package
	$Q $(GO) build \
		-tags release \
		-ldflags '-X $(MODULE)/cmd.Version=$(VERSION) -X $(MODULE)/cmd.BuildDate=$(DATE)'

# Tools

$(BIN):
	@mkdir -p $@
$(BIN)/%: | $(BIN) ; $(info $(M) building $(PACKAGE)…)
	$Q env GOBIN=$(abspath $(BIN)) $(GO) install $(PACKAGE)

GOIMPORTS = $(BIN)/goimports
$(BIN)/goimports: PACKAGE=golang.org/x/tools/cmd/goimports@latest

REVIVE = $(BIN)/revive
$(BIN)/revive: PACKAGE=github.com/mgechev/revive@latest

GOCOV = $(BIN)/gocov
$(BIN)/gocov: PACKAGE=github.com/axw/gocov/gocov@latest

GOCOVXML = $(BIN)/gocov-xml
$(BIN)/gocov-xml: PACKAGE=github.com/AlekSi/gocov-xml@latest

GOTESTSUM = $(BIN)/gotestsum
$(BIN)/gotestsum: PACKAGE=gotest.tools/gotestsum@latest

ADDLICENSE = $(BIN)/addlicense
$(BIN)/addlicense: PACKAGE=github.com/google/addlicense@latest

# Generate

# Tests

TEST_TARGETS := test-bench test-short test-verbose test-race
.PHONY: $(TEST_TARGETS) check test tests
test-bench:   ARGS=-run=__absolutelynothing__ -bench=. ## Run benchmarks
test-short:   ARGS=-short        ## Run only short tests
test-verbose: ARGS=-v            ## Run tests in verbose mode with coverage reporting
test-race:    ARGS=-race         ## Run tests with race detector
$(TEST_TARGETS): NAME=$(MAKECMDGOALS:test-%=%)
$(TEST_TARGETS): test
check test tests: check-license fmt lint $(GENERATED) | $(GOTESTSUM) ; $(info $(M) running $(NAME:%=% )tests…) @ ## Run tests
	$Q mkdir -p test
	$Q $(GOTESTSUM) --junitfile test/tests.xml -- -timeout $(TIMEOUT)s $(ARGS) $(PKGS)

COVERAGE_MODE = atomic
.PHONY: test-coverage
test-coverage: check-license fmt lint $(GENERATED)
test-coverage: | $(GOCOV) $(GOCOVXML) $(GOTESTSUM) ; $(info $(M) running coverage tests…) @ ## Run coverage tests
	$Q mkdir -p test
	$Q $(GOTESTSUM) -- \
		-coverpkg=$(shell echo $(PKGS) | tr ' ' ',') \
		-covermode=$(COVERAGE_MODE) \
		-coverprofile=test/profile.out $(PKGS)
	$Q $(GO) tool cover -html=test/profile.out -o test/coverage.html
	$Q $(GOCOV) convert test/profile.out | $(GOCOVXML) > test/coverage.xml
	@echo -n "Code coverage: "; \
		echo "scale=1;$$(sed -En 's/^<coverage line-rate="([0-9.]+)".*/\1/p' test/coverage.xml) * 100 / 1" | bc -q

.PHONY: lint
lint: | $(REVIVE) ; $(info $(M) running golint…) @ ## Run golint
	$Q $(REVIVE) -formatter friendly -set_exit_status ./...

.PHONY: fmt
fmt: | $(GOIMPORTS) ; $(info $(M) running gofmt…) @ ## Run gofmt on all source files
	$Q $(GOIMPORTS) -local $(MODULE) -w $(shell $(GO) list -f '{{$$d := .Dir}}{{range $$f := .GoFiles}}{{printf "%s/%s\n" $$d $$f}}{{end}}{{range $$f := .CgoFiles}}{{printf "%s/%s\n" $$d $$f}}{{end}}{{range $$f := .TestGoFiles}}{{printf "%s/%s\n" $$d $$f}}{{end}}' $(PKGS))

# Misc

.PHONY: clean
clean: ; $(info $(M) cleaning…)	@ ## Cleanup everything
	@rm -rf $(BIN) test $(GENERATED)

.PHONY: help
help:
	@grep -hE '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-17s\033[0m %s\n", $$1, $$2}'

.PHONY: version
version:
	@echo $(VERSION)

# See https://stackoverflow.com/questions/52242077/go-modules-finding-out-right-pseudo-version-vx-y-z-timestamp-commit-of-re
PSEUDO_VERSION ?= $(shell TZ=UTC git --no-pager show --quiet --abbrev=12 --date='format-local:%Y%m%d%H%M%S' --format="%cd-%h" 2> /dev/null || echo v0)
.PHONY: pseudo-version
pseudo-version:
	@echo $(PSEUDO_VERSION)

# License

check-license: | $(ADDLICENSE) ; $(info $(M) checking license headers…)
	$Q $(ADDLICENSE) -check -c "Alexandre Dutra" -ignore '.github/**'  -ignore '.idea/**' -ignore 'test/**' -ignore 'bin/**' . 2> /dev/null

update-license: | $(ADDLICENSE) ; $(info $(M) updating license headers…)
	$Q $(ADDLICENSE) -c "Alexandre Dutra" -ignore '.github/**'  -ignore '.idea/**' -ignore 'test/**' -ignore 'bin/**' . 2> /dev/null
