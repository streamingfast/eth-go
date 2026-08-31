package rpc

import (
	"bytes"
	"encoding/json/jsontext"
	json "encoding/json/v2"
	"math/big"
	"strings"
	"testing"

	"github.com/streamingfast/eth-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalJSONRPCWithOptions(t *testing.T) {
	t.Run("caller options override the defaults", func(t *testing.T) {
		// The JSON-RPC defaults disable HTML escaping; a caller option applied afterwards
		// must win.
		out, err := MarshalJSONRPC("<b>")
		require.NoError(t, err)
		assert.Equal(t, `"<b>"`, string(out))

		out, err = MarshalJSONRPC("<b>", jsontext.EscapeForHTML(true))
		require.NoError(t, err)
		assert.Equal(t, `"\u003cb\u003e"`, string(out))
	})

	t.Run("caller marshalers take precedence", func(t *testing.T) {
		override := json.MarshalToFunc(func(encoder *jsontext.Encoder, value eth.Uint64) error {
			return encoder.WriteToken(jsontext.String("overridden"))
		})

		out, err := MarshalJSONRPC(
			struct {
				Value eth.Uint64 `json:"value"`
				Other uint32     `json:"other"`
			}{Value: 1, Other: 2},
			json.WithMarshalers(json.JoinMarshalers(override, Marshalers(false))),
		)
		require.NoError(t, err)
		assert.Equal(t, `{"value":"overridden","other":"0x2"}`, string(out))
	})
}

func TestMarshalJSONRPCIndent(t *testing.T) {
	value := struct {
		Value uint64 `json:"value"`
	}{Value: 255}

	out, err := MarshalJSONRPCIndent(value, "\t", "  ")
	require.NoError(t, err)
	assert.Equal(t, "{\n\t  \"value\": \"0xff\"\n\t}", string(out))

	// jsontext accepts only spaces and tabs where the vendored v1 encoder took any string.
	// Report that instead of letting jsontext panic.
	_, err = MarshalJSONRPCIndent(value, ">", "  ")
	require.ErrorContains(t, err, "invalid character")
}

func TestDeprecatedIndentAndCompact(t *testing.T) {
	var indented bytes.Buffer
	require.NoError(t, Indent(&indented, []byte(`{"a":[1,2]}`), "", "\t"))
	assert.Equal(t, "{\n\t\"a\": [\n\t\t1,\n\t\t2\n\t]\n}", indented.String())

	var compacted bytes.Buffer
	require.NoError(t, Compact(&compacted, indented.Bytes()))
	assert.Equal(t, `{"a":[1,2]}`, compacted.String())

	require.Error(t, Compact(&bytes.Buffer{}, []byte(`{`)))
}

func TestDeprecatedValid(t *testing.T) {
	assert.True(t, Valid([]byte(`{"a":1}`)))
	assert.False(t, Valid([]byte(`{"a":}`)))
}

func TestDeprecatedHTMLEscape(t *testing.T) {
	var out bytes.Buffer
	HTMLEscape(&out, []byte(`{"a":"<b>&"}`))
	assert.Equal(t, `{"a":"\u003cb\u003e\u0026"}`, out.String())

	// Invalid input is passed through rather than dropped, since the signature has no way
	// to report the failure.
	var invalid bytes.Buffer
	HTMLEscape(&invalid, []byte(`{`))
	assert.Equal(t, `{`, invalid.String())
}

func TestDeprecatedEncoder(t *testing.T) {
	var out strings.Builder

	encoder := NewEncoder(&out)
	require.NoError(t, encoder.Encode(struct {
		Value    uint64   `json:"value"`
		GasPrice *big.Int `json:"gasPrice"`
	}{Value: 255, GasPrice: big.NewInt(16)}))

	assert.Equal(t, "{\"value\":\"0xff\",\"gasPrice\":\"0x10\"}\n", out.String())
}

func TestDeprecatedRawMessage(t *testing.T) {
	// RawMessage is an alias of jsontext.Value, so a raw payload round-trips untouched.
	out, err := MarshalJSONRPC(struct {
		Raw RawMessage `json:"raw"`
	}{Raw: RawMessage(`{"already":"encoded"}`)})
	require.NoError(t, err)

	assert.Equal(t, `{"raw":{"already":"encoded"}}`, string(out))
}
