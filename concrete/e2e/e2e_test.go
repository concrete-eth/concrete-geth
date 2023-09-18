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

package e2e

import (
	_ "embed"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/concrete"
	"github.com/ethereum/go-ethereum/concrete/api"
	fixture_datamod "github.com/ethereum/go-ethereum/concrete/e2e/datamod"
	"github.com/ethereum/go-ethereum/concrete/lib"
	"github.com/ethereum/go-ethereum/concrete/mock"
	"github.com/ethereum/go-ethereum/concrete/wasm"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"
	"github.com/wasmerio/wasmer-go/wasmer"
)

//go:embed build/add.wasm
var addWasm []byte

//go:embed build/kkv.wasm
var kkvWasm []byte

type pcImplementation struct {
	name    string
	address common.Address
	newPc   func() concrete.Precompile
}

func wazeroPrecompile(code []byte) concrete.Precompile {
	return wasm.NewWazeroPrecompileWithConfig(code, wazero.NewRuntimeConfigInterpreter())
}

func wasmerPrecompile(code []byte) concrete.Precompile {
	return wasm.NewWasmerPrecompileWithConfig(code, wasmer.NewConfig().UseSinglepassCompiler())
}

var addImplementations = []pcImplementation{
	{
		name:    "Native",
		address: common.BytesToAddress([]byte{130}),
		newPc:   func() concrete.Precompile { return &AdditionPrecompile{} },
	},
	{
		name:    "Wazero",
		address: common.BytesToAddress([]byte{131}),
		newPc:   func() concrete.Precompile { return wazeroPrecompile(addWasm) },
	},
	{
		name:    "Wasmer",
		address: common.BytesToAddress([]byte{132}),
		newPc:   func() concrete.Precompile { return wasmerPrecompile(addWasm) },
	},
}

func getAddABI() abi.ABI {
	addABI, err := abi.JSON(strings.NewReader(AddAbiString))
	if err != nil {
		panic(err)
	}
	return addABI
}

func TestAddPrecompileFixture(t *testing.T) {
	var (
		r        = require.New(t)
		ABI      = getAddABI()
		config   = api.EnvConfig{Trusted: true}
		meterGas = true
		gas      = uint64(1e3)
		x        = big.NewInt(1)
		y        = big.NewInt(2)
	)

	pack := func(x, y *big.Int) []byte {
		input, err := ABI.Pack("add", x, y)
		r.NoError(err)
		return input
	}

	unpack := func(output []byte) *big.Int {
		values, err := ABI.Methods["add"].Outputs.Unpack(output)
		r.NoError(err)
		value := values[0].(*big.Int)
		return value
	}

	for _, impl := range addImplementations {
		t.Run(impl.name, func(t *testing.T) {
			pc := impl.newPc()
			env := mock.NewMockEnvironment(impl.address, config, meterGas, gas)
			input := pack(x, y)
			isStatic := pc.IsStatic(input)
			r.True(isStatic)
			output, err := pc.Run(env, input)
			r.NoError(err)
			value := unpack(output)
			r.True(value.Cmp(x.Add(x, y)) == 0)
		})
	}
}

var kkvImplementations = []pcImplementation{
	{
		name:    "Native",
		address: common.BytesToAddress([]byte{140}),
		newPc:   func() concrete.Precompile { return &KeyKeyValuePrecompile{} },
	},
	{
		name:    "Wazero",
		address: common.BytesToAddress([]byte{141}),
		newPc:   func() concrete.Precompile { return wazeroPrecompile(kkvWasm) },
	},
	{
		name:    "Wasmer",
		address: common.BytesToAddress([]byte{142}),
		newPc:   func() concrete.Precompile { return wasmerPrecompile(kkvWasm) },
	},
}

func getKkvABI() abi.ABI {
	kkvABI, err := abi.JSON(strings.NewReader(KkvAbiString))
	if err != nil {
		panic(err)
	}
	return kkvABI
}

