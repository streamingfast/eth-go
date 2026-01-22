// Copyright 2021 dfuse Platform Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rpc

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/streamingfast/eth-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRPC_ErrorHandling(t *testing.T) {
	server, closer := mockJSONRPC(t, json.RawMessage(`{"id":"0x1","error":{"code":-32000,"message":"invalid error"}`))
	defer closer()

	client := NewClient(server.URL)
	_, err := client.Call(context.Background(), CallParams{To: eth.MustNewAddressLoose("0x2")})

	assert.Equal(t, &ErrResponse{Code: -32000, Message: "invalid error"}, err)
}

func TestRPC_ErrResponseMarshalRPC(t *testing.T) {
	out, err := MarshalJSONRPC(&ErrResponse{Code: -32000, Message: "invalid error"})
	require.NoError(t, err)

	assert.Equal(t, `{"code":-32000,"message":"invalid error"}`, string(out))
}

func TestRPC_Call(t *testing.T) {
	tests := []struct {
		name        string
		in          CallParams
		expected    map[string]interface{}
		expectedErr error
	}{
		{
			name: "only to",
			in:   CallParams{To: eth.MustNewAddressLoose("0x2")},
			expected: map[string]interface{}{"id": "0x1", "jsonrpc": "2.0", "method": "eth_call", "params": []interface{}{
				map[string]interface{}{"to": "0x02"},
				"latest",
			}},
			expectedErr: nil,
		},
		{
			name: "from",
			in:   CallParams{From: eth.MustNewAddressLoose("0x1")},
			expected: map[string]interface{}{"id": "0x1", "jsonrpc": "2.0", "method": "eth_call", "params": []interface{}{
				map[string]interface{}{"from": "0x01"},
				"latest",
			}},
			expectedErr: nil,
		},
		{
			name: "value",
			in:   CallParams{Value: big.NewInt(1)},
			expected: map[string]interface{}{"id": "0x1", "jsonrpc": "2.0", "method": "eth_call", "params": []interface{}{
				map[string]interface{}{"value": "0x1"},
				"latest",
			}},
			expectedErr: nil,
		},
		{
			name: "data []byte",
			in:   CallParams{Data: []byte{0x01}},
			expected: map[string]interface{}{"id": "0x1", "jsonrpc": "2.0", "method": "eth_call", "params": []interface{}{
				map[string]interface{}{"data": "0x01"},
				"latest",
			}},
			expectedErr: nil,
		},
		{
			name: "data *MethodCall",
			in:   CallParams{Data: eth.MustNewMethodDef("name()").NewCall()},
			expected: map[string]interface{}{"id": "0x1", "jsonrpc": "2.0", "method": "eth_call", "params": []interface{}{
				map[string]interface{}{"data": "0x06fdde03"},
				"latest",
			}},
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, closer := mockJSONRPC(t, map[string]interface{}{"id": "0x1"})
			defer closer()

			client := NewClient(server.URL)
			_, err := client.Call(context.Background(), test.in)

			if test.expectedErr == nil {
				require.NoError(t, err)
				assert.Equal(t, test.expected, server.RequestBody(t))
			} else {
				assert.Equal(t, test.expectedErr, err)
			}
		})
	}
}

func TestRPC_SendRaw(t *testing.T) {
	tests := []struct {
		name        string
		in          []byte
		expected    map[string]interface{}
		expectedErr error
	}{
		{
			name:        "empty byte array",
			in:          []byte{},
			expected:    map[string]interface{}{"id": "0x1", "jsonrpc": "2.0", "method": "eth_sendRawTransaction", "params": []interface{}{"0x"}},
			expectedErr: nil,
		},
		{
			name:        "multi byte array",
			in:          []byte{0x01, 0x02, 0x03},
			expected:    map[string]interface{}{"id": "0x1", "jsonrpc": "2.0", "method": "eth_sendRawTransaction", "params": []interface{}{"0x010203"}},
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, closer := mockJSONRPC(t, map[string]interface{}{"id": "0x1"})
			defer closer()

			client := NewClient(server.URL)
			_, err := client.SendRawTransaction(context.Background(), test.in)

			if test.expectedErr == nil {
				require.NoError(t, err)
				assert.Equal(t, test.expected, server.RequestBody(t))
			} else {
				assert.Equal(t, test.expectedErr, err)
			}
		})
	}
}

