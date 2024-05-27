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

package api

import (
	"bytes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/utils"
	"github.com/ethereum/go-ethereum/log"
	"github.com/holiman/uint256"
)

type Environment interface {
	Execute(op OpCode, args [][]byte) [][]byte

	// Meta
	EnableGasMetering(meter bool)
	Debug(msg string) // TODO: improve
	Debugf(msg string, ctx ...interface{})
	TimeNow() uint64

	// Utils
	Keccak256(data []byte) common.Hash
	// Flow control
	Revert(error)
	// Gas
	UseGas(amount uint64)

	// Ephemeral
	// EphemeralLoad_Unsafe(key common.Hash) common.Hash
	// EphemeralStore_Unsafe(key common.Hash, value common.Hash)

	// Local - READ
	// Address
	GetAddress() common.Address
	// Gas
	GetGasLeft() uint64
	// Block
	GetBlockNumber() uint64
	GetBlockGasLimit() uint64
	GetBlockTimestamp() uint64
	GetBlockDifficulty() *uint256.Int
	GetBlockBaseFee() *uint256.Int
	GetBlockCoinbase() common.Address
	GetPrevRandom() common.Hash
	// Block hash
	GetBlockHash(block uint64) common.Hash
	// Balance
	GetBalance(address common.Address) *uint256.Int
	// Transaction
	GetTxGasPrice() *uint256.Int
	GetTxOrigin() common.Address
	// Call
	GetCallData() []byte
	GetCallDataSize() int
	GetCaller() common.Address
	GetCallValue() *uint256.Int
	// Storage
	StorageLoad(key common.Hash) common.Hash
	// Code
	GetCode(address common.Address) []byte
	GetCodeSize() int

	// Local - WRITE
	// Storage
	StorageStore(key common.Hash, value common.Hash)
	// Log
	Log(topics []common.Hash, data []byte)

	// External - READ
	// Balance
	GetExternalBalance(address common.Address) *uint256.Int
	// Code
	GetExternalCode(address common.Address) []byte
	GetExternalCodeSize(address common.Address) int
	GetExternalCodeHash(address common.Address) common.Hash
	// Call
	CallStatic(address common.Address, data []byte, gas uint64) ([]byte, error)

	// External - WRITE
	// Call
	Call(address common.Address, data []byte, gas uint64, value *uint256.Int) ([]byte, error)
	CallDelegate(address common.Address, data []byte, gas uint64) ([]byte, error)
	// Create
	Create(data []byte, value *uint256.Int) (common.Address, error)
	Create2(data []byte, endowment *uint256.Int, salt *uint256.Int) (common.Address, error)
}

type EnvConfig struct {
	IsStatic bool
	// Ephemeral bool
	IsTrusted bool
}

type logger struct{}

func (logger) Debug(msg string) {
	log.Debug(msg)
}

var _ Logger = logger{}

type Contract struct {
	Address  common.Address
	Origin   common.Address
	Caller   common.Address
	GasPrice *uint256.Int
	Input    []byte
	Gas      uint64
	Value    *uint256.Int
}

func NewContract(origin, caller, address common.Address, gasPrice *uint256.Int) *Contract {
	return &Contract{
		Address:  address,
		Origin:   origin,
		Caller:   caller,
		GasPrice: gasPrice,
	}
}

func (c *Contract) useGas(gas uint64) bool {
	if c.Gas < gas {
		return false
	}
	c.Gas -= gas
	return true
}

type Env struct {
	table    JumpTable
	_execute func(op OpCode, env *Env, args [][]byte) ([][]byte, error)

	config   EnvConfig
	meterGas bool

	logger  Logger
	statedb StateDB
	block   BlockContext
	caller  Caller

	contract *Contract

	revertErr   error
	callGasTemp uint64
}

func NewEnvironment(
	config EnvConfig,
	meterGas bool,
	statedb StateDB,
	block BlockContext,
	caller Caller,
	contract *Contract,
) *Env {
	return &Env{
		table:    newEnvironmentMethods(),
		_execute: execute,
		config:   config,
		meterGas: meterGas,
		logger:   logger{},
		statedb:  statedb,
		block:    block,
		caller:   caller,
		contract: contract,
	}
}

func NewProxyEnvironment(execute func(op OpCode, env *Env, args [][]byte) ([][]byte, error)) *Env {
	return &Env{_execute: execute}
}

