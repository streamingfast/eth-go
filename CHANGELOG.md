# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Added the standard JSON-RPC 2.0 error codes to the `rpc` package, typed as `rpc.ErrorCode`: `rpc.ErrorCodeParseError` (`-32700`), `rpc.ErrorCodeInvalidRequest` (`-32600`), `rpc.ErrorCodeMethodNotFound` (`-32601`), `rpc.ErrorCodeInvalidParams` (`-32602`), `rpc.ErrorCodeInternalError` (`-32603`) as well as `rpc.ErrorCodeServerError` (`-32000`) for the implementation defined case where the call reached a handler and the handler itself failed.

  The pre-existing `rpc.JSON_RPC_INVALID_REQUEST_ERROR` and `rpc.JSON_RPC_INVALID_ARGUMENT_ERROR` constants are unchanged and keep working, `rpc.ErrorCodeInvalidRequest` and `rpc.ErrorCodeInvalidParams` are the preferred spelling in new code.

- Added `rpc.ErrResponse` constructors so a JSON-RPC server can build errors without assembling a struct literal, the message being formatted according to `fmt.Sprintf` rules: `rpc.NewErrResponse(code, format, args...)`, `rpc.NewParseError`, `rpc.NewInvalidRequestError`, `rpc.NewMethodNotFoundError`, `rpc.NewInvalidParamsError`, `rpc.NewInternalError` and `rpc.NewServerError`.

### Changed

- **Breaking** `eth.MustNewAddress` is now strict in argument it accepts, the received input must have exactly 20 bytes once decoded. You can find back the previous behavior by using `MustNewAddressLoose` that has been added.

- **Breaking** `eth.NewAddress` is now strict in argument it accepts, the received input must have exactly 20 bytes once decoded. You can find back the previous behavior by using `NewAddressLoose` that has been added.

- JSON-RPC code `-32602` is now treated as a deterministic error, except when the message indicates state/block unavailability (see `NON_DETERMINISTIC_STATE_MESSAGES` in the Fixed section).

- **Breaking** The `ABI` has changed so that multiple events/functions of the name or same id are parsed correctly, in order defined.

### Deprecated

- _Deprecated_ `ABI#FindLog` replaced with `ABI#FindLogByTopic`.

- _Deprecated_ `ABI#FindFunction` replaced with `ABI#FindFunctionByHash`.

### Fixed

- Fixed `rpc.IsGenericDeterministicError` (and thus `rpc.IsDeterministicError`) misclassifying `-32602` state/block unavailability errors (e.g. `Block requested not found ... historical state that is not available`, `missing trie node`) as deterministic. These are transient, node-capability dependent conditions (pruned node) that an archive node answers correctly, so they must not be cached permanently. A new `rpc.NON_DETERMINISTIC_STATE_MESSAGES` list excludes them.

- Fixed `Uint256.MarshalText()` to return hex-encoded strings (e.g., `0x1234...`) instead of decimal strings, matching Ethereum conventions and the behavior of `MarshalJSONRPC()`.

- Fixed `rpc.Do[T]` failing when `T` is `string`. The function now returns the raw result string directly without attempting JSON unmarshal for string types, since `DoRequest` already returns the unquoted string value.

- Fixed function selector (method ID) computation for nested tuples. Previously, nested tuples in function signatures were not recursively expanded, causing the literal string `"tuple"` to be used instead of the expanded type signature. For example, a function with a `SignedRAV` parameter containing a nested `RAV` tuple would incorrectly produce `recoverRAVSigner((tuple,bytes))` instead of the correct `recoverRAVSigner(((bytes32,address,address,address,uint64,uint128,bytes),bytes))`. This caused incorrect function selectors to be computed, leading to contract call failures.

- Fixed event topic ID computation for nested tuples. Similar to the function selector fix, event signatures now correctly expand nested tuple types.

- Fixed nested tuple encoding when using `map[string]interface{}` or `[]interface{}` input types. Previously, encoding nested tuples (tuples containing other tuples) would fail with an error like `struct "<unknown>" has 0 fields` because the ABI parser wasn't recursively processing nested component definitions. The `structComponent` type now includes a `Components` field and `toStructComponents` recursively populates it during ABI parsing.

- Fixed ABI encoding/decoding of tuples with dynamic components (e.g., `bytes`, `string`, or dynamic arrays). Per Solidity ABI spec, a tuple is dynamic if any of its components is dynamic. Previously, `isDynamicType()` only considered `bytes`, `string`, and arrays as dynamic, but not tuples containing dynamic fields. This caused incorrect encoding (inline instead of offset-based) for such tuples.

- `rpc.Block#BaseFee` is now correctly a `*eth.Uint256` value, you can use `(*uint256.Int)(block.BaseFee).Uint64()` to get back the `uint64` value again (you should check for `nil` value though because it **can** be `nil`).

- Encoding of `bytes` in ABI format wasn't properly left padding.

- `LogEventDef.Signature()` is now formatted to be read for `Keccak` processing

- `rpc.Block.Nonce` is now a FixedUint64 to enforce `0x0000000000000000` encoding.

- `rpc.Block.Timestamp` is now encoded as a `uint64` instead of a time.RFC3339 string

- `rpc.BlockRef` decoding fixed to support either a `BlockNumber` or a `BlockHash`

### Added

- Added support for constructor encoding in ABI. New `ConstructorDef` type represents contract constructors, with `NewCall()` method to create `ConstructorCall` instances that can be encoded. Unlike method calls, constructor encoding does not include a 4-byte method selector - it's just the ABI-encoded parameters.

- Added `ABI#FindConstructor()` to retrieve the first constructor from the ABI.

- Added `ABI#FindConstructors()` to retrieve all constructors from the ABI.

- Added `ABI#FindConstructorBySignature(signature string)` to find a constructor by its parameter signature (e.g., `"(address,uint256)"`).

- Added `NewAddressLoose` that accepts address that might contain less or more than 20 bytes.

- Added improved type information on `LogEventDef`.

- Added improved type information on `MethodDef`.

- Added improved type information on `StructComponent`.

- Added `Components` field to `StructComponent` to support nested tuple types.

- Added `StructComponent.Signature()` method that returns the canonical type signature with recursive tuple expansion.

- Added `LogParameter.Signature()` method that returns the canonical type signature with recursive tuple expansion.

- Added `ABI#FindLogsByTopic` to find all logs with a given topic.

- Added `ABI#FindLogsByName` to find all logs with a given name.

- Added `ABI#FindFunctionsByName` to find all functions with a given name.

- Added `out of gas` and `Out of gas` as deterministic error (with the constraint that all provider of `eth_call` used have a `gasCap` configured >= than `gasLimit` used for a call, which should be fixed).

- `LogEventDef.LogID()` is now exposed publicly

- Added `NewPrivateKeyFromECDSA(*ecdsa.PrivateKey)` to create an eth-go PrivateKey from a standard library ecdsa.PrivateKey, enabling integration with existing ECDSA key infrastructure.

- Added `(*PrivateKey).ToECDSA()` to convert an eth-go PrivateKey to a standard library ecdsa.PrivateKey for use with standard crypto/ecdsa functions.

- Added `Signature.RBytes()` method that returns the R component of a signature as a 32-byte array, complementing the existing `R() *big.Int` method.

- Added `Signature.SBytes()` method that returns the S component of a signature as a 32-byte array, complementing the existing `S() *big.Int` method.

- Added `NewSignatureFromComponents(v byte, r, s [32]byte)` to construct a Signature from individual V, R, S components, useful when signature components are stored separately (e.g., in protobuf messages or database schemas).

[unreleased]: https://github.com/streamingfast/eth-go