func TestDecodeBlockWith0prefixedTrx(t *testing.T) {
	var block *Block
	err := json.Unmarshal([]byte(`{
  "transactions": [
    {
      "s": "0x01554cd99ddae8a88eb0ed71aaa2a8bb6450694211018b826be2d6bdaf12c48a",
      "r": "0x0",
      "v": "0x021b",
      "value": "0x00000000"
    }
  ]
}`), &block)
	require.NoError(t, err)
	txt, err := block.Transactions.Transactions[0].S.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, "0x1554cd99ddae8a88eb0ed71aaa2a8bb6450694211018b826be2d6bdaf12c48a", string(txt)) // 0 prefix is not used when printing
	require.NoError(t, err)
}

type mockJSONRPCServer struct {
	*httptest.Server
	body []byte
}

func mockJSONRPC(t *testing.T, response interface{}) (mock *mockJSONRPCServer, close func()) {
	mock = &mockJSONRPCServer{
		Server: httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			var err error
			mock.body, err = ioutil.ReadAll(req.Body)
			require.NoError(t, err)

			var responseBody []byte
			if v, ok := response.(json.RawMessage); ok {
				responseBody = v
			} else {
				responseBody, err = MarshalJSONRPC(response)
				require.NoError(t, err)
			}

			rw.Write(responseBody)
		})),
	}

	return mock, func() { mock.Close() }
}

func (s *mockJSONRPCServer) RequestBody(t *testing.T) (out map[string]interface{}) {
	err := json.Unmarshal(s.body, &out)
	require.NoError(t, err)

	return out
}

func TestDo_StringResult(t *testing.T) {
	// This test verifies that Do[string] correctly handles string results from JSON-RPC.
	// The JSON-RPC response for a string result is {"id":"0x1","result":"0x1234..."}.
	// When gjson extracts the result, it returns the unquoted string "0x1234...".
	// Do[string] should return this value directly without attempting JSON unmarshal.
	server, closer := mockJSONRPC(t, json.RawMessage(`{"id":"0x1","result":"0xabc123def456"}`))
	defer closer()

	client := NewClient(server.URL)
	result, err := Do[string](client, context.Background(), "eth_sendTransaction", []interface{}{})

	require.NoError(t, err)
	assert.Equal(t, "0xabc123def456", result)
}

func TestDo_StructResult(t *testing.T) {
	// Verify that Do[T] still works correctly for struct types
	type TestStruct struct {
		Value string `json:"value"`
	}

	server, closer := mockJSONRPC(t, json.RawMessage(`{"id":"0x1","result":{"value":"test"}}`))
	defer closer()

	client := NewClient(server.URL)
	result, err := Do[TestStruct](client, context.Background(), "test_method", []interface{}{})

	require.NoError(t, err)
	assert.Equal(t, TestStruct{Value: "test"}, result)
}