func execute(op OpCode, env *Env, args [][]byte) ([][]byte, error) {
	operation := env.table[op]

	if !env.config.IsTrusted && operation.trusted {
		return nil, ErrEnvNotTrusted
	}
	if env.config.IsStatic && !operation.static {
		return nil, ErrWriteProtection
	}

	if env.meterGas {
		gasConst := operation.constantGas
		if ok := env.useGas(gasConst); !ok {
			return nil, ErrOutOfGas
		}
		if operation.dynamicGas != nil {
			gasDyn, err := operation.dynamicGas(env, args)
			if err != nil {
				return nil, err
			}
			if ok := env.useGas(gasDyn); !ok {
				return nil, ErrOutOfGas
			}
		}
	}

	return operation.execute(env, args)
}

func (env *Env) execute(op OpCode, args [][]byte) [][]byte {
	ret, err := env._execute(op, env, args)
	if err != nil {
		panic(err)
	}
	return ret
}

func (env *Env) Config() EnvConfig {
	return env.config
}

func (env *Env) Caller() Caller {
	return env.caller
}

func (env *Env) Contract() *Contract {
	return env.contract
}

func (env *Env) RevertError() error {
	return env.revertErr
}

func (env *Env) Gas() uint64 {
	return env.contract.Gas
}

func (env *Env) useGas(gas uint64) bool {
	return env.contract.useGas(gas)
}

func (env *Env) BlockContext() BlockContext {
	return env.block
}

func (env *Env) Execute(op OpCode, args [][]byte) [][]byte {
	return env.execute(op, args)
}

func (env *Env) EnableGasMetering(meter bool) {
	input := [][]byte{{0x00}}
	if meter {
		input[0][0] = byte(0x01)
	}
	env.execute(EnableGasMetering_OpCode, input)
}

func (env *Env) Debug(msg string) {
	input := [][]byte{[]byte(msg)}
	env.execute(Debug_OpCode, input)
}

func (env *Env) Debugf(msg string, ctx ...interface{}) {
	formattedMsg := debugfFormat(msg, ctx...)
	env.Debug(formattedMsg)
}

func debugfFormat(msg string, ctx ...interface{}) string {
	var buf bytes.Buffer
	logger := log.NewLogger(log.NewTerminalHandlerWithLevel(&buf, log.LevelDebug, true))
	logger.Debug(msg, ctx...)
	return buf.String()
}

func (env *Env) TimeNow() uint64 {
	output := env.execute(TimeNow_OpCode, nil)
	return utils.BytesToUint64(output[0])
}

func (env *Env) UseGas(gas uint64) {
	input := [][]byte{utils.Uint64ToBytes(gas)}
	env.execute(UseGas_OpCode, input)
}

func (env *Env) Revert(err error) {
	input := [][]byte{[]byte(err.Error())}
	env.execute(Revert_OpCode, input)
}

func (env *Env) Keccak256(data []byte) common.Hash {
	input := [][]byte{data}
	output := env.execute(Keccak256_OpCode, input)
	hash := common.BytesToHash(output[0])
	return hash
}

func (env *Env) GetAddress() common.Address {
	output := env.execute(GetAddress_OpCode, nil)
	return common.BytesToAddress(output[0])
}

func (env *Env) GetGasLeft() uint64 {
	output := env.execute(GetGasLeft_OpCode, nil)
	return utils.BytesToUint64(output[0])
}

func (env *Env) GetBlockNumber() uint64 {
	output := env.execute(GetBlockNumber_OpCode, nil)
	return utils.BytesToUint64(output[0])
}

func (env *Env) GetBlockGasLimit() uint64 {
	output := env.execute(GetBlockGasLimit_OpCode, nil)
	return utils.BytesToUint64(output[0])
}

func (env *Env) GetBlockTimestamp() uint64 {
	output := env.execute(GetBlockTimestamp_OpCode, nil)
	return utils.BytesToUint64(output[0])
}

func (env *Env) GetBlockDifficulty() *uint256.Int {
	output := env.execute(GetBlockDifficulty_OpCode, nil)
	return new(uint256.Int).SetBytes(output[0])
}

func (env *Env) GetBlockBaseFee() *uint256.Int {
	output := env.execute(GetBlockBaseFee_OpCode, nil)
	return new(uint256.Int).SetBytes(output[0])
}

func (env *Env) GetBlockCoinbase() common.Address {
	output := env.execute(GetBlockCoinbase_OpCode, nil)
	return common.BytesToAddress(output[0])
}

func (env *Env) GetPrevRandom() common.Hash {
	output := env.execute(GetPrevRandom_OpCode, nil)
	return common.BytesToHash(output[0])
}

func (env *Env) GetBlockHash(number uint64) common.Hash {
	input := [][]byte{utils.Uint64ToBytes(number)}
	output := env.execute(GetBlockHash_OpCode, input)
	return common.BytesToHash(output[0])
}

