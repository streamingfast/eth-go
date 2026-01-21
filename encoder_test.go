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

func TestEncoder_Write(t *testing.T) {
	type period struct {
		TokenID         *big.Int
		TokenType       Uint8
		FromBlockNumber Uint64
		ToBlockNumber   Uint64
	}

	tests := []struct {
		name        string
		typeName    string
		components  []*StructComponent
		in          interface{}
		expectError bool
		expectHex   string
	}{
		{
			name:      "bool true",
			typeName:  "bool",
			in:        true,
			expectHex: "0000000000000000000000000000000000000000000000000000000000000001",
		},
		{
			name:      "bool false",
			typeName:  "bool",
			in:        false,
			expectHex: "0000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			name:      "uint8",
			typeName:  "uint8",
			in:        uint8(7),
			expectHex: "0000000000000000000000000000000000000000000000000000000000000007",
		},
		{
			name:      "uint24",
			typeName:  "uint24",
			in:        uint32(16777215),
			expectHex: "0000000000000000000000000000000000000000000000000000000000ffffff",
		},
		{
			name:      "uint56",
			typeName:  "uint56",
			in:        uint64(72057594037927933),
			expectHex: "00000000000000000000000000000000000000000000000000fffffffffffffd",
		},
		{
			name:      "uint64",
			typeName:  "uint64",
			in:        uint64(673),
			expectHex: "00000000000000000000000000000000000000000000000000000000000002a1",
		},
		{
			name:      "uint96",
			typeName:  "uint96",
			in:        bigString(t, "79228162514264337593543950335"),
			expectHex: "0000000000000000000000000000000000000000ffffffffffffffffffffffff",
		},
		{
			name:      "uint112",
			typeName:  "uint112",
			in:        bigString(t, "5192296858534827628530496329220095"),
			expectHex: "000000000000000000000000000000000000ffffffffffffffffffffffffffff",
		},
		{
			name:      "uint256",
			typeName:  "uint256",
			in:        big.NewInt(2938),
			expectHex: "0000000000000000000000000000000000000000000000000000000000000b7a",
		},
		{
			name:      "bigger uint256",
			typeName:  "uint256",
			in:        big.NewInt(2317850009133627),
			expectHex: "00000000000000000000000000000000000000000000000000083c12826fe23b",
		},
		{
			name:      "address",
			typeName:  "address",
			in:        MustNewAddress("7d97ba95dac25316b9531152b3baa32327994da8"),
			expectHex: "0000000000000000000000007d97ba95dac25316b9531152b3baa32327994da8",
		},
		{
			name:      "method",
			typeName:  "method",
			in:        "transfer(address,uint256)",
			expectHex: "a9059cbb",
		},
		{
			name:      "string",
			typeName:  "string",
			in:        "UniswapV2: LOCKED",
			expectHex: "0000000000000000000000000000000000000000000000000000000000000011556e697377617056323a204c4f434b4544000000000000000000000000000000",
		},
		{
			name:      "bytes",
			typeName:  "bytes",
			in:        []byte{0x01, 0x03, 0xaa, 0xbb, 0xcc},
			expectHex: "00000000000000000000000000000000000000000000000000000000000000050103aabbcc000000000000000000000000000000000000000000000000000000",
		},
		{
			name:      "bytes_empty",
			typeName:  "bytes",
			in:        []byte{},
			expectHex: "0000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			name:     "bytes_flush",
			typeName: "bytes",
			in: []byte{
				0x01, 0x03, 0xaa, 0xbb, 0xcc, 0xfe, 0x45, 0xab,
				0x02, 0x03, 0xaa, 0xbb, 0xcc, 0xfe, 0x45, 0xcd,
				0x03, 0x03, 0xaa, 0xbb, 0xcc, 0xfe, 0x45, 0xef,
				0x04, 0x03, 0xaa, 0xbb, 0xcc, 0xfe, 0x45, 0xff,
			},
			expectHex: "00000000000000000000000000000000000000000000000000000000000000200103aabbccfe45ab0203aabbccfe45cd0303aabbccfe45ef0403aabbccfe45ff",
		},
		{
			name:     "bytes_flush_plus_one",
			typeName: "bytes",
			in: []byte{
				0x01, 0x03, 0xaa, 0xbb, 0xcc, 0xfe, 0x45, 0xab,
				0x02, 0x03, 0xaa, 0xbb, 0xcc, 0xfe, 0x45, 0xcd,
				0x03, 0x03, 0xaa, 0xbb, 0xcc, 0xfe, 0x45, 0xef,
				0x04, 0x03, 0xaa, 0xbb, 0xcc, 0xfe, 0x45, 0xff,
				0x01,
			},
			expectHex: "00000000000000000000000000000000000000000000000000000000000000210103aabbccfe45ab0203aabbccfe45cd0303aabbccfe45ef0403aabbccfe45ff0100000000000000000000000000000000000000000000000000000000000000",
		},
		{
			name:     "tuple from interface slice",
			typeName: "tuple",
			components: []*StructComponent{
				{Name: "tokenID", TypeName: "uint256"},
				{Name: "tokenType", TypeName: "uint256"},
				{Name: "fromBlockNumber", TypeName: "uint64"},
				{Name: "toBlockNumber", TypeName: "uint64"},
			},
			in:        []interface{}{big.NewInt(1), big.NewInt(2), uint64(3), uint64(4)},
			expectHex: "0000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000004",
		},
		{
			name:     "tuple from interface map",
			typeName: "tuple",
			components: []*StructComponent{
				{Name: "tokenID", TypeName: "uint256"},
				{Name: "tokenType", TypeName: "uint256"},
				{Name: "fromBlockNumber", TypeName: "uint64"},
				{Name: "toBlockNumber", TypeName: "uint64"},
			},
			in: map[string]interface{}{
				"tokenID":         big.NewInt(1),
				"tokenType":       big.NewInt(2),
				"fromBlockNumber": uint64(3),
				"toBlockNumber":   uint64(4),
			},
			expectHex: "0000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000004",
		},
		{
			name:     "tuple from struct",
			typeName: "tuple",
			components: []*StructComponent{
				{Name: "tokenID", TypeName: "uint256"},
				{Name: "tokenType", TypeName: "uint8"},
				{Name: "fromBlockNumber", TypeName: "uint64"},
				{Name: "toBlockNumber", TypeName: "uint64"},
			},
			in: period{
				TokenID:         big.NewInt(1),
				TokenType:       Uint8(2),
				FromBlockNumber: Uint64(3),
				ToBlockNumber:   Uint64(4),
			},
			expectHex: "0000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000004",
		},
		{
			name:     "tuple from struct ptr",
			typeName: "tuple",
			components: []*StructComponent{
				{Name: "tokenID", TypeName: "uint256"},
				{Name: "tokenType", TypeName: "uint8"},
				{Name: "fromBlockNumber", TypeName: "uint64"},
				{Name: "toBlockNumber", TypeName: "uint64"},
			},
			in: &period{
				TokenID:         big.NewInt(1),
				TokenType:       Uint8(2),
				FromBlockNumber: Uint64(3),
				ToBlockNumber:   Uint64(4),
			},
			expectHex: "0000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000004",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := NewEncoder()
			err := e.write(test.typeName, test.components, test.in)
			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectHex, hex.EncodeToString(e.buffer))
			}
		})
	}
}

