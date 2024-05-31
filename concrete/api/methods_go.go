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
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/concrete/crypto"
	"github.com/ethereum/go-ethereum/concrete/utils"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
)

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
		Revert_OpCode: {
			execute:     opRevert,
			constantGas: GasQuickStep,
			static:      true,
		},
		UseGas_OpCode: {
			execute:    opUseGas,
			dynamicGas: gasUseGas,
			static:     true,
		},
		Keccak256_OpCode: {
			execute:     opKeccak256,
			constantGas: params.Keccak256Gas,
			dynamicGas:  gasKeccak256,
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
		TLoad_OpCode: {
			execute:     opTload,
			constantGas: params.WarmStorageReadCostEIP2929,
			static:      true,
		},
		GetCode_OpCode: {
			// disabled
			// TODO: Why is this disabled?
			execute:     opGetCode,
			constantGas: 0,
			static:      true,
		},
		GetCodeSize_OpCode: {
			// disabled
			execute:     opGetCodeSize,
			constantGas: 0,
			static:      true,
		},
		TStore_OpCode: {
			execute:     opTstore,
			constantGas: params.WarmStorageReadCostEIP2929,
			static:      false,
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
			dynamicGas:  gasGetExternalBalance,
			static:      true,
		},
		GetExternalCode_OpCode: {
			execute:     opGetExternalCode,
			constantGas: params.WarmStorageReadCostEIP2929,
			dynamicGas:  gasGetExternalCode,
			static:      true,
		},
		GetExternalCodeSize_OpCode: {
			execute:     opGetExternalCodeSize,
			constantGas: params.WarmStorageReadCostEIP2929,
			dynamicGas:  gasGetExternalCodeSize,
			static:      true,
		},
		GetExternalCodeHash_OpCode: {
			execute:     opGetExternalCodeHash,
			constantGas: params.WarmStorageReadCostEIP2929,
			dynamicGas:  gasGetExternalCodeHash,
			static:      true,
		},
		CallStatic_OpCode: {
			execute:     opCallStatic,
			constantGas: params.WarmStorageReadCostEIP2929,
			dynamicGas:  gasCallStatic,
			static:      true,
		},
		Call_OpCode: {
			execute:     opCall,
			constantGas: params.WarmStorageReadCostEIP2929,
			dynamicGas:  gasCall,
			static:      false,
		},
		CallDelegate_OpCode: {
			execute:     opCallDelegate,
			constantGas: params.WarmStorageReadCostEIP2929,
			dynamicGas:  gasCallDelegate,
			static:      false,
		},
		Create_OpCode: {
			execute:     opCreate,
			constantGas: params.CreateGas,
			dynamicGas:  gasCreate,
			static:      false,
		},
		Create2_OpCode: {
			execute:     opCreate2,
			constantGas: params.Create2Gas,
			dynamicGas:  gasCreate2,
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

func toWordSize(size int) uint64 {
	return uint64((size + 31) / 32)
}

func gasAccountAccessMinusWarm(env *Env, address common.Address) (uint64, error) {
	if !env.statedb.AddressInAccessList(address) {
		env.statedb.AddAddressToAccessList(address)
		return params.ColdAccountAccessCostEIP2929 - params.WarmStorageReadCostEIP2929, nil
	}
	return 0, nil
}

func gasExternalCall(env *Env, address common.Address, callCost uint64) (uint64, error) {
	baseCost, err := gasAccountAccessMinusWarm(env, address)
	if err != nil {
		return 0, err
	}
	gasAvailable, overflow := math.SafeSub(env.contract.Gas, baseCost)
	if overflow {
		return 0, ErrGasUintOverflow
	}
	gas := gasAvailable - gasAvailable/64
	if gas < callCost {
		env.callGasTemp = gas
	} else {
		env.callGasTemp = callCost
	}
	return env.callGasTemp + baseCost, nil
}

func opUndefined(env *Env, args [][]byte) ([][]byte, error) {
	return nil, ErrInvalidOpCode
}

func opEnableGasMetering(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 1 || len(args[0]) != 1 {
		return nil, ErrInvalidInput
	}
	var meter bool
	if args[0][0]&1 == byte(0x01) {
		meter = true
	}
	env.meterGas = meter
	return nil, nil
}

func opDebug(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, ErrInvalidInput
	}
	msg := string(args[0])
	fmt.Fprint(os.Stderr, msg)
	return nil, nil
}

func opTimeNow(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	now := uint64(time.Now().UnixNano())
	return [][]byte{utils.Uint64ToBytes(now)}, nil
}

func gasKeccak256(env *Env, args [][]byte) (uint64, error) {
	if len(args) != 1 {
		return 0, ErrInvalidInput
	}
	// We assume len() to always be much smaller than MAX_UINT64 / Keccak256WordGas
	// so this cannot overflow
	wordSize := toWordSize(len(args[0]))
	gas := wordSize * params.Keccak256WordGas
	return gas, nil
}

func opKeccak256(env *Env, args [][]byte) ([][]byte, error) {
	hash := crypto.Keccak256(args[0])
	return [][]byte{hash}, nil
}

func gasUseGas(env *Env, args [][]byte) (uint64, error) {
	if len(args) != 1 {
		return 0, ErrInvalidInput
	}
	if len(args[0]) != 8 {
		return 0, ErrInvalidInput
	}
	gas := utils.BytesToUint64(args[0])
	if ok := env.useGas(gas); !ok {
		return 0, ErrOutOfGas
	}
	return 0, nil
}

func opUseGas(env *Env, args [][]byte) ([][]byte, error) {
	return nil, nil
}

func opRevert(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, ErrInvalidInput
	}
	reason := string(args[0])
	env.revertErr = errors.New(reason)
	return nil, ErrExecutionReverted
}

