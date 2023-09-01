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

.PHONY: concrete concrete-wasm concrete-solidity

concrete: concrete-wasm concrete-solidity

FIXTURE_DIR = ./concrete/fixtures
WASM_TESTDATA_DIR = ./concrete/wasm/testdata

concrete-wasm:
	mkdir -p $(FIXTURE_DIR)/build
	mkdir -p $(FIXTURE_DIR)/add
	mkdir -p $(FIXTURE_DIR)/kvv
	mkdir -p $(WASM_TESTDATA_DIR)
	tinygo build -opt=2 -o $(FIXTURE_DIR)/build/blank.wasm -target=wasi ./tinygo/precompiles/blank/blank.go
	cp $(FIXTURE_DIR)/build/blank.wasm $(WASM_TESTDATA_DIR)/blank.wasm
	tinygo build -opt=2 -o $(FIXTURE_DIR)/build/add.wasm -target=wasi ./tinygo/precompiles/add/add.go
	tinygo build -opt=2 -o $(FIXTURE_DIR)/build/kkv.wasm -target=wasi ./tinygo/precompiles/kkv/kkv.go

concrete-solidity:
	cd ./concrete/testtool/testdata && forge build