func TestEncoder_WriteArray(t *testing.T) {
	tests := []struct {
		name        string
		typeName    string
		components  []*StructComponent
		in          interface{}
		expectError bool
		expectHex   string
	}{
		{
			name:      "bool",
			typeName:  "bool[]",
			in:        []bool{true, false, true},
			expectHex: "0000000000000000000000000000000000000000000000000000000000000003000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001",
		},
		{
			name:      "uint8",
			typeName:  "uint8[]",
			in:        []uint8{uint8(7), uint8(1), uint8(2)},
			expectHex: "0000000000000000000000000000000000000000000000000000000000000003000000000000000000000000000000000000000000000000000000000000000700000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000002",
		},
		{
			name:      "uint24",
			typeName:  "uint24[]",
			in:        []uint32{3, 7, 5},
			expectHex: "0000000000000000000000000000000000000000000000000000000000000003000000000000000000000000000000000000000000000000000000000000000300000000000000000000000000000000000000000000000000000000000000070000000000000000000000000000000000000000000000000000000000000005",
		},
		{
			name:      "uint56",
			typeName:  "uint56[]",
			in:        []uint64{72057594037927933, 72057594037927933, 72057594037927933},
			expectHex: "000000000000000000000000000000000000000000000000000000000000000300000000000000000000000000000000000000000000000000fffffffffffffd00000000000000000000000000000000000000000000000000fffffffffffffd00000000000000000000000000000000000000000000000000fffffffffffffd",
		},
		{
			name:      "uint64",
			typeName:  "uint64[]",
			in:        []uint64{673, 10, 9},
			expectHex: "000000000000000000000000000000000000000000000000000000000000000300000000000000000000000000000000000000000000000000000000000002a1000000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000009",
		},
		{
			name:     "uint96",
			typeName: "uint96[]",
			in: []*big.Int{
				bigString(t, "79228162514264337593543950335"),
				bigString(t, "3"),
			},
			expectHex: "00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000ffffffffffffffffffffffff0000000000000000000000000000000000000000000000000000000000000003",
		},
		{
			name:     "uint112",
			typeName: "uint112[]",
			in: []*big.Int{
				bigString(t, "5"),
				bigString(t, "3"),
			},
			expectHex: "000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000050000000000000000000000000000000000000000000000000000000000000003",
		},
		{
			name:      "uint256",
			typeName:  "uint256[]",
			in:        []*big.Int{big.NewInt(2938)},
			expectHex: "00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000b7a",
		},
		{
			name:     "address",
			typeName: "address[]",
			in: []Address{
				MustNewAddress("7d97ba95dac25316b9531152b3baa32327994da8"),
				MustNewAddress("c778417e063141139fce010982780140aa0cd5ab"),
			},
			expectHex: "00000000000000000000000000000000000000000000000000000000000000020000000000000000000000007d97ba95dac25316b9531152b3baa32327994da8000000000000000000000000c778417e063141139fce010982780140aa0cd5ab",
		},
		{
			name:      "string",
			typeName:  "string[]",
			in:        []string{"UniswapV2: LOCKED", "hello world"},
			expectHex: "00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000011556e697377617056323a204c4f434b4544000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000b68656c6c6f20776f726c64000000000000000000000000000000000000000000",
		},
		{
			name:     "tuple",
			typeName: "tuple[]",
			components: []*StructComponent{
				{TypeName: "uint256"},
				{TypeName: "uint64"},
				{TypeName: "uint64"},
			},
			in:        []interface{}{[]interface{}{big.NewInt(0x1AB), uint64(0x1CD), uint64(0x1EF)}},
			expectHex: "000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000001ab00000000000000000000000000000000000000000000000000000000000001cd00000000000000000000000000000000000000000000000000000000000001ef",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := NewEncoder()
			err := e.write(test.typeName, test.components, test.in)
			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectHex, hex.EncodeToString(e.buffer))
			}
		})
	}
}

