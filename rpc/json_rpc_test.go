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
	"fmt"
	"math"
	"math/big"
	"strconv"
	"testing"

	"github.com/streamingfast/eth-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalJSONRPC(t *testing.T) {
	tests := []struct {
		name        string
		in          interface{}
		expected    string
		expectedErr error
	}{
		{"int8 1", int8(1), `"0x1"`, nil},
		{"int16 1", int16(1), `"0x1"`, nil},
		{"int32 1", int32(1), `"0x1"`, nil},
		{"int64 1", int64(1), `"0x1"`, nil},
		{"int 1", int(1), `"0x1"`, nil},

		{"uint8 1", uint8(1), `"0x1"`, nil},
		{"uint16 1", uint16(1), `"0x1"`, nil},
		{"uint32 1", uint32(1), `"0x1"`, nil},
		{"uint64 1", uint64(1), `"0x1"`, nil},
		{"uint 1", uint(1), `"0x1"`, nil},

		{"big.Int", *big.NewInt(1), `"0x1"`, nil},
		{"*big.Int", big.NewInt(1), `"0x1"`, nil},

		// Some wrapping tests
		{"wrapped no marshal RPC", wrappedUint64NoMarshalRPC(1), `"0x1"`, nil},
		{"wrapped no marshal RPC, custom UnmarshalJSON", wrappedUint64CustomUnmarshal(1), `"0x1"`, nil},

		// Quantity vs Data zero handling (quantities should be 0x0, data should be 0x)
		{"quantity int8 0", int8(0), `"0x0"`, nil},
		{"quantity int16 0", int16(0), `"0x0"`, nil},
		{"quantity int32 0", int32(0), `"0x0"`, nil},
		{"quantity int64 0", int64(0), `"0x0"`, nil},
		{"quantity int 0", int(0), `"0x0"`, nil},

		{"quantity uint8 0", uint8(0), `"0x0"`, nil},
		{"quantity uint16 0", uint16(0), `"0x0"`, nil},
		{"quantity uint32 0", uint32(0), `"0x0"`, nil},
		{"quantity uint64 0", uint64(0), `"0x0"`, nil},
		{"quantity uint 0", uint(0), `"0x0"`, nil},

		{"quantity bigInt nil", (*big.Int)(nil), `"0x0"`, nil},
		{"quantity bigInt 0", big.Int{}, `"0x0"`, nil},

		{"quantity trims zero prefix", int(15), `"0xf"`, nil},

		// Extremes, which fill the fixed-size buffer the quantity writers render into.
		{"int64 min", int64(math.MinInt64), `"-0x8000000000000000"`, nil},
		{"int64 max", int64(math.MaxInt64), `"0x7fffffffffffffff"`, nil},
		{"uint64 max", uint64(math.MaxUint64), `"0xffffffffffffffff"`, nil},
		{"big.Int wider than uint64", new(big.Int).Lsh(big.NewInt(1), 200), `"0x100000000000000000000000000000000000000000000000000"`, nil},

		// Negative quantities take go-ethereum's "-0x" spelling. The JSON-RPC spec has no
		// negative QUANTITY, so this only settles which form a caller passing one gets.
		{"negative int", int(-255), `"-0xff"`, nil},
		{"negative big.Int", big.NewInt(-255), `"-0xff"`, nil},
		{"negative big.Int wider than uint64", new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), 200)), `"-0x100000000000000000000000000000000000000000000000000"`, nil},

		{"data []byte nil", ([]byte)(nil), `"0x"`, nil},
		{"data []byte empty", []byte{}, `"0x"`, nil},
		{"data []byte 0x00", []byte{0x00}, `"0x00"`, nil},

		{"*eth.MethodCall", eth.MustNewMethodDef("name()").NewCall(), `"0x06fdde03"`, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := MarshalJSONRPC(test.in)
			if test.expectedErr == nil {
				require.NoError(t, err)
				assert.JSONEq(t, test.expected, string(actual))
			} else {
				assert.Equal(t, test.expectedErr, err)
			}
		})
	}
}

