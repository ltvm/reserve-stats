package contracts

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

func TestVersionedWrapperFallback_GetReserveRate(t *testing.T) {
	t.Skip("disable as this test require external resource")
	const (
		ethNodeURL      = "https://mainnet.infura.io"
		blockNumber     = 6000744
		internalReserve = "0x63825c174ab367968EC60f061753D3bbD36A0D8F"
		bnbAddr         = "0xB8c77482e45F1F44dE1745F52C74426C631bDD52"
		ethAddr         = "0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE"
	)

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	defer logger.Sync()
	sugar := logger.Sugar()

	client, err := ethclient.Dial(ethNodeURL)
	require.NoError(t, err)
	vwf, err := NewVersionedWrapperFallback(sugar, client)
	require.NoError(t, err)
	rates, sanityRates, err := vwf.GetReserveRate(
		blockNumber,
		common.HexToAddress(internalReserve),
		[]common.Address{common.HexToAddress(ethAddr), common.HexToAddress(bnbAddr)},
		[]common.Address{common.HexToAddress(bnbAddr), common.HexToAddress(ethAddr)},
	)
	require.NoError(t, err)
	require.Len(t, rates, 2)
	assert.Zero(t, rates[0].Int64())
	assert.Zero(t, rates[1].Int64())
	require.Len(t, sanityRates, 2)
	assert.Zero(t, sanityRates[0].Int64())
	assert.Zero(t, sanityRates[1].Int64())
}