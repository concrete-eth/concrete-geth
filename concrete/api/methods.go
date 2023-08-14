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
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/crypto"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

var (
	ErrInvalidOpCode   = errors.New("invalid opcode")
	ErrOutOfGas        = errors.New("out of gas")
	ErrWriteProtection = errors.New("write protection")
	ErrFeatureDisabled = errors.New("feature disabled")
	ErrEnvNotTrusted   = errors.New("environment not trusted")
)

type OpCode byte

const (
	// TODO: logic to opcodes [?]
	EnableGasMetering_OpCode          OpCode = 0x01
	Keccak256_OpCode                  OpCode = 0x10
	EphemeralStore_OpCode             OpCode = 0x11
	EphemeralLoad_OpCode              OpCode = 0x12
	PersistentPreimageStore_OpCode    OpCode = 0x13
	PersistentPreimageLoad_OpCode     OpCode = 0x14
	PersistentPreimageLoadSize_OpCode OpCode = 0x15
	EphemeralPreimageStore_OpCode     OpCode = 0x16
	EphemeralPreimageLoad_OpCode      OpCode = 0x17
	EphemeralPreimageLoadSize_OpCode  OpCode = 0x18
	GetAddress_OpCode                 OpCode = 0x19
	GetGasLeft_OpCode                 OpCode = 0x1a
	GetBlockNumber_OpCode             OpCode = 0x1b
	GetBlockGasLimit_OpCode           OpCode = 0x1c
	GetBlockTimestamp_OpCode          OpCode = 0x1d
	GetBlockDifficulty_OpCode         OpCode = 0x1e
	GetBlockBasefee_OpCode            OpCode = 0x1f
	GetBlockCoinbase_OpCode           OpCode = 0x20
	GetPrevRandao_OpCode              OpCode = 0x21
	GetBlockHash_OpCode               OpCode = 0x22
	GetBalance_OpCode                 OpCode = 0x23
	GetTxGasPrice_OpCode              OpCode = 0x24
	GetTxOrigin_OpCode                OpCode = 0x25
	GetCallData_OpCode                OpCode = 0x26
	GetCallDataSize_OpCode            OpCode = 0x27
	GetCaller_OpCode                  OpCode = 0x28
	GetCallValue_OpCode               OpCode = 0x29
	StorageLoad_OpCode                OpCode = 0x2a
	GetCode_OpCode                    OpCode = 0x2b
	GetCodeSize_OpCode                OpCode = 0x2c
	UseGas_OpCode                     OpCode = 0x2d
	StorageStore_OpCode               OpCode = 0x2e
	Log_OpCode                        OpCode = 0x2f
	GetExternalBalance_OpCode         OpCode = 0x30
	CallStatic_OpCode                 OpCode = 0x31
	GetExternalCode_OpCode            OpCode = 0x32
	GetExternalCodeSize_OpCode        OpCode = 0x33
	GetExternalCodeHash_OpCode        OpCode = 0x34
	Call_OpCode                       OpCode = 0x35
	CallDelegate_OpCode               OpCode = 0x36
	Create_OpCode                     OpCode = 0x37
	Create2_OpCode                    OpCode = 0x38
)

const (
	GasQuickStep   uint64 = 2
	GasFastestStep uint64 = 3
	GasFastStep    uint64 = 5
	GasMidStep     uint64 = 8
	GasSlowStep    uint64 = 10
	GasExtStep     uint64 = 20
)

type (
	executionFunc func(env *Env, args [][]byte) ([][]byte, error)
	gasFunc       func(env *Env, args [][]byte) (uint64, error)
)

type operation struct {
	execute     executionFunc
	constantGas uint64
	dynamicGas  gasFunc
}

type JumpTable [256]*operation

// TODO: Price operations properly
// TODO: over/under flows

