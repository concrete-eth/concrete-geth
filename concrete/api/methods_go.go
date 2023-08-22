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

// This file will be replaced by methods_tinygo.go when building with tinygo to
// prevent compatibility issues.

package api

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/crypto"
	"github.com/ethereum/go-ethereum/concrete/utils"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

// TODO: Price operations properly

func newEnvironmentMethods() JumpTable {
	tbl := JumpTable{
		EnableGasMetering_OpCode: {
			execute: opEnableGasMetering,
			trusted: true,
			static:  true,
		},
		Debug_OpCode: {
			execute: opDebug,
			trusted: true,
			static:  true,
		},
		TimeNow_OpCode: {
			execute: opTimeNow,
			trusted: true,
			static:  true,
		},
		Keccak256_OpCode: {
			execute:     opKeccak256,
			constantGas: params.Keccak256Gas,
			dynamicGas:  gasKeccak256,
			static:      true,
		},
		EphemeralStore_OpCode: {
			execute:     opEphemeralStore,
			constantGas: params.WarmStorageReadCostEIP2929,
			trusted:     true,
			static:      false,
		},
		EphemeralLoad_OpCode: {
			execute:     opEphemeralLoad,
			constantGas: params.WarmStorageReadCostEIP2929,
			trusted:     true,
			static:      true,
		},
		PersistentPreimageStore_OpCode: {
			execute:    opPersistentPreimageStore,
			dynamicGas: gasPersistentPreimageStore,
			trusted:    true,
			static:     false,
		},
		PersistentPreimageLoad_OpCode: {
			execute:     opPersistentPreimageLoad,
			constantGas: params.ColdSloadCostEIP2929,
			trusted:     true,
			static:      true,
		},
		PersistentPreimageLoadSize_OpCode: {
			execute:     opPersistentPreimageLoadSize,
			constantGas: params.ColdSloadCostEIP2929,
			trusted:     true,
			static:      true,
		},
		EphemeralPreimageStore_OpCode: {
			execute:     opEphemeralPreimageStore,
			constantGas: params.WarmStorageReadCostEIP2929,
			trusted:     true,
			static:      false,
		},
		EphemeralPreimageLoad_OpCode: {
			execute:     opEphemeralPreimageLoad,
			constantGas: params.WarmStorageReadCostEIP2929,
			trusted:     true,
			static:      true,
		},
		EphemeralPreimageLoadSize_OpCode: {
			execute:     opEphemeralPreimageLoadSize,
			constantGas: params.WarmStorageReadCostEIP2929,
			trusted:     true,
			static:      true,
		},
		GetAddress_OpCode: {
			execute:     opGetAddress,
			constantGas: GasQuickStep,
			static:      true,
		},
		GetGasLeft_OpCode: {
			execute:     opGetGasLeft,
			constantGas: GasQuickStep,
			static:      true,
		},
		GetBlockNumber_OpCode: {
			execute:     opGetBlockNumber,
			constantGas: GasQuickStep,
			static:      true,
		},
		GetBlockGasLimit_OpCode: {
			execute:     opGetBlockGasLimit,
			constantGas: GasQuickStep,
			static:      true,
		},
		GetBlockTimestamp_OpCode: {
			execute:     opGetBlockTimestamp,
			constantGas: GasQuickStep,
			static:      true,
		},
		GetBlockDifficulty_OpCode: {
			execute:     opGetBlockDifficulty,
			constantGas: GasQuickStep,
			static:      true,
		},
		GetBlockBaseFee_OpCode: {
			execute:     opGetBlockBaseFee,
			constantGas: GasQuickStep,
			static:      true,
		},
		GetBlockCoinbase_OpCode: {
			execute:     opGetBlockCoinbase,
			constantGas: GasQuickStep,
			static:      true,
		},
		GetPrevRandom_OpCode: {
			execute:     opGetPrevRandom,
			constantGas: GasQuickStep,
			static:      true,
		},
		GetBlockHash_OpCode: {
			execute:     opGetBlockHash,
			constantGas: GasExtStep,
			static:      true,
		},
		GetBalance_OpCode: {
			execute:     opGetBalance,
			constantGas: GasFastStep,
			static:      true,
		},
		GetTxGasPrice_OpCode: {
			execute:     opGetTxGasPrice,
			constantGas: GasQuickStep,
			static:      true,
		},
		GetTxOrigin_OpCode: {
			execute:     opGetTxOrigin,
			constantGas: GasQuickStep,
			static:      true,
		},
		GetCallData_OpCode: {
			execute:     opGetCallData,
			constantGas: GasFastestStep,
			static:      true,
		},
		GetCallDataSize_OpCode: {
			execute:     opGetCallDataSize,
			constantGas: GasQuickStep,
			static:      true,
		},
		GetCaller_OpCode: {
			execute:     opGetCaller,
			constantGas: GasQuickStep,
			static:      true,
		},
		GetCallValue_OpCode: {
			execute:     opGetCallValue,
			constantGas: GasQuickStep,
			static:      true,
		},
		StorageLoad_OpCode: {
			execute:    opStorageLoad,
			dynamicGas: gasStorageLoad,
			static:     true,
		},
		GetCode_OpCode: {
			execute:     opGetCode,
			constantGas: GasFastestStep,
			static:      true,
		},
		GetCodeSize_OpCode: {
			execute:     opGetCodeSize,
			constantGas: GasQuickStep,
			static:      true,
		},
		UseGas_OpCode: {
			execute: opUseGas,
			static:  true,
		},
		StorageStore_OpCode: {
			execute:    opStorageStore,
			dynamicGas: gasStorageStore,
			static:     false,
		},
		Log_OpCode: {
			execute:    opLog,
			dynamicGas: gasLog,
			static:     false,
		},
		GetExternalBalance_OpCode: {
			execute:     opGetExternalBalance,
			constantGas: params.WarmStorageReadCostEIP2929,
			static:      true,
			dynamicGas:  gasGetExternalBalance,
		},
		CallStatic_OpCode: {
			execute:     opCallStatic,
			constantGas: params.WarmStorageReadCostEIP2929,
			static:      true,
		},
		GetExternalCode_OpCode: {
			execute:     opGetExternalCode,
			constantGas: params.WarmStorageReadCostEIP2929,
			static:      true,
		},
		GetExternalCodeSize_OpCode: {
			execute:     opGetExternalCodeSize,
			constantGas: params.WarmStorageReadCostEIP2929,
			static:      true,
		},
		GetExternalCodeHash_OpCode: {
			execute:     opGetExternalCodeHash,
			constantGas: params.WarmStorageReadCostEIP2929,
			static:      true,
		},
		Call_OpCode: {
			execute:     opCall,
			constantGas: params.WarmStorageReadCostEIP2929,
			static:      false,
		},
		CallDelegate_OpCode: {
			execute:     opCallDelegate,
			constantGas: params.WarmStorageReadCostEIP2929,
			static:      false,
		},
		Create_OpCode: {
			execute:     opCreate,
			constantGas: params.CreateGas,
			static:      false,
		},
		Create2_OpCode: {
			execute:     opCreate2,
			constantGas: params.Create2Gas,
			static:      false,
		},
	}

	for i, entry := range tbl {
		if entry == nil {
			tbl[i] = &operation{
				execute: opUndefined,
				static:  true,
			}
		}
	}

	return tbl
}

