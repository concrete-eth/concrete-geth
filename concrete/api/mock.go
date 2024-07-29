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

// This file will be ignored when building with tinygo to prevent compatibility
// issues.

package api

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
)

type mockStateDB struct {
	stateDB           *state.StateDB
	code              map[common.Address][]byte
	balances          map[common.Address]*uint256.Int
	externalBalances  map[common.Address]*uint256.Int
	externalCodes     map[common.Address][]byte
	externalCodeHashes map[common.Address]common.Hash
	logs              []*types.Log
	state             map[common.Address]map[common.Hash]common.Hash
}

func NewMockStateDB() *mockStateDB {
	statedb, err := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
	if err != nil {
		panic(err)
	}
	return &mockStateDB{
		stateDB:           statedb,
		code:              make(map[common.Address][]byte),
		balances:          make(map[common.Address]*uint256.Int),
		externalBalances:  make(map[common.Address]*uint256.Int),
		externalCodes:     make(map[common.Address][]byte),
		externalCodeHashes: make(map[common.Address]common.Hash),
		logs:              []*types.Log{},
		state:             make(map[common.Address]map[common.Hash]common.Hash),
	}
}

func (m *mockStateDB) AddressInAccessList(addr common.Address) bool { return false }
func (m *mockStateDB) SlotInAccessList(addr common.Address, slot common.Hash) (bool, bool) {
	return false, false
}
func (m *mockStateDB) AddAddressToAccessList(addr common.Address)                {}
func (m *mockStateDB) AddSlotToAccessList(addr common.Address, slot common.Hash) {}
func (m *mockStateDB) GetCode(addr common.Address) []byte                        { return m.code[addr] }
func (m *mockStateDB) GetCodeSize(addr common.Address) int                       { return len(m.code[addr]) }
func (m *mockStateDB) GetCodeHash(addr common.Address) common.Hash               { return common.BytesToHash(m.code[addr]) }
func (m *mockStateDB) GetBalance(addr common.Address) *uint256.Int               { return m.balances[addr] }
func (m *mockStateDB) AddLog(log *types.Log)                                     { m.logs = append(m.logs, log) }

func (m *mockStateDB) GetCommittedState(addr common.Address, key common.Hash) common.Hash {
	return m.state[addr][key]
}

func (m *mockStateDB) SetState(addr common.Address, key common.Hash, value common.Hash) {
	if m.state[addr] == nil {
		m.state[addr] = make(map[common.Hash]common.Hash)
	}
	m.state[addr][key] = value
}

func (m *mockStateDB) GetState(addr common.Address, key common.Hash) common.Hash {
	return m.state[addr][key]
}

func (m *mockStateDB) SetTransientState(addr common.Address, key common.Hash, value common.Hash) {}
func (m *mockStateDB) GetTransientState(addr common.Address, key common.Hash) common.Hash {
	return common.Hash{}
}

func (m *mockStateDB) AddRefund(uint64)  {}
func (m *mockStateDB) SubRefund(uint64)  {}
func (m *mockStateDB) GetRefund() uint64 { return 0 }

func (m *mockStateDB) GetExternalBalance(addr common.Address) *uint256.Int {
	return m.externalBalances[addr]
}

func (m *mockStateDB) GetExternalCode(addr common.Address) []byte {
	return m.externalCodes[addr]
}

func (m *mockStateDB) GetExternalCodeSize(addr common.Address) int {
	return len(m.externalCodes[addr])
}

func (m *mockStateDB) GetExternalCodeHash(addr common.Address) common.Hash {
	return m.externalCodeHashes[addr]
}

var _ StateDB = (*mockStateDB)(nil)

type mockEnvConfig struct {
	envConfig EnvConfig
	meterGas  bool
	contract  *Contract
	db        StateDB
	blockCtx  BlockContext
	caller    Caller
}

type MockEnvOption func(*mockEnvConfig)

func WithConfig(config EnvConfig) MockEnvOption {
	return func(c *mockEnvConfig) {
		c.envConfig = config
	}
}

func WithStatic(static bool) MockEnvOption {
	return func(c *mockEnvConfig) {
		c.envConfig.IsStatic = static
	}
}

func WithTrusted(trusted bool) MockEnvOption {
	return func(c *mockEnvConfig) {
		c.envConfig.IsTrusted = trusted
	}
}

