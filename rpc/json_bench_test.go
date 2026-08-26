package rpc

import (
	"math/big"
	"testing"

	"github.com/streamingfast/eth-go"
)

func benchRequests() []*RPCRequest {
	addr := eth.MustNewAddress("0x1234567890123456789012345678901234567890")

	reqs := make([]*RPCRequest, 20)
	for i := range reqs {
		reqs[i] = &RPCRequest{
			JSONRPC: "2.0",
			Method:  "eth_call",
			ID:      eth.Int(i + 1),
			Params: []any{CallParams{
				From:     addr,
				To:       addr,
				GasLimit: 21000,
				GasPrice: big.NewInt(1_000_000_000),
				Data:     eth.Hex{0x70, 0xa0, 0x82, 0x31},
			}, BlockNumber(uint64(i))},
		}
	}

	return reqs
}

func BenchmarkMarshalJSONRPCBatch(b *testing.B) {
	reqs := benchRequests()
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		if _, err := MarshalJSONRPC(&reqs); err != nil {
			b.Fatal(err)
		}
	}
}
