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
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

/*
eth_interface methods not included
ReturnDataCopy
GetReturnDataSize
Finish
Revert
CallDataCopy
CallCode
ExternalCodeCopy
SelfDestruct

evm_interface-like methods added
GetBalance
*/

type ConcreteAPI interface {
	// Meta
	EnableGasMetering(meter bool)

	// Wrappers
	Persistent() Datastore
	Ephemeral() Datastore
	Block() Block

	// Aliases
	PersistentLoad(key common.Hash) common.Hash
	PersistentStore(key common.Hash, value common.Hash)

	// Ephemeral
	EphemeralLoad(key common.Hash) common.Hash
	EphemeralStore(key common.Hash, value common.Hash)

	// Preimage oracle
	PersistentPreimageStore(hash common.Hash, preimage []byte)
	PersistentPreimageLoad(hash common.Hash) []byte
	PersistentPreimageLoadSize(hash common.Hash) int
	EphemeralPreimageStore(hash common.Hash, preimage []byte)
	EphemeralPreimageLoad(hash common.Hash) []byte
	EphemeralPreimageLoadSize(hash common.Hash) int

	// INTERNAL - READ
	// Address
	GetAddress() common.Address
	// Gas
	GetGasLeft() uint64
	// Block
	GetBlockNumber() uint64
	GetBlockGasLimit() uint64
	GetBlockTimestamp() uint64
	GetBlockDifficulty() *big.Int
	GetBlockBasefee() *big.Int
	GetBlockCoinbase() common.Address
	GetPrevRandao() common.Hash
	// Block hash
	GetBlockHash(block uint64) common.Hash
	// Balance
	GetBalance(address common.Address) *big.Int
	// Transaction
	GetTxGasPrice() *big.Int
	GetTxOrigin() common.Address
	// Call
	GetCallData() []byte
	GetCallDataSize() int
	GetCaller() common.Address
	GetCallValue() *big.Int
	// Storage
	StorageLoad(key common.Hash) common.Hash
	// Code
	GetCode(address common.Address) []byte
	GetCodeSize() int

	// INTERNAL - WRITE
	// Gas
	UseGas(amount uint64)
	// Storage
	StorageStore(key common.Hash, value common.Hash)
	// Log
	Log(topics []common.Hash, data []byte)

	// EXTERNAL - READ
	// Balance
	GetExternalBalance(address common.Address) *big.Int
	// Call
	CallStatic(gas uint64, address common.Address, data []byte) ([]byte, uint64, error)
	// Code
	GetExternalCode(address common.Address) []byte
	GetExternalCodeSize(address common.Address) int

	// EXTERNAL - WRITE
	// Call
	Call(gas uint64, address common.Address, value *big.Int, data []byte) ([]byte, uint64, error)
	CallDelegate(gas uint64, address common.Address, data []byte) ([]byte, uint64, error)
	// Create
	Create(value *big.Int, data []byte) common.Address
	Create2(value *big.Int, data []byte, salt common.Hash) common.Address
}

type StoreConfig struct {
	HasPersistent      bool
	HasEphemeral       bool
	PersistentReadOnly bool
	EphemeralReadOnly  bool
}

type API struct {
	storageConfig         StoreConfig
	preimageConfig        StoreConfig
	meterGas              bool
	canDisableGasMetering bool
	address               common.Address
	statedb               StateDB
	table                 JumpTable
	execute               func(op OpCode, api *API, args [][]byte) ([][]byte, error)

	gas uint64
}

func NewAPI(address common.Address, statedb StateDB, storageConfig StoreConfig, preimageConfig StoreConfig, meterGas bool, canDisableGasMetering bool) *API {
	api := &API{
		storageConfig:         storageConfig,
		preimageConfig:        preimageConfig,
		meterGas:              meterGas,
		canDisableGasMetering: canDisableGasMetering,
		address:               address,
		statedb:               statedb,
	}
	api.table = NewConcreteAPIMethods()
	api.execute = func(op OpCode, api *API, args [][]byte) ([][]byte, error) {
		return api.table[op].execute(api, args)
	}
	return api
}

func (api *API) EnableGasMetering(meter bool) {
	if api.meterGas == meter {
		return
	}
	if !api.canDisableGasMetering {
		panic("cannot disable gas metering")
	}
	api.meterGas = meter
}

func (api *API) Persistent() Datastore {
	return nil
}

func (api *API) Ephemeral() Datastore {
	return nil
}

func (api *API) Block() Block {
	return &BlockData{api: api}
}

