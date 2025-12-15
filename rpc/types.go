package rpc

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/streamingfast/eth-go"
	"github.com/tidwall/gjson"
)

var LatestBlock = &BlockRef{tag: "latest"}
var PendingBlock = &BlockRef{tag: "pending"}
var EarliestBlock = &BlockRef{tag: "earliest"}
var FinalizedBlock = &BlockRef{tag: "finalized"}

type BlockRef struct {
	tag   string
	value uint64
	hash  eth.Hash
}

// BlockHash is supported on some providers like Alchemy and is based on [EIP-1898](https://eips.ethereum.org/EIPS/eip-1898)
// rules.
//
// @panics If hash == ""
func BlockHash(hash string) *BlockRef {
	ref, err := MaybeBlockHash(hash)
	if err != nil {
		panic(err)
	}

	return ref
}

// MaybeBlockHash is exactly like [BlockHash] but it does not panic and returns the error
// instead.
func MaybeBlockHash(in string) (*BlockRef, error) {
	hash, err := eth.NewHash(in)
	if err != nil {
		return nil, err
	}

	// We use the `nil` value internally in the struct to denot if the
	// hash variable is actually set or not, so for us `nil` has not the
	// same meaning as `len(x) == 0`. To be defensive, even if `eth.NewHash`
	// never generate the `nil` value, we ensure to set it correclty.
	if hash == nil {
		hash = []byte{}
	}

	return &BlockRef{hash: hash}, nil
}

func BlockNumber(number uint64) *BlockRef {
	return &BlockRef{value: number}
}

func (b *BlockRef) IsLatest() bool {
	return b == LatestBlock || b.tag == LatestBlock.tag
}

func (b *BlockRef) IsEarliest() bool {
	return b == EarliestBlock || b.tag == EarliestBlock.tag
}

func (b *BlockRef) IsFinalized() bool {
	return b == FinalizedBlock || b.tag == FinalizedBlock.tag
}

func (b *BlockRef) IsPending() bool {
	return b == PendingBlock || b.tag == PendingBlock.tag
}

func (b *BlockRef) BlockNumber() (number uint64, ok bool) {
	if b.tag != "" {
		return 0, false
	}

	if b.hash != nil {
		return 0, false
	}

	return b.value, true
}

func (b *BlockRef) BlockHash() (hash eth.Hash, ok bool) {
	if b.tag != "" {
		return nil, false
	}

	if b.hash == nil {
		return nil, false
	}

	return b.hash, true
}

type blockHashOrNumberObject struct {
	Hash   eth.Hash          `json:"blockHash"`
	Number *blockNumberOrHex `json:"blockNumber"`
}

// blockNumberOrHex handles both numeric JSON values and hex strings for EIP-1898
type blockNumberOrHex uint64

func (b *blockNumberOrHex) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a number first
	var num uint64
	if err := json.Unmarshal(data, &num); err == nil {
		*b = blockNumberOrHex(num)
		return nil
	}

	// Fall back to hex string
	var hexStr string
	if err := json.Unmarshal(data, &hexStr); err != nil {
		return fmt.Errorf("expected number or hex string: %w", err)
	}

	var value eth.Uint64
	if err := value.UnmarshalText([]byte(hexStr)); err != nil {
		return err
	}

	*b = blockNumberOrHex(value)
	return nil
}

func (b *BlockRef) UnmarshalJSON(text []byte) error {
	result := gjson.ParseBytes(text)

	if result.IsObject() {
		return b.UnmarshalText(text)
	}

	// Handle raw numeric JSON values
	if result.Type == gjson.Number {
		b.tag = ""
		b.hash = nil
		b.value = uint64(result.Uint())
		return nil
	}

	// Handle string values
	var s string
	if err := json.Unmarshal(text, &s); err != nil {
		return fmt.Errorf("unmarshal string: %w", err)
	}

	return b.UnmarshalText([]byte(s))
}

