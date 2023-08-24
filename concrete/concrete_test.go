// Copyright 2023 The concrete-geth Authors
//
// The concrete-geth library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The concrete library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the concrete library. If not, see <http://www.gnu.org/licenses/>.

package concrete

import (
	_ "embed"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/api"
	"github.com/ethereum/go-ethereum/concrete/fixtures"
	"github.com/ethereum/go-ethereum/concrete/lib"
	"github.com/ethereum/go-ethereum/concrete/mock"
	"github.com/ethereum/go-ethereum/concrete/precompiles"
	"github.com/ethereum/go-ethereum/concrete/wasm"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

//go:embed fixtures/build/add.wasm
var addWasm []byte

//go:embed fixtures/build/kkv.wasm
var kkvWasm []byte

var addImplementations = []struct {
	name    string
	address common.Address
	pc      precompiles.Precompile
}{
	{
		name:    "Native",
		address: common.BytesToAddress([]byte{128}),
		pc:      &fixtures.AdditionPrecompile{},
	},
	{
		name:    "Wasm",
		address: common.BytesToAddress([]byte{129}),
		pc:      wasm.NewWasmPrecompile(addWasm),
	},
}

func getAddABI() abi.ABI {
	addABI, err := abi.JSON(strings.NewReader(fixtures.AddAbiString))
	if err != nil {
		panic(err)
	}
	return addABI
}

func TestAddPrecompileFixture(t *testing.T) {
	var (
		ABI = getAddABI()
		x   = big.NewInt(1)
		y   = big.NewInt(2)
	)
	for _, impl := range addImplementations {
		precompiles.AddPrecompile(impl.address, impl.pc)
	}
	for _, impl := range addImplementations {
		env := mock.NewMockEnv(impl.address, api.EnvConfig{Trusted: true}, false, 0)
		t.Run(impl.name, func(t *testing.T) {
			input, err := ABI.Pack("add", x, y)
			if err != nil {
				t.Fatal(err)
			}
			isStatic := impl.pc.IsStatic(input)
			if err != nil {
				t.Fatal(err)
			}
			if !isStatic {
				t.Fatal("expected static")
			}
			output, err := impl.pc.Run(env, input)
			if err != nil {
				t.Fatal(err)
			}
			values, err := ABI.Methods["add"].Outputs.Unpack(output)
			if err != nil {
				t.Fatal(err)
			}
			value := values[0].(*big.Int)
			if value.Cmp(x.Add(x, y)) != 0 {
				t.Fatalf("expected 3, got %d", value)
			}
		})
	}
}

var kkvImplementations = []struct {
	name    string
	address common.Address
	pc      precompiles.Precompile
}{
	{
		name:    "Native",
		address: common.BytesToAddress([]byte{130}),
		pc:      &fixtures.KeyKeyValuePrecompile{},
	},
	{
		name:    "Wasm",
		address: common.BytesToAddress([]byte{131}),
		pc:      wasm.NewWasmPrecompile(kkvWasm),
	},
}

func getKkvABI() abi.ABI {
	kkvABI, err := abi.JSON(strings.NewReader(fixtures.KkvAbiString))
	if err != nil {
		panic(err)
	}
	return kkvABI
}

func TestKkvPrecompileFixture(t *testing.T) {
	var (
		ABI = getKkvABI()
		k1  = common.HexToHash("0x01")
		k2  = common.HexToHash("0x02")
		v   = common.HexToHash("0x03")
	)
	for _, impl := range kkvImplementations {
		precompiles.AddPrecompile(impl.address, impl.pc)
	}
	for _, impl := range kkvImplementations {
		env := mock.NewMockEnv(impl.address, api.EnvConfig{Trusted: true}, false, 0)
		t.Run(impl.name, func(t *testing.T) {
			input, err := ABI.Pack("set", k1, k2, v)
			if err != nil {
				t.Fatal(err)
			}
			isStatic := impl.pc.IsStatic(input)
			if err != nil {
				t.Fatal(err)
			}
			if isStatic {
				t.Fatal("expected non-static")
			}
			_, err = impl.pc.Run(env, input)
			if err != nil {
				t.Fatal(err)
			}
			input, err = ABI.Pack("get", k1, k2)
			if err != nil {
				t.Fatal(err)
			}
			isStatic = impl.pc.IsStatic(input)
			if err != nil {
				t.Fatal(err)
			}
			if !isStatic {
				t.Fatal("expected static")
			}
			output, err := impl.pc.Run(env, input)
			if err != nil {
				t.Fatal(err)
			}
			values, err := ABI.Methods["get"].Outputs.Unpack(output)
			if err != nil {
				t.Fatal(err)
			}
			value := common.Hash(values[0].([32]byte))
			if value != v {
				t.Fatalf("expected %s, got %s", v, value)
			}
		})
	}
}

var kkvImplementationsE2E = []struct {
	name    string
	address common.Address
	pc      precompiles.Precompile
}{
	{
		name:    "Native",
		address: common.BytesToAddress([]byte{132}),
		pc:      &fixtures.KeyKeyValuePrecompile{},
	},
	{
		name:    "Wasm",
		address: common.BytesToAddress([]byte{133}),
		pc:      wasm.NewWasmPrecompile(kkvWasm),
	},
}

func TestE2EKkvPrecompile(t *testing.T) {
	var (
		ABI           = getKkvABI()
		key, _        = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
		senderAddress = crypto.PubkeyToAddress(key.PublicKey)
		nBlocks       = 3
	)
	for _, impl := range kkvImplementationsE2E {
		precompiles.AddPrecompile(impl.address, impl.pc)
	}
	for _, impl := range kkvImplementationsE2E {
		t.Run(impl.name, func(t *testing.T) {
			var (
				gspec = &core.Genesis{
					Config: params.TestChainConfig,
					Alloc: core.GenesisAlloc{
						senderAddress: {Balance: big.NewInt(1e15)},
					},
				}
				signer = types.LatestSigner(gspec.Config)
			)
			db, blocks, _ := core.GenerateChainWithGenesis(gspec, ethash.NewFaker(), nBlocks, func(ii int, block *core.BlockGen) {
				k1 := common.BigToHash(big.NewInt(int64(ii)))
				k2 := common.BigToHash(big.NewInt(int64(ii + 1)))
				v := common.BigToHash(big.NewInt(int64(ii + 2)))
				input, err := ABI.Pack("set", k1, k2, v)
				if err != nil {
					t.Fatal(err)
				}
				tx := types.NewTransaction(block.TxNonce(senderAddress), impl.address, common.Big0, 1e5, block.BaseFee(), input)
				signed, err := types.SignTx(tx, signer, key)
				if err != nil {
					t.Fatal(err)
				}
				block.AddTx(signed)
			})
			root := blocks[len(blocks)-1].Root()
			statedb, err := state.New(root, state.NewDatabase(db), nil)
			if err != nil {
				t.Fatal(err)
			}
			env := api.NewNoCallEnvironment(impl.address, api.EnvConfig{}, statedb, false, 0)
			kkv := lib.NewDatastore(env).Mapping([]byte("map.kkv.v1"))
			for ii := 0; ii < nBlocks; ii++ {
				k1 := common.BigToHash(big.NewInt(int64(ii)))
				k2 := common.BigToHash(big.NewInt(int64(ii + 1)))
				v := common.BigToHash(big.NewInt(int64(ii + 2)))
				value := kkv.Mapping(k1.Bytes()).Value(k2.Bytes()).GetBytes32()
				if value != v {
					t.Fatalf("expected %s, got %s", v, value)
				}
			}
		})
	}
}
