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
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
)

func TestGas(t *testing.T) {
	var (
		r        = require.New(t)
		config   = EnvConfig{IsStatic: false, IsTrusted: false}
		meterGas = true
		gas      = uint64(1e6)
	)

	env, _, _, _ := NewMockEnvironment(WithConfig(config), WithMeterGas(meterGas))
	env.contract.Gas = gas

	// GetGasLeft() costs gas, so the cost of that operation must be subtracted
	// from the total gas.
	getGasLeftOpCost := env.table[GetGasLeft_OpCode].constantGas
	gas -= getGasLeftOpCost
	r.Equal(gas, env.GetGasLeft())
	r.Equal(gas, env.Gas())
}

func TestBlockOps_Minimal(t *testing.T) {
	var (
		r        = require.New(t)
		config   = EnvConfig{IsStatic: false, IsTrusted: false}
		meterGas = true
		gas      = uint64(1e6)
	)

	env, _, _, _ := NewMockEnvironment(WithConfig(config), WithMeterGas(meterGas))
	env.contract.Gas = gas

	r.Equal(env.block.GetHash(0), env.GetBlockHash(0))
	r.Equal(env.block.GasLimit(), env.GetBlockGasLimit())
	r.Equal(env.block.BlockNumber(), env.GetBlockNumber())
	r.Equal(env.block.Timestamp(), env.GetBlockTimestamp())
	r.Equal(env.block.Difficulty(), env.GetBlockDifficulty())
	r.Equal(env.block.BaseFee(), env.GetBlockBaseFee())
	r.Equal(env.block.Coinbase(), env.GetBlockCoinbase())
	r.Equal(env.block.Random(), env.GetPrevRandom())
}

func TestCallOps_Minimal(t *testing.T) {
	var (
		r        = require.New(t)
		config   = EnvConfig{IsStatic: false, IsTrusted: false}
		meterGas = true
		gas      = uint64(1e6)
	)

	env, _, _, _ := NewMockEnvironment(WithConfig(config), WithMeterGas(meterGas))
	env.contract.Input = []byte{0x01, 0x02, 0x03}
	env.contract.Gas = gas
	env.contract.Value = uint256.NewInt(1)

	r.Equal(env.contract.GasPrice, env.GetTxGasPrice())
	r.Equal(env.contract.Origin, env.GetTxOrigin())
	r.Equal(env.contract.Input, env.GetCallData())
	r.Equal(len(env.contract.Input), env.GetCallDataSize())
	r.Equal(env.contract.Caller, env.GetCaller())
	r.Equal(env.contract.Value, env.GetCallValue())
}

func TestTrustAndWriteProtection(t *testing.T) {
	var (
		r        = require.New(t)
		config   = EnvConfig{IsStatic: true, IsTrusted: false}
		meterGas = false
	)

	env, _, _, _ := NewMockEnvironment(WithConfig(config), WithMeterGas(meterGas))
	env.contract.Input = []byte{0x01, 0x02, 0x03}
	env.contract.Value = uint256.NewInt(1)

	table := newEnvironmentMethods()
	for opcode, method := range table {
		err := func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = r.(error)
				}
			}()
			env.execute(OpCode(opcode), nil)
			return nil
		}()
		if method.trusted {
			r.Equal(ErrEnvNotTrusted, err)
		} else if !method.static {
			r.Equal(ErrWriteProtection, err)
		} else {
			if err != nil && err != ErrInvalidOpCode {
				if err != ErrInvalidInput {
					_, err = method.dynamicGas(env, nil)
					r.Equal(ErrInvalidInput, err)
				}
			}
		}
	}
}

