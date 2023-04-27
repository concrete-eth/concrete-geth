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

package state

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type ephemeralStorage struct {
	address  common.Address
	addrHash common.Hash
	db       *StateDB
	root     common.Hash

	trie           Trie
	storage        Storage
	dirtyStorage   Storage
	pendingStorage map[common.Hash]struct{}
}

func newEphemeralStorage(db *StateDB, address common.Address) *ephemeralStorage {
	return &ephemeralStorage{
		address:        address,
		addrHash:       crypto.Keccak256Hash(address.Bytes()),
		db:             db,
		root:           types.EmptyRootHash,
		storage:        make(Storage),
		dirtyStorage:   make(Storage),
		pendingStorage: make(map[common.Hash]struct{}),
	}
}

func (s *ephemeralStorage) Address() common.Address {
	return s.address
}

func (s *ephemeralStorage) SetState(db Database, key, value common.Hash) {
	prev := s.GetState(db, key)
	if prev == value {
		return
	}
	s.db.journal.append(ephemeralStorageChange{
		account:  &s.address,
		key:      key,
		prevalue: prev,
	})
	s.setState(key, value)
}

func (s *ephemeralStorage) setState(key, value common.Hash) {
	s.dirtyStorage[key] = value
}

func (s *ephemeralStorage) GetState(db Database, key common.Hash) common.Hash {
	return s.getState(db, key)
}

func (s *ephemeralStorage) getState(db Database, key common.Hash) common.Hash {
	if value, ok := s.dirtyStorage[key]; ok {
		return value
	}
	return s.storage[key]
}

func (s *ephemeralStorage) finalise() {
	if len(s.dirtyStorage) == 0 {
		return
	}
	for key, value := range s.dirtyStorage {
		if value == s.storage[key] {
			continue
		}
		s.storage[key] = value
		s.pendingStorage[key] = struct{}{}
	}
	s.dirtyStorage = make(Storage)
}

func (s *ephemeralStorage) getTrie(db Database) (Trie, error) {
	if s.trie == nil {
		tr, err := db.OpenStorageTrie(common.Hash{}, s.addrHash, s.root)
		if err != nil {
			return nil, err
		}
		s.trie = tr
	}
	return s.trie, nil
}

func (s *ephemeralStorage) updateTrie(db Database) error {
	s.finalise()
	if len(s.pendingStorage) == 0 {
		return nil
	}
	tr, err := s.getTrie(db)
	if err != nil {
		return err
	}
	for key := range s.pendingStorage {
		value := s.storage[key]
		if (value == common.Hash{}) {
			if err := tr.TryDelete(key.Bytes()); err != nil {
				return err
			}
		} else {
			if err := tr.TryUpdate(key.Bytes(), value.Bytes()); err != nil {
				return err
			}
		}
	}
	s.pendingStorage = make(map[common.Hash]struct{})
	return nil
}

func (s *ephemeralStorage) updateRoot(db Database) error {
	err := s.updateTrie(db)
	if err != nil {
		return err
	}
	s.root = s.trie.Hash()
	return nil
}

func (s *ephemeralStorage) deepCopy(db *StateDB) *ephemeralStorage {
	store := newEphemeralStorage(db, s.address)
	if s.trie != nil {
		store.trie = db.ephemeralDB.CopyTrie(s.trie)
	}
	store.root = s.root
	store.storage = s.storage.Copy()
	store.dirtyStorage = s.dirtyStorage.Copy()
	for key := range s.pendingStorage {
		store.pendingStorage[key] = struct{}{}
	}
	return store
}