func TestEncoder_WriteMethodCall(t *testing.T) {
	method := &MethodCall{
		MethodDef: &MethodDef{
			Name: "swapExactTokensForTokens",
			Parameters: []*MethodParameter{
				{TypeName: "uint256"},
				{TypeName: "uint256"},
				{TypeName: "address[]"},
				{TypeName: "address"},
				{TypeName: "uint256"},
			},
		},
		Data: []interface{}{
			big.NewInt(100000000000000),
			big.NewInt(2317850009133627),
			[]Address{
				MustNewAddress("d24af825e38495ee362466f214946cdf53aab8c8"), // JOHNY
				MustNewAddress("c778417e063141139fce010982780140aa0cd5ab"), // WETH
				MustNewAddress("7d97ba95dac25316b9531152b3baa32327994da8"), // STEPD
			},
			MustNewAddress("40c7f627ffb69b8d8752c518f8790b04a523bee5"),
			big.NewInt(1600958277),
		},
	}

	e := Encoder{}
	err := e.WriteMethodCall(method)
	require.NoError(t, err)
	expectHex := "38ed1739" +
		"00000000000000000000000000000000000000000000000000005af3107a4000" +
		"00000000000000000000000000000000000000000000000000083c12826fe23b" +
		"00000000000000000000000000000000000000000000000000000000000000a0" +
		"00000000000000000000000040c7f627ffb69b8d8752c518f8790b04a523bee5" +
		"000000000000000000000000000000000000000000000000000000005f6caf45" +
		"0000000000000000000000000000000000000000000000000000000000000003" +
		"000000000000000000000000d24af825e38495ee362466f214946cdf53aab8c8" +
		"000000000000000000000000c778417e063141139fce010982780140aa0cd5ab" +
		"0000000000000000000000007d97ba95dac25316b9531152b3baa32327994da8"
	assert.Equal(t, expectHex, hex.EncodeToString(e.buffer))
}

