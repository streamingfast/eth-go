package rpc

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/streamingfast/eth-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTopicFilter(t *testing.T) {
	type args struct {
		exprs    []interface{}
		appender func(f *TopicFilter)
	}
	tests := []struct {
		name    string
		args    args
		wantOut *TopicFilter
	}{
		{
			"topic0",
			args{exprs: []interface{}{"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"}},
			&TopicFilter{
				topics: []TopicFilterExpr{{exact: topic("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")}},
			},
		},
		{
			"topic0 with append null",
			args{
				exprs: []interface{}{"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"},
				appender: func(f *TopicFilter) {
					f.Append(nil)
					f.Append(eth.MustNewAddress("0xFffDB7377345371817F2b4dD490319755F5899eC"))
				},
			},
			&TopicFilter{
				topics: []TopicFilterExpr{
					{exact: topic("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")},
					{exact: nil},
					{exact: topic("0x000000000000000000000000FffDB7377345371817F2b4dD490319755F5899eC")},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewTopicFilter(tt.args.exprs...)
			if tt.args.appender != nil {
				tt.args.appender(f)
			}

			if gotOut := f; !reflect.DeepEqual(gotOut, tt.wantOut) {
				t.Errorf("NewTopicFilter() = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}

func topic(in string) (out *eth.Topic) {
	var bytes [32]byte
	copy(bytes[:], eth.MustNewHash(in))
	return (*eth.Topic)(&bytes)
}

func TestBlockRef_UnmarshalJSON(t *testing.T) {
	type args struct {
		text string
	}
	tests := []struct {
		name      string
		args      args
		expected  *BlockRef
		assertion require.ErrorAssertionFunc
	}{
		{"latest all lower case", args{`"latest"`}, LatestBlock, require.NoError},
		{"latest mixed case", args{`"laTesT"`}, LatestBlock, require.NoError},
		{"latest all upper case", args{`"LATEST"`}, LatestBlock, require.NoError},

		{"earliest all lower case", args{`"earliest"`}, EarliestBlock, require.NoError},
		{"earliest mixed case", args{`"EarliesT"`}, EarliestBlock, require.NoError},
		{"earliest all upper case", args{`"EARLIEST"`}, EarliestBlock, require.NoError},

		{"pending all lower case", args{`"pending"`}, PendingBlock, require.NoError},
		{"pending mixed case", args{`"pEndIng"`}, PendingBlock, require.NoError},
		{"pending all upper case", args{`"PENDING"`}, PendingBlock, require.NoError},

		{"block number decimal zero", args{`"0"`}, BlockNumber(0), require.NoError},
		{"block number decimal value", args{`"112"`}, BlockNumber(112), require.NoError},

		{"block number hexadecimal empty", args{`"0x"`}, BlockNumber(0), require.NoError},
		{"block number hexadecimal zero", args{`"0x0"`}, BlockNumber(0), require.NoError},
		{"block number hexadecimal zero", args{`"0x0"`}, BlockNumber(0), require.NoError},
		{"block number hexadecimal value", args{`"0x123"`}, BlockNumber(291), require.NoError},

		{"block hash empty", args{`{"blockHash":"0x"}`}, BlockHash(""), require.NoError},
		{"block hash full", args{`{"blockHash":"0xf092d0fffe12ec3978b369b861121b62b37a4c1176beda7116f24ce1b7a4937e"}`}, BlockHash("0xf092d0fffe12ec3978b369b861121b62b37a4c1176beda7116f24ce1b7a4937e"), require.NoError},

		{"block number object hex string", args{`{"blockNumber":"0x808f609"}`}, BlockNumber(134805001), require.NoError},
		{"block number object hex string uppercase", args{`{"blockNumber":"0x808F609"}`}, BlockNumber(134805001), require.NoError},
		{"block number object hex string zero", args{`{"blockNumber":"0x0"}`}, BlockNumber(0), require.NoError},

		{"block number object numeric", args{`{"blockNumber":134805001}`}, BlockNumber(134805001), require.NoError},
		{"block number object numeric zero", args{`{"blockNumber":0}`}, BlockNumber(0), require.NoError},
		{"block number object numeric large", args{`{"blockNumber":5175791617}`}, BlockNumber(5175791617), require.NoError},

		{"block number raw numeric", args{`134805001`}, BlockNumber(134805001), require.NoError},
		{"block number raw numeric zero", args{`0`}, BlockNumber(0), require.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BlockRef{}
			tt.assertion(t, b.UnmarshalJSON([]byte(tt.args.text)))

			assert.Equal(t, tt.expected, b)
		})
	}
}

func TestName(t *testing.T) {
	data := []byte(`["0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3",false]`)

	v := []interface{}{new(BlockRef), new(bool)}
	err := json.Unmarshal(data, &v)
	require.NoError(t, err)

	fmt.Println(v)

}

func TestBlockRef_UnmarshalText(t *testing.T) {
	type args struct {
		text string
	}
	tests := []struct {
		name      string
		args      args
		expected  *BlockRef
		assertion require.ErrorAssertionFunc
	}{
		{"latest all lower case", args{"latest"}, LatestBlock, require.NoError},
		{"latest mixed case", args{"laTesT"}, LatestBlock, require.NoError},
		{"latest all upper case", args{"LATEST"}, LatestBlock, require.NoError},

		{"earliest all lower case", args{"earliest"}, EarliestBlock, require.NoError},
		{"earliest mixed case", args{"EarliesT"}, EarliestBlock, require.NoError},
		{"earliest all upper case", args{"EARLIEST"}, EarliestBlock, require.NoError},

		{"pending all lower case", args{"pending"}, PendingBlock, require.NoError},
		{"pending mixed case", args{"pEndIng"}, PendingBlock, require.NoError},
		{"pending all upper case", args{"PENDING"}, PendingBlock, require.NoError},

		{"block number decimal zero", args{"0"}, BlockNumber(0), require.NoError},
		{"block number decimal value", args{"112"}, BlockNumber(112), require.NoError},

		{"block number hexadecimal empty", args{"0x"}, BlockNumber(0), require.NoError},
		{"block number hexadecimal zero", args{"0x0"}, BlockNumber(0), require.NoError},
		{"block number hexadecimal zero", args{"0x0"}, BlockNumber(0), require.NoError},
		{"block number hexadecimal value", args{"0x123"}, BlockNumber(291), require.NoError},

		{"block hash empty", args{`{"blockHash":"0x"}`}, BlockHash(""), require.NoError},
		{"block hash full", args{`{"blockHash":"0xf092d0fffe12ec3978b369b861121b62b37a4c1176beda7116f24ce1b7a4937e"}`}, BlockHash("0xf092d0fffe12ec3978b369b861121b62b37a4c1176beda7116f24ce1b7a4937e"), require.NoError},

		{"block number object hex string", args{`{"blockNumber":"0x808f609"}`}, BlockNumber(134805001), require.NoError},
		{"block number object hex string uppercase", args{`{"blockNumber":"0x808F609"}`}, BlockNumber(134805001), require.NoError},
		{"block number object hex string zero", args{`{"blockNumber":"0x0"}`}, BlockNumber(0), require.NoError},

		{"block number object numeric", args{`{"blockNumber":134805001}`}, BlockNumber(134805001), require.NoError},
		{"block number object numeric zero", args{`{"blockNumber":0}`}, BlockNumber(0), require.NoError},
		{"block number object numeric large", args{`{"blockNumber":5175791617}`}, BlockNumber(5175791617), require.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BlockRef{}
			tt.assertion(t, b.UnmarshalText([]byte(tt.args.text)))

			assert.Equal(t, tt.expected, b)
		})
	}
}

func TestBlockRef_IsLatest(t *testing.T) {
	latest, earliest, pending := blockRefFromUnmarshal(t)

	tests := []struct {
		name string
		ref  *BlockRef
		want bool
	}{
		{"latest from cached one", LatestBlock, true},
		{"latest from marshalled", latest, true},

		{"not latest when earliest from cached one", EarliestBlock, false},
		{"not latest when earliest from marshalled", earliest, false},

		{"not latest when pending from cached one", PendingBlock, false},
		{"not latest when pending from marshalled", pending, false},

		{"not latest when block number", BlockNumber(10), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.ref.IsLatest())
		})
	}
}

func TestBlockRef_IsEarliest(t *testing.T) {
	latest, earliest, pending := blockRefFromUnmarshal(t)

	tests := []struct {
		name string
		ref  *BlockRef
		want bool
	}{
		{"earliest from cached one", EarliestBlock, true},
		{"earliest from marshalled", earliest, true},

		{"not earliest when latest from cached one", LatestBlock, false},
		{"not earliest when latest from marshalled", latest, false},

		{"not earliest when pending from cached one", PendingBlock, false},
		{"not earliest when pending from marshalled", pending, false},

		{"not earliest when block number", BlockNumber(10), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.ref.IsEarliest())
		})
	}
}

func TestBlockRef_IsPending(t *testing.T) {
	latest, earliest, pending := blockRefFromUnmarshal(t)

	tests := []struct {
		name string
		ref  *BlockRef
		want bool
	}{
		{"pending from cached one", PendingBlock, true},
		{"pending from marshalled", pending, true},

		{"not pending when latest from cached one", LatestBlock, false},
		{"not pending when latest from marshalled", latest, false},

		{"not pending when earliest from cached one", EarliestBlock, false},
		{"not pending when earliest from marshalled", earliest, false},

		{"not pending when block number", BlockNumber(10), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.ref.IsPending())
		})
	}
}

