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
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"math/big"
	"strconv"

	"github.com/btcsuite/btcd/btcec/v2"
	btcecdsa "github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"golang.org/x/crypto/sha3"
)

// KeyBag holds private keys in memory, for signing transactions.
type KeyBag struct {
	Keys []*PrivateKey `json:"keys"`
}

func NewKeyBag() *KeyBag {
	return &KeyBag{
		Keys: make([]*PrivateKey, 0),
	}
}

type PublicKey struct {
	inner *secp256k1.PublicKey
}

func NewPublicKeyFromECDSA(key *secp256k1.PublicKey) *PublicKey {
	return &PublicKey{inner: key}
}

func (p PublicKey) Address() Address {
	return pubkeyToAddress(p.inner)
}

type PrivateKey struct {
	inner *secp256k1.PrivateKey
}

func NewRandomPrivateKey() (*PrivateKey, error) {
	privateKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return nil, err
	}

	return &PrivateKey{inner: privateKey}, nil
}

func NewPrivateKey(rawPrivateKey string) (*PrivateKey, error) {
	keyBytes, err := newByteSlice("private key", rawPrivateKey)
	if err != nil {
		return nil, err
	}

	return privateKeyFromRawBytes(keyBytes)
}

func privateKeyFromRawBytes(privateKeyBytes []byte) (*PrivateKey, error) {
	if len(privateKeyBytes) != btcec.PrivKeyBytesLen {
		return nil, fmt.Errorf("not enough bytes, got %d bytes but secp256k1 private key must have %d bytes",
			len(privateKeyBytes), btcec.PrivKeyBytesLen)
	}

	privKey, _ := btcec.PrivKeyFromBytes(privateKeyBytes)
	return &PrivateKey{inner: (*secp256k1.PrivateKey)(privKey)}, nil
}

// NewPrivateKeyFromECDSA creates a PrivateKey from standard library ecdsa.PrivateKey
// This allows using existing ecdsa keys with eth-go signing functions.
//
// Example:
//
//	ecdsaKey, _ := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
//	ethKey := eth.NewPrivateKeyFromECDSA(ecdsaKey)
//	sig, _ := ethKey.Sign(messageHash)
func NewPrivateKeyFromECDSA(key *ecdsa.PrivateKey) (*PrivateKey, error) {
	if key == nil {
		return nil, fmt.Errorf("ecdsa private key is nil")
	}

	// Validate curve is secp256k1
	if key.Curve != btcec.S256() {
		return nil, fmt.Errorf("only secp256k1 curve is supported")
	}

	// Convert D to 32-byte representation
	privKeyBytes := key.D.Bytes()
	if len(privKeyBytes) > 32 {
		return nil, fmt.Errorf("invalid private key: too many bytes")
	}

	// Pad to 32 bytes if needed (big-endian)
	if len(privKeyBytes) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(privKeyBytes):], privKeyBytes)
		privKeyBytes = padded
	}

	return privateKeyFromRawBytes(privKeyBytes)
}

func (p *PrivateKey) String() string {
	return hex.EncodeToString(p.Bytes())
}

func (p *PrivateKey) Bytes() (out []byte) {
	return p.inner.Serialize()
}

// Sign generates the signature for the according message hash based on this private key
// using ECDSA signature rules.
//
// See Signature documentation for more info about return signature format.
func (p *PrivateKey) Sign(messageHash Hash) (out Signature, err error) {
	compressedSignature := btcecdsa.SignCompact(p.inner, messageHash, false)

	copy(out[:], compressedSignature)
	return out, nil
}

var messagePrefix = []byte("\x19Ethereum Signed Message:\n")

// SignPersonal computes the correct message from `signingData` according to [ERC-712](https://eips.ethereum.org/EIPS/eip-712)
// which is briefly `keccak256(bytesOf("\x19Ethereum Signed Message:\n") + bytesOf(toString(len(signingData))) + signingData)`.
//
// This computed generated hash is then pass directly to `privateKey.Sign(personalMessageHash)`.
//
// See Sign for more details.
func (p *PrivateKey) SignPersonal(signingData Hex) (out Signature, err error) {
	return p.Sign(computePersonalMessageHash(signingData))
}