func (b *BlockRef) UnmarshalText(text []byte) error {
	if gjson.ParseBytes(text).IsObject() {
		// Is it right to do JSON unmarshaling in the text version? Maybe we should use a pure `UnmarshalJSOM`.
		var obj blockHashOrNumberObject
		if err := json.Unmarshal(text, &obj); err != nil {
			return fmt.Errorf("invalid hash or number: %w", err)
		}

		b.tag = ""
		b.value = 0
		b.hash = obj.Hash

		// Handle blockNumber field if present (EIP-1898)
		if obj.Number != nil {
			b.value = uint64(*obj.Number)
			b.hash = nil
		}

		return nil
	}

	lowerTextString := strings.ToLower(string(text))
	if lowerTextString == LatestBlock.tag {
		*b = *LatestBlock
		return nil
	}

	if lowerTextString == EarliestBlock.tag {
		*b = *EarliestBlock
		return nil
	}

	if lowerTextString == PendingBlock.tag {
		*b = *PendingBlock
		return nil
	}
	if lowerTextString == FinalizedBlock.tag {
		*b = *FinalizedBlock
		return nil
	}

	b.tag = ""
	b.value = 0
	b.hash = nil

	if len(text) > 18 { // A hack to differentiate between Uint64 and Hash
		var value eth.Hash
		if err := value.UnmarshalText(text); err != nil {
			return fmt.Errorf("invalid block hash: %w", err)
		}
		b.hash = value
	} else {
		var value eth.Uint64
		if err := value.UnmarshalText(text); err != nil {
			return fmt.Errorf("invalid block number: %w", err)
		}
		b.value = uint64(value)
	}

	return nil
}

func (b *BlockRef) MarshalText() (string, error) {
	if b == nil {
		return "latest", nil
	}

	if b.tag != "" {
		return b.tag, nil
	}

	if b.hash != nil {
		return fmt.Sprintf("BlockHash: %s", b.hash.Pretty()), nil
	}

	return fmt.Sprintf("BlockNumber: %d", b.value), nil
}

var useShortBlockNumberNotation bool

func init() {
	useShortBlockNumberNotation = os.Getenv("ETH_RPC_SHORT_BLOCK_NUMBER_NOTATION") == "true"
}

func (b *BlockRef) MarshalJSONRPC() ([]byte, error) {
	if b == nil {
		return []byte("latest"), nil
	}

	if b.tag != "" {
		return MarshalJSONRPC(b.tag)
	}

	if b.hash != nil { // [EIP-1898](https://eips.ethereum.org/EIPS/eip-1898)
		return MarshalJSONRPC(struct {
			BlockHash string `json:"blockHash"`
		}{
			BlockHash: b.hash.Pretty(),
		})
	}

	if useShortBlockNumberNotation {
		return fmt.Appendf(nil, "\"0x%x\"", b.value), nil
	}

	return MarshalJSONRPC(struct {
		BlockNumber uint64 `json:"blockNumber"`
	}{
		BlockNumber: b.value,
	})
}

func (b *BlockRef) String() string {
	if b == nil {
		return "latest"
	}

	if b.tag != "" {
		return strings.ToUpper(string(b.tag[0])) + b.tag[1:]
	}

	return "#" + strconv.FormatUint(b.value, 10)
}

type LogEntry struct {
	Address          eth.Address `json:"address"`
	Topics           []eth.Hash  `json:"topics"`
	Data             eth.Hex     `json:"data"`
	BlockNumber      eth.Uint64  `json:"blockNumber"`
	TransactionHash  eth.Hash    `json:"transactionHash"`
	TransactionIndex eth.Uint64  `json:"transactionIndex"`
	BlockHash        eth.Hash    `json:"blockHash"`
	LogIndex         eth.Uint64  `json:"logIndex"`
	Removed          bool        `json:"removed"`
}

func (e *LogEntry) ToLog() (out eth.Log) {
	out.Address = e.Address
	if len(e.Topics) > 0 {
		out.Topics = make([][]byte, len(e.Topics))
		for i, topic := range e.Topics {
			out.Topics[i] = topic
		}
	}
	out.Data = e.Data
	out.BlockIndex = uint32(e.LogIndex)

	return
}

