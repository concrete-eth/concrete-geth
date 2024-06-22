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
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
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

	env, _, _, _ := NewMockEnvironment(config, meterGas)
	env.contract.Gas = gas

	// GetGasLeft() costs gas, so the cost of that operation must be subtracted
	// from the total gas.
	getGasLeftOpCost := env.table[GetGasLeft_OpCode].constantGas
	gas -= getGasLeftOpCost
	r.Equal(gas, env.GetGasLeft())
	r.Equal(gas, env.Gas())
}

func TestCallOps_Minimal(t *testing.T) {
	var (
		r        = require.New(t)
		config   = EnvConfig{IsStatic: false, IsTrusted: false}
		meterGas = true
		gas      = uint64(1e6)
	)

	env, _, _, _ := NewMockEnvironment(config, meterGas)
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

	env, _, _, _ := NewMockEnvironment(config, meterGas)
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

	env, _, _, _ := NewMockEnvironment(config, meterGas)

	read, write, _ := os.Pipe()

	// Capture stderr
	stderr := os.Stderr
	os.Stderr = write
	defer func() {
		os.Stderr = stderr
	}()

	// Copy catured stderr to a buffer
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

	env, _, _block, _ := NewMockEnvironment(config, meterGas)
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

	blockHash := block.GetHash(0)
	r.Equal(blockHash, env.GetBlockHash(0))

	blockGasLimit := block.GasLimit()
	r.Equal(blockGasLimit, env.GetBlockGasLimit())

	blockTimestamp := block.Timestamp()
	r.Equal(blockTimestamp, env.GetBlockTimestamp())

	blockDifficulty := block.Difficulty()
	r.Equal(blockDifficulty, env.GetBlockDifficulty())

	blockBaseFee := block.BaseFee()
	r.Equal(blockBaseFee, env.GetBlockBaseFee())

	blockCoinbase := block.Coinbase()
	r.Equal(blockCoinbase, env.GetBlockCoinbase())

	prevRandom := block.Random()
	r.Equal(prevRandom, env.GetPrevRandom())
}

func TestCallers(t *testing.T) {
	var (
		r        = require.New(t)
		config   = EnvConfig{IsStatic: false, IsTrusted: false}
		meterGas = true
		gas      = uint64(1e6)
	)

	env, _, _, _caller := NewMockEnvironment(config, meterGas)
	caller := _caller.(*mockCaller)
	env.contract.Gas = gas
	t.Run("CallStatic", func(t *testing.T) {
		input := []byte("input")

		caller.SetCallStaticFn(func(address common.Address, input []byte, gas uint64) ([]byte, uint64, error) {
			return []byte("responseCallStatic"), gas - 1, nil
		})
		expectedRet, _, expectedErr := caller.CallStatic(common.Address{}, input, gas)

		ret, err := env.CallStatic(common.Address{}, input, gas)
		r.NoError(expectedErr, err)
		r.Equal(expectedRet, ret)
		r.Less(env.Gas(), gas)

	})

	t.Run("Call", func(t *testing.T) {
		input := []byte("input")

		caller.SetCallFn(func(address common.Address, input []byte, gas uint64, value *uint256.Int) ([]byte, uint64, error) {
			return []byte("responseCall"), gas - 1, nil
		})
		expectedRet, _, expectedErr := caller.Call(common.Address{}, input, gas, new(uint256.Int))

		ret, err := env.Call(common.Address{}, input, gas, new(uint256.Int))
		r.NoError(expectedErr, err)
		r.Equal(expectedRet, ret)
		r.Less(env.Gas(), gas)
	})

	t.Run("CallDelegate", func(t *testing.T) {
		input := []byte("input")

		caller.SetCallDelegateFn(func(address common.Address, input []byte, gas uint64) ([]byte, uint64, error) {
			return []byte("responseCallDelegate"), gas - 1, nil
		})
		expectedRet, _, expectedErr := caller.CallDelegate(common.Address{}, input, gas)

		ret, err := env.CallDelegate(common.Address{}, input, gas)
		r.NoError(expectedErr, err)
		r.Equal(expectedRet, ret)
		r.Less(env.Gas(), gas)
	})

	t.Run("Create", func(t *testing.T) {
		creationCode := []byte("creationCode")

		caller.SetCreateFn(func(input []byte, gas uint64, value *uint256.Int) ([]byte, common.Address, uint64, error) {
			return []byte("runtimeCode"), common.Address{}, gas - 1, nil
		})
		expectedRet, expectedAddr, _, expectedErr := caller.Create(creationCode, gas, new(uint256.Int))

		ret, addr, err := env.Create(creationCode, new(uint256.Int))
		r.NoError(expectedErr, err)
		r.Equal(expectedRet, ret)
		r.Equal(expectedAddr, addr)
		r.Less(env.Gas(), gas)
	})

	t.Run("Create2", func(t *testing.T) {
		creationCode := []byte("creationCode")
		salt := new(uint256.Int).SetBytes([]byte("salt"))

		caller.SetCreate2Fn(func(input []byte, gas uint64, value *uint256.Int, salt *uint256.Int) ([]byte, common.Address, uint64, error) {
			return []byte("runtimeCode"), common.Address{}, gas - 1, nil
		})
		expectedRet, expectedAddr, _, expectedErr := caller.Create2(creationCode, gas, new(uint256.Int), salt)

		ret, addr, err := env.Create2(creationCode, new(uint256.Int), salt)
		r.NoError(expectedErr, err)
		r.Equal(expectedRet, ret)
		r.Equal(expectedAddr, addr)
		r.Less(env.Gas(), gas)
	})

}
