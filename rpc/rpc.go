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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/streamingfast/eth-go"
	"github.com/streamingfast/logging"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

var ErrFalseResp = errors.New("false response")

type Option func(*Client)

// TODO: refactor to use mux rpc
type Client struct {
	endpoint     string
	chainID      *big.Int
	headers      map[string]string
	httpClient   *http.Client
	useNumericID bool
}

func NewClient(endpointURL string, opts ...Option) *Client {
	c := &Client{
		endpoint: endpointURL,
		headers:  make(map[string]string),
	}

	c.httpClient = http.DefaultClient

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func WithHttpHeader(key, val string) Option {
	return func(client *Client) {
		client.headers[key] = val
	}
}

func WithHttpClient(httpClient *http.Client) Option {
	return func(client *Client) {
		client.httpClient = httpClient
	}
}

func WithNumericID(useNumericID bool) Option {
	return func(client *Client) {
		client.useNumericID = useNumericID
	}
}

type LogsParams struct {
	// FromBlock is either block number encoded as a hexadecimal or tagged value which is one of
	// "latest" (`rpc.LatestBlock`), "pending" (`rpc.PendingBlock`) or "earliest" tags (`rpc.EarliestBlock`) (optional).
	FromBlock *BlockRef `json:"fromBlock,omitempty"`

	// ToBlock is either block number encoded as a hexadecimal or tagged value which is one of
	// "latest" (`LatestBlock`), "pending" (`rpc.PendingBlock`) or "earliest" tags (`EarliestBlock`) (optional).
	ToBlock *BlockRef `json:"toBlock,omitempty"`

	// Address is the contract address or a list of addresses from which logs should originate (optional).
	Address eth.Address `json:"address,omitempty"`

	// Topics are order-dependent, each topic can also be an array of DATA with "or" options (optional).
	Topics *TopicFilter `json:"topics,omitempty"`

	// BlockHash is the block hash encoded as a hex string with prefix '0x'
	BlockHash eth.Bytes `json:"blockHash,omitempty"`
}

type TopicFilter struct {
	topics []TopicFilterExpr
}

func (f *TopicFilter) String() string {
	var elements []string
	for _, topic := range f.topics {
		elements = append(elements, topic.String())
	}

	return fmt.Sprintf("[%s]", strings.Join(elements, ", "))
}

func (f *TopicFilter) Append(in interface{}) {
	f.topics = append(f.topics, newTopicExpr(in))
}

func (f *TopicFilter) MarshalJSONRPC() ([]byte, error) {
	return MarshalJSONRPC(f.topics)
}

type TopicFilterExpr struct {
	exact *eth.Topic
	oneOf []eth.Topic
}

func (f TopicFilterExpr) String() string {
	if f.oneOf == nil {
		if f.exact == nil {
			return "null"
		}

		bytes := *f.exact
		return eth.Hex(bytes[:]).String()
	}

	var elements []string
	for _, topic := range f.oneOf {
		elements = append(elements, eth.Hex(topic[:]).String())
	}

	return fmt.Sprintf("[%s]", strings.Join(elements, ", "))
}

func (f TopicFilterExpr) MarshalJSONRPC() ([]byte, error) {
	if f.oneOf == nil {
		return MarshalJSONRPC(f.exact)
	}

	return MarshalJSONRPC(f.oneOf)
}

func NewTopicFilter(exprs ...interface{}) (out *TopicFilter) {
	if len(exprs) == 0 {
		return nil
	}

	topics := make([]TopicFilterExpr, len(exprs))
	for i, expr := range exprs {
		topics[i] = newTopicExpr(expr)
	}

	return &TopicFilter{topics: topics}
}

func newTopicExpr(expr interface{}) (out TopicFilterExpr) {
	switch v := expr.(type) {
	case TopicFilterExpr:
		return v
	case eth.Topic:
		return TopicFilterExpr{exact: &v}
	default:
		return ExactTopic(v)
	}
}

func ExactTopic(in interface{}) TopicFilterExpr {
	return TopicFilterExpr{exact: eth.LogTopic(in)}
}

func AnyTopic() TopicFilterExpr {
	return TopicFilterExpr{exact: nil}
}

func OneOfTopic(topics ...interface{}) (out TopicFilterExpr) {
	if len(topics) == 0 {
		panic("there must be at least one element to create a one of topic element")
	}

	out.oneOf = make([]eth.Topic, len(topics))
	for i, topic := range topics {
		logTopic := eth.LogTopic(topic)
		if logTopic == nil {
			panic("it's invalid to use nil value when building a one of topic element")
		}

		out.oneOf[i] = *logTopic
	}

	return
}

func (c *Client) Logs(ctx context.Context, params LogsParams) ([]*LogEntry, error) {
	params.FromBlock.shortBlockNumberNotation = true // these do not support long `{"blockNumber": 1234}` format
	params.ToBlock.shortBlockNumberNotation = true
	return Do[[]*LogEntry](c, ctx, "eth_getLogs", []interface{}{params})
}

type getBlockOptions struct {
	full bool
}

type GetBlockOption interface {
	apply(opt *getBlockOptions)
}

type withGetBlockFullTransaction bool

func (v withGetBlockFullTransaction) apply(opt *getBlockOptions) {
	opt.full = bool(v)
}

func WithGetBlockFullTransaction() GetBlockOption {
	return withGetBlockFullTransaction(true)
}

// GetBlockByNumber fetches the block by its number and optionally include full transaction receipts
// if you supply the `rpc.withGetBlockFullTransaction` option.
//
// Uses RPC call `eth_getBlockByNumber`
func (c *Client) GetBlockByNumber(ctx context.Context, ref *BlockRef, opts ...GetBlockOption) (*Block, error) {
	ref.shortBlockNumberNotation = true // this does not support long `{"blockNumber": 1234}` format
	return c.getBlock(ctx, "eth_getBlockByNumber", ref, opts)
}

// GetBlockByHash fetches the block by its hash and optionally include full transaction receipts
// if you supply the `rpc.withGetBlockFullTransaction` option.
//
// Uses RPC call `eth_getBlockByHash`.
func (c *Client) GetBlockByHash(ctx context.Context, hash eth.Hash, opts ...GetBlockOption) (*Block, error) {
	return c.getBlock(ctx, "eth_getBlockByHash", hash, opts)
}

func (c *Client) getBlock(ctx context.Context, method string, identifier interface{}, opts []GetBlockOption) (*Block, error) {
	var options getBlockOptions
	for _, opt := range opts {
		opt.apply(&options)
	}

	return Do[*Block](c, ctx, method, []interface{}{identifier, options.full})
}

func (c *Client) LatestBlockNum(ctx context.Context) (uint64, error) {
	resp, err := c.DoRequest(ctx, "eth_blockNumber", []interface{}{})
	if err != nil {
		return 0, fmt.Errorf("unable to perform eth_blockNumber request: %w", err)
	}

	value, err := strconv.ParseUint(resp, 0, 64)
	if err != nil {
		return 0, fmt.Errorf("unable to parse block number %s: %w", resp, err)
	}

	return value, nil
}

func (c *Client) FinalizeBlockNum(ctx context.Context) (uint64, error) {
	resp, err := c.DoRequest(ctx, "eth_blockNumber", []interface{}{
		FinalizedBlock.tag, // 436
		true,
	})
	if err != nil {
		return 0, fmt.Errorf("unable to perform eth_blockNumber request: %w", err)
	}

	value, err := strconv.ParseUint(resp, 0, 64)
	if err != nil {
		return 0, fmt.Errorf("unable to parse block number %s: %w", resp, err)
	}

	return value, nil
}

type CallParams struct {
	// From the address the transaction is sent from (optional).
	From eth.Address `json:"from,omitempty"`
	// To the address the transaction is directed to (required).
	To eth.Address `json:"to,omitempty"`
	// GasLimit Integer of the gas provided for the transaction execution. eth_call consumes zero gas, but this parameter may be needed by some executions (optional).
	GasLimit uint64 `json:"gas,omitempty"`
	// GasPrice big integer of the gasPrice used for each paid gas (optional).
	GasPrice *big.Int `json:"gasPrice,omitempty"`
	// Value big integer of the value sent with this transaction (optional).
	Value *big.Int `json:"value,omitempty"`
	// Hash of the method signature and encoded parameters or any object that implements `MarshalJSONRPC` and serialize to a byte array, for details see Ethereum Contract ABI in the Solidity documentation (optional).
	Data interface{} `json:"data,omitempty"`
}

func (c *Client) Call(ctx context.Context, params CallParams) (string, error) {
	return c.callAtBlock(ctx, "eth_call", params, LatestBlock)
}

func (c *Client) CallAtBlock(ctx context.Context, params CallParams, blockAt *BlockRef) (string, error) {
	return c.callAtBlock(ctx, "eth_call", params, blockAt)
}

func (c *Client) EstimateGas(ctx context.Context, params CallParams) (string, error) {
	return c.callAtBlock(ctx, "eth_estimateGas", params, LatestBlock)
}

func (c *Client) callAtBlock(ctx context.Context, method string, params interface{}, blockAt *BlockRef) (string, error) {
	return c.DoRequest(ctx, method, []interface{}{params, blockAt})
}

// Depreacted: Use SendRawTransaction instead
func (c *Client) SendRaw(ctx context.Context, rawData []byte) (string, error) {
	return c.DoRequest(ctx, "eth_sendRawTransaction", []interface{}{rawData})
}

func (c *Client) SendRawTransaction(ctx context.Context, rawData []byte) (string, error) {
	return c.DoRequest(ctx, "eth_sendRawTransaction", []interface{}{rawData})
}

func (c *Client) ChainID(ctx context.Context) (*big.Int, error) {
	if c.chainID != nil {
		return c.chainID, nil
	}

	resp, err := c.DoRequest(ctx, "eth_chainId", []interface{}{})
	if err != nil {
		return nil, fmt.Errorf("unable to perform eth_chainId request: %w", err)
	}

	i := &big.Int{}
	_, ok := i.SetString(resp, 0)
	if !ok {
		return nil, fmt.Errorf("unable to parse chain id %s: %w", resp, err)
	}
	c.chainID = i
	return c.chainID, nil
}

func (c *Client) ProtocolVersion(ctx context.Context) (string, error) {
	resp, err := c.DoRequest(ctx, "eth_protocolVersion", []interface{}{})
	if err != nil {
		return "", fmt.Errorf("unable to perform eth_protocolVersion request: %w", err)
	}

	return resp, nil
}

type SyncingResp struct {
	StartingBlockNum eth.Uint64 `json:"starting_block_num"`
	CurrentBlockNum  eth.Uint64 `json:"current_block_num"`
	HighestBlockNum  eth.Uint64 `json:"highest_block_num"`
}

func (c *Client) Syncing(ctx context.Context) (*SyncingResp, error) {
	resp, err := c.DoRequest(ctx, "eth_syncing", []interface{}{})
	if err != nil {
		return nil, fmt.Errorf("unable to perform eth_syncing request: %w", err)
	}

	if resp == "false" {
		return nil, ErrFalseResp
	}

	out := &SyncingResp{}
	if err := json.Unmarshal([]byte(resp), out); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return out, nil
}

// TransactionReceipt fetches the receipt associated with the transaction's hash received. If the
// transaction is not found by the queried node, `nil, nil` is returned. If it's found, the receipt
// is decoded and `receipt, nil` is returned. Otherwise, the RPC error is returned if something went wrong.
func (c *Client) TransactionReceipt(ctx context.Context, hash eth.Hash) (*TransactionReceipt, error) {
	return Do[*TransactionReceipt](c, ctx, "eth_getTransactionReceipt", []interface{}{hash})
}

func (c *Client) GetTransactionCount(ctx context.Context, accountAddr eth.Address, at *BlockRef) (uint64, error) {
	return c.Nonce(ctx, accountAddr, at)
}

func (c *Client) Nonce(ctx context.Context, accountAddr eth.Address, at *BlockRef) (uint64, error) {
	resp, err := c.DoRequest(ctx, "eth_getTransactionCount", []interface{}{accountAddr.Pretty(), atOrLatestIfNil(at)})
	if err != nil {
		return 0, fmt.Errorf("unable to perform eth_getTransactionCount request: %w", err)
	}

	nonce, err := strconv.ParseUint(strings.TrimPrefix(resp, "0x"), 16, 64)
	if err != nil {
		return 0, fmt.Errorf("unable to parse nonce %s: %w", resp, err)
	}
	return nonce, nil
}

func (c *Client) GetBalance(ctx context.Context, accountAddr eth.Address, at *BlockRef) (*eth.TokenAmount, error) {
	resp, err := c.DoRequest(ctx, "eth_getBalance", []interface{}{accountAddr.Pretty(), atOrLatestIfNil(at)})
	if err != nil {
		return nil, fmt.Errorf("unable to perform eth_getBalance request: %w", err)
	}

	v, ok := new(big.Int).SetString(strings.TrimPrefix(resp, "0x"), 16)
	if !ok {
		return nil, fmt.Errorf("unable to parse balance %q: %w", resp, err)
	}

	return &eth.TokenAmount{
		Amount: v,
		Token:  eth.ETHToken,
	}, nil
}

func (c *Client) GetCode(ctx context.Context, accountAddr eth.Address, at *BlockRef) (eth.Bytes, error) {
	resp, err := c.DoRequest(ctx, "eth_getCode", []interface{}{accountAddr.Pretty(), atOrLatestIfNil(at)})
	if err != nil {
		return nil, fmt.Errorf("unable to perform eth_getCode request: %w", err)
	}

	v, err := eth.NewBytes(resp)
	if err != nil {
		return nil, fmt.Errorf("unable to parse code %q: %w", resp, err)
	}

	return v, nil
}

func (c *Client) GasPrice(ctx context.Context) (*big.Int, error) {
	resp, err := c.DoRequest(ctx, "eth_gasPrice", []interface{}{})
	if err != nil {
		return nil, fmt.Errorf("unable to perform eth_gasPrice request: %w", err)
	}

	i := &big.Int{}
	_, ok := i.SetString(resp, 0)
	if !ok {
		return nil, fmt.Errorf("unable to parse gas price %s: %w", resp, err)
	}

	return i, nil
}

type RPCRequest struct {
	Params  []interface{} `json:"params"`
	Method  string        `json:"method"`
	decoder ResponseDecoder

	JSONRPC string  `json:"jsonrpc"`
	ID      eth.Int `json:"id"`
}

type RPCResponse struct {
	Content string
	ID      int
	Err     error
	decoder ResponseDecoder
}

func (r *RPCResponse) UnmarshalJSON(in []byte) error {
	response := gjson.ParseBytes(in)

	id, rawValue, err := parseResponseID(response)
	if err != nil {
		return fmt.Errorf("invalid id value %q: %w", rawValue, err)
	}

	rpcErrorResult := response.Get("error")
	if !rpcErrorResult.Exists() {
		r.ID = id
		r.Content = response.Get("result").String()
		return nil
	}

	rpcErr := &ErrResponse{}
	if err := json.Unmarshal([]byte(rpcErrorResult.Raw), rpcErr); err != nil {
		// We were not able to deserialize to RPC error, return the whole thing with an error
		return fmt.Errorf("invalid error value %q: %w", rpcErrorResult.Raw, err)
	}

	r.ID = id
	r.Err = rpcErr

	return nil
}

func (res *RPCResponse) CopyDecoder(req *RPCRequest) {
	res.decoder = req.decoder
}

func (res *RPCResponse) Empty() bool {
	return res.Content == "0x"
}

func (res *RPCResponse) Deterministic() bool {
	if res.Err == nil {
		return true
	}
	if rpcErr, ok := res.Err.(*ErrResponse); ok {
		return IsDeterministicError(rpcErr)
	}
	return false
}

func (res *RPCResponse) Decode() ([]interface{}, error) {
	if res.decoder == nil {
		return nil, fmt.Errorf("no decoder in RPC response")
	}
	if res.Err != nil {
		return nil, fmt.Errorf("error in response, cannot decode")
	}
	if res.Empty() {
		return nil, fmt.Errorf("empty response, cannot decode")
	}
	return res.decoder(eth.MustNewHex(res.Content))
}

type ETHCallOption func(*ETHCall)

func AtBlockNum(num uint64) ETHCallOption {
	return func(c *ETHCall) {
		c.atExpr = num
	}
}

func NewETHCall(to eth.Address, methodDef *eth.MethodDef, options ...ETHCallOption) *ETHCall {
	c := &ETHCall{
		params: CallParams{
			To:   to,
			Data: methodDef.NewCall().MustEncode(),
		},
		methodDef:       methodDef,
		responseDecoder: methodDef.DecodeOutput,
		atExpr:          LatestBlock,
	}
	for _, opt := range options {
		opt(c)
	}
	return c
}

// NewRawETHCall can be used to construct an ETHCall when caller has already encoded the call data
func NewRawETHCall(callParams CallParams, atBlock *BlockRef) *ETHCall {
	return &ETHCall{
		params: callParams,
		atExpr: atBlock,
	}
}

type ETHCall struct {
	params          CallParams
	methodDef       *eth.MethodDef
	atExpr          interface{}
	responseDecoder ResponseDecoder
}

func (c *ETHCall) ToRequest() *RPCRequest {
	return &RPCRequest{
		Params:  []interface{}{c.params, c.atExpr},
		decoder: c.responseDecoder,
		Method:  "eth_call",
	}
}

type ResponseDecoder func([]byte) ([]interface{}, error)

func (c *Client) DoRequests(ctx context.Context, reqs []*RPCRequest) ([]*RPCResponse, error) {
	logger := logging.Logger(ctx, zlog).With(zap.Strings("methods", methodsFromRPCRequests(reqs)))

	// sanitize reqs
	var lastID int
	// we need IDs to be sorted
	for _, req := range reqs {
		lastID++
		req.ID = eth.Int(lastID)
		req.JSONRPC = "2.0"
	}

	reqsBytes, err := MarshalJSONRPCWithOpts(&reqs, c.useNumericID)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal json_rpc requests: %w", err)
	}
	if tracer.Enabled() {
		logger.Debug("json_rpc requests", zap.Stringer("requests", eth.Hex(reqsBytes)))
	}

	resp, err := c.doRequest(ctx, logger, reqsBytes)
	if err != nil {
		return nil, err
	}

	results, err := parseRPCResults(logger, resp)
	if err != nil {
		return nil, err
	}

	if len(results) != len(reqs) {
		logger.Warn("invalid number of results", zap.Int("result_count", len(results)), zap.Int("request_count", len(reqs)))
		return nil, fmt.Errorf("invalid number of results")
	}

	return results, nil
}

