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
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/holiman/uint256"
)

type Uint8 uint8

func (b *Uint8) UnmarshalText(text []byte) error {
	value, err := parseUint(string(text), 8)
	if err != nil {
		return err
	}

	*b = Uint8(value)
	return nil
}

type Uint16 uint16

func (b *Uint16) UnmarshalText(text []byte) error {
	value, err := parseUint(string(text), 16)
	if err != nil {
		return err
	}

	*b = Uint16(value)
	return nil
}

type Uint32 uint32

func (b *Uint32) UnmarshalText(text []byte) error {
	value, err := parseUint(string(text), 32)
	if err != nil {
		return err
	}

	*b = Uint32(value)
	return nil
}

type Uint64 uint64

func (b *Uint64) UnmarshalText(text []byte) error {
	value, err := parseUint(string(text), 64)
	if err != nil {
		return err
	}

	*b = Uint64(value)
	return nil
}

type Uint256 uint256.Int

func (b *Uint256) UnmarshalText(text []byte) error {
	// we want to remove "0x0" (and any other zeroes) then we add back the leading "0x"
	noZero := bytes.TrimLeft(text, "0x")
	if len(noZero) == 0 {
		text = []byte{'0', 'x', '0'}
	} else {
		text = append([]byte{'0', 'x'}, noZero...)
	}
	return (*uint256.Int)(b).UnmarshalText(text)
}

func (b *Uint256) MarshalText() ([]byte, error) {
	return []byte((*uint256.Int)(b).Hex()), nil
}

func (b *Uint256) MarshalJSONRPC() ([]byte, error) {
	return []byte(`"` + (*uint256.Int)(b).Hex() + `"`), nil
}

// FixedUint64 is a fixed size uint64, marshalled as a fixed 8 bytes big endian
type FixedUint64 uint64

func (b *FixedUint64) UnmarshalText(text []byte) error {
	value, err := parseUint(string(text), 64)
	if err != nil {
		return err
	}

	*b = FixedUint64(value)
	return nil
}

func (n *FixedUint64) MarshalJSONRPC() ([]byte, error) {
	if n == nil {
		return []byte(`"0x0000000000000000"`), nil
	}

	v := make([]byte, 8)
	binary.BigEndian.PutUint64(v, uint64(*n))

	return Hex(v).MarshalJSONRPC()
}

// Timestamp represents a timestamp value on the Ethereum chain always in UTC
// time zone. Recorded in unix seconds
type Timestamp time.Time

func (t Timestamp) MarshalJSONRPC() ([]byte, error) {
	seconds := Uint64(time.Time(t).Unix())

	v := make([]byte, 0, 16)
	v = strconv.AppendUint(v, uint64(seconds), 16)

	return []byte(`"0x` + string(v) + `"`), nil
}

func (t Timestamp) MarshalText() ([]byte, error) {
	return []byte((time.Time)(t).Format(time.RFC3339)), nil
}

func (t *Timestamp) UnmarshalText(text []byte) error {
	// Shall we deal with an actual date time string format also here?
	value, err := parseUint(string(text), 64)
	if err != nil {
		return err
	}

	*t = Timestamp(time.Unix(int64(value), 0).UTC())
	return nil
}

func parseUint(text string, bitSize int) (uint64, error) {
	if len(text) == 0 {
		return 0, nil
	}

	// If it's a hexadecimal string, let's parse it as-is
	if strings.HasPrefix(text, "0x") || strings.HasPrefix(text, "0X") {
		text = text[2:]
		if text == "" {
			return 0, nil
		}

		value, err := strconv.ParseUint(text, 16, bitSize)
		if err != nil {
			return 0, fmt.Errorf("invalid hex uint%d number: %w", bitSize, err)
		}

		return value, nil
	}

	// Otherwise, we assume it's a decimal number
	value, err := strconv.ParseUint(text, 10, bitSize)
	if err != nil {
		return 0, fmt.Errorf("invalid uint%d number: %w", bitSize, err)
	}

	return value, nil
}

