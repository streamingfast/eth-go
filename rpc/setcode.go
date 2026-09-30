package rpc

import (
	"fmt"
	"math/big"

	"github.com/holiman/uint256"
	"github.com/streamingfast/eth-go"
	"github.com/streamingfast/eth-go/rlp"
)

// setCodeAuthorizationMagic is the EIP-7702 domain separator byte prepended to the RLP payload
// before hashing, see [SetCodeAuthorization.Authority].
const setCodeAuthorizationMagic = 0x05

type setCodeAuthorizationPayload struct {
	ChainID *big.Int
	Address []byte
	Nonce   uint64
}

// Authority recovers the address that signed this EIP-7702 set-code authorization tuple, per
// https://eips.ethereum.org/EIPS/eip-7702:
//
//	authority = ecrecover(keccak(MAGIC || rlp([chain_id, address, nonce])), y_parity, r, s)
//
// It returns an error when the signature cannot be recovered, which legitimately happens for
// authorizations the chain discarded because of an invalid (e.g. all-zero) signature.
func (a SetCodeAuthorization) Authority() (eth.Address, error) {
	payload, err := rlp.Encode(&setCodeAuthorizationPayload{
		ChainID: new(big.Int).SetUint64(uint64(a.ChainID)),
		Address: []byte(a.Address),
		Nonce:   uint64(a.Nonce),
	})
	if err != nil {
		return nil, fmt.Errorf("rlp encode authorization payload: %w", err)
	}

	hash := eth.Keccak256(append([]byte{setCodeAuthorizationMagic}, payload...))

	var sigBytes [65]byte
	rBytes := (*uint256.Int)(a.R).Bytes32()
	sBytes := (*uint256.Int)(a.S).Bytes32()
	copy(sigBytes[0:32], rBytes[:])
	copy(sigBytes[32:64], sBytes[:])
	// btcec's RecoverCompact, which backs Signature.Recover, expects the recovery byte biased by
	// Bitcoin convention (27 + recoveryID), not the raw 0/1 y_parity EIP-7702 signs over.
	sigBytes[64] = 27 + byte(a.YParity)

	sig, err := eth.NewInvertedSignatureFromBytes(sigBytes[:])
	if err != nil {
		return nil, fmt.Errorf("build signature: %w", err)
	}

	address, err := sig.Recover(hash)
	if err != nil {
		return nil, fmt.Errorf("recover authority: %w", err)
	}

	return address, nil
}