func opGetAddress(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	return [][]byte{env.contract.Address.Bytes()}, nil
}

func opGetGasLeft(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	return [][]byte{utils.Uint64ToBytes(env.Gas())}, nil
}

func opGetBlockNumber(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	number := env.block.BlockNumber()
	return [][]byte{utils.Uint64ToBytes(number)}, nil
}

func opGetBlockGasLimit(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	limit := env.block.GasLimit()
	return [][]byte{utils.Uint64ToBytes(limit)}, nil
}

func opGetBlockTimestamp(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	timestamp := env.block.Timestamp()
	return [][]byte{utils.Uint64ToBytes(timestamp)}, nil
}

func opGetBlockDifficulty(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	difficulty := env.block.Difficulty()
	return [][]byte{difficulty.Bytes()}, nil
}

func opGetBlockBaseFee(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	baseFee := env.block.BaseFee()
	return [][]byte{baseFee.Bytes()}, nil
}

func opGetBlockCoinbase(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	coinbase := env.block.Coinbase()
	return [][]byte{coinbase.Bytes()}, nil
}

func opGetPrevRandom(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	random := env.block.Random()
	return [][]byte{random.Bytes()}, nil
}

func opGetBlockHash(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, ErrInvalidInput
	}
	if len(args[0]) != 8 {
		return nil, ErrInvalidInput
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
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	balance := env.statedb.GetBalance(env.contract.Address)
	return [][]byte{balance.Bytes()}, nil
}

func opGetTxGasPrice(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	price := env.contract.GasPrice
	return [][]byte{price.Bytes()}, nil
}

func opGetTxOrigin(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	origin := env.contract.Origin
	return [][]byte{origin.Bytes()}, nil
}

func opGetCallData(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	data := env.contract.Input
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	return [][]byte{dataCopy}, nil
}

func opGetCallDataSize(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	size := len(env.contract.Input)
	return [][]byte{utils.Uint64ToBytes(uint64(size))}, nil
}

func opGetCaller(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	caller := env.contract.Caller
	return [][]byte{caller.Bytes()}, nil
}

func opGetCallValue(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	value := env.contract.Value
	return [][]byte{value.Bytes()}, nil
}