func TestEncoder_WriteTuple(t *testing.T) {
	method := &MethodCall{
		MethodDef: &MethodDef{
			Name: "getOnePeriod",
			Parameters: []*MethodParameter{
				{Name: "period", TypeName: "tuple", InternalType: "struct Period", TypeMutability: "", Components: []*StructComponent{
					{Name: "tokenID", TypeName: "uint256", InternalType: "uint256"},
					{Name: "fromBlockNum", TypeName: "uint64", InternalType: "uint64"},
					{Name: "toBlockNum", TypeName: "uint64", InternalType: "uint64"},
				}},
			},
		},
		Data: []interface{}{
			[]interface{}{
				big.NewInt(0xAB),
				uint64(0xBC),
				uint64(0x9F),
			},
		},
	}

	e := Encoder{}
	err := e.WriteMethodCall(method)
	require.NoError(t, err)
	expectHex := "7d165091" +
		"00000000000000000000000000000000000000000000000000000000000000ab" +
		"00000000000000000000000000000000000000000000000000000000000000bc" +
		"000000000000000000000000000000000000000000000000000000000000009f"
	assert.Equal(t, expectHex, hex.EncodeToString(e.buffer))
}

func TestEncoder_WriteTupleArray(t *testing.T) {
	method := &MethodCall{
		MethodDef: &MethodDef{
			Name: "tupleArray",
			Parameters: []*MethodParameter{
				{Name: "periods", TypeName: "tuple[]", InternalType: "struct ClaimPeriod[]", TypeMutability: "", Components: []*StructComponent{
					{Name: "tokenID", TypeName: "uint256", InternalType: "uint256"},
					{Name: "fromBlockNum", TypeName: "uint64", InternalType: "uint64"},
					{Name: "toBlockNum", TypeName: "uint64", InternalType: "uint64"},
				}},
			},
		},
		Data: []interface{}{
			[]interface{}{
				[]interface{}{
					big.NewInt(0x1AB),
					uint64(0x1CD),
					uint64(0x1EF),
				},
			},
		},
	}

	e := Encoder{}
	err := e.WriteMethodCall(method)
	require.NoError(t, err)
	expectHex := "15eb963d" +
		"0000000000000000000000000000000000000000000000000000000000000020" +
		"0000000000000000000000000000000000000000000000000000000000000001" +
		"00000000000000000000000000000000000000000000000000000000000001ab" +
		"00000000000000000000000000000000000000000000000000000000000001cd" +
		"00000000000000000000000000000000000000000000000000000000000001ef"
	assert.Equal(t, expectHex, hex.EncodeToString(e.buffer))
}

func TestEncoder_isArray(t *testing.T) {
	b, typeName := isArray("address[]")
	assert.Equal(t, true, b)
	assert.Equal(t, "address", typeName)

	b, typeName = isArray("address")
	assert.Equal(t, false, b)
	assert.Equal(t, "address", typeName)
}