func computePersonalMessageHash(signingData Hex) Hash {
	lengthString := strconv.FormatUint(uint64(len(signingData)), 10)
	data := make([]byte, len(messagePrefix)+len(lengthString)+len(signingData))

	copy(data, messagePrefix)
	copy(data[len(messagePrefix):], []byte(lengthString))
	copy(data[len(messagePrefix)+len(lengthString):], signingData)

	return Keccak256(data)
}

func (p *PrivateKey) MarshalJSON() ([]byte, error) {
	// The `p.String()` is guaranteed to returns only hex characters, so it's safe to wrap directly with `"` symbols
	return []byte(`"` + p.String() + `"`), nil
}

func (p *PrivateKey) UnmarshalJSON(v []byte) (err error) {
	var s string
	if err := json.Unmarshal(v, &s); err != nil {
		return err
	}

	newPrivKey, err := NewPrivateKey(s)
	if err != nil {
		return fmt.Errorf("invalid private key: %w", err)
	}

	*p = *newPrivKey
	return
}

func (p *PrivateKey) PublicKey() *PublicKey {
	return &PublicKey{inner: p.inner.PubKey()}
}

// ToECDSA converts this PrivateKey to a standard library ecdsa.PrivateKey
// This allows using eth-go keys with standard crypto/ecdsa functions.
//
// Example:
//
//	ethKey, _ := eth.NewRandomPrivateKey()
//	ecdsaKey := ethKey.ToECDSA()
//	// Use with standard library functions
func (p *PrivateKey) ToECDSA() *ecdsa.PrivateKey {
	return &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: btcec.S256(),
			X:     p.inner.PubKey().X(),
			Y:     p.inner.PubKey().Y(),
		},
		D: new(big.Int).SetBytes(p.inner.Serialize()),
	}
}

// Signature represents a btcec Signature as computed from ecdsa.SignCompact(), this signature
// is in packed form of 65 bytes with ordered V (1 byte) + R (32 bytes) + S (32 bytes).
//
// The components can be retrieved with `R()`, `S()` and `V()`.
type Signature [65]byte

func NewSignatureFromBytes(in []byte) (out Signature, err error) {
	if len(in) != 65 {
		return out, fmt.Errorf("expected signature to have 65 bytes but input has %d byte(s)", len(in))
	}

	copy(out[:], in[0:65])
	return
}

// NewSignatureFromComponents creates a Signature from V, R, S components.
// This is useful when signature components are stored separately and need
// to be reconstructed for recovery.
//
// Signature format: [V (1 byte)][R (32 bytes)][S (32 bytes)]
//
// Example:
//
//	// Components from storage/protobuf/etc
//	v := byte(0)
//	r := [32]byte{...}
//	s := [32]byte{...}
//
//	sig := eth.NewSignatureFromComponents(v, r, s)
//	address, err := sig.Recover(messageHash)
func NewSignatureFromComponents(v byte, r, s [32]byte) Signature {
	var sig Signature
	sig[0] = v
	copy(sig[1:33], r[:])
	copy(sig[33:65], s[:])
	return sig
}

// ToInverted returns the InvertedSignature version of this Signature, this is
// that the components are ordered as `R`, `S` then `V` in the inverted version.
//
// This form is used on certain Ethereum construct like when doing a personal signing
// where the `V` component must be the last component of the signature for correct
// recovery.
func (s Signature) ToInverted() (out InvertedSignature) {
	copy(out[:], s[1:65])
	out[64] = s[0]

	return
}

func (s Signature) R() *big.Int {
	return new(big.Int).SetBytes(s[1:33])
}

func (s Signature) S() *big.Int {
	return new(big.Int).SetBytes(s[33:])
}

// V returns the recovery ID according to Bitcoin rules for the signature recovery.
// Ethereum augmented recovery ID to protect agaisnt replay attacks is **not**
// applied here.
//
// See https://bitcoin.stackexchange.com/a/38909 for extra details
func (s Signature) V() byte {
	return byte(s[0])
}