func NewEnvironmentMethods() JumpTable {
	tbl := JumpTable{
		EnableGasMetering_OpCode: {
			execute: opEnableGasMetering,
		},
		Keccak256_OpCode: {
			execute:     opKeccak256,
			constantGas: params.Keccak256Gas,
			dynamicGas:  gasKeccak256,
		},
		EphemeralStore_OpCode: {
			execute:     opEphemeralStore,
			constantGas: params.WarmStorageReadCostEIP2929,
		},
		EphemeralLoad_OpCode: {
			execute:     opEphemeralLoad,
			constantGas: params.WarmStorageReadCostEIP2929,
		},
		PersistentPreimageStore_OpCode: {
			execute:    opPersistentPreimageStore,
			dynamicGas: gasPersistentPreimageStore,
		},
		PersistentPreimageLoad_OpCode: {
			execute:     opPersistentPreimageLoad,
			constantGas: params.ColdSloadCostEIP2929,
		},
		PersistentPreimageLoadSize_OpCode: {
			execute:     opPersistentPreimageLoadSize,
			constantGas: params.ColdSloadCostEIP2929,
		},
		EphemeralPreimageStore_OpCode: {
			execute:     opEphemeralPreimageStore,
			constantGas: params.WarmStorageReadCostEIP2929,
		},
		EphemeralPreimageLoad_OpCode: {
			execute:     opEphemeralPreimageLoad,
			constantGas: params.WarmStorageReadCostEIP2929,
		},
		EphemeralPreimageLoadSize_OpCode: {
			execute:     opEphemeralPreimageLoadSize,
			constantGas: params.WarmStorageReadCostEIP2929,
		},
		GetAddress_OpCode: {
			execute:     opGetAddress,
			constantGas: GasQuickStep,
		},
		GetGasLeft_OpCode: {
			execute:     opGetGasLeft,
			constantGas: GasQuickStep,
		},
		GetBlockNumber_OpCode: {
			execute:     opGetBlockNumber,
			constantGas: GasQuickStep,
		},
		GetBlockGasLimit_OpCode: {
			execute:     opGetBlockGasLimit,
			constantGas: GasQuickStep,
		},
		GetBlockTimestamp_OpCode: {
			execute:     opGetBlockTimestamp,
			constantGas: GasQuickStep,
		},
		GetBlockDifficulty_OpCode: {
			execute:     opGetBlockDifficulty,
			constantGas: GasQuickStep,
		},
		GetBlockBasefee_OpCode: {
			execute:     opGetBlockBasefee,
			constantGas: GasQuickStep,
		},
		GetBlockCoinbase_OpCode: {
			execute:     opGetBlockCoinbase,
			constantGas: GasQuickStep,
		},
		GetPrevRandao_OpCode: {
			execute:     opGetPrevRandao,
			constantGas: GasQuickStep,
		},
		GetBlockHash_OpCode: {
			execute:     opGetBlockHash,
			constantGas: GasExtStep,
		},
		GetBalance_OpCode: {
			execute:     opGetBalance,
			constantGas: GasFastStep,
		},
		GetTxGasPrice_OpCode: {
			execute:     opGetTxGasPrice,
			constantGas: GasQuickStep,
		},
		GetTxOrigin_OpCode: {
			execute:     opGetTxOrigin,
			constantGas: GasQuickStep,
		},
		GetCallData_OpCode: {
			execute:     opGetCallData,
			constantGas: GasFastestStep,
		},
		GetCallDataSize_OpCode: {
			execute:     opGetCallDataSize,
			constantGas: GasQuickStep,
		},
		GetCaller_OpCode: {
			execute:     opGetCaller,
			constantGas: GasQuickStep,
		},
		GetCallValue_OpCode: {
			execute:     opGetCallValue,
			constantGas: GasQuickStep,
		},
		StorageLoad_OpCode: {
			execute:    opStorageLoad,
			dynamicGas: gasStorageLoad,
		},
		GetCode_OpCode: {
			execute:     opGetCode,
			constantGas: GasFastestStep,
		},
		GetCodeSize_OpCode: {
			execute:     opGetCodeSize,
			constantGas: GasQuickStep,
		},
		UseGas_OpCode: {
			execute: opUseGas,
		},
		StorageStore_OpCode: {
			execute:    opStorageStore,
			dynamicGas: gasStorageStore,
		},
		Log_OpCode: {
			execute:    opLog,
			dynamicGas: gasLog,
		},
		GetExternalBalance_OpCode: {
			execute:     opGetExternalBalance,
			constantGas: params.WarmStorageReadCostEIP2929,
			dynamicGas:  gasGetExternalBalance,
		},
		CallStatic_OpCode: {
			execute:     opCallStatic,
			constantGas: params.WarmStorageReadCostEIP2929,
		},
		GetExternalCode_OpCode: {
			execute:     opGetExternalCode,
			constantGas: params.WarmStorageReadCostEIP2929,
		},
		GetExternalCodeSize_OpCode: {
			execute:     opGetExternalCodeSize,
			constantGas: params.WarmStorageReadCostEIP2929,
		},
		GetExternalCodeHash_OpCode: {
			execute:     opGetExternalCodeHash,
			constantGas: params.WarmStorageReadCostEIP2929,
		},
		Call_OpCode: {
			execute:     opCall,
			constantGas: params.WarmStorageReadCostEIP2929,
		},
		CallDelegate_OpCode: {
			execute:     opCallDelegate,
			constantGas: params.WarmStorageReadCostEIP2929,
		},
		Create_OpCode: {
			execute:     opCreate,
			constantGas: params.CreateGas,
		},
		Create2_OpCode: {
			execute:     opCreate2,
			constantGas: params.Create2Gas,
		},
	}

	for i, entry := range tbl {
		if entry == nil {
			tbl[i] = &operation{execute: opUndefined}
		}
	}

	return tbl
}