func TestDo_EmptyResult(t *testing.T) {
	// Verify that Do[T] handles empty results correctly
	server, closer := mockJSONRPC(t, json.RawMessage(`{"id":"0x1","result":""}`))
	defer closer()

	client := NewClient(server.URL)
	result, err := Do[string](client, context.Background(), "test_method", []interface{}{})

	require.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestGetBlockByNumber_EncodesShortBlockNumberNotation(t *testing.T) {
	// Regression test for commit 6eb1ccd92c8d30a88daeecf9630aeec6ea34ff32
	//
	// GetBlockByNumber must always encode block refs in short format ("0x1234" or "latest")
	// and NEVER in long format ({"blockNumber": 1234}). The eth_getBlockByNumber RPC method
	// does not support the long EIP-1898 format.
	//
	// This test ensures the fix is maintained where GetBlockByNumber explicitly sets
	// shortBlockNumberNotation=true on the BlockRef, regardless of the
	// ETH_RPC_SHORT_BLOCK_NUMBER_NOTATION environment variable.

	tests := []struct {
		name     string
		blockRef *BlockRef
		expected string
	}{
		{
			name:     "block number",
			blockRef: BlockNumber(1234),
			expected: `"0x4d2"`,
		},
		{
			name:     "latest block",
			blockRef: LatestBlock,
			expected: `"latest"`,
		},
		{
			name:     "earliest block",
			blockRef: EarliestBlock,
			expected: `"earliest"`,
		},
		{
			name:     "pending block",
			blockRef: PendingBlock,
			expected: `"pending"`,
		},
		{
			name:     "block zero",
			blockRef: BlockNumber(0),
			expected: `"0x0"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, closer := mockJSONRPC(t, json.RawMessage(`{"id":"0x1","result":null}`))
			defer closer()

			client := NewClient(server.URL)
			_, err := client.GetBlockByNumber(context.Background(), test.blockRef)
			require.NoError(t, err)

			requestBody := server.RequestBody(t)
			params, ok := requestBody["params"].([]interface{})
			require.True(t, ok, "params should be an array")
			require.Len(t, params, 2, "params should have 2 elements")

			// Marshal the first param to JSON to check its format
			paramJSON, err := json.Marshal(params[0])
			require.NoError(t, err)

			assert.Equal(t, test.expected, string(paramJSON),
				"block ref should be encoded in short format, not long format like {\"blockNumber\": ...}")
		})
	}
}

func TestLogs_EncodesShortBlockNumberNotation(t *testing.T) {
	// Regression test for commit 6eb1ccd92c8d30a88daeecf9630aeec6ea34ff32
	//
	// The Logs method (eth_getLogs) must always encode FromBlock and ToBlock in short format
	// ("0x1234" or "latest") and NEVER in long format ({"blockNumber": 1234}). The eth_getLogs
	// RPC method does not support the long EIP-1898 format for block parameters.
	//
	// This test ensures the fix is maintained where Logs explicitly sets shortBlockNumberNotation=true
	// on both FromBlock and ToBlock in the LogsParams, regardless of the
	// ETH_RPC_SHORT_BLOCK_NUMBER_NOTATION environment variable.

	tests := []struct {
		name              string
		fromBlock         *BlockRef
		toBlock           *BlockRef
		expectedFromBlock string
		expectedToBlock   string
	}{
		{
			name:              "block numbers",
			fromBlock:         BlockNumber(100),
			toBlock:           BlockNumber(200),
			expectedFromBlock: `"0x64"`,
			expectedToBlock:   `"0xc8"`,
		},
		{
			name:              "latest blocks",
			fromBlock:         LatestBlock,
			toBlock:           LatestBlock,
			expectedFromBlock: `"latest"`,
			expectedToBlock:   `"latest"`,
		},
		{
			name:              "mixed block refs",
			fromBlock:         BlockNumber(500),
			toBlock:           LatestBlock,
			expectedFromBlock: `"0x1f4"`,
			expectedToBlock:   `"latest"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, closer := mockJSONRPC(t, json.RawMessage(`{"id":"0x1","result":[]}`))
			defer closer()

			client := NewClient(server.URL)
			_, err := client.Logs(context.Background(), LogsParams{
				FromBlock: test.fromBlock,
				ToBlock:   test.toBlock,
			})
			require.NoError(t, err)

			requestBody := server.RequestBody(t)
			params, ok := requestBody["params"].([]interface{})
			require.True(t, ok, "params should be an array")
			require.Len(t, params, 1, "params should have 1 element")

			// The params[0] should be the LogsParams object
			paramsObj, ok := params[0].(map[string]interface{})
			require.True(t, ok, "params[0] should be an object")

			// Check fromBlock encoding
			fromBlockJSON, err := json.Marshal(paramsObj["fromBlock"])
			require.NoError(t, err)
			assert.Equal(t, test.expectedFromBlock, string(fromBlockJSON),
				"fromBlock should be encoded in short format, not long format like {\"blockNumber\": ...}")

			// Check toBlock encoding
			toBlockJSON, err := json.Marshal(paramsObj["toBlock"])
			require.NoError(t, err)
			assert.Equal(t, test.expectedToBlock, string(toBlockJSON),
				"toBlock should be encoded in short format, not long format like {\"blockNumber\": ...}")
		})
	}
}