// RBytes returns the R component of the signature as a 32-byte array.
// Signature format is [V (1 byte)][R (32 bytes)][S (32 bytes)]
//
// Example:
//
//	sig, _ := privateKey.Sign(messageHash)
//	r := sig.RBytes()  // Extract R component as bytes
func (s Signature) RBytes() [32]byte {
	var r [32]byte
	copy(r[:], s[1:33])
	return r
}

// SBytes returns the S component of the signature as a 32-byte array.
// Signature format is [V (1 byte)][R (32 bytes)][S (32 bytes)]
//
// Example:
//
//	sig, _ := privateKey.Sign(messageHash)
//	sVal := sig.SBytes()  // Extract S component as bytes
func (s Signature) SBytes() [32]byte {
	var sVal [32]byte
	copy(sVal[:], s[33:65])
	return sVal
}

func (s Signature) Recover(messageHash Hash) (Address, error) {
	publicKey, compressed, err := btcecdsa.RecoverCompact(s[:], messageHash)
	if err != nil {
		return nil, fmt.Errorf("ecdsa recover compact: %w", err)
	}

	// Original key was compressed, is it possible in our usage? For now, just ignore it
	_ = compressed

	return NewPublicKeyFromECDSA(publicKey).Address(), nil
}

func (s Signature) RecoverPersonal(signingData Hex) (Address, error) {
	return s.Recover(computePersonalMessageHash(signingData))
}

func (s Signature) String() string {
	return hex.EncodeToString(s[:])
}

// InvertedSignature represents a standard Signature but the order of component
// `V` is inverted, being the last byte of the bytes (where it's the first byte in the
// standard `btcec` Signature).
//
// The InverteSignature is in packed form of 65 bytes and order of the components is
// R (32 bytes) + S (32 bytes) + V (1 byte).
//
// The components can be retrieved with `R()`, `S()` and `V()`.
//
// This form is used on certain Ethereum construct like when doing a personal signing
// where the `V` component must be the last component of the signature for correct
// recovery.
type InvertedSignature [65]byte

func NewInvertedSignatureFromBytes(in []byte) (out InvertedSignature, err error) {
	if len(in) != 65 {
		return out, fmt.Errorf("expected inverted signature to have 65 bytes but input has %d byte(s)", len(in))
	}

	copy(out[:], in[0:65])
	return
}

func (s InvertedSignature) ToSignature() (out Signature) {
	out[0] = s[64]
	copy(out[1:], s[0:64])

	return
}

// R returns the R component of signature.
func (s InvertedSignature) R() *big.Int {
	return new(big.Int).SetBytes(s[0:32])
}

// S returns the R component of signature.
func (s InvertedSignature) S() *big.Int {
	return new(big.Int).SetBytes(s[32:64])
}

// V returns the recovery ID according to Bitcoin rules for the signature recovery.
// Ethereum augmented recovery ID to protect agaisnt replay attacks is **not**
// applied here.
//
// See https://bitcoin.stackexchange.com/a/38909 for extra details
func (s InvertedSignature) V() byte {
	return byte(s[64])
}

// RecoverPersonal is a shortcut method for `signature.ToSignature().Recover(messageHash)`.
func (s InvertedSignature) Recover(messageHash Hash) (Address, error) {
	return s.ToSignature().Recover(messageHash)
}

// RecoverPersonal is a shortcut method for `signature.ToSignature().RecoverPersonal(signingData)`.
func (s InvertedSignature) RecoverPersonal(signingData Hex) (Address, error) {
	return s.ToSignature().RecoverPersonal(signingData)
}

func (s InvertedSignature) String() string {
	return hex.EncodeToString(s[:])
}

type keccakState interface {
	hash.Hash
	Read([]byte) (int, error)
}

func Keccak256(data ...[]byte) []byte {
	b := make([]byte, 32)
	d := sha3.NewLegacyKeccak256().(keccakState)
	for _, b := range data {
		d.Write(b)
	}
	d.Read(b)
	return b
}

func pubkeyToAddress(p *secp256k1.PublicKey) Address {
	if p == nil {
		return nil
	}

	pubBytes := elliptic.Marshal(btcec.S256(), p.X(), p.Y())
	return Address(Keccak256(pubBytes[1:])[12:])
}
