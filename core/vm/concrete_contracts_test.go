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

package vm

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm/concrete"
	"github.com/ethereum/go-ethereum/core/vm/concrete/api"
	cc_api "github.com/ethereum/go-ethereum/core/vm/concrete/api"
	"github.com/ethereum/go-ethereum/core/vm/concrete/lib"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	TestConcretePrecompileAddress = common.BytesToAddress([]byte{255})
	EphemeralValueID              = crypto.Keccak256Hash([]byte{1})
	PersistentValueID             = crypto.Keccak256Hash([]byte{2})
)

func newTestConcreteAPI() (api.API, api.EVM, *state.StateDB) {
	stateDB, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
	evm := lib.NewMockEVM(stateDB)
	api := api.New(evm, TestConcretePrecompileAddress)
	return api, evm, stateDB
}

type testPC struct{}

func (op *testPC) MutatesStorage(input []byte) bool {
	return new(big.Int).SetBytes(input).Cmp(big.NewInt(0)) == 1
}

func (op *testPC) RequiredGas(input []byte) uint64 {
	return new(big.Int).SetBytes(input).Uint64()
}

func (op *testPC) New(api cc_api.API) error {
	api.Ephemeral().Set(EphemeralValueID, common.BytesToHash(common.Big1.Bytes()))
	return nil
}

func (op *testPC) Commit(api cc_api.API) error {
	api.Persistent().Set(PersistentValueID, common.BytesToHash(common.Big2.Bytes()))
	return nil
}

func (op *testPC) Run(db cc_api.API, input []byte) ([]byte, error) {
	return common.Big0.Bytes(), nil
}

func TestConcretePrecompiledNewCommit(t *testing.T) {

	concrete.Precompiles = make(map[common.Address]api.Precompile)
	concrete.Precompiles[TestConcretePrecompileAddress] = &testPC{}

	api, _, stateDB := newTestConcreteAPI()

	if val := api.Ephemeral().Get(EphemeralValueID); val != common.BytesToHash(common.Big1.Bytes()) {
		t.Error("Expected ephemeral value to be 1 but got", val)
	}

	if val := api.Persistent().Get(PersistentValueID); val != (common.Hash{}) {
		t.Error("Persistent value should be empty but got", val)
	}

	stateDB.Commit(false)

	if val := api.Ephemeral().Get(EphemeralValueID); val != (common.Hash{}) {
		t.Error("Ephemeral value should be empty but got", val)
	}

	if val := api.Persistent().Get(PersistentValueID); val != common.BytesToHash(common.Big2.Bytes()) {
		t.Error("Persistent value should be 2 but got", val)
	}

}

func TestConcretePrecompiledRequiredGas(t *testing.T) {

	_, evm, _ := newTestConcreteAPI()

	_, remGas, err := RunConcretePrecompile(evm, TestConcretePrecompileAddress, &testPC{}, common.Big0.Bytes(), 0, false)

	if err != nil {
		t.Error(err)
	}

	if remGas != 0 {
		t.Error("Expected remaining gas to be 0 but got", remGas)
	}

	_, remGas, err = RunConcretePrecompile(evm, TestConcretePrecompileAddress, &testPC{}, common.Big1.Bytes(), 0, false)

	if err != ErrOutOfGas {
		t.Error("Expected out of gas error but got", err)
	}

	if remGas != 0 {
		t.Error("Expected remaining gas to be 0 but got", remGas)
	}

	_, remGas, err = RunConcretePrecompile(evm, TestConcretePrecompileAddress, &testPC{}, common.Big1.Bytes(), 1, false)

	if err != nil {
		t.Error(err)
	}

	if remGas != 0 {
		t.Error("Expected remaining gas to be 0 but got", remGas)
	}

	_, remGas, err = RunConcretePrecompile(evm, TestConcretePrecompileAddress, &testPC{}, common.Big1.Bytes(), 2, false)

	if err != nil {
		t.Error(err)
	}

	if remGas != 1 {
		t.Error("Expected remaining gas to be 1 but got", remGas)
	}
}

func TestConcretePrecompiledMutatesStorage(t *testing.T) {

	_, evm, _ := newTestConcreteAPI()

	_, _, err := RunConcretePrecompile(evm, TestConcretePrecompileAddress, &testPC{}, common.Big0.Bytes(), 1, false)

	if err != nil {
		t.Error(err)
	}

	_, _, err = RunConcretePrecompile(evm, TestConcretePrecompileAddress, &testPC{}, common.Big0.Bytes(), 1, true)

	if err != nil {
		t.Error(err)
	}

	_, _, err = RunConcretePrecompile(evm, TestConcretePrecompileAddress, &testPC{}, common.Big1.Bytes(), 1, false)

	if err != nil {
		t.Error(err)
	}

	_, _, err = RunConcretePrecompile(evm, TestConcretePrecompileAddress, &testPC{}, common.Big1.Bytes(), 1, true)

	if err != ErrWriteProtection {
		t.Error("Expected write protection error but got", err)
	}
}
