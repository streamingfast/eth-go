package rpc

import (
	"bytes"
	"encoding/json/jsontext"
	json "encoding/json/v2"
	"io"
)

// This file keeps the symbols that the vendored `encoding/json` fork used to export from
// this package compiling. They are all thin wrappers over the standard library and are
// deprecated: import `encoding/json/v2` or `encoding/json/jsontext` directly instead.

// Marshaler is the interface implemented by types that can marshal themselves into valid
// JSON.
//
// Deprecated: use [json.Marshaler] from `encoding/json/v2`, or [MarshalerRPC] when the
// type needs a JSON-RPC specific representation.
type Marshaler = json.Marshaler

// RawMessage is a raw encoded JSON value.
//
// Deprecated: use [jsontext.Value].
type RawMessage = jsontext.Value

// SyntaxError describes a JSON syntax error.
//
// Deprecated: use [jsontext.SyntacticError].
type SyntaxError = jsontext.SyntacticError

// Compact appends to dst the JSON-encoded src with insignificant space characters elided.
//
// Deprecated: use [jsontext.Value.Compact].
func Compact(dst *bytes.Buffer, src []byte) error {
	value := jsontext.Value(src).Clone()
	if err := value.Compact(); err != nil {
		return err
	}

	_, err := dst.Write(value)
	return err
}

// Indent appends to dst an indented form of the JSON-encoded src.
//
// Deprecated: use [jsontext.Value.Indent].
func Indent(dst *bytes.Buffer, src []byte, prefix, indent string) error {
	opts, err := indentOptions(prefix, indent)
	if err != nil {
		return err
	}

	value := jsontext.Value(src).Clone()
	if err := value.Indent(opts...); err != nil {
		return err
	}

	_, err = dst.Write(value)
	return err
}

// Valid reports whether data is a valid JSON encoding.
//
// Deprecated: use [jsontext.Value.IsValid].
func Valid(data []byte) bool {
	return jsontext.Value(data).IsValid()
}

// HTMLEscape appends to dst the JSON-encoded src with <, >, &, U+2028 and U+2029
// characters inside string literals escaped, so that the JSON is safe to embed inside
// HTML <script> tags.
//
// Deprecated: use [jsontext.Value.Format] with [jsontext.EscapeForHTML] and
// [jsontext.EscapeForJS].
func HTMLEscape(dst *bytes.Buffer, src []byte) {
	value := jsontext.Value(src).Clone()
	if err := value.Format(jsontext.EscapeForHTML(true), jsontext.EscapeForJS(true)); err != nil {
		// The v1 signature has no way to report an error; invalid input is passed through
		// unchanged, matching how the fork behaved on malformed JSON.
		dst.Write(src)
		return
	}

	dst.Write(value)
}

// Encoder writes JSON-RPC encoded values to an output stream.
//
// Deprecated: use [MarshalJSONRPC] together with [json.MarshalWrite] options, or
// [jsontext.Encoder] for token-level control.
type Encoder struct {
	writer      io.Writer
	prefix      string
	indent      string
	escapeHTML  bool
	initialized bool
}

// NewEncoder returns a new [Encoder] that writes to w.
//
// Deprecated: see [Encoder].
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{writer: w, escapeHTML: true, initialized: true}
}

// SetIndent instructs the encoder to format each subsequent encoded value as if indented
// by [Indent].
//
// Deprecated: see [Encoder].
func (e *Encoder) SetIndent(prefix, indent string) {
	e.prefix = prefix
	e.indent = indent
}

// SetEscapeHTML controls whether HTML characters inside string values are escaped.
//
// Deprecated: see [Encoder].
func (e *Encoder) SetEscapeHTML(on bool) {
	e.escapeHTML = on
}

// Encode writes the JSON-RPC encoding of v to the stream, followed by a newline.
//
// Deprecated: see [Encoder].
func (e *Encoder) Encode(v any) error {
	opts := []Options{jsontext.EscapeForHTML(e.escapeHTML)}
	if e.indent != "" || e.prefix != "" {
		indentOpts, err := indentOptions(e.prefix, e.indent)
		if err != nil {
			return err
		}

		opts = append(opts, indentOpts...)
	}

	out, err := MarshalJSONRPC(v, opts...)
	if err != nil {
		return err
	}

	_, err = e.writer.Write(append(out, '\n'))
	return err
}
