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

package eth

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMethodCall_AppendArgFromString(t *testing.T) {
	tests := []struct {
		name             string
		signature        string
		inputs           []string
		expectMethodDef  *MethodDef
		expectMethodCall *MethodCall
	}{
		{
			name:      "testing bytes",
			signature: "method(bytes)",
			inputs:    []string{"0xaabbcc"},
			expectMethodDef: &MethodDef{
				Name:       "method",
				Parameters: []*MethodParameter{{TypeName: "bytes"}},
			},
			expectMethodCall: &MethodCall{
				MethodDef: &MethodDef{Name: "method", Parameters: []*MethodParameter{{TypeName: "bytes"}}},
				Data: []interface{}{
					Hex([]byte{0xaa, 0xbb, 0xcc}),
				},
			},
		},
		{
			name:      "testing address",
			signature: "method(address)",
			inputs:    []string{"0x5a0b54d5dc17e0aadc383d2db43b0a0d3e029c4c"},
			expectMethodDef: &MethodDef{
				Name:       "method",
				Parameters: []*MethodParameter{{TypeName: "address"}},
			},
			expectMethodCall: &MethodCall{
				MethodDef: &MethodDef{Name: "method", Parameters: []*MethodParameter{{TypeName: "address"}}},
				Data: []interface{}{
					MustNewAddress("5a0b54d5dc17e0aadc383d2db43b0a0d3e029c4c"),
				},
			},
		},
		{
			name:      "testing uint32",
			signature: "method(uint32)",
			inputs:    []string{"13"},
			expectMethodDef: &MethodDef{
				Name:       "method",
				Parameters: []*MethodParameter{{TypeName: "uint32"}},
			},
			expectMethodCall: &MethodCall{
				MethodDef: &MethodDef{Name: "method", Parameters: []*MethodParameter{{TypeName: "uint32"}}},
				Data: []interface{}{
					Uint32(13),
				},
			},
		},
		{
			name:      "testing uint64",
			signature: "method(uint64)",
			inputs:    []string{"13"},
			expectMethodDef: &MethodDef{
				Name:       "method",
				Parameters: []*MethodParameter{{TypeName: "uint64"}},
			},
			expectMethodCall: &MethodCall{
				MethodDef: &MethodDef{Name: "method", Parameters: []*MethodParameter{{TypeName: "uint64"}}},
				Data: []interface{}{
					Uint64(13),
				},
			},
		},
		{
			name:      "testing uint112",
			signature: "method(uint112)",
			inputs:    []string{"123456789"},
			expectMethodDef: &MethodDef{
				Name:       "method",
				Parameters: []*MethodParameter{{TypeName: "uint112"}},
			},
			expectMethodCall: &MethodCall{
				MethodDef: &MethodDef{Name: "method", Parameters: []*MethodParameter{{TypeName: "uint112"}}},
				Data: []interface{}{
					big.NewInt(123456789),
				},
			},
		},
		{
			name:      "testing uint256",
			signature: "method(uint256)",
			inputs:    []string{"123456789"},
			expectMethodDef: &MethodDef{
				Name:       "method",
				Parameters: []*MethodParameter{{TypeName: "uint256"}},
			},
			expectMethodCall: &MethodCall{
				MethodDef: &MethodDef{Name: "method", Parameters: []*MethodParameter{{TypeName: "uint256"}}},
				Data: []interface{}{
					big.NewInt(123456789),
				},
			},
		},
		{
			name:      "testing bool",
			signature: "method(bool)",
			inputs:    []string{"true"},
			expectMethodDef: &MethodDef{
				Name:       "method",
				Parameters: []*MethodParameter{{TypeName: "bool"}},
			},
			expectMethodCall: &MethodCall{
				MethodDef: &MethodDef{Name: "method", Parameters: []*MethodParameter{{TypeName: "bool"}}},
				Data: []interface{}{
					true,
				},
			},
		},
		{
			name:      "testing address[]",
			signature: "method(address[])",
			inputs: []string{
				"[\"0x5a0b54d5dc17e0aadc383d2db43b0a0d3e029c4c\",\"0x5a0b54d5dc17e0aadc383d2db43b0a0d3e029c4c\"]",
			},
			expectMethodDef: &MethodDef{
				Name:       "method",
				Parameters: []*MethodParameter{{TypeName: "address[]"}},
			},
			expectMethodCall: &MethodCall{
				MethodDef: &MethodDef{Name: "method", Parameters: []*MethodParameter{{TypeName: "address[]"}}},
				Data: []interface{}{
					[]Address{
						MustNewAddress("5a0b54d5dc17e0aadc383d2db43b0a0d3e029c4c"),
						MustNewAddress("5a0b54d5dc17e0aadc383d2db43b0a0d3e029c4c"),
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			methodDef, err := NewMethodDef(test.signature)
			require.NoError(t, err)
			assert.Equal(t, test.expectMethodDef, methodDef)

			methodCall := methodDef.NewCall()
			for _, input := range test.inputs {
				methodCall.AppendArgFromString(input)
			}

			assert.Equal(t, test.expectMethodCall, methodCall)
		})
	}
}

func TestMethodCallData_AppendArgFromString(t *testing.T) {
	tests := []struct {
		name         string
		methodDef    *MethodDef
		inputs       []string
		expectedData []interface{}
	}{
		{
			name: "testing tuple",
			methodDef: &MethodDef{
				Name: "tuple",
				Parameters: []*MethodParameter{
					{Name: "period", TypeName: "tuple", Components: []*StructComponent{
						{Name: "tokenID", TypeName: "uint256"},
						{Name: "fromBlockNum", TypeName: "uint64"},
						{Name: "toBlockNum", TypeName: "uint64"},
					}},
				},
			},
			inputs: []string{
				`{"tokenID":1,"fromBlockNum":"0xa","toBlockNum":"11"}`,
			},
			expectedData: []interface{}{
				[]interface{}{
					big.NewInt(1),
					Uint64(10),
					Uint64(11),
				}},
		},
		{
			name: "testing tuple[]",
			methodDef: &MethodDef{
				Name: "tupleArray",
				Parameters: []*MethodParameter{
					{Name: "periods", TypeName: "tuple[]", Components: []*StructComponent{
						{Name: "tokenID", TypeName: "uint256"},
						{Name: "fromBlockNum", TypeName: "uint64"},
						{Name: "toBlockNum", TypeName: "uint64"},
					}},
				},
			},
			inputs: []string{
				`[{"tokenID":1,"fromBlockNum":"0xa","toBlockNum":"11"},{"tokenID":2,"fromBlockNum":"0x14","toBlockNum":"21"}]`,
			},
			expectedData: []interface{}{
				[]interface{}{
					[]interface{}{
						big.NewInt(1),
						Uint64(10),
						Uint64(11),
					},
					[]interface{}{
						big.NewInt(2),
						Uint64(20),
						Uint64(21),
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			methodCall := test.methodDef.NewCall()
			for _, input := range test.inputs {
				methodCall.AppendArgFromString(input)
			}

			require.Len(t, methodCall.err, 0)
			assert.Equal(t, test.expectedData, methodCall.Data)
		})
	}
}

func TestMethodCall_AppendArgFromStringTooManyArgs(t *testing.T) {
	methodDef, err := NewMethodDef("method(address)")
	require.NoError(t, err)
	assert.Equal(t, &MethodDef{
		Name:       "method",
		Parameters: []*MethodParameter{{TypeName: "address"}},
	}, methodDef)

	methodCall := methodDef.NewCall()
	methodCall.AppendArgFromString("0x5a0b54d5dc17e0aadc383d2db43b0a0d3e029c4c")
	methodCall.AppendArgFromString("0x5a0b54d5dc17e0aadc383d2db43b0a0d3e029c4c")
	_, err = methodCall.Encode()
	assert.Error(t, err)
}

func TestNewMethodDef(t *testing.T) {
	var tests = []struct {
		name            string
		signature       string
		expectMethodDef *MethodDef
		expectError     bool
	}{
		{
			name:        "not a method",
			signature:   "not a method",
			expectError: true,
		},
		{
			name:      "method no arg",
			signature: "method()",
			expectMethodDef: &MethodDef{
				Name: "method",
			},
		},
		{
			name:      "method one arg",
			signature: "method(address)",
			expectMethodDef: &MethodDef{
				Name:       "method",
				Parameters: []*MethodParameter{{TypeName: "address"}},
			},
		},
		{
			name:      "method one array arg",
			signature: "method(address[])",
			expectMethodDef: &MethodDef{
				Name:       "method",
				Parameters: []*MethodParameter{{TypeName: "address[]"}},
			},
		},
		{
			name:      "method one named array arg",
			signature: "method(address[] ids)",
			expectMethodDef: &MethodDef{
				Name:       "method",
				Parameters: []*MethodParameter{{Name: "ids", TypeName: "address[]"}},
			},
		},
		{
			name:      "method one payable arg",
			signature: "method(address payable)",
			expectMethodDef: &MethodDef{
				Name:       "method",
				Parameters: []*MethodParameter{{TypeName: "address", Payable: true}},
			},
		},
		{
			name:      "method one payable named arg",
			signature: "method(address payable recipient)",
			expectMethodDef: &MethodDef{
				Name:       "method",
				Parameters: []*MethodParameter{{Name: "recipient", TypeName: "address", Payable: true}},
			},
		},
		{
			name:      "method one arg with returns",
			signature: "method(address) (uint256)",
			expectMethodDef: &MethodDef{
				Name:             "method",
				Parameters:       []*MethodParameter{{TypeName: "address"}},
				ReturnParameters: []*MethodParameter{{TypeName: "uint256"}},
			},
		},
		{
			name:      "method one arg with named returns",
			signature: "method(address) (uint256 value)",
			expectMethodDef: &MethodDef{
				Name:             "method",
				Parameters:       []*MethodParameter{{TypeName: "address"}},
				ReturnParameters: []*MethodParameter{{Name: "value", TypeName: "uint256"}},
			},
		},
		{
			name:      "method one arg with returns keyword",
			signature: "method(address) returns (uint256)",
			expectMethodDef: &MethodDef{
				Name:             "method",
				Parameters:       []*MethodParameter{{TypeName: "address"}},
				ReturnParameters: []*MethodParameter{{TypeName: "uint256"}},
			},
		},
		{
			name:      "method mutli arg",
			signature: "method(address,uint256)",
			expectMethodDef: &MethodDef{
				Name: "method",
				Parameters: []*MethodParameter{
					{TypeName: "address"},
					{TypeName: "uint256"},
				},
			},
		},
		{
			name:      "method mutli arg and names",
			signature: " transferFrom(address sender, address recipient, uint256 amount) ",
			expectMethodDef: &MethodDef{
				Name: "transferFrom",
				Parameters: []*MethodParameter{
					{Name: "sender", TypeName: "address"},
					{Name: "recipient", TypeName: "address"},
					{Name: "amount", TypeName: "uint256"},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			methodDef, err := NewMethodDef(test.signature)
			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectMethodDef, methodDef)
			}
		})
	}

}

func TestNewMethodParameter(t *testing.T) {
	var tests = []struct {
		name              string
		methodParameter   string
		expectMethodParam *MethodParameter
		expectError       bool
	}{
		{
			name:            "method no arg",
			methodParameter: "address",
			expectMethodParam: &MethodParameter{
				TypeName: "address",
			},
		},
		{
			name:            "method one arg",
			methodParameter: "address recipient",
			expectMethodParam: &MethodParameter{
				Name:     "recipient",
				TypeName: "address",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m, err := newMethodParameter(test.methodParameter)
			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectMethodParam, m)
			}
		})
	}

}

func TestMethodDef_Signature_NestedTuples(t *testing.T) {
	tests := []struct {
		name            string
		methodDef       *MethodDef
		expectedSig     string
		expectedMethodID string // first 4 bytes of keccak256(signature) in hex
	}{
		{
			name: "single-level tuple",
			methodDef: &MethodDef{
				Name: "encodeRAV",
				Parameters: []*MethodParameter{
					{
						Name:     "rav",
						TypeName: "tuple",
						Components: []*StructComponent{
							{Name: "collectionId", TypeName: "bytes32"},
							{Name: "payer", TypeName: "address"},
							{Name: "serviceProvider", TypeName: "address"},
							{Name: "dataService", TypeName: "address"},
							{Name: "timestampNs", TypeName: "uint64"},
							{Name: "valueAggregate", TypeName: "uint128"},
							{Name: "metadata", TypeName: "bytes"},
						},
					},
				},
			},
			expectedSig:     "encodeRAV((bytes32,address,address,address,uint64,uint128,bytes))",
			expectedMethodID: "26969c4c",
		},
		{
			name: "nested tuple - SignedRAV with nested RAV",
			methodDef: &MethodDef{
				Name: "recoverRAVSigner",
				Parameters: []*MethodParameter{
					{
						Name:     "signedRAV",
						TypeName: "tuple",
						Components: []*StructComponent{
							{
								Name:     "rav",
								TypeName: "tuple",
								Components: []*StructComponent{
									{Name: "collectionId", TypeName: "bytes32"},
									{Name: "payer", TypeName: "address"},
									{Name: "serviceProvider", TypeName: "address"},
									{Name: "dataService", TypeName: "address"},
									{Name: "timestampNs", TypeName: "uint64"},
									{Name: "valueAggregate", TypeName: "uint128"},
									{Name: "metadata", TypeName: "bytes"},
								},
							},
							{Name: "signature", TypeName: "bytes"},
						},
					},
				},
			},
			expectedSig:     "recoverRAVSigner(((bytes32,address,address,address,uint64,uint128,bytes),bytes))",
			expectedMethodID: "63648817",
		},
		{
			name: "nested tuple - TupleWithNestedDynamic",
			methodDef: &MethodDef{
				Name: "funTupleWithNestedDynamic",
				Parameters: []*MethodParameter{
					{
						Name:     "input",
						TypeName: "tuple",
						Components: []*StructComponent{
							{Name: "owner", TypeName: "address"},
							{
								Name:     "inner",
								TypeName: "tuple",
								Components: []*StructComponent{
									{Name: "id", TypeName: "uint256"},
									{Name: "description", TypeName: "string"},
								},
							},
							{Name: "timestamp", TypeName: "uint64"},
						},
					},
				},
			},
			expectedSig:      "funTupleWithNestedDynamic((address,(uint256,string),uint64))",
			expectedMethodID: "56e67bfe",
		},
		{
			name: "deeply nested tuple",
			methodDef: &MethodDef{
				Name: "deepNested",
				Parameters: []*MethodParameter{
					{
						Name:     "data",
						TypeName: "tuple",
						Components: []*StructComponent{
							{
								Name:     "level1",
								TypeName: "tuple",
								Components: []*StructComponent{
									{
										Name:     "level2",
										TypeName: "tuple",
										Components: []*StructComponent{
											{Name: "value", TypeName: "uint256"},
										},
									},
									{Name: "addr", TypeName: "address"},
								},
							},
							{Name: "data", TypeName: "bytes"},
						},
					},
				},
			},
			expectedSig:      "deepNested((((uint256),address),bytes))",
			expectedMethodID: "c2f64ed1",
		},
		{
			name: "array of tuples",
			methodDef: &MethodDef{
				Name: "batchProcess",
				Parameters: []*MethodParameter{
					{
						Name:     "items",
						TypeName: "tuple[]",
						Components: []*StructComponent{
							{Name: "id", TypeName: "uint256"},
							{Name: "owner", TypeName: "address"},
						},
					},
				},
			},
			expectedSig:      "batchProcess((uint256,address)[])",
			expectedMethodID: "584a3c48",
		},
		{
			name: "array of nested tuples",
			methodDef: &MethodDef{
				Name: "batchNested",
				Parameters: []*MethodParameter{
					{
						Name:     "items",
						TypeName: "tuple[]",
						Components: []*StructComponent{
							{
								Name:     "inner",
								TypeName: "tuple",
								Components: []*StructComponent{
									{Name: "value", TypeName: "uint256"},
									{Name: "name", TypeName: "string"},
								},
							},
							{Name: "flag", TypeName: "bool"},
						},
					},
				},
			},
			expectedSig:      "batchNested(((uint256,string),bool)[])",
			expectedMethodID: "6433c36d",
		},
		{
			name: "mixed parameters with nested tuple",
			methodDef: &MethodDef{
				Name: "mixedParams",
				Parameters: []*MethodParameter{
					{Name: "id", TypeName: "uint256"},
					{
						Name:     "data",
						TypeName: "tuple",
						Components: []*StructComponent{
							{
								Name:     "nested",
								TypeName: "tuple",
								Components: []*StructComponent{
									{Name: "a", TypeName: "address"},
									{Name: "b", TypeName: "bytes32"},
								},
							},
							{Name: "value", TypeName: "bytes"},
						},
					},
					{Name: "recipient", TypeName: "address"},
				},
			},
			expectedSig:      "mixedParams(uint256,((address,bytes32),bytes),address)",
			expectedMethodID: "fe2f7c70",
		},
		{
			name: "fixed-size array of tuples",
			methodDef: &MethodDef{
				Name: "fixedArray",
				Parameters: []*MethodParameter{
					{
						Name:     "items",
						TypeName: "tuple[3]",
						Components: []*StructComponent{
							{Name: "x", TypeName: "uint256"},
							{Name: "y", TypeName: "uint256"},
						},
					},
				},
			},
			expectedSig:      "fixedArray((uint256,uint256)[3])",
			expectedMethodID: "eb589c90",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig := tt.methodDef.Signature()
			assert.Equal(t, tt.expectedSig, sig, "signature mismatch")

			methodID := tt.methodDef.MethodID()
			assert.Equal(t, tt.expectedMethodID, hex.EncodeToString(methodID), "method ID mismatch")
		})
	}
}

