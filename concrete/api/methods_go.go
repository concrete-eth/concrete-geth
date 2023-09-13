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
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/concrete/crypto"
	"github.com/ethereum/go-ethereum/concrete/utils"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

const (
	PreimageStoreGas uint64 = 100
	PreimageDataGas  uint64 = 200
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
			execute:     opPersistentPreimageStore,
			constantGas: PreimageStoreGas,
			dynamicGas:  gasPersistentPreimageStore,
			trusted:     true,
			static:      false,
		},
		PersistentPreimageLoad_OpCode: {
			execute:    opPersistentPreimageLoad,
			dynamicGas: gasPersistentPreimageLoad,
			trusted:    true,
			static:     true,
		},
		PersistentPreimageLoadSize_OpCode: {
			execute:    opPersistentPreimageLoadSize,
			dynamicGas: gasPersistentPreimageLoadSize,
			trusted:    true,
			static:     true,
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
			// disabled
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
		UseGas_OpCode: {
			execute:     opUseGas,
			constantGas: GasQuickStep,
			static:      true,
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
	if env.statedb.AddressInAccessList(address) {
		env.statedb.AddAddressToAccessList(address)
		return params.ColdAccountAccessCostEIP2929 - params.WarmStorageReadCostEIP2929, nil
	}
	return 0, nil
}

func gasHashAccess(env *Env, hash common.Hash) (uint64, error) {
	if env.statedb.HashInAccessList(hash) {
		env.statedb.AddHashToAccessList(hash)
		return params.ColdSloadCostEIP2929, nil
	}
	return params.WarmStorageReadCostEIP2929, nil
}

func gasExternalCall(env *Env, address common.Address, callCost uint64) (uint64, error) {
	baseCost, err := gasAccountAccessMinusWarm(env, address)
	if err != nil {
		return 0, err
	}
	gasAvailable, overflow := math.SafeSub(env.gas, baseCost)
	if overflow {
		return 0, ErrGasUintOverflow
	}
	gas := gasAvailable - gasAvailable/64
	if gas < callCost {
		env.callGasTemp = gas
	} else {
		env.callGasTemp = callCost
	}
	// baseCost < MAX_UINT64 / 64, so this cannot overflow
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
	if env.meterGas == meter {
		return nil, nil
	}
	env.meterGas = meter
	return nil, nil
}

func opDebug(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, ErrInvalidInput
	}
	msg := string(args[0])
	env.logger.Debug(msg)
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

func opEphemeralStore(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 2 {
		return nil, ErrInvalidInput
	}
	if len(args[0]) != 32 || len(args[1]) != 32 {
		return nil, ErrInvalidInput
	}
	if !env.config.Ephemeral {
		return nil, ErrFeatureDisabled
	}
	key := common.BytesToHash(args[0])
	value := common.BytesToHash(args[1])
	env.statedb.SetEphemeralState(env.address, key, value)
	return nil, nil
}

func opEphemeralLoad(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, ErrInvalidInput
	}
	if len(args[0]) != 32 {
		return nil, ErrInvalidInput
	}
	if !env.config.Ephemeral {
		return nil, ErrFeatureDisabled
	}
	key := common.BytesToHash(args[0])
	value := env.statedb.GetEphemeralState(env.address, key)
	return [][]byte{value.Bytes()}, nil
}

func gasPersistentPreimageStore(env *Env, args [][]byte) (uint64, error) {
	if len(args) != 1 {
		return 0, ErrInvalidInput
	}
	preimage := args[0]
	// We assume len() to always be much smaller than MAX_UINT64 / (Keccak256WordGas + PreimageDataGas + CopyGas)
	// so this cannot overflow
	size := uint64(len(preimage))
	wordSize := toWordSize(len(preimage))
	keccakGas := wordSize * params.Keccak256WordGas
	dataGas := (size) * PreimageDataGas
	copyGas := wordSize * params.CopyGas
	return keccakGas + dataGas + copyGas, nil
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

func gasPersistentPreimageLoad(env *Env, args [][]byte) (uint64, error) {
	if len(args) != 1 {
		return 0, ErrInvalidInput
	}
	if len(args[0]) != 32 {
		return 0, ErrInvalidInput
	}
	hash := common.BytesToHash(args[0])
	return gasHashAccess(env, hash)
}

func opPersistentPreimageLoad(env *Env, args [][]byte) ([][]byte, error) {
	if !env.config.Preimages {
		return nil, ErrFeatureDisabled
	}
	key := common.BytesToHash(args[0])
	preimage := env.statedb.GetPersistentPreimage(key)
	return [][]byte{preimage}, nil
}

func gasPersistentPreimageLoadSize(env *Env, args [][]byte) (uint64, error) {
	if len(args) != 1 {
		return 0, ErrInvalidInput
	}
	if len(args[0]) != 32 {
		return 0, ErrInvalidInput
	}
	hash := common.BytesToHash(args[0])
	return gasHashAccess(env, hash)
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
	if len(args) != 1 {
		return nil, ErrInvalidInput
	}
	if !env.config.Ephemeral || !env.config.Preimages {
		return nil, ErrFeatureDisabled
	}
	preimage := args[0]
	hash := crypto.Keccak256Hash(preimage)
	env.statedb.AddEphemeralPreimage(hash, preimage)
	return [][]byte{hash.Bytes()}, nil
}

func opEphemeralPreimageLoad(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, ErrInvalidInput
	}
	if len(args[0]) != 32 {
		return nil, ErrInvalidInput
	}
	if !env.config.Ephemeral || !env.config.Preimages {
		return nil, ErrFeatureDisabled
	}
	key := common.BytesToHash(args[0])
	preimage := env.statedb.GetEphemeralPreimage(key)
	return [][]byte{preimage}, nil
}

