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
	"crypto/elliptic"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"math/big"
	"strconv"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
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
	compressedSignature := ecdsa.SignCompact(p.inner, messageHash, false)

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

func (s Signature) Recover(messageHash Hash) (Address, error) {
	publicKey, compressed, err := ecdsa.RecoverCompact(s[:], messageHash)
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
