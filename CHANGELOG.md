# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- **Breaking** `eth.MustNewAddress` is now strict in argument it accepts, the received input must have exactly 20 bytes once decoded. You can find back the previous behavior by using `MustNewAddressLoose` that has been added.

- **Breaking** `eth.NewAddress` is now strict in argument it accepts, the received input must have exactly 20 bytes once decoded. You can find back the previous behavior by using `NewAddressLoose` that has been added.

- JSON-RPC code `-32602` is now treated as a deterministic error.

- **Breaking** The `ABI` has changed so that multiple events/functions of the name or same id are parsed correctly, in order defined.

### Deprecated

- _Deprecated_ `ABI#FindLog` replaced with `ABI#FindLogByTopic`.

- _Deprecated_ `ABI#FindFunction` replaced with `ABI#FindFunctionByHash`.

### Fixed

- Fixed nested tuple encoding when using `map[string]interface{}` or `[]interface{}` input types. Previously, encoding nested tuples (tuples containing other tuples) would fail with an error like `struct "<unknown>" has 0 fields` because the ABI parser wasn't recursively processing nested component definitions. The `structComponent` type now includes a `Components` field and `toStructComponents` recursively populates it during ABI parsing.

- Fixed ABI encoding/decoding of tuples with dynamic components (e.g., `bytes`, `string`, or dynamic arrays). Per Solidity ABI spec, a tuple is dynamic if any of its components is dynamic. Previously, `isDynamicType()` only considered `bytes`, `string`, and arrays as dynamic, but not tuples containing dynamic fields. This caused incorrect encoding (inline instead of offset-based) for such tuples.

- `rpc.Block#BaseFee` is now correctly a `*eth.Uint256` value, you can use `(*uint256.Int)(block.BaseFee).Uint64()` to get back the `uint64` value again (you should check for `nil` value though because it **can** be `nil`).

- Encoding of `bytes` in ABI format wasn't properly left padding.

- `LogEventDef.Signature()` is now formatted to be read for `Keccak` processing

- `rpc.Block.Nonce` is now a FixedUint64 to enforce `0x0000000000000000` encoding.

- `rpc.Block.Timestamp` is now encoded as a `uint64` instead of a time.RFC3339 string

- `rpc.BlockRef` decoding fixed to support either a `BlockNumber` or a `BlockHash`

### Added

- Added `NewAddressLoose` that accepts address that might contain less or more than 20 bytes.

- Added improved type information on `LogEventDef`.

- Added improved type information on `MethodDef`.

- Added improved type information on `StructComponent`.

- Added `Components` field to `StructComponent` to support nested tuple types.

- Added `ABI#FindLogsByTopic` to find all logs with a given topic.

- Added `ABI#FindLogsByName` to find all logs with a given name.

- Added `ABI#FindFunctionsByName` to find all functions with a given name.

- Added `out of gas` and `Out of gas` as deterministic error (with the constraint that all provider of `eth_call` used have a `gasCap` configured >= than `gasLimit` used for a call, which should be fixed).

- `LogEventDef.LogID()` is now exposed publicly

[unreleased]: https://github.com/streamingfast/eth-go
