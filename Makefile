# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: geth android ios evm all test clean

GOBIN = ./build/bin
GO ?= latest
GORUN = env GO111MODULE=on go run

geth:
	$(GORUN) build/ci.go install ./cmd/geth
	@echo "Done building."
	@echo "Run \"$(GOBIN)/geth\" to launch geth."

all:
	$(GORUN) build/ci.go install

test: all
	$(GORUN) build/ci.go test

lint: ## Run linters.
	$(GORUN) build/ci.go lint

clean:
	env GO111MODULE=on go clean -cache
	rm -fr build/_workspace/pkg/ $(GOBIN)/*

# The devtools target installs tools required for 'go generate'.
# You need to put $GOBIN (or $GOPATH/bin) in your PATH to use 'go generate'.

devtools:
	env GOBIN= go install golang.org/x/tools/cmd/stringer@latest
	env GOBIN= go install github.com/fjl/gencodec@latest
	env GOBIN= go install github.com/golang/protobuf/protoc-gen-go@latest
	env GOBIN= go install ./cmd/abigen
	@type "solc" 2> /dev/null || echo 'Please install solc'
	@type "protoc" 2> /dev/null || echo 'Please install protoc'

.PHONY: concrete concrete-wasm concrete-solidity concrete-datamod

concrete: concrete-wasm concrete-solidity concrete-datamod

E2E_DIR = ./concrete/e2e
TINYGO_PCS_DIR = ./tinygo/precompiles
WASM_TESTDATA_DIR = ./concrete/wasm/testdata

concrete-wasm:
	mkdir -p $(E2E_DIR)/build
	tinygo build -opt=2 -o $(E2E_DIR)/build/blank.wasm -target=wasi $(TINYGO_PCS_DIR)/blank/blank.go
	tinygo build -opt=2 -o $(E2E_DIR)/build/add.wasm -target=wasi $(TINYGO_PCS_DIR)/add/add.go
	tinygo build -opt=2 -o $(E2E_DIR)/build/kkv.wasm -target=wasi $(TINYGO_PCS_DIR)/kkv/kkv.go
	tinygo build -opt=2 -o $(E2E_DIR)/build/gas.wasm -target=wasi $(TINYGO_PCS_DIR)/gas/gas.go
	mkdir -p $(WASM_TESTDATA_DIR)
	cp $(E2E_DIR)/build/blank.wasm $(WASM_TESTDATA_DIR)/blank.wasm
	cp $(E2E_DIR)/build/gas.wasm $(WASM_TESTDATA_DIR)/gas.wasm

concrete-solidity:
	cd ./concrete/testtool/testdata && forge build

DATAMOD_CMD_DIR = ./concrete/cmd/concrete
DATAMOD_DIR = ./concrete/codegen/datamod

concrete-datamod:
	go run $(DATAMOD_CMD_DIR) datamod $(DATAMOD_DIR)/testdata/good-datamod.json \
		--pkg testdata --out $(DATAMOD_DIR)/testdata --table-type-experimental
	go run $(DATAMOD_CMD_DIR) datamod concrete/e2e/datamod.json \
		--pkg datamod --out concrete/e2e/datamod
