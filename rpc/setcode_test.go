package rpc

import (
	"encoding/json"
	"testing"

	"github.com/holiman/uint256"
	"github.com/streamingfast/eth-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Fixture data captured from a live reth-dev (Amsterdam fork) node's `eth_getTransactionByHash`
// response for a SetCode (type 0x04) transaction carrying three authorization tuples.
func TestSetCodeAuthorization_Authority(t *testing.T) {
	tests := []struct {
		name              string
		auth              SetCodeAuthorization
		expectedAuthority string
		assertErr         require.ErrorAssertionFunc
	}{
		{
			name: "regular chain id",
			auth: SetCodeAuthorization{
				ChainID: 0x539,
				Address: eth.MustNewAddress("0x48b92f7af3a76831e4f5da69959554ad2c134a90"),
				Nonce:   0x1,
				YParity: 0x0,
				R:       mustUint256FromHex("0x7d229398f62c67ab5341ee6b433f3828d23f4b5315b7c7c7f5ec93bd81c16b2a"),
				S:       mustUint256FromHex("0xfb34e96be711b61f95f9a4e718b6a77a9dea9eb307dd303a194240ed49a02d5"),
			},
			expectedAuthority: "0x71562b71999873db5b286df957af199ec94617f7",
			assertErr:         require.NoError,
		},
		{
			name: "wildcard chain id (chainId 0)",
			auth: SetCodeAuthorization{
				ChainID: 0x0,
				Address: eth.MustNewAddress("0xa1128904651b17348a17344b066b4a0ae69ae06f"),
				Nonce:   0x0,
				YParity: 0x0,
				R:       mustUint256FromHex("0x699c8fd68a22cec4b65f0f5a782455d5fa3ffbc960fce1e16b904bf3c2da6ef1"),
				S:       mustUint256FromHex("0x463a614c29be603d5b99b77b7529ffb5331f008183cdaa91f0664a83cd456489"),
			},
			expectedAuthority: "0x703c4b2bd70c169f5717101caee543299fc946c7",
			assertErr:         require.NoError,
		},
		{
			name: "discarded authorization with invalid all-zero signature",
			auth: SetCodeAuthorization{
				ChainID: 0x0,
				Address: eth.MustNewAddress("0xa1128904651b17348a17344b066b4a0ae69ae06f"),
				Nonce:   0x0,
				YParity: 0x1,
				R:       mustUint256FromHex("0x0"),
				S:       mustUint256FromHex("0x0"),
			},
			assertErr: require.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authority, err := tt.auth.Authority()
			tt.assertErr(t, err)
			if tt.expectedAuthority != "" {
				assert.Equal(t, eth.MustNewAddress(tt.expectedAuthority), authority)
			}
		})
	}
}

func TestTransaction_AuthorizationList_Decode(t *testing.T) {
	payload := `{
		"hash": "0x1111111111111111111111111111111111111111111111111111111111111111",
		"nonce": "0x2a",
		"from": "0x48b92f7af3a76831e4f5da69959554ad2c134a90",
		"to": null,
		"value": "0x0",
		"gasPrice": "0x1",
		"gas": "0x5208",
		"type": "0x4",
		"authorizationList": [
			{
				"chainId": "0x539",
				"address": "0x48b92f7af3a76831e4f5da69959554ad2c134a90",
				"nonce": "0x1",
				"yParity": "0x0",
				"r": "0x7d229398f62c67ab5341ee6b433f3828d23f4b5315b7c7c7f5ec93bd81c16b2a",
				"s": "0xfb34e96be711b61f95f9a4e718b6a77a9dea9eb307dd303a194240ed49a02d5"
			}
		]
	}`

	var tx Transaction
	require.NoError(t, json.Unmarshal([]byte(payload), &tx))

	require.Len(t, tx.AuthorizationList, 1)

	auth := tx.AuthorizationList[0]
	assert.Equal(t, eth.Uint64(0x539), auth.ChainID)
	assert.Equal(t, eth.MustNewAddress("0x48b92f7af3a76831e4f5da69959554ad2c134a90"), auth.Address)
	assert.Equal(t, eth.Uint64(0x1), auth.Nonce)
	assert.Equal(t, eth.Uint64(0x0), auth.YParity)

	authority, err := auth.Authority()
	require.NoError(t, err)
	assert.Equal(t, eth.MustNewAddress("0x71562b71999873db5b286df957af199ec94617f7"), authority)
}

func mustUint256FromHex(in string) *eth.Uint256 {
	return (*eth.Uint256)(uint256.MustFromHex(in))
}
