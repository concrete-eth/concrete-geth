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

// This file will be replaced by interface_tinygo.go when building with tinygo
// to prevent compatibility issues.

package api

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
)

type StateDB interface {
	// Access list
	AddressInAccessList(addr common.Address) bool
	SlotInAccessList(addr common.Address, slot common.Hash) (addressOk bool, slotOk bool)
	AddAddressToAccessList(addr common.Address)
	AddSlotToAccessList(addr common.Address, slot common.Hash)
	// Code
	GetCode(common.Address) []byte
	GetCodeSize(common.Address) int
	GetCodeHash(common.Address) common.Hash
	// Balance
	GetBalance(addr common.Address) *uint256.Int
	// Logs
	AddLog(*types.Log)
	// Refunds
	AddRefund(uint64)
	SubRefund(uint64)
	GetRefund() uint64
	// Storage
	GetCommittedState(addr common.Address, key common.Hash) common.Hash
	SetState(addr common.Address, key common.Hash, value common.Hash)
	GetState(addr common.Address, key common.Hash) common.Hash
	// Storage -- Concrete
	// SetEphemeralState(addr common.Address, key common.Hash, value common.Hash)
	// GetEphemeralState(addr common.Address, key common.Hash) common.Hash
}