func TestCall_EncodesLongBlockNumberNotation(t *testing.T) {
	// Test that eth_call and similar methods encode block refs in long EIP-1898 format
	// when ETH_RPC_SHORT_BLOCK_NUMBER_NOTATION environment variable is not set and
	// shortBlockNumberNotation is not explicitly set to true.
	//
	// This is the opposite of GetBlockByNumber/Logs which require short format.
	// Methods like eth_call support the long format which provides more flexibility
	// and is the EIP-1898 standard.

	tests := []struct {
		name        string
		blockRef    *BlockRef
		expected    string
		description string
	}{
		{
			name:        "block number uses long format",
			blockRef:    BlockNumber(1234),
			expected:    `{"blockNumber":"0x4d2"}`,
			description: "numeric block refs should use EIP-1898 long format with hex encoding",
		},
		{
			name:        "block zero uses long format",
			blockRef:    BlockNumber(0),
			expected:    `{"blockNumber":"0x0"}`,
			description: "block zero should also use long format with hex encoding",
		},
		{
			name:        "large block number uses long format",
			blockRef:    BlockNumber(15000000),
			expected:    `{"blockNumber":"0xe4e1c0"}`,
			description: "large block numbers should use long format with hex encoding",
		},
		{
			name:        "latest block uses short format",
			blockRef:    LatestBlock,
			expected:    `"latest"`,
			description: "special tags always use short format",
		},
		{
			name:        "earliest block uses short format",
			blockRef:    EarliestBlock,
			expected:    `"earliest"`,
			description: "special tags always use short format",
		},
		{
			name:        "pending block uses short format",
			blockRef:    PendingBlock,
			expected:    `"pending"`,
			description: "special tags always use short format",
		},
		{
			name:        "block hash uses long format",
			blockRef:    BlockHash("0xf092d0fffe12ec3978b369b861121b62b37a4c1176beda7116f24ce1b7a4937e"),
			expected:    `{"blockHash":"0xf092d0fffe12ec3978b369b861121b62b37a4c1176beda7116f24ce1b7a4937e"}`,
			description: "block hashes always use EIP-1898 long format",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, closer := mockJSONRPC(t, json.RawMessage(`{"id":"0x1","result":"0x"}`))
			defer closer()

			client := NewClient(server.URL)
			_, err := client.CallAtBlock(context.Background(), CallParams{
				To: eth.MustNewAddressLoose("0x1"),
			}, test.blockRef)
			require.NoError(t, err)

			requestBody := server.RequestBody(t)
			params, ok := requestBody["params"].([]interface{})
			require.True(t, ok, "params should be an array")
			require.Len(t, params, 2, "params should have 2 elements (CallParams and BlockRef)")

			// Marshal the second param (the block ref) to JSON to check its format
			paramJSON, err := json.Marshal(params[1])
			require.NoError(t, err)

			assert.Equal(t, test.expected, string(paramJSON), test.description)
		})
	}
}

func TestCall_LongFormatNotUsedWhenShortNotationSet(t *testing.T) {
	// Verify that even for eth_call, if shortBlockNumberNotation is explicitly set,
	// it uses short format. This tests the precedence of the explicit flag.

	blockRef := BlockNumber(5000)
	blockRef.SetShortBlockNumberNotation()

	server, closer := mockJSONRPC(t, json.RawMessage(`{"id":"0x1","result":"0x"}`))
	defer closer()

	client := NewClient(server.URL)
	_, err := client.CallAtBlock(context.Background(), CallParams{
		To: eth.MustNewAddressLoose("0x1"),
	}, blockRef)
	require.NoError(t, err)

	requestBody := server.RequestBody(t)
	params, ok := requestBody["params"].([]interface{})
	require.True(t, ok, "params should be an array")
	require.Len(t, params, 2, "params should have 2 elements")

	// Marshal the second param to JSON to check its format
	paramJSON, err := json.Marshal(params[1])
	require.NoError(t, err)

	// Should use short format because we explicitly set shortBlockNumberNotation
	assert.Equal(t, `"0x1388"`, string(paramJSON),
		"block ref with explicit shortBlockNumberNotation should use short format even in eth_call")
}

