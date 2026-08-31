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
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddress_New(t *testing.T) {
	testNewStrict(t, 20, func(in string) (fmt.Stringer, error) { return NewAddress(in) })
}

func TestAddressLoose_New(t *testing.T) {
	testNew(t, func(in string) (fmt.Stringer, error) { return NewAddressLoose(in) })
}

func TestHex_New(t *testing.T) {
	testNew(t, func(in string) (fmt.Stringer, error) { return NewHex(in) })
}

func TestHash_New(t *testing.T) {
	testNew(t, func(in string) (fmt.Stringer, error) { return NewHash(in) })
}

func testNew(t *testing.T, new func(in string) (fmt.Stringer, error)) {
	tests := []struct {
		name        string
		in          string
		expected    string
		expectedErr error
	}{
		{"standard", "0xab", "ab", nil},
		{"odd length", "0xa", "0a", nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, err := new(test.in)
			if test.expectedErr == nil {
				require.NoError(t, err)
				assert.Equal(t, test.expected, value.String())
			} else {
				assert.Equal(t, test.expectedErr, err)
			}
		})
	}
}

func testNewStrict(t *testing.T, strictLength int, new func(in string) (fmt.Stringer, error)) {
	repeat := func(in string, times int) string {
		out := make([]byte, len(in)*times)
		for offset := 0; offset < len(in)*times; offset += len(in) {
			copy(out[offset:offset+len(in)], []byte(in))
		}

		return string(out)
	}

	fillerBelow := repeat("00", strictLength-1)
	fillerAbove := repeat("00", strictLength+1)

	tests := []struct {
		name        string
		in          string
		expected    string
		expectedErr error
	}{
		{"standard", "0xab" + fillerBelow, "ab" + fillerBelow, nil},
		{"odd length", "0xa" + fillerBelow, "0a" + fillerBelow, nil},
		{"bytes too low, prefixed", "0x" + fillerBelow, "", errors.New("invalid length in bytes, wanted 20 bytes once decoded but decoding gave us 19 bytes")},
		{"bytes too low, unprefixed", fillerBelow, "", errors.New("invalid length in bytes, wanted 20 bytes once decoded but decoding gave us 19 bytes")},
		{"bytes too high, prefixed", "0x" + fillerAbove, "", errors.New("invalid length in bytes, wanted 20 bytes once decoded but decoding gave us 21 bytes")},
		{"bytes too high, unprefixed", fillerAbove, "", errors.New("invalid length in bytes, wanted 20 bytes once decoded but decoding gave us 21 bytes")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, err := new(test.in)
			if test.expectedErr == nil {
				require.NoError(t, err)
				assert.Equal(t, test.expected, value.String())
			} else {
				assert.Equal(t, test.expectedErr, err)
			}
		})
	}
}

func TestAddress_Pretty(t *testing.T) {
	testPretty(t, func(in []byte) string { return Address(in).Pretty() })
}

func TestHash_Pretty(t *testing.T) {
	testPretty(t, func(in []byte) string { return Hash(in).Pretty() })
}

func TestHex_Pretty(t *testing.T) {
	testPretty(t, func(in []byte) string { return Hex(in).Pretty() })
}

func testPretty(t *testing.T, pretty func(in []byte) string) {
	tests := []struct {
		name     string
		in       []byte
		expected string
	}{
		{"standard", []byte{0xab}, "0xab"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, pretty(test.in))
		})
	}
}

func TestAddress_UnmarshalJSON(t *testing.T) {
	testUnmarshalJSON(t, func(in []byte) (fmt.Stringer, error) {
		var out Hex
		return out, json.Unmarshal(in, &out)
	})
}

func TestHash_UnmarshalJSON(t *testing.T) {
	testUnmarshalJSON(t, func(in []byte) (fmt.Stringer, error) {
		var out Hash
		return out, json.Unmarshal(in, &out)
	})
}

func TestHex_UnmarshalJSON(t *testing.T) {
	testUnmarshalJSON(t, func(in []byte) (fmt.Stringer, error) {
		var out Hex
		return out, json.Unmarshal(in, &out)
	})
}

func testUnmarshalJSON(t *testing.T, unmarshalJSON func(jsonMessage []byte) (fmt.Stringer, error)) {
	tests := []struct {
		name        string
		inJSON      string
		expected    string
		expectedErr error
	}{
		{"standard", `"ab"`, "ab", nil},
		{"odd length", `"770"`, "0770", nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, err := unmarshalJSON([]byte(test.inJSON))
			if test.expectedErr == nil {
				require.NoError(t, err)
				assert.Equal(t, test.expected, value.String())
			} else {
				assert.Equal(t, test.expectedErr, err)
			}
		})
	}
}

