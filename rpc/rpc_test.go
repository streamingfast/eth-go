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