type TransactionReceipt struct {
	// TransactionHash is the hash of the transaction.
	TransactionHash eth.Hash `json:"transactionHash"`
	// TransactionIndex is the transactions index position in the block.
	TransactionIndex eth.Uint64 `json:"transactionIndex"`
	// BlockHash is the hash of the block where this transaction was in.
	BlockHash eth.Hash `json:"blockHash"`
	// BlockNumber is the block number where this transaction was in.
	BlockNumber eth.Uint64 `json:"blockNumber"`
	// From is the address of the sender.
	From eth.Address `json:"from"`
	// To is the address of the receiver, `null` when the transaction is a contract creation transaction.
	To *eth.Address `json:"to,omitempty"`
	// CumulativeGasUsed is the the total amount of gas used when this transaction was executed in the block.
	CumulativeGasUsed eth.Uint64 `json:"cumulativeGasUsed"`
	// EffectiveGasPrice is the sum of the base fee and tip paid per unit of gas.
	EffectiveGasPrice eth.Uint64 `json:"effectiveGasPrice"`
	// GasUsed is the the amount of gas used by this specific transaction alone.
	GasUsed eth.Uint64 `json:"gasUsed"`
	// ContractAddress is the the contract address created, if the transaction was a contract creation, otherwise - null.
	ContractAddress *eth.Address `json:"contractAddress,omitempty"`
	// Logs is the Array of log objects, which this transaction generated.
	Logs []*LogEntry `json:"logs"`
	// LogsBloom is the Bloom filter for light clients to quickly retrieve related logs.
	LogsBloom eth.Hex `json:"logsBloom"`
	// Type is the transaction type, 0x0 for legacy transactions, 0x1 for access list types, 0x2 for dynamic fees
	Type eth.Uint64 `json:"type"`

	// Root: 32 bytes of post-transaction stateroot (pre Byzantium)
	Root eth.Hash `json:"root"`
	// Status is either 1 (success) or 0 (failure) (post Byzantium)
	Status *eth.Uint64 `json:"status"`
}

// Transaction retrieve from `eth_getBlockByXXX` methods.
type Transaction struct {
	// Hash is the hash of the transaction.
	Hash eth.Hash `json:"hash"`

	Nonce eth.Uint64 `json:"nonce,omitempty"`

	// BlockHash is the hash of the block where this transaction was in, none when pending.
	BlockHash eth.Hash `json:"blockHash"`

	// BlockNumber is the block number where this transaction was in, none when pending.
	BlockNumber eth.Uint64 `json:"blockNumber"`

	// TransactionIndex is the transactions index position in the block, none when pending.
	TransactionIndex eth.Uint64 `json:"transactionIndex"`

	// From is the address of the sender.
	From eth.Address `json:"from"`

	// To is the address of the receiver, `null` when the transaction is a contract creation transaction.
	To *eth.Address `json:"to"`

	// Value is the ETH transfered value to the recipient/
	Value *eth.Uint256 `json:"value"`

	// GasPrice is the ETH value of the gas sender is willing to pay, it's the effective gas price when London fork is active and transaction type is 0x02 (DynamicFee).
	GasPrice *eth.Uint256 `json:"gasPrice"`

	// Gas is the the amount of gas sender is willing to allocate for this transaction, might not be fully consumed.
	Gas eth.Uint64 `json:"gas"`

	// Input data the transaction will receive for execution of EVM.
	Input eth.Hex `json:"input,omitempty"`

	// V is the ECDSA recovery id of the transaction's signature.
	V eth.Uint64 `json:"v,omitempty"`

	// R is the ECDSA signature R point of transaction's signature.
	R *eth.Uint256 `json:"r,omitempty"`

	// S is the ECDSA signature S point of transaction's signature.
	S *eth.Uint256 `json:"s,omitempty"`

	// AccessList is the defined access list tuples when the transaction is of AccessList type (0x01), none when transaction of other types.
	AccessList AccessList `json:"accessList,omitempty"`

	// ChainID is the identifier chain the transaction was executed in, none if London fork is **not** activated
	ChainID eth.Uint64 `json:"chainId,omitempty"`

	// MaxFeePerGas is the identifier chain the transaction was executed in, none if London fork is **not** activated
	MaxFeePerGas *eth.Uint256 `json:"maxFeePerGas,omitempty"`

	// MaxPriorityFeePerGas is the identifier chain the transaction was executed in, none if London fork is **not** activated
	MaxPriorityFeePerGas *eth.Uint256 `json:"maxPriorityFeePerGas,omitempty"`

	// Type is the transaction's type
	Type eth.TransactionType `json:"type"`
}

type AccessList []AccessTuple

type AccessTuple struct {
	Address     eth.Address `json:"address"`
	StorageKeys []eth.Hash  `json:"storageKeys"`
}