func TestDebugf(t *testing.T) {
	var (
		r        = require.New(t)
		config   = EnvConfig{IsStatic: true, IsTrusted: true}
		meterGas = false
	)

	env, _, _, _ := NewMockEnvironment(WithConfig(config), WithMeterGas(meterGas))

	read, write, _ := os.Pipe()

	// Capture stderr
	stderr := os.Stderr
	os.Stderr = write
	defer func() {
		os.Stderr = stderr
	}()

	// Copy captured stderr to a buffer
	done := make(chan *bytes.Buffer)
	go func() {
		defer read.Close()
		var buf bytes.Buffer
		io.Copy(&buf, read)
		done <- &buf
	}()

	env.Debugf("Message", "arg1", 1, "arg2", "val2", "arg3", struct{ A int }{A: 3})

	write.Close()
	buf := <-done

	r.Equal("Message                                  \x1b[36marg1\x1b[0m=1 \x1b[36marg2\x1b[0m=val2 \x1b[36marg3\x1b[0m={A:3}\n", buf.String()[35:])
}

func TestBlockContextMethods(t *testing.T) {
	var (
		r        = require.New(t)
		config   = EnvConfig{IsStatic: true, IsTrusted: false}
		meterGas = true
		gas      = uint64(1e6)
	)

	env, _, _block, _ := NewMockEnvironment(WithConfig(config), WithMeterGas(meterGas))
	block := _block.(*mockBlockContext)

	t.Run("BlockHash", func(t *testing.T) {
		env.contract.Gas = gas

		var (
			currentBlockNumber = uint64(400)
			lastBlockNumber    = currentBlockNumber - 1
			recentBlockNumber  = currentBlockNumber - 2
			limitBlockNumber   = lastBlockNumber - 255
			limitBlockNumberS1 = limitBlockNumber - 1
			futureBlockNumber  = currentBlockNumber + 1
		)

		block.SetBlockNumber(currentBlockNumber)
		block.SetBlockHash(currentBlockNumber, common.Hash{0x01})
		block.SetBlockHash(lastBlockNumber, common.Hash{0x02})
		block.SetBlockHash(recentBlockNumber, common.Hash{0x03})
		block.SetBlockHash(limitBlockNumber, common.Hash{0x04})
		block.SetBlockHash(limitBlockNumberS1, common.Hash{0x05})
		block.SetBlockHash(futureBlockNumber, common.Hash{0x06})

		r.Equal(common.Hash{}, env.GetBlockHash(currentBlockNumber))
		r.Equal(block.GetHash(lastBlockNumber), env.GetBlockHash(lastBlockNumber))
		r.Equal(block.GetHash(recentBlockNumber), env.GetBlockHash(recentBlockNumber))
		r.Equal(block.GetHash(limitBlockNumber), env.GetBlockHash(limitBlockNumber))
		r.Equal(common.Hash{}, env.GetBlockHash(limitBlockNumberS1))
		r.Equal(common.Hash{}, env.GetBlockHash(futureBlockNumber))

		r.Equal(env.Gas(), gas-6*GasExtStep)
	})

	t.Run("BlockNumber", func(t *testing.T) {
		env.contract.Gas = gas

		blockNumber := uint64(123456)
		block.SetBlockNumber(blockNumber)

		r.Equal(blockNumber, env.GetBlockNumber())
		r.Equal(env.Gas(), gas-GasQuickStep)
	})

	t.Run("GasLimit", func(t *testing.T) {
		env.contract.Gas = gas

		blockGasLimit := uint64(8000000)
		block.SetGasLimit(blockGasLimit)

		r.Equal(blockGasLimit, env.GetBlockGasLimit())
		r.Equal(env.Gas(), gas-GasQuickStep)
	})

	t.Run("Timestamp", func(t *testing.T) {
		env.contract.Gas = gas

		blockTimestamp := uint64(1625097600)
		block.SetTimestamp(blockTimestamp)

		r.Equal(blockTimestamp, env.GetBlockTimestamp())
		r.Equal(env.Gas(), gas-GasQuickStep)
	})

	t.Run("Difficulty", func(t *testing.T) {
		env.contract.Gas = gas

		blockDifficulty := uint256.NewInt(5000000000)
		block.SetDifficulty(blockDifficulty)

		r.Equal(blockDifficulty, env.GetBlockDifficulty())
		r.Equal(env.Gas(), gas-GasQuickStep)
	})

	t.Run("BaseFee", func(t *testing.T) {
		env.contract.Gas = gas

		blockBaseFee := uint256.NewInt(1000000000)
		block.SetBaseFee(blockBaseFee)

		r.Equal(blockBaseFee, env.GetBlockBaseFee())
		r.Equal(env.Gas(), gas-GasQuickStep)
	})

	t.Run("Coinbase", func(t *testing.T) {
		env.contract.Gas = gas

		blockCoinbase := common.Address{0x01}
		block.SetCoinbase(blockCoinbase)

		r.Equal(blockCoinbase, env.GetBlockCoinbase())
		r.Equal(env.Gas(), gas-GasQuickStep)
	})

	t.Run("PrevRandom", func(t *testing.T) {
		env.contract.Gas = gas

		blockRandom := common.Hash{0x01}
		block.SetRandom(blockRandom)

		r.Equal(blockRandom, env.GetPrevRandom())
		r.Equal(env.Gas(), gas-GasQuickStep)
	})
}

