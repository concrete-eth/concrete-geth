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
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

type OpCode byte

const (
	EphemeralStore_OpCode             OpCode = 0x10
	EphemeralLoad_OpCode              OpCode = 0x11
	PersistentPreimageStore_OpCode    OpCode = 0x12
	PersistentPreimageLoad_OpCode     OpCode = 0x13
	PersistentPreimageLoadSize_OpCode OpCode = 0x14
	EphemeralPreimageStore_OpCode     OpCode = 0x15
	EphemeralPreimageLoad_OpCode      OpCode = 0x16
	EphemeralPreimageLoadSize_OpCode  OpCode = 0x17
	GetAddress_OpCode                 OpCode = 0x18
	GetGasLeft_OpCode                 OpCode = 0x19
	GetBlockNumber_OpCode             OpCode = 0x1a
	GetBlockGasLimit_OpCode           OpCode = 0x1b
	GetBlockTimestamp_OpCode          OpCode = 0x1c
	GetBlockDifficulty_OpCode         OpCode = 0x1d
	GetBlockBasefee_OpCode            OpCode = 0x1e
	GetBlockCoinbase_OpCode           OpCode = 0x1f
	GetPrevRandao_OpCode              OpCode = 0x20
	GetBlockHash_OpCode               OpCode = 0x21
	GetBalance_OpCode                 OpCode = 0x22
	GetTxGasPrice_OpCode              OpCode = 0x23
	GetTxOrigin_OpCode                OpCode = 0x24
	GetCallData_OpCode                OpCode = 0x25
	GetCallDataSize_OpCode            OpCode = 0x26
	GetCaller_OpCode                  OpCode = 0x27
	GetCallValue_OpCode               OpCode = 0x28
	StorageLoad_OpCode                OpCode = 0x29
	GetCode_OpCode                    OpCode = 0x2a
	GetCodeSize_OpCode                OpCode = 0x2b
	UseGas_OpCode                     OpCode = 0x2c
	StorageStore_OpCode               OpCode = 0x2d
	Log_OpCode                        OpCode = 0x2e
	GetExternalBalance_OpCode         OpCode = 0x2f
	CallStatic_OpCode                 OpCode = 0x30
	GetExternalCode_OpCode            OpCode = 0x31
	GetExternalCodeSize_OpCode        OpCode = 0x32
	Call_OpCode                       OpCode = 0x33
	CallDelegate_OpCode               OpCode = 0x34
	Create_OpCode                     OpCode = 0x35
	Create2_OpCode                    OpCode = 0x36
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
	executionFunc func(api *API, args [][]byte) ([][]byte, error)
	gasFunc       func(api *API) (uint64, error)
)

type operation struct {
	execute     executionFunc
	constantGas uint64
	dynamicGas  gasFunc
}

type JumpTable [64]*operation