func gasStorageLoad(env *Env, args [][]byte) (uint64, error) {
	if len(args) != 1 {
		return 0, ErrInvalidInput
	}
	if len(args[0]) != 32 {
		return 0, ErrInvalidInput
	}
	key := common.BytesToHash(args[0])
	if _, slotPresent := env.statedb.SlotInAccessList(env.contract.Address, key); !slotPresent {
		env.statedb.AddSlotToAccessList(env.contract.Address, key)
		return params.ColdSloadCostEIP2929, nil
	}
	return params.WarmStorageReadCostEIP2929, nil
}

func opStorageLoad(env *Env, args [][]byte) ([][]byte, error) {
	key := common.BytesToHash(args[0])
	value := env.statedb.GetState(env.contract.Address, key)
	return [][]byte{value.Bytes()}, nil
}

func opTload(env *Env, args [][]byte) ([][]byte, error) {
	key := common.BytesToHash(args[0])
	value := env.statedb.GetTransientState(env.contract.Address, key)
	return [][]byte{value.Bytes()}, nil
}

func opGetCode(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	return nil, ErrInvalidOpCode
}

func opGetCodeSize(env *Env, args [][]byte) ([][]byte, error) {
	return nil, ErrInvalidOpCode
}

func gasStorageStore(env *Env, args [][]byte) (uint64, error) {
	if len(args) != 2 {
		return 0, ErrInvalidInput
	}
	if len(args[0]) != 32 || len(args[1]) != 32 {
		return 0, ErrInvalidInput
	}
	if env.contract.Gas <= params.SstoreSentryGasEIP2200 {
		return 0, errors.New("not enough gas for reentrancy sentry")
	}
	var (
		key     = common.BytesToHash(args[0])
		current = env.statedb.GetState(env.contract.Address, key)
		cost    = uint64(0)
	)
	if _, slotPresent := env.statedb.SlotInAccessList(env.contract.Address, key); !slotPresent {
		cost = params.ColdSloadCostEIP2929
		env.statedb.AddSlotToAccessList(env.contract.Address, key)
	}
	value := common.BytesToHash(args[1])
	if current == value {
		return cost + params.WarmStorageReadCostEIP2929, nil
	}
	original := env.statedb.GetCommittedState(env.contract.Address, key)
	if original == current {
		if original == (common.Hash{}) {
			return cost + params.SstoreSetGasEIP2200, nil
		}
		if value == (common.Hash{}) {
			env.statedb.AddRefund(params.SstoreClearsScheduleRefundEIP3529)
		}
		return cost + (params.SstoreResetGasEIP2200 - params.ColdSloadCostEIP2929), nil
	}
	if original != (common.Hash{}) {
		if current == (common.Hash{}) {
			env.statedb.SubRefund(params.SstoreClearsScheduleRefundEIP3529)
		} else if value == (common.Hash{}) {
			env.statedb.AddRefund(params.SstoreClearsScheduleRefundEIP3529)
		}
	}
	if original == value {
		if original == (common.Hash{}) {
			env.statedb.AddRefund(params.SstoreSetGasEIP2200 - params.WarmStorageReadCostEIP2929)
		} else {
			env.statedb.AddRefund((params.SstoreResetGasEIP2200 - params.ColdSloadCostEIP2929) - params.WarmStorageReadCostEIP2929)
		}
	}
	return cost + params.WarmStorageReadCostEIP2929, nil
}

func opStorageStore(env *Env, args [][]byte) ([][]byte, error) {
	key := common.BytesToHash(args[0])
	value := common.BytesToHash(args[1])
	env.statedb.SetState(env.contract.Address, key, value)
	return nil, nil
}

func opTstore(env *Env, args [][]byte) ([][]byte, error) {
	key := common.BytesToHash(args[0])
	value := common.BytesToHash(args[1])
	env.statedb.SetTransientState(env.contract.Address, key, value)
	return nil, nil
}