func opUndefined(env *Env, args [][]byte) ([][]byte, error) {
	return nil, ErrInvalidOpCode
}

func opEnableGasMetering(env *Env, args [][]byte) ([][]byte, error) {
	var meter bool
	if args[0][0] == byte(0x01) {
		meter = true
	}
	if env.meterGas == meter {
		return nil, nil
	}
	if !env.config.Trusted {
		return nil, ErrEnvNotTrusted
	}
	env.meterGas = meter
	return nil, nil
}

func gasKeccak256(env *Env, args [][]byte) (uint64, error) {
	wordSize := (len(args[0]) + 31) / 32
	gas := uint64(wordSize) * params.Keccak256WordGas
	return gas, nil
}

func opKeccak256(env *Env, args [][]byte) ([][]byte, error) {
	hash := crypto.Keccak256(args[0])
	return [][]byte{hash}, nil
}

func opEphemeralStore(env *Env, args [][]byte) ([][]byte, error) {
	if !env.config.Ephemeral {
		return nil, ErrFeatureDisabled
	}
	if env.config.Static {
		return nil, ErrWriteProtection
	}
	key := common.BytesToHash(args[0])
	value := common.BytesToHash(args[1])
	env.statedb.SetEphemeralState(env.address, key, value)
	return nil, nil
}

func opEphemeralLoad(env *Env, args [][]byte) ([][]byte, error) {
	if !env.config.Ephemeral {
		return nil, ErrFeatureDisabled
	}
	key := common.BytesToHash(args[0])
	data := env.statedb.GetEphemeralState(env.address, key)
	return [][]byte{data.Bytes()}, nil
}

func gasPersistentPreimageStore(env *Env, args [][]byte) (uint64, error) {
	wordSize := (len(args[0]) + 31) / 32
	gas := uint64(wordSize) * params.CreateDataGas
	return gas, nil
}

func opPersistentPreimageStore(env *Env, args [][]byte) ([][]byte, error) {
	if !env.config.Preimages {
		return nil, ErrFeatureDisabled
	}
	if env.config.Static {
		return nil, ErrWriteProtection
	}
	value := args[0]
	key := crypto.Keccak256Hash(value)
	env.statedb.AddPersistentPreimage(key, value)
	return nil, nil
}

func opPersistentPreimageLoad(env *Env, args [][]byte) ([][]byte, error) {
	if !env.config.Preimages {
		return nil, ErrFeatureDisabled
	}
	key := common.BytesToHash(args[0])
	data := env.statedb.GetPersistentPreimage(key)
	return [][]byte{data}, nil
}

func opPersistentPreimageLoadSize(env *Env, args [][]byte) ([][]byte, error) {
	if !env.config.Preimages {
		return nil, ErrFeatureDisabled
	}
	key := common.BytesToHash(args[0])
	size := env.statedb.GetPersistentPreimageSize(key)
	return [][]byte{Uint64ToBytes(uint64(size))}, nil
}

func opEphemeralPreimageStore(env *Env, args [][]byte) ([][]byte, error) {
	if !env.config.Ephemeral || !env.config.Preimages {
		return nil, ErrFeatureDisabled
	}
	if env.config.Static {
		return nil, ErrWriteProtection
	}
	key := common.BytesToHash(args[0])
	value := args[1]
	env.statedb.AddEphemeralPreimage(key, value)
	return nil, nil
}

func opEphemeralPreimageLoad(env *Env, args [][]byte) ([][]byte, error) {
	if !env.config.Ephemeral || !env.config.Preimages {
		return nil, ErrFeatureDisabled
	}
	key := common.BytesToHash(args[0])
	data := env.statedb.GetEphemeralPreimage(key)
	return [][]byte{data}, nil
}

func opEphemeralPreimageLoadSize(env *Env, args [][]byte) ([][]byte, error) {
	if !env.config.Ephemeral || !env.config.Preimages {
		return nil, ErrFeatureDisabled
	}
	key := common.BytesToHash(args[0])
	size := env.statedb.GetEphemeralPreimageSize(key)
	return [][]byte{Uint64ToBytes(uint64(size))}, nil
}