func TestEncoder_override(t *testing.T) {
	tests := []struct {
		name        string
		buf         []byte
		offset      uint64
		data        []byte
		expectError bool
		expectBytes []byte
	}{
		{
			name:        "golden path",
			buf:         []byte{0xaa, 0x00, 0xbb},
			offset:      1,
			data:        []byte{0xcc},
			expectError: false,
			expectBytes: []byte{0xaa, 0xcc, 0xbb},
		},
		{
			name:        "overlaps with non-zero data",
			buf:         []byte{0xaa, 0x00, 0xbb},
			offset:      1,
			data:        []byte{0xcc, 0xdd},
			expectError: false,
			expectBytes: []byte{0xaa, 0xcc, 0xdd},
		},
		{
			name:        "insufficient room in buffer",
			buf:         []byte{0xaa, 0x00, 0xbb},
			offset:      1,
			data:        []byte{0xcc, 0xdd, 0xee},
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := &Encoder{buffer: test.buf}
			err := e.override(test.offset, test.data)
			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectBytes, e.buffer)
			}
		})
	}
}

// TestEncoder_WriteTupleWithDynamicComponents tests encoding tuples that contain
// dynamic types like `bytes` or `string`. Per Solidity ABI spec, such tuples are
// considered dynamic and require offset-based encoding.
//
// These tests use the EXACT SAME values as the Solidity tests in tests/src/Codec.sol
// (emitEventUTupleWithBytes, emitEventUTupleWithString, emitEventUTupleWithMultipleDynamic)
// so that the expected bytes can be directly compared with Solidity's output.
//
// The expected bytes match the event data from Solidity (without the outer 0x20 offset
// that wraps the tuple when it's a top-level event parameter).
func TestEncoder_WriteTupleWithDynamicComponents(t *testing.T) {
	tests := []struct {
		name       string
		components []*StructComponent
		in         interface{}
		expectHex  string
	}{
		{
			// Matches Solidity: emitEventUTupleWithBytes() in tests/src/Codec.sol
			// struct TupleWithBytes { address signer; bytes metadata; uint64 value; }
			// Values: (0xdB0De9288CF0713De91371969efCC9969dd94117, 0xdeadbeef, 42)
			name: "TupleWithBytes - matches Solidity emitEventUTupleWithBytes",
			components: []*StructComponent{
				{Name: "signer", TypeName: "address"},
				{Name: "metadata", TypeName: "bytes"},
				{Name: "value", TypeName: "uint64"},
			},
			in: []interface{}{
				MustNewAddress("dB0De9288CF0713De91371969efCC9969dd94117"),
				[]byte{0xde, 0xad, 0xbe, 0xef},
				uint64(42),
			},
			// address signer | offset to bytes (0x60) | uint64 value (42) | bytes length (4) | bytes data (0xdeadbeef)
			expectHex: "000000000000000000000000db0de9288cf0713de91371969efcc9969dd94117" +
				"0000000000000000000000000000000000000000000000000000000000000060" +
				"000000000000000000000000000000000000000000000000000000000000002a" +
				"0000000000000000000000000000000000000000000000000000000000000004" +
				"deadbeef00000000000000000000000000000000000000000000000000000000",
		},
		{
			// Matches Solidity: emitEventUTupleWithString() in tests/src/Codec.sol
			// struct TupleWithString { uint256 id; string name; }
			// Values: (123, "hello")
			name: "TupleWithString - matches Solidity emitEventUTupleWithString",
			components: []*StructComponent{
				{Name: "id", TypeName: "uint256"},
				{Name: "name", TypeName: "string"},
			},
			in: []interface{}{
				big.NewInt(123),
				"hello",
			},
			// uint256 id (123) | offset to string (0x40) | string length (5) | string data ("hello")
			expectHex: "000000000000000000000000000000000000000000000000000000000000007b" +
				"0000000000000000000000000000000000000000000000000000000000000040" +
				"0000000000000000000000000000000000000000000000000000000000000005" +
				"68656c6c6f000000000000000000000000000000000000000000000000000000",
		},
		{
			// Matches Solidity: emitEventUTupleWithMultipleDynamic() in tests/src/Codec.sol
			// struct TupleWithMultipleDynamic { bytes data1; uint64 value; bytes data2; }
			// Values: (0x0102, 100, 0x030405)
			name: "TupleWithMultipleDynamic - matches Solidity emitEventUTupleWithMultipleDynamic",
			components: []*StructComponent{
				{Name: "data1", TypeName: "bytes"},
				{Name: "value", TypeName: "uint64"},
				{Name: "data2", TypeName: "bytes"},
			},
			in: []interface{}{
				[]byte{0x01, 0x02},
				uint64(100),
				[]byte{0x03, 0x04, 0x05},
			},
			// offset to data1 (0x60) | uint64 value (100) | offset to data2 (0xa0) | data1 length (2) | data1 | data2 length (3) | data2
			expectHex: "0000000000000000000000000000000000000000000000000000000000000060" +
				"0000000000000000000000000000000000000000000000000000000000000064" +
				"00000000000000000000000000000000000000000000000000000000000000a0" +
				"0000000000000000000000000000000000000000000000000000000000000002" +
				"0102000000000000000000000000000000000000000000000000000000000000" +
				"0000000000000000000000000000000000000000000000000000000000000003" +
				"0304050000000000000000000000000000000000000000000000000000000000",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := NewEncoder()
			err := e.write("tuple", test.components, test.in)
			require.NoError(t, err)
			assert.Equal(t, test.expectHex, hex.EncodeToString(e.buffer))
		})
	}
}