func gasLog(env *Env, args [][]byte) (uint64, error) {
	if len(args) == 0 || len(args) > 5 {
		return 0, ErrInvalidInput
	}
	nTopics := len(args) - 1
	for _, arg := range args[:nTopics] {
		if len(arg) != 32 {
			return 0, ErrInvalidInput
		}
	}
	topicGas := uint64(nTopics) * params.LogTopicGas
	dataSize := uint64(len(args[nTopics]))
	dataGas := dataSize * params.LogDataGas
	return params.LogGas + topicGas + dataGas, nil
}

func opLog(env *Env, args [][]byte) ([][]byte, error) {
	topics := make([]common.Hash, len(args)-1)
	for i, arg := range args[:len(topics)] {
		topics[i] = common.BytesToHash(arg)
	}
	data := args[len(args)-1]
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	env.statedb.AddLog(&types.Log{
		Address:     env.contract.Address,
		Topics:      topics,
		Data:        dataCopy,
		BlockNumber: env.block.BlockNumber(),
	})
	return nil, nil
}

func gasGetExternalBalance(env *Env, args [][]byte) (uint64, error) {
	if len(args) != 1 {
		return 0, ErrInvalidInput
	}
	if len(args[0]) != 20 {
		return 0, ErrInvalidInput
	}
	address := common.BytesToAddress(args[0])
	return gasAccountAccessMinusWarm(env, address)
}

func opGetExternalBalance(env *Env, args [][]byte) ([][]byte, error) {
	address := common.BytesToAddress(args[0])
	balance := env.statedb.GetBalance(address)
	return [][]byte{balance.Bytes()}, nil
}

func gasGetExternalCode(env *Env, args [][]byte) (uint64, error) {
	if len(args) != 1 {
		return 0, ErrInvalidInput
	}
	if len(args[0]) != 20 {
		return 0, ErrInvalidInput
	}
	address := common.BytesToAddress(args[0])
	return gasAccountAccessMinusWarm(env, address)
}

func opGetExternalCode(env *Env, args [][]byte) ([][]byte, error) {
	address := common.BytesToAddress(args[0])
	code := env.statedb.GetCode(address)
	codeCopy := make([]byte, len(code))
	copy(codeCopy, code)
	return [][]byte{codeCopy}, nil
}

func gasGetExternalCodeSize(env *Env, args [][]byte) (uint64, error) {
	if len(args) != 1 {
		return 0, ErrInvalidInput
	}
	if len(args[0]) != 20 {
		return 0, ErrInvalidInput
	}
	address := common.BytesToAddress(args[0])
	return gasAccountAccessMinusWarm(env, address)
}

func opGetExternalCodeSize(env *Env, args [][]byte) ([][]byte, error) {
	address := common.BytesToAddress(args[0])
	size := env.statedb.GetCodeSize(address)
	return [][]byte{utils.Uint64ToBytes(uint64(size))}, nil
}

func gasGetExternalCodeHash(env *Env, args [][]byte) (uint64, error) {
	if len(args) != 1 {
		return 0, ErrInvalidInput
	}
	if len(args[0]) != 20 {
		return 0, ErrInvalidInput
	}
	address := common.BytesToAddress(args[0])
	return gasAccountAccessMinusWarm(env, address)
}

func opGetExternalCodeHash(env *Env, args [][]byte) ([][]byte, error) {
	address := common.BytesToAddress(args[0])
	hash := env.statedb.GetCodeHash(address)
	return [][]byte{hash.Bytes()}, nil
}

func gasCallStatic(env *Env, args [][]byte) (uint64, error) {
	if len(args) != 3 {
		return 0, ErrInvalidInput
	}
	if len(args[0]) != 20 {
		return 0, ErrInvalidInput
	}
	if len(args[2]) != 8 {
		return 0, ErrInvalidInput
	}
	address := common.BytesToAddress(args[0])
	gas := utils.BytesToUint64(args[2])
	return gasExternalCall(env, address, gas)
}

func opCallStatic(env *Env, args [][]byte) ([][]byte, error) {
	var (
		address = common.BytesToAddress(args[0])
		input   = args[1]
		gas     = env.callGasTemp
	)
	output, gasLeft, err := env.caller.CallStatic(address, input, gas)
	env.contract.Gas += gasLeft
	return [][]byte{output, utils.EncodeError(err)}, nil
}