// TODO: handle invalid args

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
	env.meterGas = meter
	return nil, nil
}

func opDebug(env *Env, args [][]byte) ([][]byte, error) {
	msg := string(args[0])
	env.logger.Debug(msg)
	return nil, nil
}

func opTimeNow(env *Env, args [][]byte) ([][]byte, error) {
	now := uint64(time.Now().UnixNano())
	return [][]byte{utils.Uint64ToBytes(now)}, nil
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
	preimage := args[0]
	hash := crypto.Keccak256Hash(preimage)
	env.statedb.AddPersistentPreimage(hash, preimage)
	return [][]byte{hash.Bytes()}, nil
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
	return [][]byte{utils.Uint64ToBytes(uint64(size))}, nil
}

func opEphemeralPreimageStore(env *Env, args [][]byte) ([][]byte, error) {
	if !env.config.Ephemeral || !env.config.Preimages {
		return nil, ErrFeatureDisabled
	}
	preimage := args[0]
	hash := crypto.Keccak256Hash(preimage)
	env.statedb.AddEphemeralPreimage(hash, preimage)
	return [][]byte{hash.Bytes()}, nil
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
	return [][]byte{utils.Uint64ToBytes(uint64(size))}, nil
}

func opGetAddress(env *Env, args [][]byte) ([][]byte, error) {
	return [][]byte{env.address.Bytes()}, nil
}

func opGetGasLeft(env *Env, args [][]byte) ([][]byte, error) {
	return [][]byte{utils.Uint64ToBytes(env.gas)}, nil
}

func opGetBlockNumber(env *Env, args [][]byte) ([][]byte, error) {
	if env.block == nil {
		return nil, ErrNoData
	}
	number := env.block.BlockNumber()
	return [][]byte{utils.Uint64ToBytes(number)}, nil
}

func opGetBlockGasLimit(env *Env, args [][]byte) ([][]byte, error) {
	if env.block == nil {
		return nil, ErrNoData
	}
	limit := env.block.GasLimit()
	return [][]byte{utils.Uint64ToBytes(limit)}, nil
}

func opGetBlockTimestamp(env *Env, args [][]byte) ([][]byte, error) {
	if env.block == nil {
		return nil, ErrNoData
	}
	timestamp := env.block.Timestamp()
	return [][]byte{utils.Uint64ToBytes(timestamp)}, nil
}

func opGetBlockDifficulty(env *Env, args [][]byte) ([][]byte, error) {
	if env.block == nil {
		return nil, ErrNoData
	}
	difficulty := env.block.Difficulty()
	return [][]byte{difficulty.Bytes()}, nil
}

func opGetBlockBaseFee(env *Env, args [][]byte) ([][]byte, error) {
	if env.block == nil {
		return nil, ErrNoData
	}
	baseFee := env.block.BaseFee()
	return [][]byte{baseFee.Bytes()}, nil
}

