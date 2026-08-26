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

package rpc

import (
	"encoding/json"
	"fmt"
	"strings"
)

type ErrorCode int

func (c ErrorCode) MarshalJSONRPC() ([]byte, error) {
	return json.Marshal(int(c))
}

type ErrResponse struct {
	Code    ErrorCode   `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitzero,omitempty"`
}

func (e *ErrResponse) Error() string {
	return fmt.Sprintf("rpc error (code %d): %s", e.Code, e.Message)
}

// Try to check if the call was reverted. The JSON-RPC response for reverts is
// not standardized, so we have ad-hoc checks for each of Geth, Parity and
// Ganache.
// These come from https://github.com/graphprotocol/graph-node/blob/94e93b07554d6e77aa618ac2f0716b70e0af671c/chain/ethereum/src/ethereum_adapter.rs#L500

const JSON_RPC_INVALID_REQUEST_ERROR = -32600
const JSON_RPC_INVALID_ARGUMENT_ERROR = -32602

var GETH_DETERMINISTIC_ERRORS = []string{
	"revert",
	"invalid jump destination",
	"invalid opcode",
	"stack limit reached 1024",
	"stack underflow (",
	"gas uint64 overflow",
	// See https://github.com/graphprotocol/graph-node/blob/94e93b07554d6e77aa618ac2f0716b70e0af671c/chain/ethereum/src/ethereum_adapter.rs#L511 for some reasoning
	// why `out of gas` is considered deterministic. One thing to remember, from an operator perspective, a minimal amount of `gasCap` is required on the node that
	// is handling `eth_call`, it must be equal or greater than `gasLimit` we use for the call.
	"out of gas",
	"evm error: invalidfeopcode",
}

var RETH_DETERMINISTIC_ERRORS = []string{
	"revert",
	"invalidjump",
	"opcodenotfound",
	"stackoverflow",
	"stackunderflow",
	"outofgas",
	"invalidfeopcode",
}

const PARITY_BAD_INSTRUCTION_FE = "Bad instruction fe"
const PARITY_BAD_INSTRUCTION_FD = "Bad instruction fd"
const PARITY_BAD_JUMP_PREFIX = "Bad jump"
const PARITY_STACK_LIMIT_PREFIX = "Out of stack"
const PARITY_OUT_OF_GAS = "Out of gas"
const PARITY_OUT_OF_BOUND = "return data out of bounds"
const PARITY_VM_EXECUTION_ERROR = -32015
const PARITY_REVERT_PREFIX = "Reverted 0x"

// const PARITY_OUT_OF_GAS = "Out of gas" // same as geth
// const XDAI_REVERT = "revert" // same as geth

const GANACHE_VM_EXECUTION_ERROR = -32000
const GANACHE_REVERT_MESSAGE = "VM Exception while processing transaction: revert"

// IsDeterministicError determines if an error is deterministic or not according to
// some heuristic based on the error's message.
//
// The rules used here are the one used by `graph-node` software to determine if the
// error is determinsitic or not.
//
// See https://github.com/graphprotocol/graph-node/blob/581ff5cf2978af66c49c91d4bf819b6647173776/chain/ethereum/src/ethereum_adapter.rs#L486
func IsDeterministicError(err *ErrResponse) bool {
	return IsGenericDeterministicError(err) ||
		IsGethDeterministicError(err) ||
		IsRethDeterministicError(err) ||
		IsParityDeterministicError(err) ||
		IsGanacheDeterministicError(err)
}

func IsGenericDeterministicError(err *ErrResponse) bool {
	return err.Code == JSON_RPC_INVALID_ARGUMENT_ERROR || err.Code == JSON_RPC_INVALID_REQUEST_ERROR
}

func IsGethDeterministicError(err *ErrResponse) bool {
	msg := strings.ToLower(err.Message)
	for _, e := range GETH_DETERMINISTIC_ERRORS {
		if strings.Contains(msg, e) {
			return true
		}
	}
	return false
}

func IsRethDeterministicError(err *ErrResponse) bool {
	msg := strings.ToLower(err.Message)
	for _, e := range RETH_DETERMINISTIC_ERRORS {
		if strings.Contains(msg, e) {
			return true
		}
	}
	return false
}

func IsParityDeterministicError(err *ErrResponse) bool {
	if err.Code == PARITY_VM_EXECUTION_ERROR {
		return true
	}

	if err.Message == PARITY_OUT_OF_GAS {
		return true
	}

	if err.Code == -32000 && err.Message == PARITY_OUT_OF_BOUND {
		return true
	}

	if err.Message == PARITY_BAD_INSTRUCTION_FD ||
		err.Message == PARITY_BAD_INSTRUCTION_FE {
		return true
	}

	if strings.HasPrefix(err.Message, PARITY_BAD_JUMP_PREFIX) ||
		strings.HasPrefix(err.Message, PARITY_STACK_LIMIT_PREFIX) ||
		strings.HasPrefix(err.Message, PARITY_REVERT_PREFIX) {
		return true
	}

	return false
}

func IsGanacheDeterministicError(err *ErrResponse) bool {
	return err.Code == GANACHE_VM_EXECUTION_ERROR && strings.HasPrefix(err.Message, GANACHE_REVERT_MESSAGE)
}