func TestCallMethods(t *testing.T) {
	var (
		r        = require.New(t)
		config   = EnvConfig{IsStatic: false, IsTrusted: false}
		meterGas = true
	)

	t.Run("CallStatic", func(t *testing.T) {
		var (
			env, _, _, _caller = NewMockEnvironment(WithConfig(config), WithMeterGas(meterGas))
			caller             = _caller.(*mockCaller)
			gas                = uint64(1e6)
			callAddr           = common.Address{0x01}
			callInput          = []byte("input")
			useGas             = uint64(123)
			callGas            = uint64(456)
			callOutput         = []byte("outputCallStatic")
		)

		env.contract.Gas = gas

		caller.SetCallStaticFn(func(addr common.Address, input []byte, gas uint64) ([]byte, uint64, error) {
			r.Equal(callAddr, addr)
			r.Equal(callInput, input)
			r.Equal(callGas, gas)
			return callOutput, gas - useGas, nil
		})

		ret, err := env.CallStatic(callAddr, callInput, callGas)
		r.NoError(err)
		r.Equal(callOutput, ret)

		r.Equal(env.Gas(), gas-params.ColdAccountAccessCostEIP2929-useGas)
	})

	t.Run("CallStaticInsufficientGas", func(t *testing.T) {
		var (
			env, _, _, _caller = NewMockEnvironment(WithConfig(config), WithMeterGas(meterGas))
			caller             = _caller.(*mockCaller)
			gas                = uint64(3000)
			availableGas       = gas - params.ColdAccountAccessCostEIP2929
			expectedGas        = availableGas - availableGas/64
			callGas            = uint64(1000)
		)

		env.contract.Gas = gas

		caller.SetCallStaticFn(func(addr common.Address, input []byte, gas uint64) ([]byte, uint64, error) {
			r.Equal(expectedGas, gas)
			return nil, gas, nil
		})

		env.CallStatic(common.Address{}, nil, callGas)
	})

	t.Run("Call", func(t *testing.T) {
		var (
			env, _, _, _caller = NewMockEnvironment(WithConfig(config), WithMeterGas(meterGas))
			caller             = _caller.(*mockCaller)
			gas                = uint64(1e6)
			callAddr           = common.Address{0x01}
			callInput          = []byte("input")
			useGas             = uint64(123)
			callGas            = uint64(456)
			callValue          = uint256.NewInt(1)
			callOutput         = []byte("outputCall")
		)

		env.contract.Gas = gas

		caller.SetCallFn(func(addr common.Address, input []byte, gas uint64, value *uint256.Int) ([]byte, uint64, error) {
			r.Equal(callAddr, addr)
			r.Equal(callInput, input)
			r.Equal(callGas, gas)
			r.Equal(callValue, value)
			return callOutput, gas - useGas, nil
		})

		ret, err := env.Call(callAddr, callInput, callGas, callValue)
		r.NoError(err)
		r.Equal(callOutput, ret)

		r.Equal(env.Gas(), gas-params.ColdAccountAccessCostEIP2929-useGas)
	})

	t.Run("CallInsufficientGas", func(t *testing.T) {
		var (
			env, _, _, _caller = NewMockEnvironment(WithConfig(config), WithMeterGas(meterGas))
			caller             = _caller.(*mockCaller)
			gas                = uint64(3000)
			availableGas       = gas - params.ColdAccountAccessCostEIP2929
			expectedGas        = availableGas - availableGas/64
			callGas            = uint64(1000)
		)

		env.contract.Gas = gas

		caller.SetCallFn(func(addr common.Address, input []byte, gas uint64, value *uint256.Int) ([]byte, uint64, error) {
			r.Equal(expectedGas, gas)
			return nil, gas, nil
		})

		env.Call(common.Address{}, nil, callGas, new(uint256.Int))
	})

	t.Run("CallDelegate", func(t *testing.T) {
		var (
			env, _, _, _caller = NewMockEnvironment(WithConfig(config), WithMeterGas(meterGas))
			caller             = _caller.(*mockCaller)
			gas                = uint64(1e6)
			callAddr           = common.Address{0x01}
			callInput          = []byte("input")
			useGas             = uint64(123)
			callGas            = uint64(456)
			callOutput         = []byte("outputCallDelegate")
		)

		env.contract.Gas = gas

		caller.SetCallDelegateFn(func(addr common.Address, input []byte, gas uint64) ([]byte, uint64, error) {
			r.Equal(callAddr, addr)
			r.Equal(callInput, input)
			r.Equal(callGas, gas)
			return callOutput, gas - useGas, nil
		})

		ret, err := env.CallDelegate(callAddr, callInput, callGas)
		r.NoError(err)
		r.Equal(callOutput, ret)

		r.Equal(env.Gas(), gas-params.ColdAccountAccessCostEIP2929-useGas)
	})

	t.Run("CallDelegateInsufficientGas", func(t *testing.T) {
		var (
			env, _, _, _caller = NewMockEnvironment(WithConfig(config), WithMeterGas(meterGas))
			caller             = _caller.(*mockCaller)
			gas                = uint64(3000)
			availableGas       = gas - params.ColdAccountAccessCostEIP2929
			expectedGas        = availableGas - availableGas/64
			callGas            = uint64(1000)
		)

		env.contract.Gas = gas

		caller.SetCallDelegateFn(func(addr common.Address, input []byte, gas uint64) ([]byte, uint64, error) {
			r.Equal(expectedGas, gas)
			return nil, gas, nil
		})

		env.CallDelegate(common.Address{}, nil, callGas)
	})

	t.Run("Create", func(t *testing.T) {
		var (
			env, _, _, _caller = NewMockEnvironment(WithConfig(config), WithMeterGas(meterGas))
			caller             = _caller.(*mockCaller)
			gas                = uint64(1e6)
			createInput        = []byte("input")
			availableGas       = gas - params.CreateGas - (toWordSize(len(createInput)) * params.InitCodeWordGas)
			usedGas            = availableGas - availableGas/64
			createValue        = uint256.NewInt(1)
			createOutput       = []byte("createOutput")
			createAddr         = common.Address{0x01}
		)

		env.contract.Gas = gas

		caller.SetCreateFn(func(input []byte, gas uint64, value *uint256.Int) ([]byte, common.Address, uint64, error) {
			r.Equal(createInput, input)
			r.Equal(usedGas, gas)
			r.Equal(createValue, value)
			return createOutput, createAddr, gas, nil
		})

		ret, addr, err := env.Create(createInput, createValue)
		r.NoError(err)
		r.Equal(createOutput, ret)
		r.Equal(createAddr, addr)

		r.Equal(env.Gas(), availableGas)
	})

	t.Run("CreateInsufficientGas", func(t *testing.T) {
		var (
			env, _, _, _caller = NewMockEnvironment(WithConfig(config), WithMeterGas(meterGas))
			caller             = _caller.(*mockCaller)
			gas                = uint64(32005)
			createInput        = []byte("input")
			availableGas       = gas - params.CreateGas - (toWordSize(len(createInput)) * params.InitCodeWordGas)
			expectedGas        = availableGas - availableGas/64
		)

		env.contract.Gas = gas

		caller.SetCreateFn(func(input []byte, gas uint64, value *uint256.Int) ([]byte, common.Address, uint64, error) {
			r.Equal(expectedGas, gas)
			return nil, common.Address{}, gas, nil
		})

		env.Create(createInput, new(uint256.Int))
	})

	t.Run("Create2", func(t *testing.T) {
		var (
			env, _, _, _caller = NewMockEnvironment(WithConfig(config), WithMeterGas(meterGas))
			caller             = _caller.(*mockCaller)
			gas                = uint64(1e6)
			createInput        = []byte("input")
			availableGas       = gas - params.Create2Gas - (toWordSize(len(createInput)) * (params.InitCodeWordGas + params.Keccak256WordGas))
			usedGas            = availableGas - availableGas/64
			createValue        = uint256.NewInt(1)
			createSalt         = new(uint256.Int).SetBytes([]byte("salt"))
			createOutput       = []byte("create2Output")
			createAddr         = common.Address{0x01}
		)

		env.contract.Gas = gas

		caller.SetCreate2Fn(func(input []byte, gas uint64, value *uint256.Int, salt *uint256.Int) ([]byte, common.Address, uint64, error) {
			r.Equal(createInput, input)
			r.Equal(usedGas, gas)
			r.Equal(createValue, value)
			r.Equal(createSalt, salt)
			return createOutput, createAddr, gas, nil
		})

		ret, addr, err := env.Create2(createInput, createValue, createSalt)
		r.NoError(err)
		r.Equal(createOutput, ret)
		r.Equal(createAddr, addr)

		r.Equal(env.Gas(), availableGas)
	})

	t.Run("Create2InsufficientGas", func(t *testing.T) {
		var (
			env, _, _, _caller = NewMockEnvironment(WithConfig(config), WithMeterGas(meterGas))
			caller             = _caller.(*mockCaller)
			gas                = uint64(32010)
			createInput        = []byte("input")
			availableGas       = gas - params.Create2Gas - (toWordSize(len(createInput)) * (params.InitCodeWordGas + params.Keccak256WordGas))
			expectedGas        = availableGas - availableGas/64
		)

		env.contract.Gas = gas

		caller.SetCreate2Fn(func(input []byte, gas uint64, value *uint256.Int, salt *uint256.Int) ([]byte, common.Address, uint64, error) {
			r.Equal(expectedGas, gas)
			return nil, common.Address{}, gas, nil
		})

		env.Create2(createInput, new(uint256.Int), new(uint256.Int))
	})
}