func (api *API) PersistentLoad(key common.Hash) common.Hash {
	return api.StorageLoad(key)
}

func (api *API) PersistentStore(key common.Hash, value common.Hash) {
	api.StorageStore(key, value)
}

func (api *API) EphemeralLoad(key common.Hash) common.Hash {
	input := [][]byte{key.Bytes()}
	output, _ := api.execute(EphemeralLoad_OpCode, api, input)
	hash := common.BytesToHash(output[0])
	return hash
}

func (api *API) EphemeralStore(key common.Hash, value common.Hash) {
	input := [][]byte{key.Bytes(), value.Bytes()}
	api.execute(EphemeralStore_OpCode, api, input)
}

func (api *API) PersistentPreimageStore(hash common.Hash, preimage []byte) {
	input := [][]byte{hash.Bytes(), preimage}
	api.execute(PersistentPreimageStore_OpCode, api, input)
}

func (api *API) PersistentPreimageLoad(hash common.Hash) []byte {
	input := [][]byte{hash.Bytes()}
	output, _ := api.execute(PersistentPreimageLoad_OpCode, api, input)
	return output[0]
}

func (api *API) PersistentPreimageLoadSize(hash common.Hash) int {
	input := [][]byte{hash.Bytes()}
	output, _ := api.execute(PersistentPreimageLoadSize_OpCode, api, input)
	return int(BytesToUint64(output[0]))
}

func (api *API) EphemeralPreimageStore(hash common.Hash, preimage []byte) {
	input := [][]byte{hash.Bytes(), preimage}
	api.execute(EphemeralPreimageStore_OpCode, api, input)
}

func (api *API) EphemeralPreimageLoad(hash common.Hash) []byte {
	input := [][]byte{hash.Bytes()}
	output, _ := api.execute(EphemeralPreimageLoad_OpCode, api, input)
	return output[0]
}

func (api *API) EphemeralPreimageLoadSize(hash common.Hash) int {
	input := [][]byte{hash.Bytes()}
	output, _ := api.execute(EphemeralPreimageLoadSize_OpCode, api, input)
	return int(BytesToUint64(output[0]))
}

func (api *API) GetAddress() common.Address {
	output, _ := api.execute(GetAddress_OpCode, api, nil)
	return common.BytesToAddress(output[0])
}

func (api *API) GetGasLeft() uint64 {
	output, _ := api.execute(GetGasLeft_OpCode, api, nil)
	return BytesToUint64(output[0])
}

func (api *API) GetBlockNumber() uint64 {
	output, _ := api.execute(GetBlockNumber_OpCode, api, nil)
	return BytesToUint64(output[0])
}

func (api *API) GetBlockGasLimit() uint64 {
	output, _ := api.execute(GetBlockGasLimit_OpCode, api, nil)
	return BytesToUint64(output[0])
}

func (api *API) GetBlockTimestamp() uint64 {
	output, _ := api.execute(GetBlockTimestamp_OpCode, api, nil)
	return BytesToUint64(output[0])
}

func (api *API) GetBlockDifficulty() *big.Int {
	output, _ := api.execute(GetBlockDifficulty_OpCode, api, nil)
	return new(big.Int).SetBytes(output[0])
}

func (api *API) GetBlockBasefee() *big.Int {
	output, _ := api.execute(GetBlockBasefee_OpCode, api, nil)
	return new(big.Int).SetBytes(output[0])
}

func (api *API) GetBlockCoinbase() common.Address {
	output, _ := api.execute(GetBlockCoinbase_OpCode, api, nil)
	return common.BytesToAddress(output[0])
}

func (api *API) GetPrevRandao() common.Hash {
	output, _ := api.execute(GetPrevRandao_OpCode, api, nil)
	return common.BytesToHash(output[0])
}

func (api *API) GetBlockHash(number uint64) common.Hash {
	input := [][]byte{Uint64ToBytes(number)}
	output, _ := api.execute(GetBlockHash_OpCode, api, input)
	return common.BytesToHash(output[0])
}

func (api *API) GetBalance(address common.Address) *big.Int {
	input := [][]byte{address.Bytes()}
	output, _ := api.execute(GetBalance_OpCode, api, input)
	return new(big.Int).SetBytes(output[0])
}

func (api *API) GetTxGasPrice() *big.Int {
	output, _ := api.execute(GetTxGasPrice_OpCode, api, nil)
	return new(big.Int).SetBytes(output[0])
}