func opGetBlockCoinbase(env *Env, args [][]byte) ([][]byte, error) {
	if env.block == nil {
		return nil, ErrNoData
	}
	coinbase := env.block.Coinbase()
	return [][]byte{coinbase.Bytes()}, nil
}

func opGetPrevRandom(env *Env, args [][]byte) ([][]byte, error) {
	if env.block == nil {
		return nil, ErrNoData
	}
	random := env.block.Random()
	return [][]byte{random.Bytes()}, nil
}

func opGetBlockHash(env *Env, args [][]byte) ([][]byte, error) {
	if env.block == nil {
		return nil, ErrNoData
	}
	number := utils.BytesToUint64(args[0])
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
	if env.call == nil {
		return nil, ErrNoData
	}
	price := env.call.TxGasPrice()
	return [][]byte{price.Bytes()}, nil
}

func opGetTxOrigin(env *Env, args [][]byte) ([][]byte, error) {
	if env.call == nil {
		return nil, ErrNoData
	}
	origin := env.call.TxOrigin()
	return [][]byte{origin.Bytes()}, nil
}

func opGetCallData(env *Env, args [][]byte) ([][]byte, error) {
	if env.call == nil {
		return nil, ErrNoData
	}
	data := env.call.CallData()
	return [][]byte{data}, nil
}

func opGetCallDataSize(env *Env, args [][]byte) ([][]byte, error) {
	if env.call == nil {
		return nil, ErrNoData
	}
	size := env.call.CallDataSize()
	return [][]byte{utils.Uint64ToBytes(uint64(size))}, nil
}

func opGetCaller(env *Env, args [][]byte) ([][]byte, error) {
	if env.call == nil {
		return nil, ErrNoData
	}
	caller := env.call.Caller()
	return [][]byte{caller.Bytes()}, nil
}

func opGetCallValue(env *Env, args [][]byte) ([][]byte, error) {
	if env.call == nil {
		return nil, ErrNoData
	}
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
	return [][]byte{utils.Uint64ToBytes(uint64(size))}, nil
}

func opUseGas(env *Env, args [][]byte) ([][]byte, error) {
	gas := utils.BytesToUint64(args[0])
	if ok := env.useGas(gas); !ok {
		return nil, ErrOutOfGas
	}
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
	if env.caller == nil {
		return nil, ErrNoData
	}
	address := common.BytesToAddress(args[0])
	input := args[1]
	gas := utils.BytesToUint64(args[2])
	output, gasLeft, err := env.caller.CallStatic(address, input, gas)
	env.gas += gasLeft
	return [][]byte{output, utils.EncodeError(err)}, nil
}

func opGetExternalCode(env *Env, args [][]byte) ([][]byte, error) {
	address := common.BytesToAddress(args[0])
	code := env.statedb.GetCode(address)
	return [][]byte{code}, nil
}

func opGetExternalCodeSize(env *Env, args [][]byte) ([][]byte, error) {
	address := common.BytesToAddress(args[0])
	size := env.statedb.GetCodeSize(address)
	return [][]byte{utils.Uint64ToBytes(uint64(size))}, nil
}

func opGetExternalCodeHash(env *Env, args [][]byte) ([][]byte, error) {
	address := common.BytesToAddress(args[0])
	hash := env.statedb.GetCodeHash(address)
	return [][]byte{hash.Bytes()}, nil
}

func opCall(env *Env, args [][]byte) ([][]byte, error) {
	if env.caller == nil {
		return nil, ErrNoData
	}
	address := common.BytesToAddress(args[0])
	input := args[1]
	gas := utils.BytesToUint64(args[2])
	value := new(big.Int).SetBytes(args[3])
	output, gasLeft, err := env.caller.Call(address, input, gas, value)
	env.gas += gasLeft
	return [][]byte{output, utils.EncodeError(err)}, nil
}

func opCallDelegate(env *Env, args [][]byte) ([][]byte, error) {
	if env.caller == nil {
		return nil, ErrNoData
	}
	address := common.BytesToAddress(args[0])
	input := args[1]
	gas := utils.BytesToUint64(args[2])
	output, gasLeft, err := env.caller.CallDelegate(address, input, gas)
	env.gas += gasLeft
	return [][]byte{output, utils.EncodeError(err)}, nil
}

func opCreate(env *Env, args [][]byte) ([][]byte, error) {
	if env.caller == nil {
		return nil, ErrNoData
	}
	input := args[0]
	value := new(big.Int).SetBytes(args[1])
	address, gasLeft, err := env.caller.Create(input, value)
	env.gas += gasLeft
	return [][]byte{address.Bytes(), utils.EncodeError(err)}, nil
}

func opCreate2(env *Env, args [][]byte) ([][]byte, error) {
	if env.caller == nil {
		return nil, ErrNoData
	}
	input := args[0]
	value := new(big.Int).SetBytes(args[1])
	salt := common.BytesToHash(args[2])
	address, gasLeft, err := env.caller.Create2(input, salt, value)
	env.gas += gasLeft
	return [][]byte{address.Bytes(), utils.EncodeError(err)}, nil
}