type Int int

func (b *Int) UnmarshalText(text []byte) error {
	// strconv.IntSize, not 64, because int is platform sized.
	value, err := parseInt(string(text), strconv.IntSize)
	if err != nil {
		return err
	}

	*b = Int(value)
	return nil
}

type Int8 int8

func (b *Int8) UnmarshalText(text []byte) error {
	value, err := parseInt(string(text), 8)
	if err != nil {
		return err
	}

	*b = Int8(value)
	return nil
}

type Int16 int16

func (b *Int16) UnmarshalText(text []byte) error {
	value, err := parseInt(string(text), 16)
	if err != nil {
		return err
	}

	*b = Int16(value)
	return nil
}

type Int32 int32

func (b *Int32) UnmarshalText(text []byte) error {
	value, err := parseInt(string(text), 32)
	if err != nil {
		return err
	}

	*b = Int32(value)
	return nil
}

type Int64 int64

func (b *Int64) UnmarshalText(text []byte) error {
	value, err := parseInt(string(text), 64)
	if err != nil {
		return err
	}

	*b = Int64(value)
	return nil
}

func parseInt(text string, bitSize int) (int64, error) {
	if len(text) == 0 {
		return 0, nil
	}

	// A negative hexadecimal is written "-0xff", as go-ethereum's hexutil does it.
	sign := ""
	if rest, found := strings.CutPrefix(text, "-"); found {
		sign, text = "-", rest
	}

	// If it's a hexadecimal string, let's parse it as-is
	if digits, found := cutHexPrefix(text); found {
		if digits == "" {
			return 0, nil
		}

		// Older versions of this package wrote the sign after the prefix instead, as
		// "0x-ff". Keep reading that form.
		if rest, found := strings.CutPrefix(digits, "-"); found {
			sign, digits = "-", rest
		}

		// The sign goes back on before parsing so that strconv range-checks against the
		// signed bounds, where -0x80 fits an int8 but 0x80 does not.
		value, err := strconv.ParseInt(sign+digits, 16, bitSize)
		if err != nil {
			return 0, fmt.Errorf("invalid hex int%d number: %w", bitSize, err)
		}

		return value, nil
	}

	// Otherwise, we assume it's a decimal number
	value, err := strconv.ParseInt(sign+text, 10, bitSize)
	if err != nil {
		return 0, fmt.Errorf("invalid int%d number: %w", bitSize, err)
	}

	return value, nil
}

func cutHexPrefix(text string) (digits string, found bool) {
	if digits, found = strings.CutPrefix(text, "0x"); found {
		return digits, true
	}

	return strings.CutPrefix(text, "0X")
}

type Bytes []byte

func MustNewBytes(input string) Bytes {
	return Bytes(mustNewByteSlice("bytes", input))
}

func NewBytes(input string) (Bytes, error) {
	out, err := newByteSlice("bytes", input)
	if err != nil {
		return nil, err
	}

	return Bytes(out), nil
}

func (h Bytes) String() string                   { return byteSlice(h).String() }
func (h Bytes) Pretty() string                   { return byteSlice(h).Pretty() }
func (h Bytes) Bytes() []byte                    { return h[:] }
func (h Bytes) IsZero() bool                     { return len(h) == 0 }
func (h Bytes) MarshalText() ([]byte, error)     { return byteSlice(h).MarshalText() }
func (h Bytes) ID() uint64                       { return byteSlice(h).ID() }
func (h Bytes) MarshalJSON() ([]byte, error)     { return byteSlice(h).MarshalJSON() }
func (h Bytes) MarshalJSONRPC() ([]byte, error)  { return byteSlice(h).MarshalJSONRPC() }
func (h *Bytes) UnmarshalJSON(data []byte) error { return (*byteSlice)(h).UnmarshalJSON(data) }
func (h *Bytes) UnmarshalText(data []byte) error { return (*byteSlice)(h).UnmarshalText(data) }

type Hex []byte

func MustNewHex(input string) Hex {
	return Hex(mustNewByteSlice("hex", input))
}

