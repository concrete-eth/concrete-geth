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

//go:build !tinygo

// This file will ignored when building with tinygo to prevent compatibility
// issues.

package api

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
)

func NewMockEnvironment(config EnvConfig, meterGas bool, contract *Contract) *Env {
	return NewEnvironment(
		config,
		meterGas,
		NewMockStateDB(),
		NewMockBlockContext(),
		NewMockCaller(),
		contract,
	)
}

type mockStateDB struct{}

func NewMockStateDB() *mockStateDB { return &mockStateDB{} }

func (m *mockStateDB) AddressInAccessList(addr common.Address) bool { return false }
func (m *mockStateDB) SlotInAccessList(addr common.Address, slot common.Hash) (bool, bool) {
	return false, false
}
func (m *mockStateDB) AddAddressToAccessList(addr common.Address)                {}
func (m *mockStateDB) AddSlotToAccessList(addr common.Address, slot common.Hash) {}
func (m *mockStateDB) GetCode(addr common.Address) []byte                        { return []byte{} }
func (m *mockStateDB) GetCodeSize(addr common.Address) int                       { return 0 }
func (m *mockStateDB) GetCodeHash(addr common.Address) common.Hash               { return common.Hash{} }
func (m *mockStateDB) GetBalance(addr common.Address) *uint256.Int               { return uint256.NewInt(0) }
func (m *mockStateDB) AddLog(*types.Log)                                         {}

func (m *mockStateDB) GetCommittedState(addr common.Address, key common.Hash) common.Hash {
	return common.Hash{}
}
func (m *mockStateDB) SetState(addr common.Address, key common.Hash, value common.Hash) {}
func (m *mockStateDB) GetState(addr common.Address, key common.Hash) common.Hash {
	return common.Hash{}
}

func (m *mockStateDB) AddRefund(uint64)  {}
func (m *mockStateDB) SubRefund(uint64)  {}
func (m *mockStateDB) GetRefund() uint64 { return 0 }

var _ StateDB = (*mockStateDB)(nil)

type mockBlockContext struct{}

func NewMockBlockContext() *mockBlockContext { return &mockBlockContext{} }

func (m *mockBlockContext) GetHash(uint64) common.Hash { return common.Hash{} }
func (m *mockBlockContext) GasLimit() uint64           { return 0 }
func (m *mockBlockContext) BlockNumber() uint64        { return 0 }
func (m *mockBlockContext) Timestamp() uint64          { return 0 }
func (m *mockBlockContext) Difficulty() *uint256.Int   { return uint256.NewInt(0) }
func (m *mockBlockContext) BaseFee() *uint256.Int      { return uint256.NewInt(0) }
func (m *mockBlockContext) Coinbase() common.Address   { return common.Address{} }
func (m *mockBlockContext) Random() common.Hash        { return common.Hash{} }

var _ BlockContext = (*mockBlockContext)(nil)

type mockCaller struct{}

func NewMockCaller() *mockCaller {
	return &mockCaller{}
}

func (c *mockCaller) CallStatic(addr common.Address, input []byte, gas uint64) ([]byte, uint64, error) {
	return nil, 0, nil
}

func (c *mockCaller) Call(addr common.Address, input []byte, gas uint64, value *uint256.Int) ([]byte, uint64, error) {
	return nil, 0, nil
}

func (c *mockCaller) CallDelegate(addr common.Address, input []byte, gas uint64) ([]byte, uint64, error) {
	return nil, 0, nil
}

func (c *mockCaller) Create(input []byte, gas uint64, value *uint256.Int) ([]byte, common.Address, uint64, error) {
	return nil, common.Address{}, 0, nil
}

func (c *mockCaller) Create2(input []byte, gas uint64, endowment *uint256.Int, salt *uint256.Int) ([]byte, common.Address, uint64, error) {
	return nil, common.Address{}, 0, nil
}

var _ Caller = (*mockCaller)(nil)