func TestIsDynamicType(t *testing.T) {
	tests := []struct {
		name       string
		typeName   string
		components []*StructComponent
		expected   bool
	}{
		{"bytes is dynamic", "bytes", nil, true},
		{"string is dynamic", "string", nil, true},
		{"uint256 is static", "uint256", nil, false},
		{"address is static", "address", nil, false},
		{"uint256[] is dynamic", "uint256[]", nil, true},
		{"address[] is dynamic", "address[]", nil, true},
		{"tuple with all static is static", "tuple", []*StructComponent{
			{Name: "a", TypeName: "uint256"},
			{Name: "b", TypeName: "address"},
		}, false},
		{"tuple with bytes is dynamic", "tuple", []*StructComponent{
			{Name: "a", TypeName: "uint256"},
			{Name: "b", TypeName: "bytes"},
		}, true},
		{"tuple with string is dynamic", "tuple", []*StructComponent{
			{Name: "a", TypeName: "string"},
		}, true},
		{"tuple with array is dynamic", "tuple", []*StructComponent{
			{Name: "a", TypeName: "uint256[]"},
		}, true},
		{"tuple[] is always dynamic", "tuple[]", []*StructComponent{
			{Name: "a", TypeName: "uint256"},
		}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := isDynamicType(test.typeName, test.components)
			assert.Equal(t, test.expected, result)
		})
	}
}

