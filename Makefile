.PHONY: test oracle-test oracle-compare bench build

test:
	go test ./...

oracle-test:
	test -n "$$ARK_ORACLE_SAVE"
	go test ./arksave -run TestOracleSaveEnumeratesObjects -count=1
	if [ -n "$$ARK_ORACLE_TRIBUTE" ]; then go test ./arktribute -run TestOracleTributeParsesLocalIndex -count=1; else echo "ARK_ORACLE_TRIBUTE not set; skipping tribute oracle"; fi

oracle-compare:
	test -n "$$ARK_ORACLE_SAVE"
	.oracle/venv/bin/python scripts/oracle_compare.py

bench:
	go test ./arkapi -run '^$$' -bench . -benchmem

build:
	mkdir -p bin
	CGO_ENABLED=0 go build -o bin/arksave ./cmd/arksave