func TestMethodDef_Signature_FromParsedABI(t *testing.T) {
	abi, err := ParseABI("testdata/nested_tuple.abi.json")
	require.NoError(t, err)

	t.Run("recoverRAVSigner from parsed ABI", func(t *testing.T) {
		fn := abi.FindFunctionByName("recoverRAVSigner")
		require.NotNil(t, fn, "function recoverRAVSigner should exist in ABI")

		expectedSig := "recoverRAVSigner(((bytes32,address,address,address,uint64,uint128,bytes),bytes))"
		assert.Equal(t, expectedSig, fn.Signature(), "signature should recursively expand nested tuples")

		// Expected selector: keccak256("recoverRAVSigner(((bytes32,address,address,address,uint64,uint128,bytes),bytes))")[:4]
		assert.Equal(t, "63648817", hex.EncodeToString(fn.MethodID()), "method ID should match correct selector")
	})

	t.Run("funTupleWithNestedDynamic from parsed ABI", func(t *testing.T) {
		fn := abi.FindFunctionByName("funTupleWithNestedDynamic")
		require.NotNil(t, fn, "function funTupleWithNestedDynamic should exist in ABI")

		expectedSig := "funTupleWithNestedDynamic((address,(uint256,string),uint64))"
		assert.Equal(t, expectedSig, fn.Signature(), "signature should recursively expand nested tuples")

		// Expected selector: keccak256("funTupleWithNestedDynamic((address,(uint256,string),uint64))")[:4]
		assert.Equal(t, "56e67bfe", hex.EncodeToString(fn.MethodID()), "method ID should match correct selector")
	})
}