func TestUint8_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		inJSON      string
		expected    uint8
		expectedErr error
	}{
		{"empty", `""`, 0, nil},
		{"hex empty", `"0x"`, 0, nil},
		{"hex odd length", `"0x1"`, 1, nil},
		{"hex mixed case", `"0xAc"`, 172, nil},
		{"hex boundary", `"0xff"`, math.MaxUint8, nil},
		{"hex outside boundary", `"0x0100"`, 0, errors.New(`invalid hex uint8 number: strconv.ParseUint: parsing "0100": value out of range`)},
		{"decimal", `"101"`, 101, nil},
		{"decimal boundary", `"255"`, math.MaxUint8, nil},
		{"decimal outside boundary", `"256"`, 0, errors.New(`invalid uint8 number: strconv.ParseUint: parsing "256": value out of range`)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var value Uint8
			err := json.Unmarshal([]byte(test.inJSON), &value)
			if test.expectedErr == nil {
				require.NoError(t, err)
				assert.Equal(t, test.expected, uint8(value))
			} else {
				assert.EqualError(t, err, test.expectedErr.Error())
			}
		})
	}
}

func TestUint32_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		inJSON      string
		expected    uint32
		expectedErr error
	}{
		{"empty", `""`, 0, nil},
		{"hex empty", `"0x"`, 0, nil},
		{"hex odd length", `"0x1"`, 1, nil},
		{"hex mixed case", `"0x0AbC"`, 2748, nil},
		{"hex boundary", `"0xffffffff"`, math.MaxUint32, nil},
		{"hex outside boundary", `"0x0100000000"`, 0, errors.New(`invalid hex uint32 number: strconv.ParseUint: parsing "0100000000": value out of range`)},
		{"decimal", `"101"`, 101, nil},
		{"decimal boundary", `"4294967295"`, math.MaxUint32, nil},
		{"decimal outside boundary", `"4294967296"`, 0, errors.New(`invalid uint32 number: strconv.ParseUint: parsing "4294967296": value out of range`)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var value Uint32
			err := json.Unmarshal([]byte(test.inJSON), &value)
			if test.expectedErr == nil {
				require.NoError(t, err)
				assert.Equal(t, test.expected, uint32(value))
			} else {
				assert.EqualError(t, err, test.expectedErr.Error())
			}
		})
	}
}

func TestUint64_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		inJSON      string
		expected    uint64
		expectedErr error
	}{
		{"empty", `""`, 0, nil},
		{"hex empty", `"0x"`, 0, nil},
		{"hex odd length", `"0x1"`, 1, nil},
		{"hex mixed case", `"0x0AbC"`, 2748, nil},
		{"hex boundary", `"0xffffffffffffffff"`, math.MaxUint64, nil},
		{"hex outside boundary", `"0x010000000000000000"`, 0, errors.New(`invalid hex uint64 number: strconv.ParseUint: parsing "010000000000000000": value out of range`)},
		{"decimal", `"101"`, 101, nil},
		{"decimal boundary", `"18446744073709551615"`, math.MaxUint64, nil},
		{"decimal outside boundary", `"18446744073709551616"`, 0, errors.New(`invalid uint64 number: strconv.ParseUint: parsing "18446744073709551616": value out of range`)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var value Uint64
			err := json.Unmarshal([]byte(test.inJSON), &value)
			if test.expectedErr == nil {
				require.NoError(t, err)
				assert.Equal(t, test.expected, uint64(value))
			} else {
				assert.EqualError(t, err, test.expectedErr.Error())
			}
		})
	}
}

// TODO: add more test cases
func TestFixedUint64_MarshalJSONRPC(t *testing.T) {
	nonce := FixedUint64(66)

	b, err := nonce.MarshalJSONRPC()

	require.NoError(t, err)
	assert.Equal(t, `"0x0000000000000042"`, string(b))
}

// TODO: add more test cases
func TestTimestamp_MarshalJSONRPC(t *testing.T) {
	ts := Timestamp(time.Unix(0, 0))

	b, err := ts.MarshalJSONRPC()

	require.NoError(t, err)
	assert.Equal(t, `"0x0"`, string(b))
}