func opEphemeralPreimageLoadSize(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, ErrInvalidInput
	}
	if len(args[0]) != 32 {
		return nil, ErrInvalidInput
	}
	if !env.config.Ephemeral || !env.config.Preimages {
		return nil, ErrFeatureDisabled
	}
	key := common.BytesToHash(args[0])
	size := env.statedb.GetEphemeralPreimageSize(key)
	return [][]byte{utils.Uint64ToBytes(uint64(size))}, nil
}

func opGetAddress(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	return [][]byte{env.address.Bytes()}, nil
}

func opGetGasLeft(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	return [][]byte{utils.Uint64ToBytes(env.gas)}, nil
}

func opGetBlockNumber(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	if env.block == nil {
		return nil, ErrNoData
	}
	number := env.block.BlockNumber()
	return [][]byte{utils.Uint64ToBytes(number)}, nil
}

func opGetBlockGasLimit(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	if env.block == nil {
		return nil, ErrNoData
	}
	limit := env.block.GasLimit()
	return [][]byte{utils.Uint64ToBytes(limit)}, nil
}

func opGetBlockTimestamp(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	if env.block == nil {
		return nil, ErrNoData
	}
	timestamp := env.block.Timestamp()
	return [][]byte{utils.Uint64ToBytes(timestamp)}, nil
}

func opGetBlockDifficulty(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	if env.block == nil {
		return nil, ErrNoData
	}
	difficulty := env.block.Difficulty()
	return [][]byte{difficulty.Bytes()}, nil
}

func opGetBlockBaseFee(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	if env.block == nil {
		return nil, ErrNoData
	}
	baseFee := env.block.BaseFee()
	return [][]byte{baseFee.Bytes()}, nil
}

func opGetBlockCoinbase(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	if env.block == nil {
		return nil, ErrNoData
	}
	coinbase := env.block.Coinbase()
	return [][]byte{coinbase.Bytes()}, nil
}

func opGetPrevRandom(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	if env.block == nil {
		return nil, ErrNoData
	}
	random := env.block.Random()
	return [][]byte{random.Bytes()}, nil
}

func opGetBlockHash(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, ErrInvalidInput
	}
	if env.block == nil {
		return nil, ErrNoData
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
	balance := env.statedb.GetBalance(env.address)
	return [][]byte{balance.Bytes()}, nil
}

func opGetTxGasPrice(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	if env.call == nil {
		return nil, ErrNoData
	}
	price := env.call.TxGasPrice()
	return [][]byte{price.Bytes()}, nil
}

func opGetTxOrigin(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
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
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	return [][]byte{dataCopy}, nil
}

func opGetCallDataSize(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	if env.call == nil {
		return nil, ErrNoData
	}
	size := env.call.CallDataSize()
	return [][]byte{utils.Uint64ToBytes(uint64(size))}, nil
}

func opGetCaller(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	if env.call == nil {
		return nil, ErrNoData
	}
	caller := env.call.Caller()
	return [][]byte{caller.Bytes()}, nil
}

func opGetCallValue(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	if env.call == nil {
		return nil, ErrNoData
	}
	value := env.call.CallValue()
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
	if _, slotPresent := env.statedb.SlotInAccessList(env.address, key); slotPresent {
		env.statedb.AddSlotToAccessList(env.address, key)
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
	if len(args) != 0 {
		return nil, ErrInvalidInput
	}
	return nil, ErrInvalidOpCode
}

func opGetCodeSize(env *Env, args [][]byte) ([][]byte, error) {
	return nil, ErrInvalidOpCode
}

func opUseGas(env *Env, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, ErrInvalidInput
	}
	if len(args[0]) != 8 {
		return nil, ErrInvalidInput
	}
	gas := utils.BytesToUint64(args[0])
	if ok := env.useGas(gas); !ok {
		return nil, ErrOutOfGas
	}
	return nil, nil
}

