package vm

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
)

func newTestBlockContext() BlockContext {
	var (
		block0Hash = crypto.Keccak256Hash([]byte("block0"))
		block1Hash = crypto.Keccak256Hash([]byte("block1"))
		randomHash = crypto.Keccak256Hash([]byte("random"))
	)

	getHash := func(blockNumber uint64) common.Hash {
		switch blockNumber {
		case 0:
			return block0Hash
		case 1:
			return block1Hash
		}
		return common.Hash{}
	}

	blockCtx := BlockContext{
		CanTransfer: func(StateDB, common.Address, *uint256.Int) bool { return true },
		Transfer:    func(StateDB, common.Address, common.Address, *uint256.Int) {},
		BlockNumber: big.NewInt(1),
		GetHash:     getHash,
		Time:        50,
		GasLimit:    1000,
		Difficulty:  big.NewInt(1234),
		BaseFee:     big.NewInt(5678),
		Coinbase:    common.HexToAddress("0x0854167430392BBc2D15Dd1Cc17e761897AF31C9"),
		Random:      &randomHash,
	}

	return blockCtx
}

func TestConcreteBlockContext(t *testing.T) {
	r := require.New(t)
	blockCtx := newTestBlockContext()
	ccBlockCtx := concreteBlockContext{ctx: &blockCtx}

	var (
		block0Hash = blockCtx.GetHash(0)
		block1Hash = blockCtx.GetHash(1)
	)

	t.Run("GetHash", func(t *testing.T) {
		r.Equal(block0Hash, ccBlockCtx.GetHash(0))
		r.Equal(block1Hash, ccBlockCtx.GetHash(1))
	})
	t.Run("Timestamp", func(t *testing.T) {
		r.Equal(blockCtx.Time, ccBlockCtx.Timestamp())
	})
	t.Run("BlockNumber", func(t *testing.T) {
		r.Equal(blockCtx.BlockNumber.Uint64(), ccBlockCtx.BlockNumber())
	})
	t.Run("GasLimit", func(t *testing.T) {
		r.Equal(blockCtx.GasLimit, ccBlockCtx.GasLimit())
	})
	t.Run("Difficulty", func(t *testing.T) {
		r.Equal(uint256.MustFromBig(blockCtx.Difficulty), ccBlockCtx.Difficulty())
	})
	t.Run("BaseFee", func(t *testing.T) {
		r.Equal(uint256.MustFromBig(blockCtx.BaseFee), ccBlockCtx.BaseFee())
	})
	t.Run("Coinbase", func(t *testing.T) {
		r.Equal(blockCtx.Coinbase, ccBlockCtx.Coinbase())
	})
	t.Run("Random", func(t *testing.T) {
		r.Equal(*blockCtx.Random, ccBlockCtx.Random())
	})
}