func TestStateMethods(t *testing.T) {
	r := require.New(t)
	config := EnvConfig{IsStatic: false, IsTrusted: false}
	meterGas := true

	env, statedb, _, _ := NewMockEnvironment(WithConfig(config), WithMeterGas(meterGas))
	env.contract.Gas = uint64(1e6)

	t.Run("Storage", func(t *testing.T) {
		key := common.HexToHash("0x01")
		value := common.HexToHash("0x02")
		initialGas := env.contract.Gas

		env.StorageStore(key, value)
		r.True(initialGas > env.contract.Gas)
		usedGas := initialGas - env.contract.Gas
		loadedValue := env.StorageLoad(key)
		r.Equal(value, loadedValue)
		storedValue := statedb.GetState(env.Contract().Address, key)
		r.Equal(value, storedValue)
		t.Logf("Gas used for Storage: %d", usedGas)
	})

	t.Run("Log", func(t *testing.T) {
		topics := []common.Hash{common.HexToHash("0x01"), common.HexToHash("0x02")}
		data := []byte("log data")
		initialGas := env.contract.Gas

		env.Log(topics, data)
		r.True(initialGas > env.contract.Gas)
		usedGas := initialGas - env.contract.Gas

		log := &types.Log{
			Address: env.Contract().Address,
			Topics:  topics,
			Data:    data,
		}
		logs := statedb.(*state.StateDB).Logs()
		r.Len(logs, 1)
		r.Equal(log, logs[0])
		t.Logf("Gas used for Log: %d", usedGas)
	})

	t.Run("GetCode", func(t *testing.T) {
		code := []byte{0x60, 0x60, 0x60}
		statedb.(*state.StateDB).SetCode(env.Contract().Address, code)
		initialGas := env.contract.Gas

		loadedCode := env.GetCode()
		r.True(initialGas > env.contract.Gas)
		usedGas := initialGas - env.contract.Gas
		r.Equal(code, loadedCode)

		t.Logf("Gas used for GetCode: %d", usedGas)
	})

	t.Run("GetCodeSize", func(t *testing.T) {
		code := []byte{0x60, 0x60, 0x60}
		statedb.(*state.StateDB).SetCode(env.Contract().Address, code)
		initialGas := env.contract.Gas

		loadedCodeSize := env.GetCodeSize()
		codeSizeGasUsed := initialGas - env.contract.Gas
		expectedCodeSizeGas := GasQuickStep
		r.Equal(expectedCodeSizeGas, codeSizeGasUsed)
		r.Equal(len(code), loadedCodeSize)
	}) //worked

	t.Run("GetExternalBalance", func(t *testing.T) {
		address := common.HexToAddress("0x12345")
		balance := uint256.NewInt(5000)
		statedb.(*state.StateDB).SetBalance(address, balance)
		initialGas := env.contract.Gas

		loadedBalance := env.GetExternalBalance(address)
		r.True(initialGas > env.contract.Gas)
		usedGas := initialGas - env.contract.Gas
		r.Equal(balance, loadedBalance)
		t.Logf("Gas used for GetExternalBalance: %d", usedGas)
	}) //didnt work

	t.Run("GetExternalCode", func(t *testing.T) {
		address := common.HexToAddress("0x12345")
		code := []byte{0x61, 0x61, 0x61, 0x61}
		statedb.(*state.StateDB).SetCode(address, code)
		initialGas := env.contract.Gas

		loadedCode := env.GetExternalCode(address)
		r.True(initialGas > env.contract.Gas)
		usedGas := initialGas - env.contract.Gas

		r.Equal(code, loadedCode)
		t.Logf("Gas used for GetExternalCode: %d", usedGas)
	})

	t.Run("GetExternalCodeSize", func(t *testing.T) {
		address := common.HexToAddress("0x12345")
		code := []byte{0x61, 0x61, 0x61, 0x61}
		statedb.(*state.StateDB).SetCode(address, code)
		initialGas := env.contract.Gas

		loadedCodeSize := env.GetExternalCodeSize(address)
		r.True(initialGas > env.contract.Gas)
		usedGas := initialGas - env.contract.Gas
		r.Equal(len(code), loadedCodeSize)
		t.Logf("Gas used for GetExternalCodeSize: %d", usedGas)
	})

	t.Run("GetExternalCodeHash", func(t *testing.T) {
		address := common.HexToAddress("0x12345")
		code := []byte{0x61, 0x61, 0x61, 0x61}
		statedb.(*state.StateDB).SetCode(address, code)
		initialGas := env.contract.Gas

		loadedCodeHash := env.GetExternalCodeHash(address)
		r.True(initialGas > env.contract.Gas)
		usedGas := initialGas - env.contract.Gas
		r.Equal(crypto.Keccak256Hash(code), loadedCodeHash)
		t.Logf("Gas used for GetExternalCodeHash: %d", usedGas)
	})
}