func (c *Client) DoRequest(ctx context.Context, method string, params []interface{}) (string, error) {
	logger := logging.Logger(ctx, zlog).With(zap.String("method", method))

	req := RPCRequest{
		Params:  params,
		JSONRPC: "2.0",
		Method:  method,
		ID:      eth.Int(1),
	}
	reqCnt, err := MarshalJSONRPCWithOpts(&req, c.useNumericID)
	if err != nil {
		return "", fmt.Errorf("unable to marshal json_rpc request: %w", err)
	}

	if tracer.Enabled() {
		logger.Debug("json_rpc request", zap.String("request", string(reqCnt)))
	}

	resp, err := c.doRequest(ctx, logger, reqCnt)
	if err != nil {
		return "", err
	}

	results, err := parseRPCResults(logger, resp)
	if err != nil {
		return "", err
	}
	if len(results) != 1 {
		return "", fmt.Errorf("received no result than number of requests")
	}

	return results[0].Content, results[0].Err
}

// Do performs an RPC request and unmarshals the response into the type T.
// It wraps DoRequest and handles JSON unmarshalling automatically.
//
// When T is string, the result is returned directly without JSON unmarshalling
// since DoRequest already returns the unquoted string value from the JSON response.
func Do[T any](c *Client, ctx context.Context, method string, params []interface{}) (out T, err error) {
	result, err := c.DoRequest(ctx, method, params)
	if err != nil {
		return out, err
	}

	if result == "" {
		return out, nil
	}

	// Handle string type directly without JSON unmarshal.
	// DoRequest returns the unquoted string value (e.g., "0x1234..." becomes 0x1234...),
	// so we can't use json.Unmarshal on it for string types.
	if s, ok := any(&out).(*string); ok {
		*s = result
		return out, nil
	}

	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return out, fmt.Errorf("unmarshal %s response: %w", method, err)
	}

	return out, nil
}

