.PHONY: test oracle-test build

test:
	go test ./...

oracle-test:
	test -n "$$ARK_ORACLE_SAVE"
	go test ./arksave -run TestOracleSaveEnumeratesObjects -count=1

build:
	mkdir -p bin
	CGO_ENABLED=0 go build -o bin/arksave ./cmd/arksave
