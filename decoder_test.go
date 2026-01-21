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
	"errors"
	"math/big"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecoder_Read(t *testing.T) {
	tests := []struct {
		name        string
		typeName    string
		in          string
		expectOut   interface{}
		expectError bool
	}{
		{
			name:      "bool true",
			typeName:  "bool",
			in:        "0x0000000000000000000000000000000000000000000000000000000000000001",
			expectOut: true,
		},
		{
			name:      "bool false",
			typeName:  "bool",
			in:        "0x0000000000000000000000000000000000000000000000000000000000000000",
			expectOut: false,
		},
		{
			name:      "uint8",
			typeName:  "uint8",
			in:        "0x0000000000000000000000000000000000000000000000000000000000000007",
			expectOut: uint8(7),
		},
		{
			name:      "uint24",
			typeName:  "uint24",
			in:        "0x0000000000000000000000000000000000000000000000000000000000ffffff",
			expectOut: uint32(16777215),
		},
		{
			name:      "uint56",
			typeName:  "uint56",
			in:        "0x00000000000000000000000000000000000000000000000000fffffffffffffd",
			expectOut: uint64(72057594037927933),
		},
		{
			name:      "uint64",
			typeName:  "uint64",
			in:        "0x00000000000000000000000000000000000000000000000000000000000002a1",
			expectOut: uint64(673),
		},
		{
			name:      "uint96",
			typeName:  "uint96",
			in:        "0x0000000000000000000000000000000000000000ffffffffffffffffffffffff",
			expectOut: bigString(t, "79228162514264337593543950335"),
		},
		{
			name:      "uint112",
			typeName:  "uint112",
			in:        "0x000000000000000000000000000000000000ffffffffffffffffffffffffffff",
			expectOut: bigString(t, "5192296858534827628530496329220095"),
		},
		{
			name:      "uint256",
			typeName:  "uint256",
			in:        "0x0000000000000000000000000000000000000000000000000000000000000b7a",
			expectOut: big.NewInt(2938),
		},
		{
			name:      "bigger uint256",
			typeName:  "uint256",
			in:        "0x00000000000000000000000000000000000000000000000000003c12826fe23b",
			expectOut: big.NewInt(66050195448379),
		},
		{
			name:      "address",
			typeName:  "address",
			in:        "0x0000000000000000000000007d97ba95dac25316b9531152b3baa32327994da8",
			expectOut: MustNewAddress("7d97ba95dac25316b9531152b3baa32327994da8"),
		},
		{
			name:      "method",
			typeName:  "method",
			in:        "0xa9059cbb",
			expectOut: "transfer(address recipient,uint256 amount)",
		},
		{
			name:      "string",
			typeName:  "string",
			in:        "0x0000000000000000000000000000000000000000000000000000000000000011556e697377617056323a204c4f434b4544000000000000000000000000000000",
			expectOut: "UniswapV2: LOCKED",
		},
		{
			name:      "bytes",
			typeName:  "bytes",
			in:        "0x00000000000000000000000000000000000000000000000000000000000000050103aabbcc",
			expectOut: []byte{0x01, 0x03, 0xaa, 0xbb, 0xcc},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := NewDecoderFromString(test.in)
			require.NoError(t, err)

			out, err := d.Read(test.typeName)
			if test.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectOut, out)
			}
		})
	}
}