func opGetAddress(env *Env, args [][]byte) ([][]byte, error) {
	return [][]byte{env.address.Bytes()}, nil
}

func opGetGasLeft(env *Env, args [][]byte) ([][]byte, error) {
	return [][]byte{Uint64ToBytes(env.gas)}, nil
}

func opGetBlockNumber(env *Env, args [][]byte) ([][]byte, error) {
	number := env.block.BlockNumber()
	return [][]byte{Uint64ToBytes(number)}, nil
}

func opGetBlockGasLimit(env *Env, args [][]byte) ([][]byte, error) {
	limit := env.block.GasLimit()
	return [][]byte{Uint64ToBytes(limit)}, nil
}

func opGetBlockTimestamp(env *Env, args [][]byte) ([][]byte, error) {
	timestamp := env.block.Timestamp()
	return [][]byte{Uint64ToBytes(timestamp)}, nil
}

func opGetBlockDifficulty(env *Env, args [][]byte) ([][]byte, error) {
	difficulty := env.block.Difficulty()
	return [][]byte{difficulty.Bytes()}, nil
}

func opGetBlockBasefee(env *Env, args [][]byte) ([][]byte, error) {
	basefee := env.block.BaseFee()
	return [][]byte{basefee.Bytes()}, nil
}

func opGetBlockCoinbase(env *Env, args [][]byte) ([][]byte, error) {
	coinbase := env.block.Coinbase()
	return [][]byte{coinbase.Bytes()}, nil
}

func opGetPrevRandao(env *Env, args [][]byte) ([][]byte, error) {
	randao := env.block.Random()
	return [][]byte{randao.Bytes()}, nil
}

func opGetBlockHash(env *Env, args [][]byte) ([][]byte, error) {
	number := BytesToUint64(args[0])
	var upper, lower uint64
	upper = env.block.BlockNumber()
	if upper < 257 {
		lower = 0
	} else {
		lower = upper - 256
	}
	if number < lower || number > upper {
		return nil, nil
	}
	hash := env.block.GetHash(number)
	return [][]byte{hash.Bytes()}, nil
}

func opGetBalance(env *Env, args [][]byte) ([][]byte, error) {
	balance := env.statedb.GetBalance(env.address)
	return [][]byte{balance.Bytes()}, nil
}

func opGetTxGasPrice(env *Env, args [][]byte) ([][]byte, error) {
	price := env.call.TxGasPrice()
	return [][]byte{price.Bytes()}, nil
}

func opGetTxOrigin(env *Env, args [][]byte) ([][]byte, error) {
	origin := env.call.TxOrigin()
	return [][]byte{origin.Bytes()}, nil
}

func opGetCallData(env *Env, args [][]byte) ([][]byte, error) {
	data := env.call.CallData()
	return [][]byte{data}, nil
}

func opGetCallDataSize(env *Env, args [][]byte) ([][]byte, error) {
	size := env.call.CallDataSize()
	return [][]byte{Uint64ToBytes(uint64(size))}, nil
}

func opGetCaller(env *Env, args [][]byte) ([][]byte, error) {
	caller := env.call.Caller()
	return [][]byte{caller.Bytes()}, nil
}

func opGetCallValue(env *Env, args [][]byte) ([][]byte, error) {
	value := env.call.CallValue()
	return [][]byte{value.Bytes()}, nil
}

func gasStorageLoad(env *Env, args [][]byte) (uint64, error) {
	statedb := env.statedb
	address := env.address
	key := common.BytesToHash(args[0])
	if _, slotPresent := statedb.SlotInAccessList(address, key); !slotPresent {
		env.statedb.AddSlotToAccessList(address, key)
		return params.ColdSloadCostEIP2929, nil
	}
	return params.WarmStorageReadCostEIP2929, nil
}

func opStorageLoad(env *Env, args [][]byte) ([][]byte, error) {
	if env.config.Static {
		return nil, ErrWriteProtection
	}
	key := common.BytesToHash(args[0])
	value := env.statedb.GetPersistentState(env.address, key)
	return [][]byte{value.Bytes()}, nil
}

func opGetCode(env *Env, args [][]byte) ([][]byte, error) {
	code := env.statedb.GetCode(env.address)
	return [][]byte{code}, nil
}

func opGetCodeSize(env *Env, args [][]byte) ([][]byte, error) {
	size := env.statedb.GetCodeSize(env.address)
	return [][]byte{Uint64ToBytes(uint64(size))}, nil
}