func WithMeterGas(meterGas bool) MockEnvOption {
	return func(c *mockEnvConfig) {
		c.meterGas = meterGas
	}
}

func WithContract(contract *Contract) MockEnvOption {
	return func(c *mockEnvConfig) {
		c.contract = contract
	}
}

func WithStateDB(statedb StateDB) MockEnvOption {
	return func(c *mockEnvConfig) {
		c.db = statedb
	}
}

func WithBlockCtx(blockCtx BlockContext) MockEnvOption {
	return func(c *mockEnvConfig) {
		c.blockCtx = blockCtx
	}
}

func WithCaller(caller Caller) MockEnvOption {
	return func(c *mockEnvConfig) {
		c.caller = caller
	}
}

func NewMockEnvironment(opts ...MockEnvOption) (*Env, StateDB, BlockContext, Caller) {
	mockConfig := &mockEnvConfig{
		envConfig: EnvConfig{},
		meterGas:  false,
	}

	for _, opt := range opts {
		opt(mockConfig)
	}

	if mockConfig.contract == nil {
		mockConfig.contract = &Contract{GasPrice: uint256.NewInt(0), Value: uint256.NewInt(0), Input: []byte{}}
	}
	if mockConfig.db == nil {
		mockConfig.db = NewMockStateDB()
	}
	if mockConfig.blockCtx == nil {
		mockConfig.blockCtx = NewMockBlockContext()
	}
	if mockConfig.caller == nil {
		mockConfig.caller = NewMockCaller()
	}

	env := NewEnvironment(mockConfig.envConfig, mockConfig.meterGas, mockConfig.db, mockConfig.blockCtx, mockConfig.caller, mockConfig.contract)

	env._execute = func(op OpCode, env *Env, args [][]byte) ([][]byte, error) {
		switch op {
		case GetBalance_OpCode:
			address := common.BytesToAddress(args[0])
			balance := env.statedb.GetBalance(address)
			return [][]byte{balance.Bytes()}, nil
		case GetCode_OpCode:
			address := common.BytesToAddress(args[0])
			code := env.statedb.GetCode(address)
			return [][]byte{code}, nil
		case GetCodeSize_OpCode:
			address := common.BytesToAddress(args[0])
			code := env.statedb.GetCode(address)
			size := uint64(len(code))
			return [][]byte{uint256.NewInt(size).Bytes()}, nil
		case GetExternalBalance_OpCode:
			address := common.BytesToAddress(args[0])
			balance := env.statedb.GetExternalBalance(address)
			return [][]byte{balance.Bytes()}, nil
		case GetExternalCode_OpCode:
			address := common.BytesToAddress(args[0])
			code := env.statedb.GetExternalCode(address)
			return [][]byte{code}, nil
		case GetExternalCodeSize_OpCode:
			address := common.BytesToAddress(args[0])
			code := env.statedb.GetExternalCode(address)
			size := uint64(len(code))
			return [][]byte{uint256.NewInt(size).Bytes()}, nil
		case GetExternalCodeHash_OpCode:
			address := common.BytesToAddress(args[0])
			hash := env.statedb.GetExternalCodeHash(address)
			return [][]byte{hash.Bytes()}, nil
		case StorageLoad_OpCode:
			key := common.BytesToHash(args[0])
			value := env.statedb.GetState(env.contract.Address, key)
			return [][]byte{value.Bytes()}, nil
		case StorageStore_OpCode:
			key := common.BytesToHash(args[0])
			value := common.BytesToHash(args[1])
			env.statedb.SetState(env.contract.Address, key, value)
			return nil, nil
		case Log_OpCode:
			topics := make([]common.Hash, len(args)-1)
			for i := 0; i < len(args)-1; i++ {
				topics[i] = common.BytesToHash(args[i])
			}
			data := args[len(args)-1]
			env.statedb.AddLog(&types.Log{
				Address: env.contract.Address,
				Topics:  topics,
				Data:    data,
			})
			return nil, nil
		}
		return nil, nil
	}

	return env, mockConfig.db, mockConfig.blockCtx, mockConfig.caller
}

type mockBlockContext struct {
	coinbase    common.Address
	gasLimit    uint64
	blockNumber uint64
	time        uint64
	difficulty  *uint256.Int
	baseFee     *uint256.Int
	random      common.Hash
	blockHashes map[uint64]common.Hash
}

