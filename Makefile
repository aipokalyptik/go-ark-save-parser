.PHONY: test oracle-test bench build

test:
	go test ./...

oracle-test:
	test -n "$$ARK_ORACLE_SAVE"
	go test ./arksave -run TestOracleSaveEnumeratesObjects -count=1

bench:
	go test ./arkapi -run '^$$' -bench . -benchmem

build:
	mkdir -p bin
	CGO_ENABLED=0 go build -o bin/arksave ./cmd/arksave
