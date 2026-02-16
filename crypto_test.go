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
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrivateKey_Generate63BytesPrivateKey(t *testing.T) {
	t.Skip("Used to generate a random private key < 32 bytes")

	for {
		privKey, err := NewRandomPrivateKey()
		require.NoError(t, err)

		ecdsaKey := privKey.inner.ToECDSA()

		if len(ecdsaKey.D.Bytes()) < 32 {
			fmt.Printf("D value (Hex %x, Text %s, Bytes %s) has less than 32 bytes\n", ecdsaKey.D, ecdsaKey.D.Text(10), hex.EncodeToString(ecdsaKey.D.Bytes()))
			require.NoError(t, errors.New("bytes < 32"))
		}
	}
}

func TestNewPrivateKey(t *testing.T) {
	tests := []struct {
		name           string
		in             string
		expectedString string
		expectedErr    string
	}{
		{
			name:           "valid hex without prefix",
			in:             "52e1cc4b9c8b4fc9b202adf06462bdcc248e170c9abd56b2adb84c8d87bee674",
			expectedString: "52e1cc4b9c8b4fc9b202adf06462bdcc248e170c9abd56b2adb84c8d87bee674",
		},
		{
			name:           "valid hex with 0x prefix",
			in:             "0x52e1cc4b9c8b4fc9b202adf06462bdcc248e170c9abd56b2adb84c8d87bee674",
			expectedString: "52e1cc4b9c8b4fc9b202adf06462bdcc248e170c9abd56b2adb84c8d87bee674",
		},
		{
			name:           "valid hex uppercase with 0x prefix",
			in:             "0x52E1CC4B9C8B4FC9B202ADF06462BDCC248E170C9ABD56B2ADB84C8D87BEE674",
			expectedString: "52e1cc4b9c8b4fc9b202adf06462bdcc248e170c9abd56b2adb84c8d87bee674",
		},
		{
			name:           "key < 32 bytes left padded",
			in:             "0033752648ef6373b2904568fc7452957b73bbb4f91657735bb53e633513b805",
			expectedString: "0033752648ef6373b2904568fc7452957b73bbb4f91657735bb53e633513b805",
		},
		{
			name:           "key < 32 bytes left padded with 0x prefix",
			in:             "0x0033752648ef6373b2904568fc7452957b73bbb4f91657735bb53e633513b805",
			expectedString: "0033752648ef6373b2904568fc7452957b73bbb4f91657735bb53e633513b805",
		},
		{
			name:        "invalid hex characters",
			in:          "gg33752648ef6373b2904568fc7452957b73bbb4f91657735bb53e633513b805",
			expectedErr: "invalid private key",
		},
		{
			name:        "too short",
			in:          "52e1cc4b9c8b4fc9b202adf06462bdcc248e170c9abd56b2adb84c8d87bee6",
			expectedErr: "not enough bytes",
		},
		{
			name:        "too long",
			in:          "52e1cc4b9c8b4fc9b202adf06462bdcc248e170c9abd56b2adb84c8d87bee67400",
			expectedErr: "not enough bytes",
		},
		{
			name:        "empty string",
			in:          "",
			expectedErr: "not enough bytes",
		},
		{
			name:        "only 0x prefix",
			in:          "0x",
			expectedErr: "not enough bytes",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := NewPrivateKey(test.in)
			if test.expectedErr == "" {
				require.NoError(t, err)
				require.NotNil(t, actual, "invalid private for input %q", test.in)
				assert.Equal(t, test.expectedString, actual.String())
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedErr)
			}
		})
	}
}

func TestPrivateKey_String(t *testing.T) {
	tests := []struct {
		in          string
		expectedErr error
	}{
		{in: "52e1cc4b9c8b4fc9b202adf06462bdcc248e170c9abd56b2adb84c8d87bee674", expectedErr: nil},
		{in: "2f6e6a9af650c60e4b2c6a0a1b440cec202182bed3fc15bc5b8eec1132b2d6ad", expectedErr: nil},
		{in: "cb3c1ca36610c116e9aa478102faeaf100cc79c462f0d25a631110e36a2868c8", expectedErr: nil},
		{in: "b1c73d7b1103387b725df649a70d8c2c3cca61a60781f19d0a91ced1ea1be35b", expectedErr: nil},
		{in: "72194c8757fde016fa20cbc1ec09b8e64413c4ad7aba23bdb12664311e03f419", expectedErr: nil},

		// Key < 32 bytes (64 characters) left padded here
		{in: "0033752648ef6373b2904568fc7452957b73bbb4f91657735bb53e633513b805", expectedErr: nil},
	}

	for _, test := range tests {
		t.Run(test.in, func(t *testing.T) {
			actual, err := NewPrivateKey(test.in)
			if test.expectedErr == nil {
				require.NoError(t, err)
				require.NotNil(t, actual, "invalid private for input %q", test.in)
				assert.Equal(t, test.in, actual.String())
			} else {
				assert.Equal(t, test.expectedErr, err)
			}
		})
	}
}