func TestBlockRef_BlockNumber(t *testing.T) {
	latest, earliest, pending := blockRefFromUnmarshal(t)

	tests := []struct {
		name       string
		ref        *BlockRef
		wantNumber uint64
		wantOk     bool
	}{
		{"not block number when latest from cached one", LatestBlock, 0, false},
		{"not block number when latest from marshalled", latest, 0, false},

		{"not block number when earliest from cached one", EarliestBlock, 0, false},
		{"not block number when earliest from marshalled", earliest, 0, false},

		{"not block number when pending from cached one", PendingBlock, 0, false},
		{"not block number when pending from marshalled", pending, 0, false},

		{"block number", BlockNumber(10), 10, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualNum, actualOk := tt.ref.BlockNumber()

			assert.Equal(t, tt.wantNumber, actualNum)
			assert.Equal(t, tt.wantOk, actualOk)
		})
	}
}

func blockRefFromUnmarshal(t *testing.T) (*BlockRef, *BlockRef, *BlockRef) {
	t.Helper()

	latest := &BlockRef{}
	require.NoError(t, latest.UnmarshalText([]byte("latest")))

	earliest := &BlockRef{}
	require.NoError(t, earliest.UnmarshalText([]byte("earliest")))

	pending := &BlockRef{}
	require.NoError(t, pending.UnmarshalText([]byte("pending")))

	return latest, earliest, pending
}