func NewMockBlockContext() *mockBlockContext {
	return &mockBlockContext{
		coinbase:    common.Address{},
		gasLimit:    0,
		blockNumber: 0,
		time:        0,
		difficulty:  uint256.NewInt(0),
		baseFee:     uint256.NewInt(0),
		random:      common.Hash{},
		blockHashes: make(map[uint64]common.Hash),
	}
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

func (m *mockBlockContext) SetBlockHash(blockNumber uint64, hash common.Hash) {
	m.blockHashes[blockNumber] = hash
}

func (m *mockBlockContext) GetHash(blockNumber uint64) common.Hash { return m.blockHashes[blockNumber] }
func (m *mockBlockContext) GasLimit() uint64                       { return m.gasLimit }
func (m *mockBlockContext) BlockNumber() uint64                    { return m.blockNumber }
func (m *mockBlockContext) Timestamp() uint64                      { return m.time }
func (m *mockBlockContext) Difficulty() *uint256.Int               { return m.difficulty }
func (m *mockBlockContext) BaseFee() *uint256.Int                  { return m.baseFee }
func (m *mockBlockContext) Coinbase() common.Address               { return m.coinbase }
func (m *mockBlockContext) Random() common.Hash                    { return m.random }

var _ BlockContext = (*mockBlockContext)(nil)

type mockCaller struct {
	callFn         func(common.Address, []byte, uint64, *uint256.Int) ([]byte, uint64, error)
	callStaticFn   func(common.Address, []byte, uint64) ([]byte, uint64, error)
	callDelegateFn func(common.Address, []byte, uint64) ([]byte, uint64, error)
	createFn       func([]byte, uint64, *uint256.Int) ([]byte, common.Address, uint64, error)
	create2Fn      func([]byte, uint64, *uint256.Int, *uint256.Int) ([]byte, common.Address, uint64, error)
}

func NewMockCaller() *mockCaller {
	return &mockCaller{}
}

func (c *mockCaller) SetCallFn(fn func(common.Address, []byte, uint64, *uint256.Int) ([]byte, uint64, error)) {
	c.callFn = fn
}

func (c *mockCaller) SetCallStaticFn(fn func(common.Address, []byte, uint64) ([]byte, uint64, error)) {
	c.callStaticFn = fn
}

func (c *mockCaller) SetCallDelegateFn(fn func(common.Address, []byte, uint64) ([]byte, uint64, error)) {
	c.callDelegateFn = fn
}

func (c *mockCaller) SetCreateFn(fn func([]byte, uint64, *uint256.Int) ([]byte, common.Address, uint64, error)) {
	c.createFn = fn
}

func (c *mockCaller) SetCreate2Fn(fn func([]byte, uint64, *uint256.Int, *uint256.Int) ([]byte, common.Address, uint64, error)) {
	c.create2Fn = fn
}

func (c *mockCaller) Call(addr common.Address, input []byte, gas uint64, value *uint256.Int) ([]byte, uint64, error) {
	if c.callFn == nil {
		return nil, 0, nil
	}
	return c.callFn(addr, input, gas, value)
}

func (c *mockCaller) CallStatic(addr common.Address, input []byte, gas uint64) ([]byte, uint64, error) {
	if c.callStaticFn == nil {
		return nil, 0, nil
	}
	return c.callStaticFn(addr, input, gas)
}

func (c *mockCaller) CallDelegate(addr common.Address, input []byte, gas uint64) ([]byte, uint64, error) {
	if c.callDelegateFn == nil {
		return nil, 0, nil
	}
	return c.callDelegateFn(addr, input, gas)
}

func (c *mockCaller) Create(input []byte, gas uint64, value *uint256.Int) ([]byte, common.Address, uint64, error) {
	if c.createFn == nil {
		return nil, common.Address{}, 0, nil
	}
	return c.createFn(input, gas, value)
}

func (c *mockCaller) Create2(input []byte, gas uint64, endowment *uint256.Int, salt *uint256.Int) ([]byte, common.Address, uint64, error) {
	if c.create2Fn == nil {
		return nil, common.Address{}, 0, nil
	}
	return c.create2Fn(input, gas, endowment, salt)
}

var _ Caller = (*mockCaller)(nil)