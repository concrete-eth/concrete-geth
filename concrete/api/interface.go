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
	GetBalance(addr common.Address) *big.Int

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
