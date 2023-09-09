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
	fixture_datamod "github.com/ethereum/go-ethereum/concrete/fixtures/datamod"
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
	"github.com/stretchr/testify/require"
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
		address: common.BytesToAddress([]byte{130}),
		pc:      &fixtures.AdditionPrecompile{},
	},
	{
		name:    "Wazero",
		address: common.BytesToAddress([]byte{131}),
		pc:      wasm.NewWazeroPrecompile(addWasm),
	},
	{
		name:    "Wasmer",
		address: common.BytesToAddress([]byte{132}),
		pc:      wasm.NewWasmerPrecompile(addWasm),
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
		r   = require.New(t)
		ABI = getAddABI()
		x   = big.NewInt(1)
		y   = big.NewInt(2)
	)
	for _, impl := range addImplementations {
		precompiles.AddPrecompile(impl.address, impl.pc)
	}
	for _, impl := range addImplementations {
		env := mock.NewMockEnvironment(impl.address, api.EnvConfig{Trusted: true}, false, 0)
		t.Run(impl.name, func(t *testing.T) {
			input, err := ABI.Pack("add", x, y)
			r.NoError(err)
			isStatic := impl.pc.IsStatic(input)
			r.NoError(err)
			r.True(isStatic)
			output, err := impl.pc.Run(env, input)
			r.NoError(err)
			values, err := ABI.Methods["add"].Outputs.Unpack(output)
			r.NoError(err)
			value := values[0].(*big.Int)
			r.True(value.Cmp(x.Add(x, y)) == 0)
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
		address: common.BytesToAddress([]byte{140}),
		pc:      &fixtures.KeyKeyValuePrecompile{},
	},
	{
		name:    "Wazero",
		address: common.BytesToAddress([]byte{141}),
		pc:      wasm.NewWazeroPrecompile(kkvWasm),
	},
	{
		name:    "Wasmer",
		address: common.BytesToAddress([]byte{142}),
		pc:      wasm.NewWasmerPrecompile(kkvWasm),
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
		r   = require.New(t)
		ABI = getKkvABI()
		k1  = common.HexToHash("0x01")
		k2  = common.HexToHash("0x02")
		v   = common.HexToHash("0x03")
	)
	for _, impl := range kkvImplementations {
		precompiles.AddPrecompile(impl.address, impl.pc)
	}
	for _, impl := range kkvImplementations {
		env := mock.NewMockEnvironment(impl.address, api.EnvConfig{Trusted: true}, false, 0)
		t.Run(impl.name, func(t *testing.T) {
			input, err := ABI.Pack("set", k1, k2, v)
			r.NoError(err)
			isStatic := impl.pc.IsStatic(input)
			r.NoError(err)
			r.False(isStatic)
			_, err = impl.pc.Run(env, input)
			r.NoError(err)
			input, err = ABI.Pack("get", k1, k2)
			r.NoError(err)
			isStatic = impl.pc.IsStatic(input)
			r.NoError(err)
			r.True(isStatic)
			output, err := impl.pc.Run(env, input)
			r.NoError(err)
			values, err := ABI.Methods["get"].Outputs.Unpack(output)
			r.NoError(err)
			value := common.Hash(values[0].([32]byte))
			r.Equal(v, value)
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
		address: common.BytesToAddress([]byte{150}),
		pc:      &fixtures.KeyKeyValuePrecompile{},
	},
	{
		name:    "Wazero",
		address: common.BytesToAddress([]byte{151}),
		pc:      wasm.NewWazeroPrecompile(kkvWasm),
	},
	{
		name:    "Wasmer",
		address: common.BytesToAddress([]byte{151}),
		pc:      wasm.NewWasmerPrecompile(kkvWasm),
	},
}

func TestE2EKkvPrecompile(t *testing.T) {
	var (
		r             = require.New(t)
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
				r.NoError(err)
				tx := types.NewTransaction(block.TxNonce(senderAddress), impl.address, common.Big0, 1e5, block.BaseFee(), input)
				signed, err := types.SignTx(tx, signer, key)
				r.NoError(err)
				block.AddTx(signed)
			})
			root := blocks[len(blocks)-1].Root()
			statedb, err := state.New(root, state.NewDatabase(db), nil)
			r.NoError(err)
			env := api.NewNoCallEnvironment(impl.address, api.EnvConfig{}, statedb, false, 0)
			kkv := fixture_datamod.NewKkv(lib.NewDatastore(env))
			for ii := 0; ii < nBlocks; ii++ {
				k1 := common.BigToHash(big.NewInt(int64(ii)))
				k2 := common.BigToHash(big.NewInt(int64(ii + 1)))
				v := common.BigToHash(big.NewInt(int64(ii + 2)))
				value := kkv.Get(k1, k2).GetValue()
				r.Equal(v, value)
			}
		})
	}
}