func TestDecoder_ReadArray(t *testing.T) {
	tests := []struct {
		name        string
		typeName    string
		in          string
		expectOut   interface{}
		expectError bool
	}{
		{
			name:      "bool",
			typeName:  "bool[]",
			in:        "0x0000000000000000000000000000000000000000000000000000000000000003000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001",
			expectOut: BoolArray{true, false, true},
		},
		{
			name:     "uint8",
			typeName: "uint8[]",
			in:       "0x0000000000000000000000000000000000000000000000000000000000000003000000000000000000000000000000000000000000000000000000000000000700000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000002",

			expectOut: Uint8Array{7, 1, 2},
		},
		{
			name:      "uint24",
			typeName:  "uint24[]",
			in:        "0x0000000000000000000000000000000000000000000000000000000000000003000000000000000000000000000000000000000000000000000000000000000300000000000000000000000000000000000000000000000000000000000000070000000000000000000000000000000000000000000000000000000000000005",
			expectOut: Uint32Array{3, 7, 5},
		},
		{
			name:      "uint56",
			typeName:  "uint56[]",
			in:        "0x000000000000000000000000000000000000000000000000000000000000000300000000000000000000000000000000000000000000000000fffffffffffffd00000000000000000000000000000000000000000000000000fffffffffffffd00000000000000000000000000000000000000000000000000fffffffffffffd",
			expectOut: Uint64Array{72057594037927933, 72057594037927933, 72057594037927933},
		},
		{
			name:      "uint64",
			typeName:  "uint64[]",
			in:        "0x000000000000000000000000000000000000000000000000000000000000000300000000000000000000000000000000000000000000000000000000000002a1000000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000009",
			expectOut: Uint64Array{673, 10, 9},
		},
		{
			name:      "uint96",
			typeName:  "uint96[]",
			in:        "0x00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000ffffffffffffffffffffffff0000000000000000000000000000000000000000000000000000000000000003",
			expectOut: BigIntArray{bigString(t, "79228162514264337593543950335"), bigString(t, "3")},
		},
		{
			name:      "uint112",
			typeName:  "uint112[]",
			in:        "0x000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000050000000000000000000000000000000000000000000000000000000000000003",
			expectOut: BigIntArray{bigString(t, "5"), bigString(t, "3")},
		},
		{
			name:      "uint256",
			typeName:  "uint256[]",
			in:        "0x00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000b7a",
			expectOut: BigIntArray{bigString(t, "2938")},
		},
		{
			name:     "address",
			typeName: "address[]",
			in:       "0x00000000000000000000000000000000000000000000000000000000000000020000000000000000000000007d97ba95dac25316b9531152b3baa32327994da8000000000000000000000000c778417e063141139fce010982780140aa0cd5ab",
			expectOut: AddressArray{
				MustNewAddress("7d97ba95dac25316b9531152b3baa32327994da8"),
				MustNewAddress("c778417e063141139fce010982780140aa0cd5ab"),
			},
		},
		{
			name:      "string",
			typeName:  "string[]",
			in:        "0x00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000011556e697377617056323a204c4f434b4544000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000b68656c6c6f20776f726c64000000000000000000000000000000000000000000",
			expectOut: StringArray{"UniswapV2: LOCKED", "hello world"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := NewDecoderFromString(test.in)
			require.NoError(t, err)

			out, err := d.Read(test.typeName)
			if test.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectOut, out)
			}
		})
	}
}