func gasStorageStore(env *Env, args [][]byte) (uint64, error) {
	if len(args) != 2 {
		return 0, ErrInvalidInput
	}
	if len(args[0]) != 32 || len(args[1]) != 32 {
		return 0, ErrInvalidInput
	}
	if env.gas <= params.SstoreSentryGasEIP2200 {
		return 0, errors.New("not enough gas for reentrancy sentry")
	}
	var (
		key     = common.BytesToHash(args[0])
		current = env.statedb.GetPersistentState(env.address, key)
		cost    = uint64(0)
	)
	if _, slotPresent := env.statedb.SlotInAccessList(env.address, key); !slotPresent {
		cost = params.ColdSloadCostEIP2929
		env.statedb.AddSlotToAccessList(env.address, key)
	}
	value := common.BytesToHash(args[1])
	if current == value {
		return cost + params.WarmStorageReadCostEIP2929, nil
	}
	original := env.statedb.GetCommittedState(env.address, key)
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
	env.statedb.SetPersistentState(env.address, key, value)
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
	// We assume len() to always be much smaller than (MAX_UINT64 - LogGas - 4 * LogTopicGas) / LogDataGas
	// so this cannot overflow
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
		Address:     env.address,
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
	if env.caller == nil {
		return nil, ErrNoData
	}
	var (
		address = common.BytesToAddress(args[0])
		input   = args[1]
		gas     = env.callGasTemp
	)
	inputCopy := make([]byte, len(input))
	copy(inputCopy, input)
	output, gasLeft, err := env.caller.CallStatic(address, inputCopy, gas)
	env.gas += gasLeft
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
	if env.caller == nil {
		return nil, ErrNoData
	}
	var (
		address = common.BytesToAddress(args[0])
		input   = args[1]
		gas     = env.callGasTemp
		value   = new(big.Int).SetBytes(args[3])
	)
	inputCopy := make([]byte, len(input))
	copy(inputCopy, input)
	output, gasLeft, err := env.caller.Call(address, inputCopy, gas, value)
	env.gas += gasLeft
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
	if env.caller == nil {
		return nil, ErrNoData
	}
	var (
		address = common.BytesToAddress(args[0])
		input   = args[1]
		gas     = env.callGasTemp
	)
	inputCopy := make([]byte, len(input))
	copy(inputCopy, input)
	output, gasLeft, err := env.caller.CallDelegate(address, inputCopy, gas)
	env.gas += gasLeft
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
	if env.caller == nil {
		return nil, ErrNoData
	}
	var (
		input = args[0]
		value = new(big.Int).SetBytes(args[1])
		gas   = env.gas
	)
	inputCopy := make([]byte, len(input))
	copy(inputCopy, input)
	gas -= gas / 64
	env.useGas(gas) // This will always return true since we are using a fraction of the gas left
	address, gasLeft, err := env.caller.Create(inputCopy, gas, value)
	env.gas += gasLeft
	return [][]byte{address.Bytes(), utils.EncodeError(err)}, nil
}

func gasCreate2(env *Env, args [][]byte) (uint64, error) {
	if len(args) != 3 {
		return 0, ErrInvalidInput
	}
	if len(args[1]) != 32 || len(args[2]) != 32 {
		return 0, ErrInvalidInput
	}
	if len(args) != 2 {
		return 0, ErrInvalidInput
	}
	if len(args[1]) != 32 {
		return 0, ErrInvalidInput
	}
	// We assume len() to always be much smaller than 32 * MAX_UINT64 / (InitCodeWordGas + Keccak256WordGas)
	// so this cannot overflow
	wordSize := toWordSize(len(args[0]))
	gas := wordSize * (params.InitCodeWordGas + params.Keccak256WordGas)
	return gas, nil
}

func opCreate2(env *Env, args [][]byte) ([][]byte, error) {
	if env.caller == nil {
		return nil, ErrNoData
	}
	var (
		input = args[0]
		value = new(big.Int).SetBytes(args[1])
		salt  = common.BytesToHash(args[2])
		gas   = env.gas
	)
	inputCopy := make([]byte, len(input))
	copy(inputCopy, input)
	gas -= gas / 64
	env.useGas(gas) // This will always return true since we are using a fraction of the gas left
	address, gasLeft, err := env.caller.Create2(inputCopy, salt, gas, value)
	env.gas += gasLeft
	return [][]byte{address.Bytes(), utils.EncodeError(err)}, nil
}