func (env *Env) GetBalance(address common.Address) *uint256.Int {
	input := [][]byte{address.Bytes()}
	output := env.execute(GetBalance_OpCode, input)
	return new(uint256.Int).SetBytes(output[0])
}

func (env *Env) GetTxGasPrice() *uint256.Int {
	output := env.execute(GetTxGasPrice_OpCode, nil)
	return new(uint256.Int).SetBytes(output[0])
}

func (env *Env) GetTxOrigin() common.Address {
	output := env.execute(GetTxOrigin_OpCode, nil)
	return common.BytesToAddress(output[0])
}

func (env *Env) GetCallData() []byte {
	output := env.execute(GetCallData_OpCode, nil)
	return output[0]
}

func (env *Env) GetCallDataSize() int {
	output := env.execute(GetCallDataSize_OpCode, nil)
	return int(utils.BytesToUint64(output[0]))
}

func (env *Env) GetCaller() common.Address {
	output := env.execute(GetCaller_OpCode, nil)
	return common.BytesToAddress(output[0])
}

func (env *Env) GetCallValue() *uint256.Int {
	output := env.execute(GetCallValue_OpCode, nil)
	return new(uint256.Int).SetBytes(output[0])
}

func (env *Env) StorageLoad(key common.Hash) common.Hash {
	input := [][]byte{key.Bytes()}
	output := env.execute(StorageLoad_OpCode, input)
	return common.BytesToHash(output[0])
}

func (env *Env) GetCode(address common.Address) []byte {
	input := [][]byte{address.Bytes()}
	output := env.execute(GetCode_OpCode, input)
	return output[0]
}

func (env *Env) GetCodeSize() int {
	output := env.execute(GetCodeSize_OpCode, nil)
	return int(utils.BytesToUint64(output[0]))
}

func (env *Env) StorageStore(key common.Hash, value common.Hash) {
	input := [][]byte{key.Bytes(), value.Bytes()}
	env.execute(StorageStore_OpCode, input)
}

func (env *Env) Log(topics []common.Hash, data []byte) {
	input := make([][]byte, len(topics)+1)
	for i := 0; i < len(topics); i++ {
		input[i] = topics[i].Bytes()
	}
	input[len(topics)] = data
	env.execute(Log_OpCode, input)
}

func (env *Env) GetExternalBalance(address common.Address) *uint256.Int {
	input := [][]byte{address.Bytes()}
	output := env.execute(GetExternalBalance_OpCode, input)
	return new(uint256.Int).SetBytes(output[0])
}

func (env *Env) CallStatic(address common.Address, data []byte, gas uint64) ([]byte, error) {
	input := [][]byte{utils.Uint64ToBytes(gas), address.Bytes(), data}
	output := env.execute(CallStatic_OpCode, input)
	return output[0], utils.DecodeError(output[1])
}

func (env *Env) GetExternalCode(address common.Address) []byte {
	input := [][]byte{address.Bytes()}
	output := env.execute(GetExternalCode_OpCode, input)
	return output[0]
}

func (env *Env) GetExternalCodeSize(address common.Address) int {
	input := [][]byte{address.Bytes()}
	output := env.execute(GetExternalCodeSize_OpCode, input)
	return int(utils.BytesToUint64(output[0]))
}

func (env *Env) GetExternalCodeHash(address common.Address) common.Hash {
	input := [][]byte{address.Bytes()}
	output := env.execute(GetExternalCodeHash_OpCode, input)
	return common.BytesToHash(output[0])
}

func (env *Env) Call(address common.Address, data []byte, gas uint64, value *uint256.Int) ([]byte, error) {
	input := [][]byte{utils.Uint64ToBytes(gas), address.Bytes(), value.Bytes(), data}
	output := env.execute(Call_OpCode, input)
	return output[0], utils.DecodeError(output[1])
}

func (env *Env) CallDelegate(address common.Address, data []byte, gas uint64) ([]byte, error) {
	input := [][]byte{utils.Uint64ToBytes(gas), address.Bytes(), data}
	output := env.execute(CallDelegate_OpCode, input)
	return output[0], utils.DecodeError(output[1])
}

func (env *Env) Create(data []byte, value *uint256.Int) (common.Address, error) {
	input := [][]byte{data, value.Bytes()}
	output := env.execute(Create_OpCode, input)
	return common.BytesToAddress(output[0]), utils.DecodeError(output[1])
}

func (env *Env) Create2(data []byte, endowment *uint256.Int, salt *uint256.Int) (common.Address, error) {
	input := [][]byte{data, endowment.Bytes(), salt.Bytes()}
	output := env.execute(Create2_OpCode, input)
	return common.BytesToAddress(output[0]), utils.DecodeError(output[1])
}

var _ Environment = (*Env)(nil)
