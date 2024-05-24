package vm

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
)

// Common ConcreteBlockContext object used by some tests
func newConcreteBlockContext() *concreteBlockContext {
	hash := crypto.Keccak256Hash([]byte("test coin base"))
	coinbase := common.BytesToAddress(hash[:])
	randomHash := crypto.Keccak256Hash([]byte("random hash"))

	vmctx := BlockContext{
		BlockNumber: big.NewInt(0),
		GetHash:     func(uint64) common.Hash { return crypto.Keccak256Hash([]byte("test")) },
		Time:        50,
		GasLimit:    1000,
		Difficulty:  big.NewInt(23456146),
		BaseFee:     big.NewInt(739657255),
		Coinbase:    coinbase,
		Random:      &randomHash,
	}
	return &concreteBlockContext{ctx: &vmctx}
}

func TestConcreteGetHash(t *testing.T) {
	r := require.New(t)
	concreteBlockContext := *newConcreteBlockContext()
	expectedHash := common.HexToHash("0x9c22ff5f21f0b81b113e63f7db6da94fedef11b2119b4088b89664fb9a3cb658")

	hash := concreteBlockContext.GetHash(0)

	r.Equal(hash, expectedHash)
}

func TestConcreteTimeStamp(t *testing.T) {
	r := require.New(t)
	concreteBlockContext := *newConcreteBlockContext()
	expectedTime := uint64(50)

	time := concreteBlockContext.Timestamp()

	r.Equal(time, expectedTime)
}

func TestConcreteBlockNumber(t *testing.T) {
	r := require.New(t)
	concreteBlockContext := *newConcreteBlockContext()
	expectedBlockNumber := uint64(0)

	blockNumber := concreteBlockContext.BlockNumber()

	r.Equal(expectedBlockNumber, blockNumber)
}

func TestConcreteGasLimit(t *testing.T) {
	r := require.New(t)
	concreteBlockContext := *newConcreteBlockContext()
	expectedGasLimit := uint64(1000)

	gasLimit := concreteBlockContext.GasLimit()

	r.Equal(gasLimit, expectedGasLimit)
}

func TestConcreteDifficulty(t *testing.T) {
	r := require.New(t)
	concreteBlockContext := *newConcreteBlockContext()
	expectedDifficulty := uint256.MustFromBig(big.NewInt(23456146))

	difficulty := concreteBlockContext.Difficulty()

	r.Equal(difficulty, expectedDifficulty)
}

func TestConcreteBaseFee(t *testing.T) {
	r := require.New(t)
	concreteBlockContext := *newConcreteBlockContext()
	expectedBaseFee := uint256.MustFromBig(big.NewInt(739657255))

	baseFee := concreteBlockContext.BaseFee()

	r.Equal(baseFee, expectedBaseFee)
}

func TestConcreteCoinbase(t *testing.T) {
	r := require.New(t)
	concreteBlockContext := *newConcreteBlockContext()
	expectedCoinbase := common.HexToAddress("0x0854167430392BBc2D15Dd1Cc17e761897AF31C9")

	coinbase := concreteBlockContext.Coinbase()

	r.Equal(coinbase, expectedCoinbase)
}
func TestConcreteRandom(t *testing.T) {
	r := require.New(t)
	concreteBlockContext := *newConcreteBlockContext()
	expectedRandomHash := common.HexToHash("0xa91de29564820ffde10c59f8f429df5e29433da59643ad800bbe4e3959d936d7")

	randomHash := concreteBlockContext.Random()

	r.Equal(randomHash, expectedRandomHash)
}