func Test_padTo32Bytes(t *testing.T) {
	type args struct {
		in []byte
	}

	tests := []struct {
		name    string
		args    args
		wantOut *Topic
	}{
		{
			"empty",
			args{in: nil},
			topic("0x0000000000000000000000000000000000000000000000000000000000000000"),
		},
		{
			"address",
			args{in: MustNewAddress("0xFffDB7377345371817F2b4dD490319755F5899eC")},
			topic("0x000000000000000000000000FffDB7377345371817F2b4dD490319755F5899eC"),
		},
		{
			"flush",
			args{in: MustNewHash("0x1111111111111111111111111111111111111111111111111111111111111111")},
			topic("0x1111111111111111111111111111111111111111111111111111111111111111"),
		},
		{
			"over",
			args{in: MustNewHash("0x111111111111111111111111111111111111111111111111111111111111111100000000")},
			topic("0x1111111111111111111111111111111111111111111111111111111111111111"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotOut := padToTopic(tt.args.in); !reflect.DeepEqual(gotOut, tt.wantOut) {
				t.Errorf("padTo32Bytes() = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}

func topic(in string) (out *Topic) {
	var bytes [32]byte
	copy(bytes[:], MustNewHash(in))
	return (*Topic)(&bytes)
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		name        string
		in          string
		bitSize     int
		expected    int64
		expectedErr string
	}{
		{name: "empty", in: "", bitSize: 64},
		{name: "hex", in: "0xff", bitSize: 64, expected: 255},
		{name: "hex uppercase prefix", in: "0Xff", bitSize: 64, expected: 255},
		{name: "hex prefix only", in: "0x", bitSize: 64},
		{name: "decimal", in: "255", bitSize: 64, expected: 255},
		{name: "decimal negative", in: "-255", bitSize: 64, expected: -255},

		// The form MarshalJSONRPC emits, matching go-ethereum's hexutil.
		{name: "hex negative", in: "-0xff", bitSize: 64, expected: -255},
		{name: "hex negative uppercase prefix", in: "-0Xff", bitSize: 64, expected: -255},

		// The form this package used to emit, still accepted on input.
		{name: "hex negative, legacy sign placement", in: "0x-ff", bitSize: 64, expected: -255},

		// The sign is restored before parsing so bounds are checked against the signed
		// range rather than the unsigned one.
		{name: "hex int8 lower bound", in: "-0x80", bitSize: 8, expected: -128},
		{name: "hex int8 upper bound", in: "0x7f", bitSize: 8, expected: 127},
		{name: "hex int8 overflow", in: "0x80", bitSize: 8, expectedErr: "invalid hex int8 number"},

		{name: "hex int64 lower bound", in: "-0x8000000000000000", bitSize: 64, expected: math.MinInt64},
		{name: "hex int64 upper bound", in: "0x7fffffffffffffff", bitSize: 64, expected: math.MaxInt64},

		{name: "not a number", in: "0xzz", bitSize: 64, expectedErr: "invalid hex int64 number"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := parseInt(test.in, test.bitSize)
			if test.expectedErr != "" {
				require.ErrorContains(t, err, test.expectedErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}

// TestSignedQuantityRoundTrip checks that what rpc.MarshalJSONRPC writes for a signed value
// is what parseInt reads back.
func TestSignedQuantityRoundTrip(t *testing.T) {
	for _, value := range []int64{0, 1, -1, 255, -255, math.MaxInt64, math.MinInt64} {
		t.Run(strconv.FormatInt(value, 10), func(t *testing.T) {
			var number Int64
			require.NoError(t, number.UnmarshalText([]byte(encodeSignedQuantity(value))))

			assert.Equal(t, value, int64(number))
		})
	}
}

// encodeSignedQuantity mirrors what the rpc package writes, kept here so the eth package
// does not have to import it.
func encodeSignedQuantity(value int64) string {
	magnitude := uint64(value)
	if value < 0 {
		return "-0x" + strconv.FormatUint(-magnitude, 16)
	}

	return "0x" + strconv.FormatUint(magnitude, 16)
}

// TestIntUnmarshalText covers the full range of Int, which used to be parsed as if it were
// an Int8 and so rejected anything above 127.
func TestIntUnmarshalText(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		expected Int
	}{
		{name: "zero", in: "0x0", expected: 0},
		{name: "within int8", in: "0x7f", expected: 127},
		{name: "past int8", in: "0x1ff", expected: 511},
		{name: "large", in: "0x7fffffff", expected: 2147483647},
		{name: "negative past int8", in: "-0x1ff", expected: -511},
		{name: "decimal", in: "511", expected: 511},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var actual Int
			require.NoError(t, actual.UnmarshalText([]byte(test.in)))

			assert.Equal(t, test.expected, actual)
		})
	}

	t.Run("out of range for int", func(t *testing.T) {
		var actual Int
		require.ErrorContains(t, actual.UnmarshalText([]byte("0x1"+strings.Repeat("0", 16))), "out of range")
	})
}