func TestDecoder_ReadMethodCall(t *testing.T) {
	tests := []struct {
		name           string
		inStr          string
		expectedMethod *MethodCall
		expectedErr    error
	}{
		{
			name: "transfer(address,uint256)",
			inStr: `
				a9059cbb
				000000000000000000000000aadf939f53a1b9a3df7082bdc47d01083e8ebfad
				00000000000000000000000000000000000000000000003635c9adc5dea00000
			`,
			expectedMethod: &MethodCall{
				MethodDef: &MethodDef{
					Name: "transfer",
					Parameters: []*MethodParameter{
						{TypeName: "address", Name: "recipient"},
						{TypeName: "uint256", Name: "amount"},
					},
				},
				Data: []interface{}{
					MustNewAddress("aadf939f53a1b9a3df7082bdc47d01083e8ebfad"),
					bigString(t, "1000000000000000000000"),
				},
			},
		},
		{
			name: "sendVote(string) with last valid offset",
			inStr: `
				0146d0ca
				0000000000000000000000000000000000000000000000000000000000000040
				000000000000000000000000000000000000000000000000000000000000FFFF
				0000000000000000000000000000000000000000000000000000000000000000
			`,
			expectedMethod: &MethodCall{
				MethodDef: &MethodDef{
					Name: "sendVote",
					Parameters: []*MethodParameter{
						{TypeName: "string"},
					},
				},
				Data: []interface{}{""},
			},
		},
		{
			name: "sendVote(string) with first invalid offset",
			inStr: `
				0146d0ca
				0000000000000000000000000000000000000000000000000000000000000041
				000000000000000000000000000000000000000000000000000000000000FFFF
				0000000000000000000000000000000000000000000000000000000000000000
			`,
			expectedErr: errors.New(`read parameters: invalid offset value 69 (max possible value 68) for type "string" (element #0) at offset 36`),
		},
		{
			name:  "entry(uint256,uint256,bool,address,address,bytes)",
			inStr: "0x3840c6280000000000000000000000000000000000000000000000000de0b6b3a7640000000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000100000000000000000000000032851cb39b6c2bf9c6e3f50451528f1e177b4c86000000000000000000000000160bf3d59811b6fb44fbc360c5bfa40be9de9b9f00000000000000000000000000000000000000000000000000000000000000c0000000000000000000000000000000000000000000000000000000000000014438ed17390000000000000000000000000000000000000000000000000de0b6b3a7640000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000004f48d8af9a9d119314bc1a364c0432562d249863000000000000000000000000000000000000000000000000000000005f7fcfbd000000000000000000000000000000000000000000000000000000000000000400000000000000000000000032851cb39b6c2bf9c6e3f50451528f1e177b4c860000000000000000000000003e8caad87c7c94973bf70038c51ee40ffb37930d000000000000000000000000a48762c9c79e5d700bb164c6e5dfaf22f3a4de0600000000000000000000000032851cb39b6c2bf9c6e3f50451528f1e177b4c86",
			expectedMethod: &MethodCall{
				MethodDef: &MethodDef{
					Name: "entry",
					Parameters: []*MethodParameter{
						{Name: "loanAmount", TypeName: "uint256"},
						{Name: "loanTokenIndex", TypeName: "uint256"},
						{Name: "burnChi", TypeName: "bool"},
						{Name: "loanTokenAddress", TypeName: "address"},
						{Name: "pairAddress", TypeName: "address"},
						{Name: "data", TypeName: "bytes"},
					},
				},
				Data: []interface{}{
					bigString(t, "1000000000000000000"),
					bigString(t, "10"),
					true,
					MustNewAddress("32851cb39b6c2bf9c6e3f50451528f1e177b4c86"),
					MustNewAddress("160bf3d59811b6fb44fbc360c5bfa40be9de9b9f"),
					B("38ed17390000000000000000000000000000000000000000000000000de0b6b3a7640000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000004f48d8af9a9d119314bc1a364c0432562d249863000000000000000000000000000000000000000000000000000000005f7fcfbd000000000000000000000000000000000000000000000000000000000000000400000000000000000000000032851cb39b6c2bf9c6e3f50451528f1e177b4c860000000000000000000000003e8caad87c7c94973bf70038c51ee40ffb37930d000000000000000000000000a48762c9c79e5d700bb164c6e5dfaf22f3a4de0600000000000000000000000032851cb39b6c2bf9c6e3f50451528f1e177b4c86"),
				},
			},
		},
	}

	spaceRegex := regexp.MustCompile(`\s+`)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			in := test.inStr
			if !strings.HasPrefix(in, "0x") {
				in = "0x" + in
			}

			dec, err := NewDecoderFromString(spaceRegex.ReplaceAllString(in, ""))
			require.NoError(t, err)

			m, err := dec.ReadMethodCall()
			if test.expectedErr != nil {
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectedMethod, m)
			}
		})
	}
}