func NewHex(input string) (Hex, error) {
	out, err := newByteSlice("hex", input)
	if err != nil {
		return nil, err
	}

	return Hex(out), nil
}

func (h Hex) String() string                   { return byteSlice(h).String() }
func (h Hex) Pretty() string                   { return byteSlice(h).Pretty() }
func (h Hex) Bytes() []byte                    { return h[:] }
func (h Hex) IsZero() bool                     { return len(h) == 0 }
func (h Hex) MarshalText() ([]byte, error)     { return byteSlice(h).MarshalText() }
func (h Hex) ID() uint64                       { return byteSlice(h).ID() }
func (h Hex) MarshalJSON() ([]byte, error)     { return byteSlice(h).MarshalJSON() }
func (h Hex) MarshalJSONRPC() ([]byte, error)  { return byteSlice(h).MarshalJSONRPC() }
func (h *Hex) UnmarshalJSON(data []byte) error { return (*byteSlice)(h).UnmarshalJSON(data) }
func (h *Hex) UnmarshalText(data []byte) error { return (*byteSlice)(h).UnmarshalText(data) }

type Hash []byte

func MustNewHash(input string) Hash {
	return Hash(mustNewByteSlice("hash", input))
}

func NewHash(input string) (Hash, error) {
	out, err := newByteSlice("hash", input)
	if err != nil {
		return nil, err
	}

	return Hash(out), nil
}

func (h Hash) String() string                   { return byteSlice(h).String() }
func (h Hash) Pretty() string                   { return byteSlice(h).Pretty() }
func (h Hash) Bytes() []byte                    { return h[:] }
func (h Hash) IsZero() bool                     { return len(h) == 0 }
func (h Hash) MarshalText() ([]byte, error)     { return byteSlice(h).MarshalText() }
func (h Hash) ID() uint64                       { return byteSlice(h).ID() }
func (h Hash) MarshalJSON() ([]byte, error)     { return byteSlice(h).MarshalJSON() }
func (h Hash) MarshalJSONRPC() ([]byte, error)  { return byteSlice(h).MarshalJSONRPC() }
func (h *Hash) UnmarshalJSON(data []byte) error { return (*byteSlice)(h).UnmarshalJSON(data) }
func (h *Hash) UnmarshalText(data []byte) error { return (*byteSlice)(h).UnmarshalText(data) }

type Address []byte

// MustNewAddress is exactly like [NewAddress] but it panics if an error occurrs.
//
// See [NewAdress] for detaiks.
func MustNewAddress(input string) Address {
	out, err := NewAddress(input)
	if err != nil {
		panic(err)
	}

	return out
}

// MustNewAddressLoose is exactly like [NewAddressLoose] but it panics if an error occurrs.
//
// See [NewAddressLoose] for detaiks.
func MustNewAddressLoose(input string) Address {
	out, err := NewAddressLoose(input)
	if err != nil {
		panic(err)
	}

	return out
}

// NewAddress creates an address where the input **must** contain exactly 20 bytes.
//
// If the actual input is not an hexadecimal string, an error is returned or the decoded
// lenght is not exactly 20 bytes, an error is returned.
func NewAddress(input string) (Address, error) {
	out, err := newByteSlice("address", input)
	if err != nil {
		return nil, err
	}

	if len(out) != 20 {
		return nil, fmt.Errorf("invalid length in bytes, wanted 20 bytes once decoded but decoding gave us %d bytes", len(out))
	}

	return Address(out), nil
}

// NewAddressLoose creates an address where the input can contain more than 20 bytes (truncated) or
// less than 20 bytes (left as-is).
//
// If the actual input is not an hexadecimal string, an error is returned.
func NewAddressLoose(input string) (Address, error) {
	out, err := newByteSlice("address", input)
	if err != nil {
		return nil, err
	}

	byteCount := len(out)
	if byteCount > 20 {
		out = out[byteCount-20:]
	}

	return Address(out), nil
}