type wrappedUint64NoMarshalRPC uint64

type wrappedUint64CustomUnmarshal uint64

func (u *wrappedUint64CustomUnmarshal) UnmarshalJSON(data []byte) error {
	if len(data) < 2 {
		return fmt.Errorf("invalid input: %s", string(data))
	}

	value, err := strconv.ParseUint(string(data[1:len(data)-1]), 10, 64)
	if err != nil {
		return err
	}

	*u = wrappedUint64CustomUnmarshal(value)
	return nil
}

// namedUint64 has no JSON methods at all, standing in for a quantity type defined outside
// this module.
type namedUint64 uint64

// namedUint64WithMarshalJSON spells out its own JSON form, which must be preserved.
type namedUint64WithMarshalJSON uint64

func (u namedUint64WithMarshalJSON) MarshalJSON() ([]byte, error) {
	return []byte(`"from-marshal-json"`), nil
}

// namedUint64WithMarshalText likewise, one level down in precedence.
type namedUint64WithMarshalText uint64

func (u namedUint64WithMarshalText) MarshalText() ([]byte, error) {
	return []byte("from-marshal-text"), nil
}

// namedUint64WithBoth exposes a JSON-RPC form on top of a plain JSON one; the JSON-RPC
// form wins.
type namedUint64WithBoth uint64

func (u namedUint64WithBoth) MarshalJSON() ([]byte, error) { return []byte(`"from-marshal-json"`), nil }
func (u namedUint64WithBoth) MarshalJSONRPC() ([]byte, error) {
	return []byte(`"from-marshal-jsonrpc"`), nil
}

func TestMarshalJSONRPCPrecedence(t *testing.T) {
	tests := []struct {
		name     string
		in       any
		expected string
	}{
		{"no json methods, encoded as a quantity", namedUint64(255), `"0xff"`},
		{"MarshalJSON wins over the quantity rule", namedUint64WithMarshalJSON(255), `"from-marshal-json"`},
		{"MarshalText wins over the quantity rule", namedUint64WithMarshalText(255), `"from-marshal-text"`},
		{"MarshalJSONRPC wins over MarshalJSON", namedUint64WithBoth(255), `"from-marshal-jsonrpc"`},

		// Nested, since a value reaches the marshalers differently at the root.
		{"nested", struct {
			A namedUint64                `json:"a"`
			B namedUint64WithMarshalJSON `json:"b"`
			C namedUint64WithMarshalText `json:"c"`
			D namedUint64WithBoth        `json:"d"`
		}{255, 255, 255, 255}, `{"a":"0xff","b":"from-marshal-json","c":"from-marshal-text","d":"from-marshal-jsonrpc"}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := MarshalJSONRPC(test.in)
			require.NoError(t, err)
			assert.Equal(t, test.expected, string(actual))
		})
	}
}

// TestFastPathsMatchMarshalJSONRPC guards the allocation-free fast paths registered for the
// eth byte-slice types: they bypass MarshalJSONRPC, so they must produce the same bytes it
// would have.
func TestFastPathsMatchMarshalJSONRPC(t *testing.T) {
	for _, payload := range [][]byte{nil, {}, {0x00}, {0xde, 0xad, 0xbe, 0xef}} {
		values := []MarshalerRPC{
			eth.Bytes(payload),
			eth.Hex(payload),
			eth.Hash(payload),
			eth.Address(payload),
		}

		for _, value := range values {
			t.Run(fmt.Sprintf("%T/%x", value, payload), func(t *testing.T) {
				expected, err := value.MarshalJSONRPC()
				require.NoError(t, err)

				actual, err := MarshalJSONRPC(value)
				require.NoError(t, err)

				assert.Equal(t, string(expected), string(actual))
			})
		}
	}
}
