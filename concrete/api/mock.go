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
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/holiman/uint256"
)

func NewMockStateDB() StateDB {
	statedb, err := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
	if err != nil {
		panic(err)
	}
	return statedb
}

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
