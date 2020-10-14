package eth

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestEncodeFullMethod(t *testing.T) {
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
	err := e.WriteMethod(method)
	require.NoError(t, err)
	assert.Equal(t, []byte{
		0x38, 0xed, 0x17, 0x39,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x5a, 0xf3, 0x10, 0x7a, 0x40, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x08, 0x3c, 0x12, 0x82, 0x6f, 0xe2, 0x3b,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xa0,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x40, 0xc7, 0xf6, 0x27,
		0xff, 0xb6, 0x9b, 0x8d, 0x87, 0x52, 0xc5, 0x18,
		0xf8, 0x79, 0x0b, 0x04, 0xa5, 0x23, 0xbe, 0xe5,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x5f, 0x6c, 0xaf, 0x45,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0xd2, 0x4a, 0xf8, 0x25,
		0xe3, 0x84, 0x95, 0xee, 0x36, 0x24, 0x66, 0xf2,
		0x14, 0x94, 0x6c, 0xdf, 0x53, 0xaa, 0xb8, 0xc8,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0xc7, 0x78, 0x41, 0x7e,
		0x06, 0x31, 0x41, 0x13, 0x9f, 0xce, 0x01, 0x09,
		0x82, 0x78, 0x01, 0x40, 0xaa, 0x0c, 0xd5, 0xab,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x7d, 0x97, 0xba, 0x95,
		0xda, 0xc2, 0x53, 0x16, 0xb9, 0x53, 0x11, 0x52,
		0xb3, 0xba, 0xa3, 0x23, 0x27, 0x99, 0x4d, 0xa8,
	}, e.buffer)
}

func TestEncoder_Write(t *testing.T) {
	tests := []struct {
		name        string
		typeName    string
		in          interface{}
		expectError bool
		expectBytes []byte
	}{
		{
			name:        "simple method",
			typeName:    "method",
			in:          "transfer(address,uint256)",
			expectError: false,
			expectBytes: []byte{0xa9, 0x05, 0x9c, 0xbb},
		},
		{
			name:        "another method",
			typeName:    "method",
			in:          "swapExactTokensForTokens(uint256,uint256,address[],address,uint256)",
			expectError: false,
			expectBytes: []byte{0x38, 0xed, 0x17, 0x39},
		},
		{
			name:        "simple address 7d97ba95...",
			typeName:    "address",
			in:          MustNewAddress("7d97ba95dac25316b9531152b3baa32327994da8"),
			expectError: false,
			expectBytes: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x7d, 0x97, 0xba, 0x95,
				0xda, 0xc2, 0x53, 0x16, 0xb9, 0x53, 0x11, 0x52,
				0xb3, 0xba, 0xa3, 0x23, 0x27, 0x99, 0x4d, 0xa8,
			},
		},
		{
			name:        "simple address d24af825...",
			typeName:    "address",
			in:          MustNewAddress("d24af825e38495ee362466f214946cdf53aab8c8"),
			expectError: false,
			expectBytes: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0xd2, 0x4a, 0xf8, 0x25,
				0xe3, 0x84, 0x95, 0xee, 0x36, 0x24, 0x66, 0xf2,
				0x14, 0x94, 0x6c, 0xdf, 0x53, 0xaa, 0xb8, 0xc8,
			},
		},
		{
			name:        "simple address c778417...",
			typeName:    "address",
			in:          MustNewAddress("c778417e063141139fce010982780140aa0cd5ab"),
			expectError: false,
			expectBytes: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0xc7, 0x78, 0x41, 0x7e,
				0x6, 0x31, 0x41, 0x13, 0x9f, 0xce, 0x1, 0x9,
				0x82, 0x78, 0x1, 0x40, 0xaa, 0xc, 0xd5, 0xab,
			},
		},
		{
			name:     "address array",
			typeName: "address[]",
			in: []Address{
				MustNewAddress("d24af825e38495ee362466f214946cdf53aab8c8"),
				MustNewAddress("c778417e063141139fce010982780140aa0cd5ab"),
				MustNewAddress("7d97ba95dac25316b9531152b3baa32327994da8"),
			},
			expectError: false,
			expectBytes: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03,

				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0xd2, 0x4a, 0xf8, 0x25,
				0xe3, 0x84, 0x95, 0xee, 0x36, 0x24, 0x66, 0xf2,
				0x14, 0x94, 0x6c, 0xdf, 0x53, 0xaa, 0xb8, 0xc8,

				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0xc7, 0x78, 0x41, 0x7e,
				0x6, 0x31, 0x41, 0x13, 0x9f, 0xce, 0x1, 0x9,
				0x82, 0x78, 0x1, 0x40, 0xaa, 0xc, 0xd5, 0xab,

				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x7d, 0x97, 0xba, 0x95,
				0xda, 0xc2, 0x53, 0x16, 0xb9, 0x53, 0x11, 0x52,
				0xb3, 0xba, 0xa3, 0x23, 0x27, 0x99, 0x4d, 0xa8,
			},
		},
		{
			name:        "simple uint256",
			typeName:    "uint256",
			in:          big.NewInt(2938),
			expectError: false,
			expectBytes: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0b, 0x7a,
			},
		},
		{
			name:        "bigger uint256",
			typeName:    "uint256",
			in:          big.NewInt(2317850009133627),
			expectError: false,
			expectBytes: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x08, 0x3c, 0x12, 0x82, 0x6f, 0xe2, 0x3b,
			},
		},
		{
			name:        "simple uint64",
			typeName:    "uint64",
			in:          uint64(10),
			expectError: false,
			expectBytes: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0a,
			},
		},
		{
			name:        "simple uint64",
			typeName:    "uint64",
			in:          uint64(2923872023),
			expectError: false,
			expectBytes: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0xae, 0x46, 0xbf, 0x17,
			},
		},
		{
			name:        "bytes",
			typeName:    "bytes",
			in:          []byte{0x01, 0x03, 0xaa, 0xbb, 0xcc},
			expectError: false,
			expectBytes: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x05,
				0x01, 0x03, 0xaa, 0xbb, 0xcc,
			},
		},
		{
			name:        "bool",
			typeName:    "bool",
			in:          true,
			expectError: false,
			expectBytes: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
			},
		},
		{
			name:        "bool",
			typeName:    "bool",
			in:          false,
			expectError: false,
			expectBytes: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := NewEncoder()
			err := e.Write(test.typeName, test.in)
			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectBytes, e.buffer)
			}
		})

	}
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
