package rpc

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/streamingfast/eth-go"
	"github.com/tidwall/gjson"
)

var LatestBlock = &BlockRef{tag: "latest"}
var PendingBlock = &BlockRef{tag: "pending"}
var EarliestBlock = &BlockRef{tag: "earliest"}

type BlockRef struct {
	tag   string
	value uint64
}

func BlockNumber(number uint64) *BlockRef {
	return &BlockRef{tag: "", value: number}
}

func (b *BlockRef) IsLatest() bool {
	return b == LatestBlock || b.tag == LatestBlock.tag
}

func (b *BlockRef) IsEarliest() bool {
	return b == EarliestBlock || b.tag == EarliestBlock.tag
}

func (b *BlockRef) IsPending() bool {
	return b == PendingBlock || b.tag == PendingBlock.tag
}

func (b *BlockRef) BlockNumber() (number uint64, ok bool) {
	if b.tag != "" {
		return 0, false
	}

	return b.value, true
}

func (b *BlockRef) UnmarshalText(text []byte) error {
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

	var value eth.Uint64
	if err := value.UnmarshalText(text); err != nil {
		return err
	}

	b.tag = ""
	b.value = uint64(value)
	return nil
}

func (b BlockRef) MarshalJSONRPC() ([]byte, error) {
	if b.tag != "" {
		return MarshalJSONRPC(b.tag)
	}

	return MarshalJSONRPC(b.value)
}

func (b BlockRef) String() string {
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
	// GasUsed is the the amount of gas used by this specific transaction alone.
	GasUsed eth.Uint64 `json:"gasUsed"`
	// ContractAddress is the the contract address created, if the transaction was a contract creation, otherwise - null.
	ContractAddress *eth.Address `json:"contractAddress,omitempty"`
	// Logs is the Array of log objects, which this transaction generated.
	Logs []*LogEntry `json:"logs"`
	// LogsBloom is the Bloom filter for light clients to quickly retrieve related logs.
	LogsBloom eth.Hex `json:"logsBloom"`
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
	V eth.Uint8 `json:"v,omitempty"`

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
	TotalDifficult   *eth.Uint256       `json:"totalDifficulty"`
	Miner            eth.Address        `json:"miner"`
	Nonce            eth.Uint64         `json:"nonce,omitempty"`
	LogsBloom        eth.Hex            `json:"logsBloom"`
	ExtraData        eth.Hex            `json:"extraData"`
	BaseFeePerGas    eth.Uint64         `json:"baseFeePerGas,omitempty"`
	BlockSize        eth.Uint64         `json:"size,omitempty"`
	Transactions     *BlockTransactions `json:"transactions,omitempty"`
	UnclesSHA3       eth.Hash           `json:"sha3Uncles,omitempty"`
	Uncles           []eth.Hash         `json:"uncles,omitempty"`
}

// BlockTransactions is a dynamic types and can be either a list of transactions hashes,
// retrievable via `Hashes()` getter when `GetBlockBy{Hash|Number}` is called without full transaction
// and it a list of transaction receipts if it's called with full transaction (option
// `rpc.WithGetBlockFullTransaction`).
type BlockTransactions struct {
	hashes       []eth.Hash
	transactions []Transaction
}

func (txs *BlockTransactions) MarshalJSON() ([]byte, error) {
	return txs.marshalJSON(json.Marshal)
}

func (txs *BlockTransactions) MarshalJSONRPC() ([]byte, error) {
	return txs.marshalJSON(MarshalJSONRPC)
}

func (txs *BlockTransactions) marshalJSON(marshaller func(v interface{}) ([]byte, error)) ([]byte, error) {
	if len(txs.hashes) == 0 {
		if len(txs.transactions) == 0 {
			return []byte(`[]`), nil
		}

		return marshaller(txs.transactions)
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
		return json.Unmarshal(data, &txs.transactions)
	}

	return fmt.Errorf("expected JSON array of either string or JSON object, got JSON array of %s", result.Type)
}

func (txs *BlockTransactions) Hashes() (out []eth.Hash) {
	if len(txs.hashes) == 0 {
		if len(txs.transactions) == 0 {
			return nil
		}

		out = make([]eth.Hash, len(txs.transactions))
		for i, receipt := range txs.transactions {
			out[i] = receipt.Hash
		}
		return
	}

	return txs.hashes
}

func (txs *BlockTransactions) Receipts() (out []Transaction, found bool) {
	if len(txs.transactions) == 0 {
		// We assume we were full is there is no hashes neither, in which case we assume it's ok to say we were full
		return nil, len(txs.hashes) == 0
	}

	// If we have receipts, it's sure we have full state
	return txs.transactions, true
}