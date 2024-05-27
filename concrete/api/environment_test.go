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
		contract = NewContract(common.Address{}, common.Address{}, common.HexToAddress("0xc0ffee0001"), new(uint256.Int))
		gas      = uint64(1e6)
	)

	env := NewMockEnvironment(config, meterGas, NewMockBlockContext(), NewMockCaller(), contract)
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
		contract = NewContract(common.Address{}, common.Address{}, common.HexToAddress("0xc0ffee0001"), new(uint256.Int))
		gas      = uint64(1e6)
	)

	env := NewMockEnvironment(config, meterGas, NewMockBlockContext(), NewMockCaller(), contract)
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
		contract = NewContract(common.Address{}, common.Address{}, common.HexToAddress("0xc0ffee0001"), new(uint256.Int))
		gas      = uint64(1e6)
	)

	env := NewMockEnvironment(config, meterGas, NewMockBlockContext(), NewMockCaller(), contract)
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
		contract = NewContract(common.Address{}, common.Address{}, common.HexToAddress("0xc0ffee0001"), new(uint256.Int))
	)

	env := NewMockEnvironment(config, meterGas, NewMockBlockContext(), NewMockCaller(), contract)
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
		contract = NewContract(common.Address{}, common.Address{}, common.HexToAddress("0xc0ffee0001"), new(uint256.Int))
	)

	env := NewMockEnvironment(config, meterGas, NewMockBlockContext(), NewMockCaller(), contract)

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