// TestEncoder_WriteNestedTupleFromParsedABI tests encoding nested tuples using
// an ABI that was parsed from JSON. This exercises the full code path including
// the `toStructComponents` function which must properly recursively parse nested
// component definitions.
//
// This test reproduces the bug reported where nested tuple encoding fails with:
//
//	"struct has 0 fields"
//
// The root cause is that `toStructComponents` wasn't recursively processing nested
// components, so the inner tuple's component definitions were not populated.
func TestEncoder_WriteNestedTupleFromParsedABI(t *testing.T) {
	// Parse the ABI containing nested tuple functions
	abi, err := ParseABI("testdata/nested_tuple.abi.json")
	require.NoError(t, err)

	// Test funTupleWithNestedDynamic: (address, (uint256, string), uint64)
	t.Run("funTupleWithNestedDynamic with map input", func(t *testing.T) {
		fn := abi.FindFunctionByName("funTupleWithNestedDynamic")
		require.NotNil(t, fn, "function funTupleWithNestedDynamic should exist in ABI")

		// Verify that the inner tuple components were properly parsed
		require.Len(t, fn.Parameters, 1, "should have 1 parameter")
		require.Equal(t, "tuple", fn.Parameters[0].TypeName)
		require.Len(t, fn.Parameters[0].Components, 3, "outer tuple should have 3 components")
		require.Equal(t, "tuple", fn.Parameters[0].Components[1].TypeName, "second component should be a tuple")
		require.Len(t, fn.Parameters[0].Components[1].Components, 2,
			"inner tuple should have 2 components - this will fail if nested components weren't parsed")

		// Now test encoding with map[string]interface{}
		input := map[string]interface{}{
			"owner": MustNewAddress("aAaAaAaaAaAaAaaAaAAAAAAAAaaaAaAaAaaAaaAa"),
			"inner": map[string]interface{}{
				"id":          big.NewInt(456),
				"description": "nested",
			},
			"timestamp": uint64(789),
		}

		call := fn.NewCall(input)
		data, err := call.Encode()
		require.NoError(t, err, "encoding should succeed")

		// Expected encoding uses the function's computed method ID + tuple data
		// The tuple data matches Solidity's encoding exactly
		expectedDataHex :=
			"0000000000000000000000000000000000000000000000000000000000000020" + // offset to tuple
				"000000000000000000000000aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" + // address owner
				"0000000000000000000000000000000000000000000000000000000000000060" + // offset to inner tuple
				"0000000000000000000000000000000000000000000000000000000000000315" + // uint64 timestamp = 789
				"00000000000000000000000000000000000000000000000000000000000001c8" + // uint256 id = 456
				"0000000000000000000000000000000000000000000000000000000000000040" + // offset to string
				"0000000000000000000000000000000000000000000000000000000000000006" + // string length = 6
				"6e65737465640000000000000000000000000000000000000000000000000000" // "nested"

		// Verify method selector prefix and data
		assert.Equal(t, fn.MethodID(), data[:4], "method selector should match")
		assert.Equal(t, expectedDataHex, hex.EncodeToString(data[4:]), "tuple data should match Solidity encoding")
	})

	t.Run("funTupleWithNestedDynamic with slice input", func(t *testing.T) {
		fn := abi.FindFunctionByName("funTupleWithNestedDynamic")
		require.NotNil(t, fn)

		// Test encoding with []interface{}
		input := []interface{}{
			MustNewAddress("aAaAaAaaAaAaAaaAaAAAAAAAAaaaAaAaAaaAaaAa"),
			[]interface{}{
				big.NewInt(456),
				"nested",
			},
			uint64(789),
		}

		call := fn.NewCall(input)
		data, err := call.Encode()
		require.NoError(t, err, "encoding should succeed")

		expectedDataHex :=
			"0000000000000000000000000000000000000000000000000000000000000020" +
				"000000000000000000000000aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
				"0000000000000000000000000000000000000000000000000000000000000060" +
				"0000000000000000000000000000000000000000000000000000000000000315" +
				"00000000000000000000000000000000000000000000000000000000000001c8" +
				"0000000000000000000000000000000000000000000000000000000000000040" +
				"0000000000000000000000000000000000000000000000000000000000000006" +
				"6e65737465640000000000000000000000000000000000000000000000000000"

		assert.Equal(t, fn.MethodID(), data[:4], "method selector should match")
		assert.Equal(t, expectedDataHex, hex.EncodeToString(data[4:]), "tuple data should match Solidity encoding")
	})

	// Test recoverRAVSigner: original bug report case
	// SignedRAV: (RAV rav, bytes signature)
	// RAV: (bytes32 collectionId, address payer, address serviceProvider, address dataService, uint64 timestampNs, uint128 valueAggregate, bytes metadata)
	t.Run("recoverRAVSigner nested tuple encoding", func(t *testing.T) {
		fn := abi.FindFunctionByName("recoverRAVSigner")
		require.NotNil(t, fn, "function recoverRAVSigner should exist in ABI")

		// Verify nested components were parsed
		require.Len(t, fn.Parameters, 1)
		require.Equal(t, "tuple", fn.Parameters[0].TypeName)
		require.Len(t, fn.Parameters[0].Components, 2, "SignedRAV should have 2 components")
		require.Equal(t, "tuple", fn.Parameters[0].Components[0].TypeName, "first component (rav) should be a tuple")
		require.Len(t, fn.Parameters[0].Components[0].Components, 7,
			"RAV tuple should have 7 components - this will fail if nested components weren't parsed")

		// Create test data matching the original bug report
		collectionIDBytes := make([]byte, 32)
		collectionIDBytes[31] = 0x01 // Just set last byte to 1

		payerAddr := MustNewAddress("1111111111111111111111111111111111111111")
		spAddr := MustNewAddress("2222222222222222222222222222222222222222")
		dsAddr := MustNewAddress("3333333333333333333333333333333333333333")

		ravTuple := map[string]interface{}{
			"collectionId":    collectionIDBytes,
			"payer":           payerAddr,
			"serviceProvider": spAddr,
			"dataService":     dsAddr,
			"timestampNs":     uint64(123),
			"valueAggregate":  big.NewInt(1000),
			"metadata":        []byte{},
		}

		signedRAVTuple := map[string]interface{}{
			"rav":       ravTuple,
			"signature": []byte{0xaa, 0xbb, 0xcc}, // dummy signature
		}

		call := fn.NewCall(signedRAVTuple)
		data, err := call.Encode()
		require.NoError(t, err, "encoding should succeed - original bug caused 'struct has 0 fields' error here")

		// Just verify it encoded something reasonable (method selector + some data)
		require.True(t, len(data) > 4, "should have encoded data beyond method selector")
	})
}