func (a Address) String() string                   { return byteSlice(a).String() }
func (a Address) Pretty() string                   { return byteSlice(a).Pretty() }
func (a Address) Bytes() []byte                    { return a[:] }
func (a Address) IsZero() bool                     { return len(a) == 0 }
func (a Address) MarshalText() ([]byte, error)     { return byteSlice(a).MarshalText() }
func (a Address) ID() uint64                       { return byteSlice(a).ID() }
func (a Address) MarshalJSON() ([]byte, error)     { return byteSlice(a).MarshalJSON() }
func (a Address) MarshalJSONRPC() ([]byte, error)  { return byteSlice(a).MarshalJSONRPC() }
func (a *Address) UnmarshalJSON(data []byte) error { return (*byteSlice)(a).UnmarshalJSON(data) }
func (a *Address) UnmarshalText(data []byte) error { return (*byteSlice)(a).UnmarshalText(data) }

type byteSlice []byte

func mustNewByteSlice(tag string, input string) byteSlice {
	out, err := newByteSlice(tag, input)
	if err != nil {
		panic(err)
	}

	return out
}

func newByteSlice(tag string, input string) (out byteSlice, err error) {
	bytes, err := hex.DecodeString(SanitizeHex(input))
	if err != nil {
		return out, fmt.Errorf("invalid %s %q: %w", tag, input, err)
	}

	return bytes, nil
}

func (b byteSlice) String() string {
	return hex.EncodeToString(b)
}

func (b byteSlice) Pretty() string {
	return "0x" + hex.EncodeToString(b)
}

func (b byteSlice) IsZero() bool {
	return len(b) == 0
}

func (b byteSlice) Bytes() []byte {
	return b
}

func (b byteSlice) MarshalText() ([]byte, error) {
	return []byte(b.String()), nil
}

func (b byteSlice) ID() uint64 {
	return binary.LittleEndian.Uint64(b)
}

func (b byteSlice) MarshalJSON() ([]byte, error) {
	return []byte(`"` + hex.EncodeToString([]byte(b)) + `"`), nil
}

func (b byteSlice) MarshalJSONRPC() ([]byte, error) {
	return []byte(`"` + b.Pretty() + `"`), nil
}

func (b *byteSlice) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	return b.UnmarshalText([]byte(s))
}

func (b *byteSlice) UnmarshalText(text []byte) error {
	s := strings.TrimPrefix(string(text), "0x")
	if len(s)%2 != 0 {
		s = "0" + s
	}

	var err error
	if *b, err = hex.DecodeString(s); err != nil {
		return err
	}

	return nil
}

type Topic [32]byte

func (f Topic) MarshalJSONRPC() ([]byte, error) {
	return []byte(`"0x` + hex.EncodeToString(f[:]) + `"`), nil
}

func LogTopic(in interface{}) *Topic {
	switch v := in.(type) {
	case string:
		return padToTopic(MustNewHash(v))
	case []byte:
		return padToTopic(v)
	case Hash:
		return padToTopic(v)
	case Hex:
		return padToTopic(v)
	case Address:
		return padToTopic(v)
	case nil:
		return nil
	default:
		valueOf := reflect.ValueOf(v)
		if valueOf.Kind() == reflect.Ptr && valueOf.IsNil() {
			return nil
		}

		panic(fmt.Errorf("don't know how to turn %T into a LogTopic", in))
	}
}

func padToTopic(in []byte) (out *Topic) {
	startOffset := 32 - len(in)
	if startOffset < 0 {
		startOffset = 0
	}

	var topic Topic
	copy(topic[startOffset:], in)

	return &topic
}

//go:generate go-enum -f=$GOFILE --noprefix --prefix TxType --lower --names

// ENUM(
//
//	Legacy
//	AccessList
//	DynamicFee
//
// )
type TransactionType uint8

func (b *TransactionType) UnmarshalText(text []byte) error {
	value, err := parseUint(string(text), 8)
	if err != nil {
		return err
	}

	*b = TransactionType(value)
	return nil
}
