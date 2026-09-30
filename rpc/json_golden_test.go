package rpc

import (
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/streamingfast/eth-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// goldenFile records the exact output of [MarshalJSONRPC] for every construct the
// JSON-RPC encoder treats specially. It exists to pin the wire format across the
// migration away from the vendored `encoding/json` fork, and stays afterwards as
// a regression net.
//
// Regenerate with `GOLDEN_UPDATE=true go test ./rpc -run TestMarshalJSONRPCGolden`.
const goldenFile = "testdata/json_rpc_golden.json"

type goldenPlainInts struct {
	Int     int     `json:"int"`
	IntNeg  int     `json:"int_neg"`
	IntZero int     `json:"int_zero"`
	Int8    int8    `json:"int8"`
	Int16   int16   `json:"int16"`
	Int32   int32   `json:"int32"`
	Int64   int64   `json:"int64"`
	Uint    uint    `json:"uint"`
	Uint8   uint8   `json:"uint8"`
	Uint16  uint16  `json:"uint16"`
	Uint32  uint32  `json:"uint32"`
	Uint64  uint64  `json:"uint64"`
	Uintptr uintptr `json:"uintptr"`
}

type goldenEthInts struct {
	Int    eth.Int    `json:"eth_int"`
	Int8   eth.Int8   `json:"eth_int8"`
	Int16  eth.Int16  `json:"eth_int16"`
	Int32  eth.Int32  `json:"eth_int32"`
	Int64  eth.Int64  `json:"eth_int64"`
	Uint8  eth.Uint8  `json:"eth_uint8"`
	Uint16 eth.Uint16 `json:"eth_uint16"`
	Uint32 eth.Uint32 `json:"eth_uint32"`
	Uint64 eth.Uint64 `json:"eth_uint64"`
	TxType eth.TransactionType
}

type goldenBigInts struct {
	Value    big.Int  `json:"value"`
	Ptr      *big.Int `json:"ptr"`
	Nil      *big.Int `json:"nil"`
	Zero     *big.Int `json:"zero"`
	Negative *big.Int `json:"negative"`
	Huge     *big.Int `json:"huge"`
}

type goldenBytes struct {
	Slice      []byte       `json:"slice"`
	SliceNil   []byte       `json:"slice_nil"`
	SliceEmpty []byte       `json:"slice_empty"`
	Bytes      eth.Bytes    `json:"bytes"`
	BytesNil   eth.Bytes    `json:"bytes_nil"`
	Hex        eth.Hex      `json:"hex"`
	Hash       eth.Hash     `json:"hash"`
	Address    eth.Address  `json:"address"`
	AddressPtr *eth.Address `json:"address_ptr"`
	AddressNil *eth.Address `json:"address_nil"`
	Topic      eth.Topic    `json:"topic"`
	Array      [3]byte      `json:"array"`
	ArrayInts  [2]uint64    `json:"array_ints"`
}

type goldenMisc struct {
	Bool      bool              `json:"bool"`
	BoolFalse bool              `json:"bool_false"`
	String    string            `json:"string"`
	HTML      string            `json:"html"`
	Unicode   string            `json:"unicode"`
	Float32   float32           `json:"float32"`
	Float64   float64           `json:"float64"`
	Iface     any               `json:"iface"`
	IfaceNil  any               `json:"iface_nil"`
	Map       map[string]uint64 `json:"map"`
	MapNil    map[string]uint64 `json:"map_nil"`
	Slice     []uint64          `json:"slice"`
	SliceNil  []uint64          `json:"slice_nil"`
	Nested    *goldenPlainInts  `json:"nested"`
	NestedNil *goldenPlainInts  `json:"nested_nil"`
	Ints      []eth.Uint64      `json:"ints"`
	Uint256   *eth.Uint256      `json:"uint256"`
	Fixed     *eth.FixedUint64  `json:"fixed"`
	Timestamp eth.Timestamp     `json:"timestamp"`
	unexposed string            //lint:ignore U1000 exercises unexported-field skipping
}

type goldenOmitEmpty struct {
	IntZero     uint64         `json:"int_zero,omitzero,omitempty"`
	IntSet      uint64         `json:"int_set,omitzero,omitempty"`
	EthIntZero  eth.Uint64     `json:"eth_int_zero,omitzero,omitempty"`
	EthIntSet   eth.Uint64     `json:"eth_int_set,omitzero,omitempty"`
	StringEmpty string         `json:"string_empty,omitzero,omitempty"`
	StringSet   string         `json:"string_set,omitzero,omitempty"`
	BoolFalse   bool           `json:"bool_false,omitzero,omitempty"`
	PtrNil      *big.Int       `json:"ptr_nil,omitzero,omitempty"`
	PtrSet      *big.Int       `json:"ptr_set,omitzero,omitempty"`
	SliceNil    []uint64       `json:"slice_nil,omitzero,omitempty"`
	SliceEmpty  []uint64       `json:"slice_empty,omitzero,omitempty"`
	SliceSet    []uint64       `json:"slice_set,omitzero,omitempty"`
	BytesNil    eth.Bytes      `json:"bytes_nil,omitzero,omitempty"`
	BytesEmpty  eth.Bytes      `json:"bytes_empty,omitzero,omitempty"`
	AccessNil   AccessList     `json:"access_nil,omitzero,omitempty"`
	AccessEmpty AccessList     `json:"access_empty,omitzero,omitempty"`
	MapNil      map[string]int `json:"map_nil,omitzero,omitempty"`
}

type goldenTagged struct {
	Renamed  uint64 `json:"renamed"`
	Skipped  uint64 `json:"-"`
	Untagged uint64
	Quoted   uint64 `json:"quoted,string"`
}

type goldenEmbedded struct {
	goldenPlainInts
	Extra uint64 `json:"extra"`
}

// goldenCase pairs a stable name with the value to encode. `numericID` mirrors
// the `useNumericID` client option.
type goldenCase struct {
	name      string
	value     any
	numericID bool
}

func goldenCases(t *testing.T) []goldenCase {
	t.Helper()

	addr := eth.MustNewAddress("0x1234567890123456789012345678901234567890")
	fixed := eth.FixedUint64(0xdeadbeef)
	uint256 := &eth.Uint256{}
	require.NoError(t, uint256.UnmarshalText([]byte("0x1f")))

	topics := &TopicFilter{}
	topics.Append(eth.MustNewHash("0xaabbcc"))
	topics.Append(OneOfTopic(eth.MustNewHash("0xdd"), eth.MustNewHash("0xee")))
	topics.Append(AnyTopic())

	huge, ok := new(big.Int).SetString("123456789012345678901234567890abcdef", 16)
	require.True(t, ok)

	return []goldenCase{
		{name: "nil", value: nil},
		{name: "plain_ints", value: goldenPlainInts{
			Int: 255, IntNeg: -255, IntZero: 0,
			Int8: 127, Int16: 4096, Int32: 1 << 20, Int64: 1 << 40,
			Uint: 255, Uint8: 255, Uint16: 4096, Uint32: 1 << 20, Uint64: 1 << 40, Uintptr: 16,
		}},
		{name: "eth_ints", value: goldenEthInts{
			Int: 1, Int8: 2, Int16: 3, Int32: 4, Int64: 5,
			Uint8: 6, Uint16: 7, Uint32: 8, Uint64: 9, TxType: 2,
		}},
		{name: "big_ints", value: goldenBigInts{
			Value:    *big.NewInt(31),
			Ptr:      big.NewInt(255),
			Nil:      nil,
			Zero:     big.NewInt(0),
			Negative: big.NewInt(-255),
			Huge:     huge,
		}},
		{name: "bytes", value: goldenBytes{
			Slice:      []byte{0xaa, 0xbb},
			SliceNil:   nil,
			SliceEmpty: []byte{},
			Bytes:      eth.Bytes{0x01, 0x02},
			BytesNil:   nil,
			Hex:        eth.Hex{0x0f},
			Hash:       eth.MustNewHash("0xaabbccdd"),
			Address:    addr,
			AddressPtr: &addr,
			AddressNil: nil,
			Topic:      *eth.LogTopic("0xaabb"),
			Array:      [3]byte{1, 2, 3},
			ArrayInts:  [2]uint64{0, 255},
		}},
		// `unicode` is the one place the output deliberately differs from the vendored fork.
		// It escaped U+2028 and U+2029, as `encoding/json` v1 does; json/v2 leaves them as
		// raw UTF-8 and this package follows it. See marshalOptions.
		{name: "misc", value: goldenMisc{
			Bool: true, BoolFalse: false,
			String:   "hello",
			HTML:     `<a href="x">&</a>`,
			Unicode:  "h\u00e9llo\u2028w\u00f6rld\u2029!",
			Float32:  1.5,
			Float64:  2.25,
			Iface:    eth.Uint64(255),
			IfaceNil: nil,
			Map:      map[string]uint64{"b": 2, "a": 1, "c": 3},
			MapNil:   nil,
			Slice:    []uint64{1, 2, 3},
			SliceNil: nil,
			Nested:   &goldenPlainInts{Int: 1},
			Ints:     []eth.Uint64{1, 2},
			Uint256:  uint256,
			Fixed:    &fixed,
			// Fixed instant so the golden file stays stable.
			Timestamp: eth.Timestamp(time.Unix(1700000000, 0).UTC()),
			unexposed: "ignored",
		}},
		{name: "omit_empty", value: goldenOmitEmpty{
			IntSet: 1, EthIntSet: 2, StringSet: "set",
			PtrSet:      big.NewInt(3),
			SliceEmpty:  []uint64{},
			SliceSet:    []uint64{4},
			BytesEmpty:  eth.Bytes{},
			AccessEmpty: AccessList{},
		}},
		{name: "tagged", value: goldenTagged{Renamed: 1, Skipped: 2, Untagged: 3, Quoted: 4}},
		{name: "embedded", value: goldenEmbedded{goldenPlainInts: goldenPlainInts{Int: 1}, Extra: 2}},
		{name: "err_response", value: &ErrResponse{Code: -32000, Message: "invalid error"}},
		{name: "err_response_data", value: &ErrResponse{Code: 3, Message: "reverted", Data: eth.Hex{0x08, 0xc3}}},
		{name: "block_ref_latest", value: LatestBlock},
		{name: "block_ref_number", value: BlockNumber(1234)},
		{name: "block_ref_hash", value: BlockHash("0xaabbcc")},
		{name: "topic_filter", value: topics},
		{name: "call_params", value: CallParams{
			From:     addr,
			To:       addr,
			GasLimit: 21000,
			GasPrice: big.NewInt(1_000_000_000),
			Value:    big.NewInt(0),
			Data:     eth.Hex{0x70, 0xa0},
		}},
		{name: "call_params_empty", value: CallParams{}},
		{name: "logs_params", value: LogsParams{
			FromBlock: BlockNumber(1),
			ToBlock:   LatestBlock,
			Address:   addr,
			Topics:    topics,
			BlockHash: eth.Bytes{0xaa},
		}},
		{name: "request_hex_id", value: &RPCRequest{
			Params:  []any{addr, LatestBlock},
			Method:  "eth_getBalance",
			JSONRPC: "2.0",
			ID:      eth.Int(255),
		}},
		{name: "request_numeric_id", numericID: true, value: &RPCRequest{
			Params:  []any{addr, LatestBlock},
			Method:  "eth_getBalance",
			JSONRPC: "2.0",
			ID:      eth.Int(255),
		}},
		{name: "request_batch", value: []*RPCRequest{
			{Params: []any{}, Method: "eth_blockNumber", JSONRPC: "2.0", ID: eth.Int(1)},
			{Params: nil, Method: "eth_chainId", JSONRPC: "2.0", ID: eth.Int(2)},
		}},
		{name: "top_level_uint64", value: eth.Uint64(255)},
		{name: "top_level_bytes", value: []byte{0xde, 0xad}},
		{name: "top_level_slice", value: []eth.Uint64{0, 1, 255}},
		{name: "top_level_map", value: map[string]eth.Uint64{"z": 1, "a": 2}},
		{name: "top_level_string", value: "plain"},
	}
}

func TestMarshalJSONRPCGolden(t *testing.T) {
	cases := goldenCases(t)

	actual := make(map[string]string, len(cases))
	for _, c := range cases {
		out, err := MarshalJSONRPCWithOpts(c.value, c.numericID)
		require.NoError(t, err, "case %q", c.name)
		actual[c.name] = string(out)
	}

	if os.Getenv("GOLDEN_UPDATE") == "true" {
		require.NoError(t, os.MkdirAll(filepath.Dir(goldenFile), 0755))

		content, err := json.MarshalIndent(actual, "", "  ")
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(goldenFile, append(content, '\n'), 0644))

		t.Logf("golden file %s updated with %d cases", goldenFile, len(actual))
		return
	}

	content, err := os.ReadFile(goldenFile)
	require.NoError(t, err, "run with GOLDEN_UPDATE=true to create %s", goldenFile)

	var expected map[string]string
	require.NoError(t, json.Unmarshal(content, &expected))

	for _, c := range cases {
		want, found := expected[c.name]
		if !assert.True(t, found, "case %q missing from golden file", c.name) {
			continue
		}

		assert.Equal(t, want, actual[c.name], "case %q", c.name)
	}

	assert.Len(t, actual, len(expected), "golden file and cases are out of sync")
}