func (c *Client) doRequest(ctx context.Context, logger *zap.Logger, reqsBytes []byte) ([]byte, error) {
	body := bytes.NewBuffer(reqsBytes)

	resp, err := c.post(ctx, c.endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("sending request to json_rpc endpoint: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("error in response: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read json_rpc response body: %w", err)
	}

	if tracer.Enabled() {
		logger.Debug("json_rpc call response", zap.String("response_body", string(bodyBytes)))
	}

	return bodyBytes, nil
}

func (c *Client) post(ctx context.Context, url string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, err
	}

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	req.Header.Set("Content-Type", "application/json")
	return c.httpClient.Do(req)
}

func (c *Client) String() string {
	return c.endpoint
}

func parseRPCResults(logger *zap.Logger, in []byte) ([]*RPCResponse, error) {
	responses := []gjson.Result{}

	// we may receive `[{resp1},{resp2}]` OR `{resp}`
	parsed := gjson.ParseBytes(in)
	if parsed.IsArray() {
		responses = parsed.Array()
	} else {
		responses = append(responses, parsed)
	}

	var out []*RPCResponse
	for i, response := range responses {
		rpcErrorResult := response.Get("error")
		if !rpcErrorResult.Exists() {
			id, rawValue, err := parseResponseID(response)
			if err != nil {
				return nil, fmt.Errorf("json_rpc invalid id value %q for request at index %d: %s", rawValue, i, err)
			}

			out = append(out, &RPCResponse{Content: response.Get("result").String(), ID: int(id)})
			continue
		}

		content := rpcErrorResult.Raw
		if tracer.Enabled() {
			logger.Debug("json_rpc call response error",
				zap.String("response_body", string(in)),
				zap.String("error", content),
			)
		}

		rpcErr := &ErrResponse{}
		err := json.Unmarshal([]byte(content), rpcErr)
		if err != nil {
			// We were not able to deserialize to RPC error, return the whole thing with an error
			return nil, fmt.Errorf("json_rpc returned error: %s", rpcErrorResult)
		}

		id, rawValue, err := parseResponseID(response)
		if err != nil {
			return nil, fmt.Errorf("json_rpc invalid id value %q for request at index %d: %s", rawValue, i, err)
		}

		out = append(out, &RPCResponse{Err: rpcErr, ID: id})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})

	return out, nil
}

func methodsFromRPCRequests(requests []*RPCRequest) (out []string) {
	out = make([]string, len(requests))
	for i, v := range requests {
		out[i] = v.Method
	}

	return
}

func parseResponseID(result gjson.Result) (int, string, error) {
	idResult := result.Get("id")
	idValue := idResult.String()

	// Handle numeric ids directly
	if idResult.Type == gjson.Number {
		return int(idResult.Int()), idValue, nil
	}

	// Handle hex string ids
	var id eth.Uint64
	if err := id.UnmarshalText([]byte(idValue)); err != nil {
		return 0, idValue, err
	}

	return int(id), idValue, nil
}

func atOrLatestIfNil(ref *BlockRef) *BlockRef {
	if ref == nil {
		return LatestBlock
	}

	return ref
}