func TestPrivateKey_PublicKey(t *testing.T) {
	tests := []struct {
		in          string
		expected    string
		expectedErr error
	}{
		{
			in:          "52e1cc4b9c8b4fc9b202adf06462bdcc248e170c9abd56b2adb84c8d87bee674",
			expected:    "821b55d8abe79bc98f05eb675fdc50dfe796b7ab",
			expectedErr: nil,
		},
		{
			in:          "2f6e6a9af650c60e4b2c6a0a1b440cec202182bed3fc15bc5b8eec1132b2d6ad",
			expected:    "403a09cd493e41d381ebff4ffb01de4c9f2ff1dc",
			expectedErr: nil,
		},
		{
			in:          "cb3c1ca36610c116e9aa478102faeaf100cc79c462f0d25a631110e36a2868c8",
			expected:    "dfe15828305e9968835e7797f063bd8bbec038c0",
			expectedErr: nil,
		},
		{
			in:          "b1c73d7b1103387b725df649a70d8c2c3cca61a60781f19d0a91ced1ea1be35b",
			expected:    "ccd3dd1db00b7c5b493b879f04379287227d200d",
			expectedErr: nil,
		},
		{
			in:          "72194c8757fde016fa20cbc1ec09b8e64413c4ad7aba23bdb12664311e03f419",
			expected:    "e8c4b557fb717f7a366d4cb7f8fe41ef1e108c16",
			expectedErr: nil,
		},
		{
			in:          "72194c8757fde016fa20cbc1ec09b8e64413c4ad7aba23bdb12664311e03f419",
			expected:    "e8c4b557fb717f7a366d4cb7f8fe41ef1e108c16",
			expectedErr: nil,
		},

		// Key < 32 bytes (64 characters) left padded here
		{
			in:          "0033752648ef6373b2904568fc7452957b73bbb4f91657735bb53e633513b805",
			expected:    "43bf67eae59c0d0eca283839e5cd9b22ca89d530",
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.in, func(t *testing.T) {
			actual, err := NewPrivateKey(test.in)
			if test.expectedErr == nil {
				require.NoError(t, err)
				require.NotNil(t, actual, "invalid private for input %q", test.in)
				assert.Equal(t, test.expected, actual.PublicKey().Address().String())
			} else {
				assert.Equal(t, test.expectedErr, err)
			}
		})
	}
}

func TestPrivateKey_Bytes(t *testing.T) {
	tests := []struct {
		in          string
		expectedErr error
	}{
		{in: "52e1cc4b9c8b4fc9b202adf06462bdcc248e170c9abd56b2adb84c8d87bee674", expectedErr: nil},
		{in: "2f6e6a9af650c60e4b2c6a0a1b440cec202182bed3fc15bc5b8eec1132b2d6ad", expectedErr: nil},
		{in: "cb3c1ca36610c116e9aa478102faeaf100cc79c462f0d25a631110e36a2868c8", expectedErr: nil},
		{in: "b1c73d7b1103387b725df649a70d8c2c3cca61a60781f19d0a91ced1ea1be35b", expectedErr: nil},
		{in: "72194c8757fde016fa20cbc1ec09b8e64413c4ad7aba23bdb12664311e03f419", expectedErr: nil},

		// Key < 32 bytes (64 characters) left padded here
		{in: "0033752648ef6373b2904568fc7452957b73bbb4f91657735bb53e633513b805", expectedErr: nil},
	}

	for _, test := range tests {
		t.Run(test.in, func(t *testing.T) {
			actual, err := NewPrivateKey(test.in)

			if test.expectedErr == nil {
				require.NoError(t, err)
				require.NotNil(t, actual, "invalid private for input %q", test.in)

				actualBytes := hex.EncodeToString(actual.Bytes())
				assert.Equal(t, test.in, actualBytes)
			} else {
				assert.Equal(t, test.expectedErr, err)
			}
		})
	}
}