func NewConcreteAPIMethods() JumpTable {
	tbl := JumpTable{
		EphemeralLoad_OpCode: {
			execute:     opEphemeralLoad,
			constantGas: params.WarmStorageReadCostEIP2929,
		},
		EphemeralStore_OpCode: {
			execute:     opEphemeralStore,
			constantGas: params.WarmStorageReadCostEIP2929,
		},
		PersistentPreimageStore_OpCode: {
			execute: opPersistentPreimageStore,
		},
		PersistentPreimageLoad_OpCode: {
			execute:     opPersistentPreimageLoad,
			constantGas: params.ExtcodeCopyBaseEIP150,
		},
		PersistentPreimageLoadSize_OpCode: {
			execute:     opPersistentPreimageLoadSize,
			constantGas: params.ExtcodeSizeGasEIP150,
		},
		EphemeralPreimageStore_OpCode: {
			execute: opEphemeralPreimageStore,
		},
		EphemeralPreimageLoad_OpCode: {
			execute: opEphemeralPreimageLoad,
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
			execute:     opStorageLoad,
			constantGas: params.SloadGasEIP150,
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
			execute: opStorageStore,
		},
		Log_OpCode: {
			execute: opLog,
		},
		GetExternalBalance_OpCode: {
			execute:     opGetExternalBalance,
			constantGas: params.BalanceGasEIP150,
		},
		CallStatic_OpCode: {
			execute:     opCallStatic,
			constantGas: params.CallGasEIP150,
		},
		GetExternalCode_OpCode: {
			execute:     opGetExternalCode,
			constantGas: params.ExtcodeCopyBaseEIP150,
		},
		GetExternalCodeSize_OpCode: {
			execute:     opGetExternalCodeSize,
			constantGas: params.ExtcodeSizeGasEIP150,
		},
		Call_OpCode: {
			execute:     opCall,
			constantGas: params.CallGasEIP150,
		},
		CallDelegate_OpCode: {
			execute:     opCallDelegate,
			constantGas: params.CallGasEIP150,
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

func opUndefined(api *API, args [][]byte) ([][]byte, error) {
	return nil, fmt.Errorf("undefined opcode")
}

// Permission checks panic on failure because the imply an error in implementation, not in
// execution. A functioning VM should never fail a permission check.

func opEphemeralStore(api *API, args [][]byte) ([][]byte, error) {
	if !api.storageConfig.HasEphemeral || api.storageConfig.EphemeralReadOnly {
		panic("ephemeral store is disabled or read-only")
	}
	key := common.BytesToHash(args[0])
	value := common.BytesToHash(args[1])
	api.statedb.SetEphemeralState(api.address, key, value)
	return nil, nil
}

func opEphemeralLoad(api *API, args [][]byte) ([][]byte, error) {
	if !api.storageConfig.HasEphemeral {
		panic("ephemeral store is disabled")
	}
	key := common.BytesToHash(args[0])
	data := api.statedb.GetEphemeralState(api.address, key)
	return [][]byte{data.Bytes()}, nil
}

func opPersistentPreimageStore(api *API, args [][]byte) ([][]byte, error) {
	if !api.preimageConfig.HasPersistent || api.preimageConfig.PersistentReadOnly {
		panic("persistent store is disabled or read-only")
	}
	key := common.BytesToHash(args[0])
	value := args[1]
	api.statedb.AddPersistentPreimage(key, value)
	return nil, nil
}

func opPersistentPreimageLoad(api *API, args [][]byte) ([][]byte, error) {
	if !api.preimageConfig.HasPersistent {
		panic("persistent store is disabled")
	}
	key := common.BytesToHash(args[0])
	data := api.statedb.GetPersistentPreimage(key)
	return [][]byte{data}, nil
}

func opPersistentPreimageLoadSize(api *API, args [][]byte) ([][]byte, error) {
	if !api.preimageConfig.HasPersistent {
		panic("persistent store is disabled")
	}
	key := common.BytesToHash(args[0])
	size := api.statedb.GetPersistentPreimageSize(key)
	return [][]byte{Uint64ToBytes(uint64(size))}, nil
}

func opEphemeralPreimageStore(api *API, args [][]byte) ([][]byte, error) {
	if !api.preimageConfig.HasEphemeral || api.preimageConfig.EphemeralReadOnly {
		panic("ephemeral store is disabled or read-only")
	}
	key := common.BytesToHash(args[0])
	value := args[1]
	api.statedb.AddEphemeralPreimage(key, value)
	return nil, nil
}

func opEphemeralPreimageLoad(api *API, args [][]byte) ([][]byte, error) {
	if !api.preimageConfig.HasEphemeral {
		panic("ephemeral store is disabled")
	}
	key := common.BytesToHash(args[0])
	data := api.statedb.GetEphemeralPreimage(key)
	return [][]byte{data}, nil
}

func opEphemeralPreimageLoadSize(api *API, args [][]byte) ([][]byte, error) {
	if !api.preimageConfig.HasEphemeral {
		panic("ephemeral store is disabled")
	}
	key := common.BytesToHash(args[0])
	size := api.statedb.GetEphemeralPreimageSize(key)
	return [][]byte{Uint64ToBytes(uint64(size))}, nil
}

func opGetAddress(api *API, args [][]byte) ([][]byte, error) {
	return [][]byte{api.address.Bytes()}, nil
}

func opGetGasLeft(api *API, args [][]byte) ([][]byte, error) {
	return [][]byte{Uint64ToBytes(api.gas)}, nil
}

func opGetBlockNumber(api *API, args [][]byte) ([][]byte, error) {
	return nil, nil
}

func opGetBlockGasLimit(api *API, args [][]byte) ([][]byte, error) {
	return nil, nil
}

func opGetBlockTimestamp(api *API, args [][]byte) ([][]byte, error) {
	return nil, nil
}

func opGetBlockDifficulty(api *API, args [][]byte) ([][]byte, error) {
	return nil, nil
}

func opGetBlockBasefee(api *API, args [][]byte) ([][]byte, error) {
	return nil, nil
}

func opGetBlockCoinbase(api *API, args [][]byte) ([][]byte, error) {
	return nil, nil
}

func opGetPrevRandao(api *API, args [][]byte) ([][]byte, error) {
	return nil, nil
}

func opGetBlockHash(api *API, args [][]byte) ([][]byte, error) {
	return nil, nil
}

func opGetBalance(api *API, args [][]byte) ([][]byte, error) {
	balance := api.statedb.GetBalance(api.address)
	return [][]byte{balance.Bytes()}, nil
}

func opGetTxGasPrice(api *API, args [][]byte) ([][]byte, error) {
	return nil, nil
}

func opGetTxOrigin(api *API, args [][]byte) ([][]byte, error) {
	return nil, nil
}

func opGetCallData(api *API, args [][]byte) ([][]byte, error) {
	return nil, nil
}

func opGetCallDataSize(api *API, args [][]byte) ([][]byte, error) {
	return nil, nil
}

func opGetCaller(api *API, args [][]byte) ([][]byte, error) {
	return nil, nil
}

func opGetCallValue(api *API, args [][]byte) ([][]byte, error) {
	return nil, nil
}

func opStorageLoad(api *API, args [][]byte) ([][]byte, error) {
	return nil, nil
}

func opGetCode(api *API, args [][]byte) ([][]byte, error) {
	return nil, nil
}

func opGetCodeSize(api *API, args [][]byte) ([][]byte, error) {
	return nil, nil
}

func opUseGas(api *API, args [][]byte) ([][]byte, error) {
	gas := BytesToUint64(args[0])
	api.gas -= gas
	return nil, nil
}

func opStorageStore(api *API, args [][]byte) ([][]byte, error) {
	return nil, nil
}

func opLog(api *API, args [][]byte) ([][]byte, error) {
	return nil, nil
}

func opGetExternalBalance(api *API, args [][]byte) ([][]byte, error) {
	return nil, nil
}

func opCallStatic(api *API, args [][]byte) ([][]byte, error) {
	return nil, nil
}

func opGetExternalCode(api *API, args [][]byte) ([][]byte, error) {
	return nil, nil
}

func opGetExternalCodeSize(api *API, args [][]byte) ([][]byte, error) {
	return nil, nil
}

func opCall(api *API, args [][]byte) ([][]byte, error) {
	return nil, nil
}

func opCallDelegate(api *API, args [][]byte) ([][]byte, error) {
	return nil, nil
}

func opCreate(api *API, args [][]byte) ([][]byte, error) {
	return nil, nil
}

func opCreate2(api *API, args [][]byte) ([][]byte, error) {
	return nil, nil
}