type Block struct {
	Number           eth.Uint64         `json:"number"`
	Hash             eth.Hash           `json:"hash"`
	ParentHash       eth.Hash           `json:"parentHash"`
	Timestamp        eth.Timestamp      `json:"timestamp"`
	StateRoot        eth.Hash           `json:"stateRoot"`
	TransactionsRoot eth.Hash           `json:"transactionsRoot"`
	ReceiptsRoot     eth.Hash           `json:"receiptsRoot"`
	MixHash          eth.Hash           `json:"mixHash"`
	GasLimit         eth.Uint64         `json:"gasLimit"`
	GasUsed          eth.Uint64         `json:"gasUsed"`
	Difficulty       *eth.Uint256       `json:"difficulty"`
	TotalDifficulty  *eth.Uint256       `json:"totalDifficulty"`
	Miner            eth.Address        `json:"miner"`
	Nonce            eth.FixedUint64    `json:"nonce,omitempty"`
	LogsBloom        eth.Hex            `json:"logsBloom"`
	ExtraData        eth.Hex            `json:"extraData"`
	BaseFeePerGas    *eth.Uint256       `json:"baseFeePerGas,omitempty"`
	BlockSize        eth.Uint64         `json:"size,omitempty"`
	Transactions     *BlockTransactions `json:"transactions,omitempty"`
	UnclesSHA3       eth.Hash           `json:"sha3Uncles,omitempty"`
	Uncles           []eth.Hash         `json:"uncles,omitempty"`

	BlobGasUsed           *eth.Uint64  `json:"blobGasUsed,omitempty"`           // EIP-4844
	ExcessBlobGas         *eth.Uint64  `json:"excessBlobGas,omitempty"`         // EIP-4844
	ParentBeaconBlockRoot *eth.Hash    `json:"parentBeaconBlockRoot,omitempty"` // EIP-4844
	WithdrawalsHash       *eth.Hash    `json:"withdrawalsRoot,omitempty"`       // EIP-4895
	Withdrawals           []Withdrawal `json:"withdrawals,omitempty"`           // EIP-4895
}

type Withdrawal struct {
	Index     eth.Uint64  `json:"index"`          // monotonically increasing identifier issued by consensus layer
	Validator eth.Uint64  `json:"validatorIndex"` // index of validator associated with withdrawal
	Address   eth.Address `json:"address"`        // target address for withdrawn ether
	Amount    eth.Uint64  `json:"amount"`         // value of withdrawal in Gwei
}

// BlockTransactions is a dynamic types and can be either a list of transactions hashes,
// retrievable via `Hashes()` getter when `GetBlockBy{Hash|Number}` is called without full transaction
// and it a list of transaction receipts if it's called with full transaction (option
// `rpc.WithGetBlockFullTransaction`).
type BlockTransactions struct {
	hashes       []eth.Hash
	Transactions []Transaction
}

func NewBlockTransactions() *BlockTransactions {
	return &BlockTransactions{
		Transactions: make([]Transaction, 0),
	}
}

func (txs *BlockTransactions) MarshalJSON() ([]byte, error) {
	return txs.marshalJSON(json.Marshal)
}

func (txs *BlockTransactions) MarshalJSONRPC() ([]byte, error) {
	return txs.marshalJSON(MarshalJSONRPC)
}

func (txs *BlockTransactions) marshalJSON(marshaller func(v interface{}) ([]byte, error)) ([]byte, error) {
	if len(txs.hashes) == 0 {
		if len(txs.Transactions) == 0 {
			return []byte(`[]`), nil
		}

		return marshaller(txs.Transactions)
	}

	return marshaller(txs.hashes)
}

func (txs *BlockTransactions) UnmarshalJSON(data []byte) error {
	rootResult := gjson.ParseBytes(data)
	if !rootResult.IsArray() {
		return fmt.Errorf("expected JSON array, got %s", rootResult.Type)
	}

	result := rootResult.Get("0")
	if result.Type == gjson.Null {
		// No transactions in this block
		return nil
	}

	if result.Type == gjson.String {
		return json.Unmarshal(data, &txs.hashes)
	}

	if result.IsObject() {
		return json.Unmarshal(data, &txs.Transactions)
	}

	return fmt.Errorf("expected JSON array of either string or JSON object, got JSON array of %s", result.Type)
}

func (txs *BlockTransactions) Hashes() (out []eth.Hash) {
	if len(txs.hashes) == 0 {
		if len(txs.Transactions) == 0 {
			return nil
		}

		out = make([]eth.Hash, len(txs.Transactions))
		for i, receipt := range txs.Transactions {
			out[i] = receipt.Hash
		}
		return
	}

	return txs.hashes
}

func (txs *BlockTransactions) Receipts() (out []Transaction, found bool) {
	if len(txs.Transactions) == 0 {
		// We assume we were full is there is no hashes neither, in which case we assume it's ok to say we were full
		return nil, len(txs.hashes) == 0
	}

	// If we have receipts, it's sure we have full state
	return txs.Transactions, true
}