func TestPrivateKey_ToECDSA(t *testing.T) {
	tests := []struct {
		in          string
		expectedD   string
		expectedX   string
		expectedY   string
		expectedErr error
	}{
		{
			in:          "52e1cc4b9c8b4fc9b202adf06462bdcc248e170c9abd56b2adb84c8d87bee674",
			expectedX:   "2c8f6d4764c3aca75696e18aeef683932a2bfa0be1603adb54f30dfad8e5cf23",
			expectedY:   "72a9d6eeb0e5caffba1fca22e12878c450e6ef09434888f04c6a97b6f50c75d4",
			expectedErr: nil,
		},
		{
			in:          "2f6e6a9af650c60e4b2c6a0a1b440cec202182bed3fc15bc5b8eec1132b2d6ad",
			expectedX:   "73523c72e55da219c8650a9fe47f83477958a6185b2322b2c948b88e1dcea92",
			expectedY:   "c7e872d816c129b6cf51b05b9c5bc22ddd34fc7d7e35da17fcdbe5598fee56a1",
			expectedErr: nil,
		},
		{
			in:          "cb3c1ca36610c116e9aa478102faeaf100cc79c462f0d25a631110e36a2868c8",
			expectedX:   "b235c1450002e43d49f25de3b479aa7ce98c8b369ebf6ba9dcfee1e8e1ff1a1a",
			expectedY:   "d5f52b3d4347659c35d1d9a44e8bf0b1a70f52d99b85d716e36b74dfe602da26",
			expectedErr: nil,
		},
		{
			in:          "b1c73d7b1103387b725df649a70d8c2c3cca61a60781f19d0a91ced1ea1be35b",
			expectedX:   "ab23cd020e638a977d57907668bf4923108797a8f5860e4c412598bdc2f484b",
			expectedY:   "cda179d557bbcddc83174f47ec45b1e49bd9dbd6bcea26932fc3c25ade0c72b9",
			expectedErr: nil,
		},
		{
			in:          "72194c8757fde016fa20cbc1ec09b8e64413c4ad7aba23bdb12664311e03f419",
			expectedX:   "70203b56d3f40f1823f2de383bd90891d75c00b151e65f4b01d70202f38a9f8",
			expectedY:   "e7d116e7ad511340119b69208c9a78776a8b72ab87359e7a365f0d7f68e313bb",
			expectedErr: nil,
		},

		// Key < 32 bytes (64 characters) left padded here
		{
			in:          "0033752648ef6373b2904568fc7452957b73bbb4f91657735bb53e633513b805",
			expectedD:   "33752648ef6373b2904568fc7452957b73bbb4f91657735bb53e633513b805",
			expectedX:   "685ed2e19b95b5aee50709c579300f93960e3750a106859fd02e75ad2a9d481f",
			expectedY:   "f114535858b7b4856a6e62332599c94a5a751cb31aaea58fad6765c4cf8e786c",
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.in, func(t *testing.T) {
			actual, err := NewPrivateKey(test.in)
			if test.expectedErr == nil {
				require.NoError(t, err)
				require.NotNil(t, actual, "invalid private for input %q", test.in)

				ecdsaPk := actual.inner.ToECDSA()

				expectedD := test.expectedD
				if expectedD == "" {
					expectedD = test.in
				}

				assert.Equal(t, expectedD, ecdsaPk.D.Text(16), "private coordinate D")
				assert.Equal(t, test.expectedX, ecdsaPk.X.Text(16), "public point X")
				assert.Equal(t, test.expectedY, ecdsaPk.Y.Text(16), "public point Y")
			} else {
				assert.Equal(t, test.expectedErr, err)
			}
		})
	}
}

func TestPrivateKeySignPersonal(t *testing.T) {
	// Private key of public key 0xfffdb7377345371817f2b4dd490319755f5899ec
	priv, err := NewPrivateKey("db4c20e40f4049efa3c0d3added58dc171ccda274a96a9b9313b305a22841a5d")
	require.NoError(t, err)

	// This is exercised in `eth-go/tests/src/PersonalSigning.sol` (and `eth-go/tests/src/test/PersonalSigning.sol`) in `testRecoverPersonalSigner`
	signature, err := priv.SignPersonal(MustNewHex("0xcf36ac4f97dc10d91fc2cbb20d718e94a8cbfe0f82eaedc6a4aa38946fb797cd"))
	require.NoError(t, err)
	require.Equal(t,
		"cfc0c9160c1fbe884f02298b76e194611904a4f18b6814dfd22f95b659b2f90b2564e455543316f125c3c51ec72b22917401d4872ddabb9a7a2d0bea1a827db31b",
		signature.ToInverted().String(),
	)
}

func TestKeccak256(t *testing.T) {
	hexData := MustNewHex("00000000000000000000000000000000000000000000000000000000000000ab00000000000000000000000000000000000000000000000000000000000000bc000000000000000000000000000000000000000000000000000000000000009f")

	tests := []struct {
		in        string
		expectOut string
	}{
		{
			in:        "Pregnant(address,uint256,uint256,uint256)",
			expectOut: "241ea03ca20251805084d27d4440371c34a0b85ff108f6bb5611248f73818b80",
		},
		{
			in:        "Transfer(address,address,uint256)",
			expectOut: "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
		},
		{
			in:        "Approval(address,address,uint256)",
			expectOut: "8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925",
		},
		{
			in:        "Birth(address,uint256,uint256,uint256,uint256)",
			expectOut: "0a5311bd2a6608f08a180df2ee7c5946819a649b204b554bb8e39825b2c50ad5",
		},
		{
			in:        string(hexData),
			expectOut: "29f854483f7f0bbfe56b2e12b8da6cc2caf951abb1777b75bace664018c1085a",
		},
	}

	for _, test := range tests {
		t.Run(test.in, func(t *testing.T) {
			out := Keccak256([]byte(test.in))
			assert.Equal(t, test.expectOut, hex.EncodeToString(out))
		})
	}
}

