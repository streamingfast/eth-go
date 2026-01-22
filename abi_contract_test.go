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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestABIContract_Parse(t *testing.T) {
	var arrayOfStructType = ArrayType{ElementType: StructType{}}
	var uint64Type = UnsignedIntegerType{BitsSize: 64, ByteSize: 8}
	var uint256Type = UnsignedIntegerType{BitsSize: 256, ByteSize: 32}

	tests := []struct {
		name        string
		expected    *ABI
		expectedErr error
	}{
		{
			"log event indexed",
			&ABI{
				FunctionsMap:       map[string][]*MethodDef{},
				FunctionsByNameMap: map[string][]*MethodDef{},
				LogEventsMap: map[string][]*LogEventDef{
					string(b(t, "b14a725aeeb25d591b81b16b4c5b25403dd8867bdd1876fa787867f566206be1")): {{
						Name: "PairCreated",
						Parameters: []*LogParameter{
							{Name: "token0", TypeName: "address", Type: AddressType{}, Indexed: true},
						},
					}},
				},
				LogEventsByNameMap: map[string][]*LogEventDef{
					"PairCreated": {{
						Name: "PairCreated",
						Parameters: []*LogParameter{
							{Name: "token0", TypeName: "address", Type: AddressType{}, Indexed: true},
						},
					}},
				},
			},
			nil,
		},

		{
			"object format with abi key",
			&ABI{
				FunctionsMap:       map[string][]*MethodDef{},
				FunctionsByNameMap: map[string][]*MethodDef{},
				LogEventsMap: map[string][]*LogEventDef{
					string(b(t, "b14a725aeeb25d591b81b16b4c5b25403dd8867bdd1876fa787867f566206be1")): {{
						Name: "PairCreated",
						Parameters: []*LogParameter{
							{Name: "token0", TypeName: "address", Type: AddressType{}, Indexed: true},
						},
					}},
				},
				LogEventsByNameMap: map[string][]*LogEventDef{
					"PairCreated": {{
						Name: "PairCreated",
						Parameters: []*LogParameter{
							{Name: "token0", TypeName: "address", Type: AddressType{}, Indexed: true},
						},
					}},
				},
			},
			nil,
		},

		{
			"struct tuple alone",
			&ABI{
				FunctionsMap: map[string][]*MethodDef{
					string(b(t, "b961c32d")): {{
						Name: "tupleAlone",
						Parameters: []*MethodParameter{
							{Name: "period", TypeName: "tuple", Type: StructType{}, InternalType: "struct ClaimPeriod", TypeMutability: "", Components: []*StructComponent{
								{Name: "tokenID", TypeName: "uint256", Type: uint256Type, InternalType: "uint256"},
								{Name: "fromBlockNum", TypeName: "uint64", Type: uint64Type, InternalType: "uint64"},
								{Name: "toBlockNum", TypeName: "uint64", Type: uint64Type, InternalType: "uint64"},
							}},
						},
						StateMutability: StateMutabilityPure,
					}},
				},
				FunctionsByNameMap: map[string][]*MethodDef{
					"tupleAlone": {{
						Name: "tupleAlone",
						Parameters: []*MethodParameter{
							{Name: "period", TypeName: "tuple", Type: StructType{}, InternalType: "struct ClaimPeriod", TypeMutability: "", Components: []*StructComponent{
								{Name: "tokenID", TypeName: "uint256", Type: uint256Type, InternalType: "uint256"},
								{Name: "fromBlockNum", TypeName: "uint64", Type: uint64Type, InternalType: "uint64"},
								{Name: "toBlockNum", TypeName: "uint64", Type: uint64Type, InternalType: "uint64"},
							}},
						},
						StateMutability: StateMutabilityPure,
					}},
				},
			},
			nil,
		},

		{
			"struct tuple array",
			&ABI{
				FunctionsMap: map[string][]*MethodDef{
					string(b(t, "15eb963d")): {{
						Name: "tupleArray",
						Parameters: []*MethodParameter{
							{Name: "periods", TypeName: "tuple[]", Type: arrayOfStructType, InternalType: "struct ClaimPeriod[]", TypeMutability: "", Components: []*StructComponent{
								{Name: "tokenID", TypeName: "uint256", Type: uint256Type, InternalType: "uint256"},
								{Name: "fromBlockNum", TypeName: "uint64", Type: uint64Type, InternalType: "uint64"},
								{Name: "toBlockNum", TypeName: "uint64", Type: uint64Type, InternalType: "uint64"},
							}},
						},
						StateMutability: StateMutabilityView,
					}},
				},
				FunctionsByNameMap: map[string][]*MethodDef{
					"tupleArray": {{
						Name: "tupleArray",
						Parameters: []*MethodParameter{
							{Name: "periods", TypeName: "tuple[]", Type: arrayOfStructType, InternalType: "struct ClaimPeriod[]", TypeMutability: "", Components: []*StructComponent{
								{Name: "tokenID", TypeName: "uint256", Type: uint256Type, InternalType: "uint256"},
								{Name: "fromBlockNum", TypeName: "uint64", Type: uint64Type, InternalType: "uint64"},
								{Name: "toBlockNum", TypeName: "uint64", Type: uint64Type, InternalType: "uint64"},
							}},
						},
						StateMutability: StateMutabilityView,
					}},
				},
			},
			nil,
		},

		{
			"log event multiple same name",
			&ABI{
				FunctionsMap:       map[string][]*MethodDef{},
				FunctionsByNameMap: map[string][]*MethodDef{},
				LogEventsMap: map[string][]*LogEventDef{
					string(b(t, "02508328cccf110f3046b1f9c5a6a27376177a48224fb84fae5af4956a3e95b7")): {
						{
							Name: "PairCreated",
							Parameters: []*LogParameter{
								{Name: "second", TypeName: "bytes32", Type: FixedSizeBytesType{ByteSize: 32}, Indexed: false},
							},
						},
					},
					string(b(t, "b14a725aeeb25d591b81b16b4c5b25403dd8867bdd1876fa787867f566206be1")): {
						{
							Name: "PairCreated",
							Parameters: []*LogParameter{
								{Name: "first", TypeName: "address", Type: AddressType{}, Indexed: true},
							},
						},
						{
							Name: "PairCreated",
							Parameters: []*LogParameter{
								{Name: "third", TypeName: "address", Type: AddressType{}, Indexed: false},
							},
						},
					},
				},
				LogEventsByNameMap: map[string][]*LogEventDef{
					"PairCreated": {
						{
							Name: "PairCreated",
							Parameters: []*LogParameter{
								{Name: "first", TypeName: "address", Type: AddressType{}, Indexed: true},
							},
						},
						{
							Name: "PairCreated",
							Parameters: []*LogParameter{
								{Name: "second", TypeName: "bytes32", Type: FixedSizeBytesType{ByteSize: 32}, Indexed: false},
							},
						},
						{
							Name: "PairCreated",
							Parameters: []*LogParameter{
								{Name: "third", TypeName: "address", Type: AddressType{}, Indexed: false},
							},
						},
					},
				},
			},
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			abiFile := filepath.Join("testdata", strings.ReplaceAll(test.name, " ", "_")+".abi.json")
			stat, err := os.Stat(abiFile)
			require.NoError(t, err, "unable to stat abi file %q (inferred from %s space replaced by _)", abiFile, test.name)
			require.Equal(t, true, !stat.IsDir(), "abi file %q is a directory (inferred from %s space replaced by _)", abiFile, test.name)

			abi, err := ParseABI(abiFile)
			if test.expectedErr == nil {
				require.NoError(t, err)
				abiEquals(t, test.expected, abi)
			} else {
				assert.Equal(t, test.expectedErr, err)
			}
		})
	}
}

