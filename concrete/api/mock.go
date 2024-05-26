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

type (
	// CanTransferFunc is the signature of a transfer guard function
	CanTransferFunc func(mockStateDB, common.Address, *uint256.Int) bool
	// TransferFunc is the signature of a transfer function
	TransferFunc func(mockStateDB, common.Address, common.Address, *uint256.Int)
	// GetHashFunc returns the n'th block hash in the blockchain
	// and is used by the BLOCKHASH EVM op code.
	GetHashFunc func(uint64) common.Hash
)

func NewMockEnvironment(config EnvConfig, meterGas bool, contract *Contract) *Env {
	return NewEnvironment(
		config,
		meterGas,
		NewMockStateDB(),
		NewMockBlockContext(
			common.Address{},
			0,
			0,
			0,
			uint256.NewInt(0),
			uint256.NewInt(0),
			common.Hash{},
		),
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

func (m *mockStateDB) SubBalance(common.Address, *uint256.Int) {}
func (m *mockStateDB) AddBalance(common.Address, *uint256.Int) {}

var _ StateDB = (*mockStateDB)(nil)

type mockBlockContext struct {
	CanTransfer CanTransferFunc
	Transfer    TransferFunc
	getHash     GetHashFunc
	L1CostFunc  types.L1CostFunc
	coinbase    common.Address
	gasLimit    uint64
	blockNumber uint64
	time        uint64
	difficulty  *uint256.Int
	baseFee     *uint256.Int
	random      common.Hash
}

func NewMockBlockContext(
	coinBase common.Address,
	gasLimit uint64,
	blockNumber uint64,
	time uint64,
	difficulty *uint256.Int,
	baseFee *uint256.Int,
	random common.Hash,
) *mockBlockContext {
	return &mockBlockContext{
		CanTransfer: CanTransfer,
		Transfer:    Transfer,
		getHash:     GetHash,
		L1CostFunc:  nil,
		coinbase:    coinBase,
		gasLimit:    gasLimit,
		blockNumber: blockNumber,
		time:        time,
		difficulty:  difficulty,
		baseFee:     baseFee,
		random:      random,
	}
}

func CanTransfer(db mockStateDB, addr common.Address, amount *uint256.Int) bool {
	return db.GetBalance(addr).Cmp(amount) >= 0
}

func Transfer(db mockStateDB, sender, recipient common.Address, amount *uint256.Int) {
	db.SubBalance(sender, amount)
	db.AddBalance(recipient, amount)
}

func GetHash(uint64) common.Hash {
	return common.Hash{}
}

func (m *mockBlockContext) SetGasLimit(gasLimit uint64) {
	m.gasLimit = gasLimit
}
func (m *mockBlockContext) SetBlockNumber(blockNumber uint64) {
	m.blockNumber = blockNumber
}
func (m *mockBlockContext) SetTimestamp(timeStamp uint64) {
	m.time = timeStamp
}
func (m *mockBlockContext) SetDifficulty(difficulty *uint256.Int) {
	m.difficulty = difficulty
}
func (m *mockBlockContext) SetBaseFee(baseFee *uint256.Int) {
	m.baseFee = baseFee
}
func (m *mockBlockContext) SetCoinbase(coinBase common.Address) {
	m.coinbase = coinBase
}
func (m *mockBlockContext) SetRandom(random common.Hash) {
	m.random = random
}

func (m *mockBlockContext) GetHash(blockNumber uint64) common.Hash { return m.getHash(blockNumber) }
func (m *mockBlockContext) GasLimit() uint64                       { return m.gasLimit }
func (m *mockBlockContext) BlockNumber() uint64                    { return m.blockNumber }
func (m *mockBlockContext) Timestamp() uint64                      { return m.time }

func (m *mockBlockContext) Difficulty() *uint256.Int { return m.difficulty }
func (m *mockBlockContext) BaseFee() *uint256.Int    { return m.baseFee }
func (m *mockBlockContext) Coinbase() common.Address { return m.coinbase }
func (m *mockBlockContext) Random() common.Hash      { return m.random }

var _ BlockContext = (*mockBlockContext)(nil)

type mockCallerOut struct {
	output []byte
	gas    uint64
	err    error
}

type mockCallerCreateOut struct {
	output []byte
	addr   common.Address
	gas    uint64
	err    error
}

type mockCaller struct {
	expectedCallOut         *mockCallerOut
	expectedCallStaticOut   *mockCallerOut
	expectedCallDelegateOut *mockCallerOut

	expectedCallCreateOut  *mockCallerCreateOut
	expectedCallCreate2Out *mockCallerCreateOut
}

func NewMockCaller() *mockCaller {
	return &mockCaller{
		expectedCallOut:         nil,
		expectedCallStaticOut:   nil,
		expectedCallDelegateOut: nil,
		expectedCallCreateOut:   nil,
		expectedCallCreate2Out:  nil,
	}
}

func (c *mockCaller) SetExpectedCallOut(out *mockCallerOut) {
	c.expectedCallOut = out
}

func (c *mockCaller) GetExpectedCallOut() *mockCallerOut {
	return c.expectedCallOut
}

func (c *mockCaller) SetExpectedCallStaticOut(out *mockCallerOut) {
	c.expectedCallStaticOut = out
}

func (c *mockCaller) GetExpectedCallStaticOut() *mockCallerOut {
	return c.expectedCallStaticOut
}

func (c *mockCaller) SetExpectedCallDelegateOut(out *mockCallerOut) {
	c.expectedCallDelegateOut = out
}

func (c *mockCaller) GetExpectedCallDelegateOut() *mockCallerOut {
	return c.expectedCallDelegateOut
}

func (c *mockCaller) SetExpectedCallCreateOut(out *mockCallerCreateOut) {
	c.expectedCallCreateOut = out
}

func (c *mockCaller) GetExpectedCallCreateOut() *mockCallerCreateOut {
	return c.expectedCallCreateOut
}

func (c *mockCaller) SetExpectedCallCreate2Out(out *mockCallerCreateOut) {
	c.expectedCallCreate2Out = out
}

func (c *mockCaller) GetExpectedCallCreate2Out() *mockCallerCreateOut {
	return c.expectedCallCreate2Out
}

func (c *mockCaller) CallStatic(addr common.Address, input []byte, gas uint64) ([]byte, uint64, error) {
	return c.expectedCallStaticOut.output, c.expectedCallStaticOut.gas, c.expectedCallStaticOut.err
}

func (c *mockCaller) Call(addr common.Address, input []byte, gas uint64, value *uint256.Int) ([]byte, uint64, error) {
	return c.expectedCallOut.output, c.expectedCallOut.gas, c.expectedCallOut.err
}

func (c *mockCaller) CallDelegate(addr common.Address, input []byte, gas uint64) ([]byte, uint64, error) {
	return c.expectedCallDelegateOut.output, c.expectedCallDelegateOut.gas, c.expectedCallDelegateOut.err
}

func (c *mockCaller) Create(input []byte, gas uint64, value *uint256.Int) ([]byte, common.Address, uint64, error) {
	return c.expectedCallCreateOut.output, c.expectedCallCreateOut.addr, c.expectedCallCreateOut.gas, c.expectedCallCreateOut.err
}

func (c *mockCaller) Create2(input []byte, gas uint64, endowment *uint256.Int, salt *uint256.Int) ([]byte, common.Address, uint64, error) {
	return c.expectedCallCreate2Out.output, c.expectedCallCreate2Out.addr, c.expectedCallCreate2Out.gas, c.expectedCallCreate2Out.err
}

var _ Caller = (*mockCaller)(nil)
