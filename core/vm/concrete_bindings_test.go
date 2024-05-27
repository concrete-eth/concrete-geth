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

func newBlockContext() BlockContext {
	var (
		block0Hash = crypto.Keccak256Hash([]byte("test0"))
		block1Hash = crypto.Keccak256Hash([]byte("test1"))
		randomHash = crypto.Keccak256Hash([]byte("random hash"))
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

	vmctx := BlockContext{
		CanTransfer: func(StateDB, common.Address, *uint256.Int) bool { return true },
		Transfer:    func(StateDB, common.Address, common.Address, *uint256.Int) {},
		BlockNumber: big.NewInt(0),
		GetHash:     getHash,
		Time:        50,
		GasLimit:    1000,
		Difficulty:  big.NewInt(23456146),
		BaseFee:     big.NewInt(739657255),
		Coinbase:    common.HexToAddress("0x0854167430392BBc2D15Dd1Cc17e761897AF31C9"),
		Random:      &randomHash,
	}

	return vmctx
}

func TestConcreteBlockContext(t *testing.T) {
	r := require.New(t)
	blockCtx := newBlockContext()
	ccBlockCtx := concreteBlockContext{ctx: &blockCtx}
	var (
		block0Hash = crypto.Keccak256Hash([]byte("test0"))
		block1Hash = crypto.Keccak256Hash([]byte("test1"))
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
		r          = require.New(t)
		address    = common.BytesToAddress([]byte("CallStaticTestContract"))
		statedb, _ = state.New(types.EmptyRootHash, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
		config     = Config{}
		vmctx      = newBlockContext()
	)
	statedb.CreateAccount(address)
	// This example bytecode pushes 1 and 2 onto the stack, adds them, and returns the result. (Does NOT modify state)
	statedb.SetCode(address, hexutil.MustDecode("0x600260010160005260206000F3"))
	statedb.Finalise(true)
	vmenv := NewEVM(vmctx, TxContext{}, statedb, params.AllEthashProtocolChanges, config)
	ccVmenv := concreteEVM{
		evm:    vmenv,
		caller: common.Address{},
	}

	var (
		gas            = uint64(100)
		expectedReturn = []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3}
	)

	ret, remainingGas, err := ccVmenv.CallStatic(address, nil, gas)

	r.NoError(err, "Call should not return an error")
	r.NotNil(ret, "Return data should not be nil")
	r.Less(remainingGas, gas, "Gas used should be less than the provided gas")

	r.Equal(ret, expectedReturn, "Return data should be the result of the addition of 2 + 1")

}

func TestEVMCall(t *testing.T) {
	var (
		r          = require.New(t)
		address    = common.BytesToAddress([]byte("CallTestContract"))
		statedb, _ = state.New(types.EmptyRootHash, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
		config     = Config{}
		vmctx      = newBlockContext()
	)
	statedb.CreateAccount(address)
	statedb.AddAddressToAccessList(address)
	//This example bytecode saves value 1 in storage (Modify state)
	statedb.SetCode(address, hexutil.MustDecode("0x6001600055600054602060005260206000f3"))
	statedb.Finalise(true)
	vmenv := NewEVM(vmctx, TxContext{}, statedb, params.AllEthashProtocolChanges, config)
	ccVmenv := concreteEVM{
		evm:    vmenv,
		caller: common.Address{},
	}

	var (
		gas           = uint64(10000000)
		expectedState = common.HexToHash("0x1")
	)

	ret, remainingGas, err := ccVmenv.Call(address, nil, gas, new(uint256.Int))

	r.NoError(err, "Call should not return an error")
	r.NotNil(ret, "Return data should not be nil")
	r.Less(remainingGas, gas, "Gas used should be less than the provided gas")

	r.Equal(expectedState, statedb.GetState(address, common.Hash{}), "State should have changed by storing value 1 in Storage")

}

func TestEVMCreate(t *testing.T) {
	var (
		r          = require.New(t)
		statedb, _ = state.New(types.EmptyRootHash, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
		config     = Config{}
		vmctx      = newBlockContext()
		gas        = uint64(10000)
		code       = hexutil.MustDecode("0x600260010160005260206000F3")
	)

	statedb.Finalise(true)
	vmenv := NewEVM(vmctx, TxContext{}, statedb, params.AllEthashProtocolChanges, config)
	ccVmenv := concreteEVM{
		evm:    vmenv,
		caller: common.Address{},
	}

	t.Run("Create", func(t *testing.T) {
		expectedContractAddr := crypto.CreateAddress(ccVmenv.caller, statedb.GetNonce(ccVmenv.caller))

		ret, addr, remainingGas, err := ccVmenv.Create(code, gas, new(uint256.Int))

		r.NoError(err, "Call should not return an error")
		r.NotNil(ret, "Return data should not be nil")
		r.Less(remainingGas, gas, "Gas used should be less than the provided gas")
		r.Equal(expectedContractAddr, addr)
		r.Equal(statedb.GetCode(addr), ret)
	})

	t.Run("Create2", func(t *testing.T) {
		var (
			salt                 = uint256.NewInt(3521)
			codeAndHash          = &codeAndHash{code: code}
			expectedContractAddr = crypto.CreateAddress2(ccVmenv.caller, salt.Bytes32(), codeAndHash.Hash().Bytes())
		)

		ret, addr, remainingGas, err := ccVmenv.Create2(code, gas, new(uint256.Int), uint256.NewInt(3521))

		r.NoError(err, "Call should not return an error")
		r.NotNil(ret, "Return data should not be nil")
		r.Less(remainingGas, gas, "Gas used should be less than the provided gas")
		r.Equal(expectedContractAddr, addr)
		r.Equal(statedb.GetCode(addr), ret)
	})
}