func (api *API) GetTxOrigin() common.Address {
	output, _ := api.execute(GetTxOrigin_OpCode, api, nil)
	return common.BytesToAddress(output[0])
}

func (api *API) GetCallData() []byte {
	output, _ := api.execute(GetCallData_OpCode, api, nil)
	return output[0]
}

func (api *API) GetCallDataSize() int {
	output, _ := api.execute(GetCallDataSize_OpCode, api, nil)
	return int(BytesToUint64(output[0]))
}

func (api *API) GetCaller() common.Address {
	output, _ := api.execute(GetCaller_OpCode, api, nil)
	return common.BytesToAddress(output[0])
}

func (api *API) GetCallValue() *big.Int {
	output, _ := api.execute(GetCallValue_OpCode, api, nil)
	return new(big.Int).SetBytes(output[0])
}

func (api *API) StorageLoad(key common.Hash) common.Hash {
	input := [][]byte{key.Bytes()}
	output, _ := api.execute(StorageLoad_OpCode, api, input)
	return common.BytesToHash(output[0])
}

func (api *API) GetCode(address common.Address) []byte {
	input := [][]byte{address.Bytes()}
	output, _ := api.execute(GetCode_OpCode, api, input)
	return output[0]
}

func (api *API) GetCodeSize() int {
	output, _ := api.execute(GetCodeSize_OpCode, api, nil)
	return int(BytesToUint64(output[0]))
}

func (api *API) UseGas(gas uint64) {
	input := [][]byte{Uint64ToBytes(gas)}
	api.execute(UseGas_OpCode, api, input)
}

func (api *API) StorageStore(key common.Hash, value common.Hash) {
	input := [][]byte{key.Bytes(), value.Bytes()}
	api.execute(StorageStore_OpCode, api, input)
}

func (api *API) Log(topics []common.Hash, data []byte) {
	input := make([][]byte, len(topics)+1)
	for i := 0; i < len(topics); i++ {
		input[i] = topics[i].Bytes()
	}
	input[len(topics)] = data
	api.execute(Log_OpCode, api, input)
}

func (api *API) GetExternalBalance(address common.Address) *big.Int {
	input := [][]byte{address.Bytes()}
	output, _ := api.execute(GetExternalBalance_OpCode, api, input)
	return new(big.Int).SetBytes(output[0])
}

func (api *API) CallStatic(gas uint64, address common.Address, data []byte) ([]byte, uint64, error) {
	input := [][]byte{Uint64ToBytes(gas), address.Bytes(), data}
	output, err := api.execute(CallStatic_OpCode, api, input)
	return output[0], BytesToUint64(output[1]), err
}

func (api *API) GetExternalCode(address common.Address) []byte {
	input := [][]byte{address.Bytes()}
	output, _ := api.execute(GetExternalCode_OpCode, api, input)
	return output[0]
}

func (api *API) GetExternalCodeSize(address common.Address) int {
	input := [][]byte{address.Bytes()}
	output, _ := api.execute(GetExternalCodeSize_OpCode, api, input)
	return int(BytesToUint64(output[0]))
}

func (api *API) Call(gas uint64, address common.Address, value *big.Int, data []byte) ([]byte, uint64, error) {
	input := [][]byte{Uint64ToBytes(gas), address.Bytes(), value.Bytes(), data}
	output, err := api.execute(Call_OpCode, api, input)
	return output[0], BytesToUint64(output[1]), err
}

func (api *API) CallDelegate(gas uint64, address common.Address, data []byte) ([]byte, uint64, error) {
	input := [][]byte{Uint64ToBytes(gas), address.Bytes(), data}
	output, err := api.execute(CallDelegate_OpCode, api, input)
	return output[0], BytesToUint64(output[1]), err
}

func (api *API) Create(value *big.Int, data []byte) common.Address {
	input := [][]byte{value.Bytes(), data}
	output, _ := api.execute(Create_OpCode, api, input)
	return common.BytesToAddress(output[0])
}

func (api *API) Create2(value *big.Int, data []byte, salt common.Hash) common.Address {
	input := [][]byte{value.Bytes(), data, salt.Bytes()}
	output, _ := api.execute(Create2_OpCode, api, input)
	return common.BytesToAddress(output[0])
}

var _ ConcreteAPI = (*API)(nil)

type Precompile interface {
	IsReadOnly(input []byte) bool
	Finalise(api ConcreteAPI) error
	Commit(api ConcreteAPI) error
	Run(api ConcreteAPI, input []byte) ([]byte, error)
}