func TestEVMCallStatic(t *testing.T) {
	var (
		r            = require.New(t)
		statedb, _   = state.New(types.EmptyRootHash, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
		callerAddr   = common.BytesToAddress([]byte("caller"))
		selfAddr     = common.BytesToAddress([]byte("self"))
		externalAddr = common.BytesToAddress([]byte("external"))
	)

	evm := NewEVM(BlockContext{}, TxContext{}, statedb, params.TestChainConfig, Config{})
	contract := NewContract(AccountRef(callerAddr), AccountRef(selfAddr), new(uint256.Int), 10_000_000)
	ccEvm := concreteEVM{evm: evm, contract: contract}

	gas := uint64(10_000)
	slot := common.BytesToHash([]byte("key"))
	value := common.BytesToHash([]byte("value"))

	// Bytecode to load the passed slot address from storage and return it
	statedb.SetCode(externalAddr, hexutil.MustDecode("0x6000355460005260206000F3"))
	statedb.SetState(externalAddr, slot, value)

	ret, remainingGas, err := ccEvm.CallStatic(externalAddr, slot.Bytes(), gas)

	r.NoError(err)
	r.Equal(value.Bytes(), ret)
	r.Less(remainingGas, gas)
}

func TestEVMCall(t *testing.T) {
	var (
		r            = require.New(t)
		statedb, _   = state.New(types.EmptyRootHash, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
		callerAddr   = common.BytesToAddress([]byte("caller"))
		selfAddr     = common.BytesToAddress([]byte("self"))
		externalAddr = common.BytesToAddress([]byte("external"))
	)

	evm := NewEVM(newTestBlockContext(), TxContext{}, statedb, params.TestChainConfig, Config{})
	contract := NewContract(AccountRef(callerAddr), AccountRef(selfAddr), new(uint256.Int), 10_000_000)
	ccEvm := concreteEVM{evm: evm, contract: contract}

	gas := uint64(10_000_000)
	slot := common.BytesToHash([]byte("key"))
	ogValue := common.BytesToHash([]byte("original"))
	newValue := common.BytesToHash([]byte("new"))
	input := append(slot.Bytes(), newValue.Bytes()...)
	// Replace the value at the passed slot with the passed value and return the original value
	statedb.SetCode(externalAddr, hexutil.MustDecode("0x600035546020356000355560005260206000F3"))
	statedb.SetState(externalAddr, slot, ogValue)
	statedb.AddAddressToAccessList(externalAddr)

	ret, remainingGas, err := ccEvm.Call(externalAddr, input, gas, new(uint256.Int))
	r.NoError(err)
	r.Equal(ogValue.Bytes(), ret)
	r.Equal(newValue, statedb.GetState(externalAddr, slot))
	r.Less(remainingGas, gas)
}

func TestEVMCallDelegate(t *testing.T) {
	var (
		r            = require.New(t)
		statedb, _   = state.New(types.EmptyRootHash, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
		callerAddr   = common.BytesToAddress([]byte("caller"))
		selfAddr     = common.BytesToAddress([]byte("self"))
		externalAddr = common.BytesToAddress([]byte("external"))
	)

	evm := NewEVM(BlockContext{}, TxContext{}, statedb, params.TestChainConfig, Config{})
	contract := NewContract(AccountRef(callerAddr), AccountRef(selfAddr), new(uint256.Int), 10_000_000)
	ccEvm := concreteEVM{evm: evm, contract: contract}

	gas := uint64(10_000)
	slot := common.BytesToHash([]byte("key"))
	value := common.BytesToHash([]byte("value"))

	// Bytecode to load the passed slot address from storage and return it
	statedb.SetCode(externalAddr, hexutil.MustDecode("0x6000355460005260206000F3"))
	statedb.SetState(selfAddr, slot, value)

	ret, remainingGas, err := ccEvm.CallDelegate(externalAddr, slot.Bytes(), gas)

	r.NoError(err)
	r.Equal(value.Bytes(), ret)
	r.Less(remainingGas, gas)
}

func TestEVMCreate(t *testing.T) {
	var (
		r          = require.New(t)
		statedb, _ = state.New(types.EmptyRootHash, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
		callerAddr = common.BytesToAddress([]byte("caller"))
		selfAddr   = common.BytesToAddress([]byte("self"))
	)

	evm := NewEVM(newTestBlockContext(), TxContext{}, statedb, params.TestChainConfig, Config{})
	contract := NewContract(AccountRef(callerAddr), AccountRef(selfAddr), new(uint256.Int), 10_000_000)
	ccEvm := concreteEVM{
		evm:      evm,
		contract: contract,
	}

	gas := uint64(10_000)
	creationCodePrefix := hexutil.MustDecode("0x600A600C600039600A6000F3") // Return the runtime code
	runtimeCode := hexutil.MustDecode("0x607B60005260206000F3")            // Return 123
	creationCode := append(creationCodePrefix, runtimeCode...)

	t.Run("Create", func(t *testing.T) {
		ret, addr, remainingGas, err := ccEvm.Create(creationCode, gas, new(uint256.Int))

		r.NoError(err, "Call should not return an error")
		r.Equal(runtimeCode, ret, "Runtime code not matching")
		r.Equal(runtimeCode, statedb.GetCode(addr), "Runtime code not matching against the statedb code")
		r.Equal(crypto.CreateAddress(selfAddr, 0), addr, "Created address not matching returned address")
		r.Less(remainingGas, gas, "Gas used should be less than the provided gas")
	})

	t.Run("Create2", func(t *testing.T) {
		salt := new(uint256.Int).SetBytes([]byte("salt"))
		ret, addr, remainingGas, err := ccEvm.Create2(creationCode, gas, new(uint256.Int), salt)

		r.NoError(err, "Call should not return an error")
		r.Equal(runtimeCode, ret, "Runtime code not matching")
		r.Equal(runtimeCode, statedb.GetCode(addr), "Runtime code not matching against the statedb code")
		r.Equal(crypto.CreateAddress2(selfAddr, salt.Bytes32(), crypto.Keccak256(creationCode)), addr, "Created address not matching returned address")
		r.Less(remainingGas, gas, "Gas used should be less than the provided gas")
	})
}