func TestABIContract_ParseFile(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		expectedErr error
	}{
		{"standard", "testdata/uniswap_v2_factory.abi.json", nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseABI(test.filename)
			if test.expectedErr == nil {
				require.NoError(t, err)
			} else {
				assert.Equal(t, test.expectedErr, err)
			}
		})
	}
}

func TestABIContract_ParseConstructor(t *testing.T) {
	t.Run("simple constructor with address", func(t *testing.T) {
		abi, err := ParseABI("testdata/uniswap_v2_factory.abi.json")
		require.NoError(t, err)

		constructor := abi.FindConstructor()
		require.NotNil(t, constructor, "constructor should be parsed")
		require.Len(t, constructor.Parameters, 1)
		assert.Equal(t, "address", constructor.Parameters[0].TypeName)
		assert.Equal(t, "_feeToSetter", constructor.Parameters[0].Name)
		assert.Equal(t, StateMutabilityNonPayable, constructor.StateMutability)
		assert.Equal(t, "(address)", constructor.Signature())
	})

	t.Run("constructor with tuple", func(t *testing.T) {
		abi, err := ParseABI("testdata/constructor_with_tuple.abi.json")
		require.NoError(t, err)

		constructor := abi.FindConstructor()
		require.NotNil(t, constructor, "constructor should be parsed")
		require.Len(t, constructor.Parameters, 1)
		assert.Equal(t, "tuple", constructor.Parameters[0].TypeName)
		assert.Equal(t, "config", constructor.Parameters[0].Name)

		// Verify nested components were parsed
		require.Len(t, constructor.Parameters[0].Components, 3)
		assert.Equal(t, "address", constructor.Parameters[0].Components[0].TypeName)
		assert.Equal(t, "token", constructor.Parameters[0].Components[0].Name)
		assert.Equal(t, "uint256", constructor.Parameters[0].Components[1].TypeName)
		assert.Equal(t, "amount", constructor.Parameters[0].Components[1].Name)
		assert.Equal(t, "string", constructor.Parameters[0].Components[2].TypeName)
		assert.Equal(t, "name", constructor.Parameters[0].Components[2].Name)

		assert.Equal(t, "((address,uint256,string))", constructor.Signature())
	})

	t.Run("constructor lookup methods", func(t *testing.T) {
		abi, err := ParseABI("testdata/uniswap_v2_factory.abi.json")
		require.NoError(t, err)

		// Test FindConstructor
		constructor := abi.FindConstructor()
		require.NotNil(t, constructor)

		// Test FindConstructors (should return all constructors)
		constructors := abi.FindConstructors()
		require.Len(t, constructors, 1)

		// Test FindConstructorBySignature
		constructorBySignature := abi.FindConstructorBySignature("(address)")
		require.NotNil(t, constructorBySignature)
		assert.Equal(t, constructor, constructorBySignature)

		// Test with non-existent signature
		constructorNotFound := abi.FindConstructorBySignature("(uint256)")
		assert.Nil(t, constructorNotFound)
	})
}