func TestKkvPrecompileFixture(t *testing.T) {
	var (
		r   = require.New(t)
		ABI = getKkvABI()
		k1  = common.HexToHash("0x01")
		k2  = common.HexToHash("0x02")
		v   = common.HexToHash("0x03")
	)

	packSet := func(k1, k2, v common.Hash) []byte {
		input, err := ABI.Pack("set", k1, k2, v)
		r.NoError(err)
		return input
	}

	packGet := func(k1, k2 common.Hash) []byte {
		input, err := ABI.Pack("get", k1, k2)
		r.NoError(err)
		return input
	}

	unpackGet := func(output []byte) common.Hash {
		values, err := ABI.Methods["get"].Outputs.Unpack(output)
		r.NoError(err)
		value := common.Hash(values[0].([32]byte))
		return value
	}

	for _, impl := range kkvImplementations {
		t.Run(impl.name, func(t *testing.T) {
			env := mock.NewMockEnvironment(impl.address, api.EnvConfig{Trusted: true}, true, 1e5)
			pc := impl.newPc()
			var (
				err              error
				isStatic         bool
				input, output    []byte
				gasLeft, gasUsed uint64
			)

			// Test Set
			input = packSet(k1, k2, v)
			isStatic = pc.IsStatic(input)
			r.False(isStatic)
			gasLeft = env.Gas()
			_, _, err = concrete.RunPrecompile(pc, env, input, false)
			r.NoError(err)
			gasUsed = gasLeft - env.Gas()
			r.Equal(params.ColdSloadCostEIP2929+params.SstoreSetGasEIP2200, gasUsed) // Cold SSTORE

			// Test Get
			input = packGet(k1, k2)
			isStatic = pc.IsStatic(input)
			r.True(isStatic)
			gasLeft = env.Gas()
			output, _, err = concrete.RunPrecompile(pc, env, input, true)
			r.NoError(err)
			gasUsed = gasLeft - env.Gas()
			r.Equal(params.WarmStorageReadCostEIP2929, gasUsed) // Warm SLOAD

			value := unpackGet(output)
			r.Equal(v, value)

			kkv := fixture_datamod.NewKkv(lib.NewDatastore(env))
			value = kkv.Get(k1, k2).GetValue()
			r.Equal(v, value)
		})
	}
}

func TestE2EKkvPrecompile(t *testing.T) {
	var (
		r             = require.New(t)
		ABI           = getKkvABI()
		key, _        = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
		senderAddress = crypto.PubkeyToAddress(key.PublicKey)
		gspec         = &core.Genesis{
			Config:   params.TestChainConfig,
			GasLimit: 30_000_000,
			Alloc: core.GenesisAlloc{
				senderAddress: {Balance: math.MaxBig256},
			},
		}
		signer     = types.LatestSigner(gspec.Config)
		nBlocks    = 3
		txGasLimit = uint64(1e5)
	)

	pcArgs := func(ii int) (k1, k2, v common.Hash) {
		k1 = common.BigToHash(big.NewInt(int64(ii)))
		k2 = common.BigToHash(big.NewInt(int64(ii + 1)))
		v = common.BigToHash(big.NewInt(int64(ii + 2)))
		return
	}

	for _, impl := range kkvImplementations {
		t.Run(impl.name, func(t *testing.T) {
			// Create registry with precompile implementation
			concreteRegistry := concrete.NewRegistry()
			concreteRegistry.AddPrecompile(0, impl.address, impl.newPc())

			// Generate chain calling precompile every block
			db, blocks, receipts := core.GenerateChainWithGenesisWithConcrete(gspec, ethash.NewFaker(), nBlocks, concreteRegistry, func(ii int, block *core.BlockGen) {
				// Create, sign and add tx calling precompile
				k1, k2, v := pcArgs(ii)
				input, err := ABI.Pack("set", k1, k2, v)
				r.NoError(err)
				tx := types.NewTransaction(block.TxNonce(senderAddress), impl.address, common.Big0, txGasLimit, block.BaseFee(), input)
				signed, err := types.SignTx(tx, signer, key)
				r.NoError(err)
				block.AddTx(signed)
			})

			// Check receipt status
			for _, blockReceipts := range receipts {
				for _, receipt := range blockReceipts {
					r.Equal(types.ReceiptStatusSuccessful, receipt.Status)
				}
			}

			// Re-create kkv at last block
			root := blocks[len(blocks)-1].Root()
			statedb, err := state.New(root, state.NewDatabase(db), nil)
			r.NoError(err)
			env := api.NewNoCallEnvironment(impl.address, api.EnvConfig{}, statedb, false, 0)
			kkv := fixture_datamod.NewKkv(lib.NewDatastore(env))

			// Check state
			for ii := 0; ii < nBlocks; ii++ {
				k1, k2, v := pcArgs(ii)
				value := kkv.Get(k1, k2).GetValue()
				r.Equal(v, value)
			}
		})
	}
}