func gasCall(env *Env, args [][]byte) (uint64, error) {
	if len(args) != 4 {
		return 0, ErrInvalidInput
	}
	if len(args[0]) != 20 {
		return 0, ErrInvalidInput
	}
	if len(args[2]) != 8 {
		return 0, ErrInvalidInput
	}
	if len(args[3]) != 32 {
		return 0, ErrInvalidInput
	}
	address := common.BytesToAddress(args[0])
	gas := utils.BytesToUint64(args[2])
	return gasExternalCall(env, address, gas)
}

func opCall(env *Env, args [][]byte) ([][]byte, error) {
	var (
		address = common.BytesToAddress(args[0])
		input   = args[1]
		gas     = env.callGasTemp
		value   = new(uint256.Int).SetBytes(args[3])
	)
	output, gasLeft, err := env.caller.Call(address, input, gas, value)
	env.contract.Gas += gasLeft
	return [][]byte{output, utils.EncodeError(err)}, nil
}

func gasCallDelegate(env *Env, args [][]byte) (uint64, error) {
	if len(args) != 3 {
		return 0, ErrInvalidInput
	}
	if len(args[0]) != 20 {
		return 0, ErrInvalidInput
	}
	if len(args[2]) != 8 {
		return 0, ErrInvalidInput
	}
	address := common.BytesToAddress(args[0])
	gas := utils.BytesToUint64(args[2])
	return gasExternalCall(env, address, gas)
}

func opCallDelegate(env *Env, args [][]byte) ([][]byte, error) {
	var (
		address = common.BytesToAddress(args[0])
		input   = args[1]
		gas     = env.callGasTemp
	)
	output, gasLeft, err := env.caller.CallDelegate(address, input, gas)
	env.contract.Gas += gasLeft
	return [][]byte{output, utils.EncodeError(err)}, nil
}

func gasCreate(env *Env, args [][]byte) (uint64, error) {
	if len(args) != 2 {
		return 0, ErrInvalidInput
	}
	if len(args[1]) != 32 {
		return 0, ErrInvalidInput
	}
	// We assume len() to always be much smaller than 32 * MAX_UINT64 / InitCodeWordGas
	// so this cannot overflow
	wordSize := toWordSize(len(args[0]))
	gas := wordSize * params.InitCodeWordGas
	return gas, nil
}

func opCreate(env *Env, args [][]byte) ([][]byte, error) {
	var (
		input = args[0]
		value = new(uint256.Int).SetBytes(args[1])
		gas   = env.contract.Gas
	)
	gas -= gas / 64
	env.useGas(gas) // This will always return true since we are using a fraction of the gas left
	ret, address, gasLeft, err := env.caller.Create(input, gas, value)
	env.contract.Gas += gasLeft
	return [][]byte{ret, address.Bytes(), utils.EncodeError(err)}, nil
}

func gasCreate2(env *Env, args [][]byte) (uint64, error) {
	if len(args) != 3 {
		return 0, ErrInvalidInput
	}
	if len(args[1]) != 32 || len(args[2]) != 32 {
		return 0, ErrInvalidInput
	}
	// We assume len() to always be much smaller than 32 * MAX_UINT64 / (InitCodeWordGas + Keccak256WordGas)
	// so this cannot overflow
	wordSize := toWordSize(len(args[0]))
	gas := wordSize * (params.InitCodeWordGas + params.Keccak256WordGas)
	return gas, nil
}

func opCreate2(env *Env, args [][]byte) ([][]byte, error) {
	var (
		input = args[0]
		value = new(uint256.Int).SetBytes(args[1])
		salt  = new(uint256.Int).SetBytes(args[2])
		gas   = env.contract.Gas
	)
	gas -= gas / 64
	env.useGas(gas) // This will always return true since we are using a fraction of the gas left
	ret, address, gasLeft, err := env.caller.Create2(input, gas, value, salt)
	env.contract.Gas += gasLeft
	return [][]byte{ret, address.Bytes(), utils.EncodeError(err)}, nil
}
