package interfaces

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