func TestSignature(t *testing.T) {
	type args struct {
		in string
	}

	type outs struct {
		R              string
		S              string
		V              byte
		String         string
		InvertedString string
	}

	tests := []struct {
		name      string
		args      args
		wantOut   outs
		assertion require.ErrorAssertionFunc
	}{
		{
			"standard",
			args{"1bcfc0c9160c1fbe884f02298b76e194611904a4f18b6814dfd22f95b659b2f90b2564e455543316f125c3c51ec72b22917401d4872ddabb9a7a2d0bea1a827db3"},
			outs{
				R:              "cfc0c9160c1fbe884f02298b76e194611904a4f18b6814dfd22f95b659b2f90b",
				S:              "2564e455543316f125c3c51ec72b22917401d4872ddabb9a7a2d0bea1a827db3",
				V:              0x1b,
				String:         "1bcfc0c9160c1fbe884f02298b76e194611904a4f18b6814dfd22f95b659b2f90b2564e455543316f125c3c51ec72b22917401d4872ddabb9a7a2d0bea1a827db3",
				InvertedString: "cfc0c9160c1fbe884f02298b76e194611904a4f18b6814dfd22f95b659b2f90b2564e455543316f125c3c51ec72b22917401d4872ddabb9a7a2d0bea1a827db31b",
			},
			require.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in, err := hex.DecodeString(tt.args.in)
			require.NoError(t, err)

			gotOut, err := NewSignatureFromBytes(in)
			tt.assertion(t, err)

			assert.Equal(t, tt.wantOut.R, gotOut.R().Text(16))
			assert.Equal(t, tt.wantOut.S, gotOut.S().Text(16))
			assert.Equal(t, tt.wantOut.V, gotOut.V())
			assert.Equal(t, tt.wantOut.String, gotOut.String())

			inverted := gotOut.ToInverted()
			assert.Equal(t, tt.wantOut.InvertedString, inverted.String())
			assert.Equal(t, tt.wantOut.R, inverted.R().Text(16))
			assert.Equal(t, tt.wantOut.S, inverted.S().Text(16))
			assert.Equal(t, tt.wantOut.V, inverted.V())
		})
	}
}

// func TestRecoverPersonalSigner(t *testing.T) {
// 	type args struct {
// 		signature eth.Hex
// 		hash      eth.Hash
// 	}

// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    eth.Address
// 		wantErr bool
// 	}{
// 		{
// 			"standard",
// 			args{
// 				signature: eth.MustNewHex("cfc0c9160c1fbe884f02298b76e194611904a4f18b6814dfd22f95b659b2f90b2564e455543316f125c3c51ec72b22917401d4872ddabb9a7a2d0bea1a827db31b"),
// 				hash:      eth.MustNewHash("0xcf36ac4f97dc10d91fc2cbb20d718e94a8cbfe0f82eaedc6a4aa38946fb797cd"),
// 			},
// 			eth.MustNewAddress("0xfffdb7377345371817f2b4dd490319755f5899ec"),
// 			false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := RecoverPersonalSigner(tt.args.signature, tt.args.hash)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("RecoverPersonalSigner() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}

// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("RecoverPersonalSigner() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func TestSignature_RecoverPersonal(t *testing.T) {
	type args struct {
		signingData string
	}

	tests := []struct {
		name      string
		signature string
		args      args
		want      string
		assertion assert.ErrorAssertionFunc
	}{
		{
			"standard",
			"cfc0c9160c1fbe884f02298b76e194611904a4f18b6814dfd22f95b659b2f90b2564e455543316f125c3c51ec72b22917401d4872ddabb9a7a2d0bea1a827db31b",
			args{
				signingData: "0xcf36ac4f97dc10d91fc2cbb20d718e94a8cbfe0f82eaedc6a4aa38946fb797cd",
			},
			"0xfffdb7377345371817f2b4dd490319755f5899ec",
			assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signatureBytes, err := hex.DecodeString(tt.signature)
			require.NoError(t, err)

			signature, err := NewInvertedSignatureFromBytes(signatureBytes)
			require.NoError(t, err)

			got, err := signature.ToSignature().RecoverPersonal(MustNewHex(tt.args.signingData))
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got.Pretty())
		})
	}
}
