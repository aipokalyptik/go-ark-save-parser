.PHONY: test oracle-test oracle-compare bench build

test:
	go test ./...

oracle-test:
	test -n "$$ARK_ORACLE_SAVE"
	go test ./arksave -run TestOracleSaveEnumeratesObjects -count=1

oracle-compare:
	test -n "$$ARK_ORACLE_SAVE"
	.oracle/venv/bin/python scripts/oracle_compare.py

bench:
	go test ./arkapi -run '^$$' -bench . -benchmem

build:
	mkdir -p bin
	CGO_ENABLED=0 go build -o bin/arksave ./cmd/arksave