func abiEquals(t *testing.T, expected *ABI, actual *ABI) {
	if len(expected.LogEventsByNameMap) != len(actual.LogEventsByNameMap) {
		require.Equal(t, expected.LogEventsByNameMap, actual.LogEventsByNameMap)
	} else {
		for key, value := range expected.LogEventsByNameMap {
			assert.Contains(t, actual.LogEventsByNameMap, key, "log event name %s", key)
			assert.Equal(t, value, actual.LogEventsByNameMap[key])
		}
	}

	if len(expected.LogEventsMap) != len(actual.LogEventsMap) {
		require.Equal(t, expected.LogEventsMap, actual.LogEventsMap)
	} else {
		for key, value := range expected.LogEventsMap {
			assert.Contains(t, actual.LogEventsMap, key, "log event id %s", key)
			assert.Equal(t, value, actual.LogEventsMap[key])
		}
	}

	if len(expected.FunctionsByNameMap) != len(actual.FunctionsByNameMap) {
		require.Equal(t, expected.FunctionsByNameMap, actual.FunctionsByNameMap)
	} else {
		for key, value := range expected.FunctionsByNameMap {
			assert.Contains(t, actual.FunctionsByNameMap, key, "method name %s", key)
			assert.Equal(t, value, actual.FunctionsByNameMap[key])
		}
	}

	if len(expected.FunctionsMap) != len(actual.FunctionsMap) {
		require.Equal(t, expected.FunctionsMap, actual.FunctionsMap)
	} else {
		for key, value := range expected.FunctionsMap {
			assert.Contains(t, actual.FunctionsMap, key, "method id %s", key)
			assert.Equal(t, value, actual.FunctionsMap[key])
		}
	}
}