// TestDecoder_ReadTupleWithDynamicComponents tests decoding tuples that contain
// dynamic types like `bytes` or `string`. Per Solidity ABI spec, such tuples are
// considered dynamic and use offset-based encoding.
func TestDecoder_ReadTupleWithDynamicComponents(t *testing.T) {
	tests := []struct {
		name        string
		components  []*StructComponent
		hexData     string
		expectValue []interface{}
	}{
		{
			name: "tuple with bytes field",
			components: []*StructComponent{
				{Name: "signer", TypeName: "address"},
				{Name: "metadata", TypeName: "bytes"},
				{Name: "value", TypeName: "uint64"},
			},
			// This is the encoded form of (address, bytes, uint64) where:
			// - address = 0x1234567890123456789012345678901234567890
			// - bytes = 0xdeadbeef
			// - uint64 = 42
			hexData: "0x" +
				// address signer (inline)
				"0000000000000000000000001234567890123456789012345678901234567890" +
				// offset to bytes = 96 (0x60)
				"0000000000000000000000000000000000000000000000000000000000000060" +
				// uint64 value = 42
				"000000000000000000000000000000000000000000000000000000000000002a" +
				// bytes length = 4
				"0000000000000000000000000000000000000000000000000000000000000004" +
				// bytes data = 0xdeadbeef (padded)
				"deadbeef00000000000000000000000000000000000000000000000000000000",
			expectValue: []interface{}{
				MustNewAddress("1234567890123456789012345678901234567890"),
				[]byte{0xde, 0xad, 0xbe, 0xef},
				uint64(42),
			},
		},
		{
			name: "tuple with string field",
			components: []*StructComponent{
				{Name: "id", TypeName: "uint256"},
				{Name: "name", TypeName: "string"},
			},
			hexData: "0x" +
				// uint256 id = 123
				"000000000000000000000000000000000000000000000000000000000000007b" +
				// offset to string = 64 (0x40)
				"0000000000000000000000000000000000000000000000000000000000000040" +
				// string length = 5
				"0000000000000000000000000000000000000000000000000000000000000005" +
				// string data = "hello" (padded)
				"68656c6c6f000000000000000000000000000000000000000000000000000000",
			expectValue: []interface{}{
				bigString(t, "123"),
				"hello",
			},
		},
		{
			name: "tuple with multiple dynamic fields",
			components: []*StructComponent{
				{Name: "data1", TypeName: "bytes"},
				{Name: "value", TypeName: "uint64"},
				{Name: "data2", TypeName: "bytes"},
			},
			hexData: "0x" +
				// offset to data1 = 96 (0x60)
				"0000000000000000000000000000000000000000000000000000000000000060" +
				// uint64 value = 100
				"0000000000000000000000000000000000000000000000000000000000000064" +
				// offset to data2 = 160 (0xa0)
				"00000000000000000000000000000000000000000000000000000000000000a0" +
				// data1 length = 2
				"0000000000000000000000000000000000000000000000000000000000000002" +
				// data1 = 0x0102
				"0102000000000000000000000000000000000000000000000000000000000000" +
				// data2 length = 3
				"0000000000000000000000000000000000000000000000000000000000000003" +
				// data2 = 0x030405
				"0304050000000000000000000000000000000000000000000000000000000000",
			expectValue: []interface{}{
				[]byte{0x01, 0x02},
				uint64(100),
				[]byte{0x03, 0x04, 0x05},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dec, err := NewDecoderFromString(test.hexData)
			require.NoError(t, err)

			result, err := dec.readTuple(test.components)
			require.NoError(t, err)

			assert.Equal(t, len(test.expectValue), len(result))
			for i, expected := range test.expectValue {
				assert.Equal(t, expected, result[i], "mismatch at index %d", i)
			}
		})
	}
}

// TestEncoderDecoderRoundTrip_TupleWithDynamicComponents tests that encoding and
// then decoding a tuple with dynamic components produces the original values.
func TestEncoderDecoderRoundTrip_TupleWithDynamicComponents(t *testing.T) {
	components := []*StructComponent{
		{Name: "id", TypeName: "uint256"},
		{Name: "metadata", TypeName: "bytes"},
		{Name: "owner", TypeName: "address"},
		{Name: "name", TypeName: "string"},
	}

	originalValues := []interface{}{
		big.NewInt(12345),
		[]byte{0xca, 0xfe, 0xba, 0xbe},
		MustNewAddress("abcdefabcdefabcdefabcdefabcdefabcdefabcd"),
		"test name",
	}

	// Encode
	encoder := NewEncoder()
	err := encoder.write("tuple", components, originalValues)
	require.NoError(t, err)

	// Decode
	decoder := NewDecoder(encoder.Buffer())
	decoded, err := decoder.readTuple(components)
	require.NoError(t, err)

	// Compare
	require.Equal(t, len(originalValues), len(decoded))
	assert.Equal(t, originalValues[0].(*big.Int).Cmp(decoded[0].(*big.Int)), 0, "id mismatch")
	assert.Equal(t, originalValues[1], decoded[1], "metadata mismatch")
	assert.Equal(t, originalValues[2], decoded[2], "owner mismatch")
	assert.Equal(t, originalValues[3], decoded[3], "name mismatch")
}