// TestEncoder_WriteNestedTuple tests encoding tuples that contain nested tuples.
// This matches the Solidity tests in tests/src/test/Codec.sol for:
// - testFunTupleWithNestedDynamic
// - testEmitEventUTupleWithNestedDynamic
//
// The structs are:
//
//	struct InnerDynamic {
//	    uint256 id;
//	    string description;
//	}
//
//	struct TupleWithNestedDynamic {
//	    address owner;
//	    InnerDynamic inner;
//	    uint64 timestamp;
//	}
//
// Values: (0xaAaAaAaaAaAaAaaAaAAAAAAAAaaaAaAaAaaAaaAa, (456, "nested"), 789)
func TestEncoder_WriteNestedTuple(t *testing.T) {
	// Define components for the inner tuple (InnerDynamic)
	innerComponents := []*StructComponent{
		{Name: "id", TypeName: "uint256"},
		{Name: "description", TypeName: "string"},
	}

	// Define components for the outer tuple (TupleWithNestedDynamic)
	outerComponents := []*StructComponent{
		{Name: "owner", TypeName: "address"},
		{Name: "inner", TypeName: "tuple", Components: innerComponents},
		{Name: "timestamp", TypeName: "uint64"},
	}

	tests := []struct {
		name      string
		in        interface{}
		expectHex string
	}{
		{
			name: "nested tuple using []interface{}",
			in: []interface{}{
				MustNewAddress("aAaAaAaaAaAaAaaAaAAAAAAAAaaaAaAaAaaAaaAa"),
				[]interface{}{
					big.NewInt(456),
					"nested",
				},
				uint64(789),
			},
			// address owner | offset to inner tuple (0x60) | uint64 timestamp (789) |
			// uint256 id (456) | offset to string (0x40) | string length (6) | string data ("nested")
			expectHex: "000000000000000000000000aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
				"0000000000000000000000000000000000000000000000000000000000000060" +
				"0000000000000000000000000000000000000000000000000000000000000315" +
				"00000000000000000000000000000000000000000000000000000000000001c8" +
				"0000000000000000000000000000000000000000000000000000000000000040" +
				"0000000000000000000000000000000000000000000000000000000000000006" +
				"6e65737465640000000000000000000000000000000000000000000000000000",
		},
		{
			name: "nested tuple using map[string]interface{}",
			in: map[string]interface{}{
				"owner": MustNewAddress("aAaAaAaaAaAaAaaAaAAAAAAAAaaaAaAaAaaAaaAa"),
				"inner": map[string]interface{}{
					"id":          big.NewInt(456),
					"description": "nested",
				},
				"timestamp": uint64(789),
			},
			expectHex: "000000000000000000000000aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
				"0000000000000000000000000000000000000000000000000000000000000060" +
				"0000000000000000000000000000000000000000000000000000000000000315" +
				"00000000000000000000000000000000000000000000000000000000000001c8" +
				"0000000000000000000000000000000000000000000000000000000000000040" +
				"0000000000000000000000000000000000000000000000000000000000000006" +
				"6e65737465640000000000000000000000000000000000000000000000000000",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := NewEncoder()
			err := e.write("tuple", outerComponents, test.in)
			require.NoError(t, err)
			assert.Equal(t, test.expectHex, hex.EncodeToString(e.buffer))
		})
	}
}
