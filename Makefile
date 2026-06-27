GO_CACHE ?= $(CURDIR)/.cache/go-build
GO_MOD_CACHE ?= $(CURDIR)/.cache/go-mod
PY_CACHE ?= $(CURDIR)/.cache/pycache

.PHONY: test verify e2e-test oracle-test oracle-compare bench build

test:
	GOCACHE="$(GO_CACHE)" go test ./...

verify:
	GOCACHE="$(GO_CACHE)" go test ./... -count=1
	PYTHONPYCACHEPREFIX="$(PY_CACHE)" python3 -m py_compile scripts/*.py
	PYTHONPYCACHEPREFIX="$(PY_CACHE)" python3 -m unittest discover -s scripts -p '*_test.py'
	$(MAKE) build

e2e-test:
	GOCACHE="$(GO_CACHE)" go test ./arkapi ./cmd/arksave -run TestProvidedDataReadOnlyE2E -count=1

oracle-test:
	test -n "$$ARK_ORACLE_SAVE"
	GOCACHE="$(GO_CACHE)" go test ./arksave -run TestOracleSaveEnumeratesObjects -count=1
	if [ -n "$$ARK_ORACLE_TRIBUTE" ]; then GOCACHE="$(GO_CACHE)" go test ./arktribute -run TestOracleTributeParsesLocalIndex -count=1; else echo "ARK_ORACLE_TRIBUTE not set; skipping tribute oracle"; fi

oracle-compare:
	test -n "$$ARK_ORACLE_SAVE"
	.oracle/venv/bin/python scripts/oracle_compare.py

bench:
	GOCACHE="$(GO_CACHE)" go test ./arkapi -run '^$$' -bench . -benchmem

build:
	mkdir -p bin
	GOCACHE="$(GO_CACHE)" GOMODCACHE="$(GO_MOD_CACHE)" CGO_ENABLED=0 go build -o bin/arksave ./cmd/arksave