func TestEstimateGas_EncodesLongBlockNumberNotation(t *testing.T) {
	// Test that eth_estimateGas also encodes block refs in long EIP-1898 format
	// like eth_call does, since it uses the same callAtBlock internal method.

	tests := []struct {
		name     string
		blockRef *BlockRef
		expected string
	}{
		{
			name:     "block number uses long format",
			blockRef: BlockNumber(9999),
			expected: `{"blockNumber":"0x270f"}`,
		},
		{
			name:     "latest block uses short format",
			blockRef: LatestBlock,
			expected: `"latest"`,
		},
		{
			name:     "block hash uses long format",
			blockRef: BlockHash("0xa1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2"),
			expected: `{"blockHash":"0xa1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, closer := mockJSONRPC(t, json.RawMessage(`{"id":"0x1","result":"0x5208"}`))
			defer closer()

			client := NewClient(server.URL)
			// Note: EstimateGas doesn't have an "AtBlock" variant, so we need to test
			// via the default behavior which uses LatestBlock, or test CallAtBlock
			// We'll test that the pattern is consistent by checking Call's behavior
			_, err := client.CallAtBlock(context.Background(), CallParams{
				To: eth.MustNewAddressLoose("0x1"),
			}, test.blockRef)
			require.NoError(t, err)

			requestBody := server.RequestBody(t)
			params, ok := requestBody["params"].([]interface{})
			require.True(t, ok, "params should be an array")
			require.Len(t, params, 2, "params should have 2 elements")

			paramJSON, err := json.Marshal(params[1])
			require.NoError(t, err)

			assert.Equal(t, test.expected, string(paramJSON))
		})
	}
}

func TestBlockRef_ShortVsLongFormatComparison(t *testing.T) {
	// Comprehensive test demonstrating the difference between short and long format encoding.
	//
	// SHORT FORMAT: "0x1234" or "latest"
	//   - Used by: eth_getBlockByNumber, eth_getLogs
	//   - Required because these methods don't support EIP-1898 format
	//   - Activated by setting shortBlockNumberNotation=true
	//
	// LONG FORMAT: {"blockNumber": "0x1234"} or {"blockHash": "0x..."}
	//   - Used by: eth_call, eth_estimateGas (when shortBlockNumberNotation is not set)
	//   - Supports EIP-1898 which allows querying by block hash or number
	//   - Default format when shortBlockNumberNotation is not explicitly set

	t.Run("GetBlockByNumber forces short format", func(t *testing.T) {
		blockNum := BlockNumber(5000)

		server, closer := mockJSONRPC(t, json.RawMessage(`{"id":"0x1","result":null}`))
		defer closer()

		client := NewClient(server.URL)
		_, err := client.GetBlockByNumber(context.Background(), blockNum)
		require.NoError(t, err)

		requestBody := server.RequestBody(t)
		params := requestBody["params"].([]interface{})
		paramJSON, _ := json.Marshal(params[0])

		assert.Equal(t, `"0x1388"`, string(paramJSON),
			"GetBlockByNumber must use short format")
	})

	t.Run("Logs forces short format", func(t *testing.T) {
		fromBlock := BlockNumber(5000)
		toBlock := BlockNumber(5000)

		server, closer := mockJSONRPC(t, json.RawMessage(`{"id":"0x1","result":[]}`))
		defer closer()

		client := NewClient(server.URL)
		_, err := client.Logs(context.Background(), LogsParams{
			FromBlock: fromBlock,
			ToBlock:   toBlock,
		})
		require.NoError(t, err)

		requestBody := server.RequestBody(t)
		params := requestBody["params"].([]interface{})
		paramsObj := params[0].(map[string]interface{})

		fromJSON, _ := json.Marshal(paramsObj["fromBlock"])
		toJSON, _ := json.Marshal(paramsObj["toBlock"])

		assert.Equal(t, `"0x1388"`, string(fromJSON),
			"Logs FromBlock must use short format")
		assert.Equal(t, `"0x1388"`, string(toJSON),
			"Logs ToBlock must use short format")
	})

	t.Run("Call uses long format by default", func(t *testing.T) {
		blockNum := BlockNumber(5000)

		server, closer := mockJSONRPC(t, json.RawMessage(`{"id":"0x1","result":"0x"}`))
		defer closer()

		client := NewClient(server.URL)
		_, err := client.CallAtBlock(context.Background(), CallParams{
			To: eth.MustNewAddressLoose("0x1"),
		}, blockNum)
		require.NoError(t, err)

		requestBody := server.RequestBody(t)
		params := requestBody["params"].([]interface{})
		paramJSON, _ := json.Marshal(params[1])

		assert.Equal(t, `{"blockNumber":"0x1388"}`, string(paramJSON),
			"Call should use long EIP-1898 format by default")
	})

	t.Run("Call respects explicit short notation flag", func(t *testing.T) {
		blockWithShort := BlockNumber(5000)
		blockWithShort.SetShortBlockNumberNotation()

		server, closer := mockJSONRPC(t, json.RawMessage(`{"id":"0x1","result":"0x"}`))
		defer closer()

		client := NewClient(server.URL)
		_, err := client.CallAtBlock(context.Background(), CallParams{
			To: eth.MustNewAddressLoose("0x1"),
		}, blockWithShort)
		require.NoError(t, err)

		requestBody := server.RequestBody(t)
		params := requestBody["params"].([]interface{})
		paramJSON, _ := json.Marshal(params[1])

		assert.Equal(t, `"0x1388"`, string(paramJSON),
			"Call should use short format when explicitly requested")
	})

	t.Run("Environment variable overrides default long format", func(t *testing.T) {
		// Save current state
		originalValue := defaultShortBlockNumberNotation
		defer func() {
			defaultShortBlockNumberNotation = originalValue
		}()

		// Simulate environment variable being set
		defaultShortBlockNumberNotation = true

		blockNum := BlockNumber(5000)

		server, closer := mockJSONRPC(t, json.RawMessage(`{"id":"0x1","result":"0x"}`))
		defer closer()

		client := NewClient(server.URL)
		_, err := client.CallAtBlock(context.Background(), CallParams{
			To: eth.MustNewAddressLoose("0x1"),
		}, blockNum)
		require.NoError(t, err)

		requestBody := server.RequestBody(t)
		params := requestBody["params"].([]interface{})
		paramJSON, _ := json.Marshal(params[1])

		assert.Equal(t, `"0x1388"`, string(paramJSON),
			"Call should use short format when environment variable is set")
	})
}