func opUseGas(env *Env, args [][]byte) ([][]byte, error) {
	gas := BytesToUint64(args[0])
	env.gas -= gas
	return nil, nil
}

func gasStorageStore(env *Env, args [][]byte) (uint64, error) {
	return 0, nil
}

func opStorageStore(env *Env, args [][]byte) ([][]byte, error) {
	if env.config.Static {
		return nil, ErrWriteProtection
	}
	key := common.BytesToHash(args[0])
	value := common.BytesToHash(args[1])
	env.statedb.SetPersistentState(env.address, key, value)
	return nil, nil
}

func gasLog(env *Env, args [][]byte) (uint64, error) {
	nTopics := len(args) - 1
	size := len(args[nTopics])
	return params.LogGas + params.LogTopicGas*uint64(nTopics) + params.LogDataGas*uint64(size), nil
}

func opLog(env *Env, args [][]byte) ([][]byte, error) {
	if env.config.Static {
		return nil, ErrWriteProtection
	}
	topics := make([]common.Hash, len(args)-1)
	for i, arg := range args[1:] {
		topics[i] = common.BytesToHash(arg)
	}
	data := args[len(args)-1]
	env.statedb.AddLog(&types.Log{
		Address:     env.address,
		Topics:      topics,
		Data:        data,
		BlockNumber: env.block.BlockNumber(),
	})
	return nil, nil
}

func gasGetExternalBalance(env *Env, args [][]byte) (uint64, error) {
	address := common.BytesToAddress(args[0])
	if !env.statedb.AddressInAccessList(address) {
		env.statedb.AddAddressToAccessList(address)
		return params.ColdAccountAccessCostEIP2929 - params.WarmStorageReadCostEIP2929, nil
	}
	return 0, nil
}

func opGetExternalBalance(env *Env, args [][]byte) ([][]byte, error) {
	address := common.BytesToAddress(args[0])
	balance := env.statedb.GetBalance(address)
	return [][]byte{balance.Bytes()}, nil
}

func opCallStatic(env *Env, args [][]byte) ([][]byte, error) {
	address := common.BytesToAddress(args[0])
	input := args[1]
	gas := BytesToUint64(args[2])
	output, gasLeft, err := env.caller.CallStatic(address, input, gas)
	env.gas += gasLeft
	return [][]byte{output, EncodeError(err)}, nil
}

func opGetExternalCode(env *Env, args [][]byte) ([][]byte, error) {
	address := common.BytesToAddress(args[0])
	code := env.statedb.GetCode(address)
	return [][]byte{code}, nil
}

func opGetExternalCodeSize(env *Env, args [][]byte) ([][]byte, error) {
	address := common.BytesToAddress(args[0])
	size := env.statedb.GetCodeSize(address)
	return [][]byte{Uint64ToBytes(uint64(size))}, nil
}

func opGetExternalCodeHash(env *Env, args [][]byte) ([][]byte, error) {
	address := common.BytesToAddress(args[0])
	hash := env.statedb.GetCodeHash(address)
	return [][]byte{hash.Bytes()}, nil
}

func opCall(env *Env, args [][]byte) ([][]byte, error) {
	address := common.BytesToAddress(args[0])
	input := args[1]
	gas := BytesToUint64(args[2])
	value := new(big.Int).SetBytes(args[3])
	output, gasLeft, err := env.caller.Call(address, input, gas, value)
	env.gas += gasLeft
	return [][]byte{output, EncodeError(err)}, nil
}

func opCallDelegate(env *Env, args [][]byte) ([][]byte, error) {
	address := common.BytesToAddress(args[0])
	input := args[1]
	gas := BytesToUint64(args[2])
	output, gasLeft, err := env.caller.CallDelegate(address, input, gas)
	env.gas += gasLeft
	return [][]byte{output, EncodeError(err)}, nil
}

func opCreate(env *Env, args [][]byte) ([][]byte, error) {
	input := args[0]
	value := new(big.Int).SetBytes(args[1])
	address, ret, gasLeft, err := env.caller.Create(input, value)
	env.gas += gasLeft
	return [][]byte{address.Bytes(), EncodeError(err)}, nil
}

func opCreate2(env *Env, args [][]byte) ([][]byte, error) {
	input := args[0]
	value := new(big.Int).SetBytes(args[1])
	salt := common.BytesToHash(args[2])
	address, ret, gasLeft, err := env.caller.Create2(input, salt, value)
	env.gas += gasLeft
	if err != nil && ret != nil {
		err = fmt.Errorf("%w: %s", err, string(ret))
	}
	return [][]byte{address.Bytes(), EncodeError(err)}, nil
}