.PHONY: test build demo site fire-drill benches tidy examples clean evaluate install doctor

GO ?= go
ROOT ?= .sde
PREFIX ?= ./dist
SDE := $(PREFIX)/sde

build:
	@mkdir -p $(PREFIX)
	$(GO) build -ldflags "-s -w" -o $(SDE)$(shell go env GOEXE) ./cmd/sde

install: build
	@$(SDE)$(shell go env GOEXE) version
	@echo "binary: $(SDE)$(shell go env GOEXE)"
	@echo "tip: ./scripts/install.ps1  or  ./scripts/install.sh"

doctor: build
	$(SDE)$(shell go env GOEXE) doctor --root $(ROOT)

test:
	$(GO) test ./...

evaluate: build
	$(GO) run ./cmd/evaluate-harness -out evaluation -run evaluation/.run
	$(GO) test ./... -count=1
	@echo "See evaluation/SUMMARY.md (generate via evaluate.ps1 / evaluate.sh for full SUMMARY)"

tidy:
	$(GO) mod tidy

demo: build
	$(SDE)$(shell go env GOEXE) demo --root $(ROOT)

fire-drill: build
	@echo "Usage: make fire-drill ARCHIVE=path/to/psa (requires prior export/escrow)"
	@test -n "$(ARCHIVE)" || (echo "set ARCHIVE=..." && exit 1)
	$(SDE)$(shell go env GOEXE) fire-drill --root $(ROOT) --archive "$(ARCHIVE)"
	$(SDE)$(shell go env GOEXE) readiness --root $(ROOT)

benches: build
	$(SDE)$(shell go env GOEXE) benchmark --root $(ROOT)-bench
	$(GO) test ./internal/chunk -run TestWriteDifferentialBenchResults -count=1
	$(GO) test ./benchmarks -run TestPerfRegressionGates -count=1

examples:
	$(GO) run ./examples/acquisition-demo
	$(GO) run ./examples/clone-redact
	$(GO) run ./examples/sqlite-restore-cycle

site:
	@echo "Static site lives in site/ — served by GitHub Pages workflow"

clean:
	rm -rf $(ROOT) $(ROOT)-bench $(PREFIX)/sde $(PREFIX)/sde.exe evaluation/.run
