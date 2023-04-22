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

type StateDB interface {
	SetPersistentState(addr common.Address, key common.Hash, value common.Hash)
	GetPersistentState(addr common.Address, key common.Hash) common.Hash
	SetEphemeralState(addr common.Address, key common.Hash, value common.Hash)
	GetEphemeralState(addr common.Address, key common.Hash) common.Hash
	AddPersistentPreimage(hash common.Hash, preimage []byte)
	GetPersistentPreimage(hash common.Hash) []byte
	GetPersistentPreimageSize(hash common.Hash) int
	AddEphemeralPreimage(hash common.Hash, preimage []byte)
	GetEphemeralPreimage(hash common.Hash) []byte
	GetEphemeralPreimageSize(hash common.Hash) int
}

type EVM interface {
	StateDB() StateDB
	BlockHash(block uint64) common.Hash
	BlockTimestamp() uint64
	BlockNumber() *big.Int
	BlockDifficulty() *big.Int
	BlockGasLimit() uint64
	BlockCoinbase() common.Address
}

type ReadOnlyEVM struct {
	EVM
}

func NewReadOnlyEVM(evm EVM) EVM {
	return &ReadOnlyEVM{evm}
}

func (evm *ReadOnlyEVM) StateDB() StateDB {
	return &ReadOnlyStateDB{evm.EVM.StateDB()}
}

var _ EVM = &ReadOnlyEVM{}

type ReadOnlyStateDB struct {
	StateDB
}

func (db *ReadOnlyStateDB) SetPersistentState(addr common.Address, key common.Hash, value common.Hash) {
	panic("stateDB write protection")
}

func (db *ReadOnlyStateDB) SetEphemeralState(addr common.Address, key common.Hash, value common.Hash) {
	panic("stateDB write protection")
}

func (db *ReadOnlyStateDB) AddPersistentPreimage(hash common.Hash, preimage []byte) {
	panic("stateDB write protection")
}

func (db *ReadOnlyStateDB) AddEphemeralPreimage(hash common.Hash, preimage []byte) {
	panic("stateDB write protection")
}

var _ StateDB = &ReadOnlyStateDB{}

type Storage interface {
	Set(key common.Hash, value common.Hash)
	Get(key common.Hash) common.Hash
	AddPreimage(hash common.Hash, preimage []byte)
	GetPreimage(hash common.Hash) []byte
	GetPreimageSize(hash common.Hash) int
}

type PersistentStorage struct {
	address common.Address
	db      StateDB
}

func (s *PersistentStorage) Set(key common.Hash, value common.Hash) {
	s.db.SetPersistentState(s.address, key, value)
}

func (s *PersistentStorage) Get(key common.Hash) common.Hash {
	return s.db.GetPersistentState(s.address, key)
}

func (s *PersistentStorage) AddPreimage(hash common.Hash, preimage []byte) {
	s.db.AddPersistentPreimage(hash, preimage)
}

func (s *PersistentStorage) GetPreimage(hash common.Hash) []byte {
	return s.db.GetPersistentPreimage(hash)
}

func (s *PersistentStorage) GetPreimageSize(hash common.Hash) int {
	return s.db.GetPersistentPreimageSize(hash)
}

var _ Storage = (*PersistentStorage)(nil)

type EphemeralStorage struct {
	address common.Address
	db      StateDB
}

func (s *EphemeralStorage) Set(key common.Hash, value common.Hash) {
	s.db.SetEphemeralState(s.address, key, value)
}

func (s *EphemeralStorage) Get(key common.Hash) common.Hash {
	return s.db.GetEphemeralState(s.address, key)
}

func (s *EphemeralStorage) AddPreimage(hash common.Hash, preimage []byte) {
	s.db.AddEphemeralPreimage(hash, preimage)
}

func (s *EphemeralStorage) GetPreimage(hash common.Hash) []byte {
	return s.db.GetEphemeralPreimage(hash)
}

func (s *EphemeralStorage) GetPreimageSize(hash common.Hash) int {
	return s.db.GetEphemeralPreimageSize(hash)
}

var _ Storage = (*EphemeralStorage)(nil)

type Block interface {
	Timestamp() uint64
	Number() *big.Int
	Difficulty() *big.Int
	GasLimit() uint64
	Coinbase() common.Address
}

type block struct {
	evm EVM
}

func (b *block) Timestamp() uint64        { return b.evm.BlockTimestamp() }
func (b *block) Number() *big.Int         { return b.evm.BlockNumber() }
func (b *block) Difficulty() *big.Int     { return b.evm.BlockDifficulty() }
func (b *block) GasLimit() uint64         { return b.evm.BlockGasLimit() }
func (b *block) Coinbase() common.Address { return b.evm.BlockCoinbase() }

var _ Block = (*block)(nil)

type API interface {
	Address() common.Address
	EVM() EVM
	StateDB() StateDB
	Persistent() Datastore
	Ephemeral() Datastore
	BlockHash(block uint64) common.Hash
	Block() Block
}

type stateApi struct {
	address common.Address
	db      StateDB
}

func NewStateAPI(db StateDB, address common.Address) API {
	return &stateApi{
		address: address,
		db:      db,
	}
}

func (s *stateApi) Address() common.Address {
	return s.address
}

func (s *stateApi) EVM() EVM {
	return nil
}

func (s *stateApi) StateDB() StateDB {
	return s.db
}

func (s *stateApi) Persistent() Datastore {
	return &datastore{&PersistentStorage{
		address: s.address,
		db:      s.db,
	}}
}

func (s *stateApi) Ephemeral() Datastore {
	return &datastore{&EphemeralStorage{
		address: s.address,
		db:      s.db,
	}}
}

func (s *stateApi) BlockHash(block uint64) common.Hash {
	panic("API method not available")
}

func (s *stateApi) Block() Block {
	panic("API method not available")
}

var _ API = (*stateApi)(nil)

type api struct {
	stateApi
	evm EVM
}

func New(evm EVM, address common.Address) API {
	return &api{
		stateApi: stateApi{address: address, db: evm.StateDB()},
		evm:      evm,
	}
}

func (a *api) EVM() EVM {
	return a.evm
}

func (a *api) BlockHash(block uint64) common.Hash {
	return a.evm.BlockHash(block)
}

func (a *api) Block() Block {
	return &block{evm: a.evm}
}

var _ API = (*api)(nil)

type Precompile interface {
	MutatesStorage(input []byte) bool
	RequiredGas(input []byte) uint64
	New(api API) error
	Commit(api API) error
	Run(api API, input []byte) ([]byte, error)
}
