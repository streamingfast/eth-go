// Marshalling of Go values into the shape the Ethereum JSON-RPC API expects.
//
// The spec encodes every quantity and byte string as a "0x"-prefixed hexadecimal string
// rather than as a JSON number or a base64 string. These rules are expressed as
// `encoding/json/v2` type-specific marshalers; they used to live in a hand-maintained fork
// of the `encoding/json` v1 encoder.

package rpc

import (
	"encoding"
	"encoding/hex"
	"encoding/json/jsontext"
	json "encoding/json/v2"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"github.com/streamingfast/eth-go"
)

// Options is an alias of the `encoding/json/v2` option type. It is re-exported so callers
// can tune [MarshalJSONRPC] with any standard json/v2 or jsontext option without importing
// those packages directly:
//
//	rpc.MarshalJSONRPC(v, jsontext.WithIndent("  "))
type Options = json.Options

// MarshalerRPC is the interface implemented by types that can marshal themselves into
// valid JSON-RPC format.
//
// It has precedence over [json.Marshaler] (`MarshalJSON`) and over
// [encoding.TextMarshaler] (`MarshalText`), which lets a type expose a plain JSON form and
// a JSON-RPC form at the same time.
type MarshalerRPC interface {
	MarshalJSONRPC() ([]byte, error)
}

// Marshalers returns the type-specific marshalers implementing the Ethereum JSON-RPC
// encoding rules. Use it to layer your own marshalers on top of them:
//
//	opts := json.WithMarshalers(json.JoinMarshalers(myMarshalers, rpc.Marshalers(false)))
//	out, err := rpc.MarshalJSONRPC(v, opts)
//
// Marshalers earlier in a [json.JoinMarshalers] list win, so the ones passed first take
// precedence over the JSON-RPC rules.
//
// When useNumericID is true, [eth.Int] encodes as a plain JSON number instead of a hex
// string. That type is used for the `id` member of a JSON-RPC request, and some nodes
// reject the hex form.
func Marshalers(useNumericID bool) *json.Marshalers {
	if useNumericID {
		return numericIDMarshalers
	}

	return hexIDMarshalers
}

// MarshalJSONRPC returns the JSON-RPC encoding of v. It behaves like [json.Marshal] from
// `encoding/json/v2` except that:
//   - Every signed and unsigned integer type encodes as a "0x"-prefixed hex string.
//   - Byte slices encode as a "0x"-prefixed hex string rather than base64.
//   - [big.Int] encodes as a "0x"-prefixed hex string, taking precedence over its own
//     `MarshalJSON` method. A nil *big.Int encodes as "0x0".
//   - A type implementing [MarshalerRPC] uses that method, in preference to `MarshalJSON`
//     and `MarshalText`.
//   - HTML characters are not escaped.
//   - Map keys are emitted in sorted order.
//
// Additional options are applied after the JSON-RPC defaults, so they can override them.
func MarshalJSONRPC(v any, opts ...Options) ([]byte, error) {
	return MarshalJSONRPCWithOpts(v, false, opts...)
}

// MarshalJSONRPCWithOpts is [MarshalJSONRPC] with control over how the JSON-RPC request
// `id` member is encoded. See [Marshalers] for what useNumericID does.
func MarshalJSONRPCWithOpts(v any, useNumericID bool, opts ...Options) ([]byte, error) {
	// json/v2 resolves a nil pointer at the root of the value to `null` without consulting
	// any marshaler. Nested nil *big.Int values do reach [marshalQuantityOrData] and come
	// out as "0x0", so handle the root the same way.
	if value, ok := v.(*big.Int); ok && value == nil {
		return []byte(`"0x0"`), nil
	}

	return json.Marshal(v, append(marshalOptions(useNumericID), opts...)...)
}

// MarshalJSONRPCIndent is [MarshalJSONRPC] with each element on its own line, indented by
// one or more copies of indent according to the nesting depth and preceded by prefix.
//
// Both prefix and indent must be composed only of spaces and tabs; anything else is an
// error. The vendored v1 encoder this replaced accepted arbitrary strings.
func MarshalJSONRPCIndent(v any, prefix, indent string) ([]byte, error) {
	indentOpts, err := indentOptions(prefix, indent)
	if err != nil {
		return nil, err
	}

	return MarshalJSONRPC(v, indentOpts...)
}

// indentOptions validates prefix and indent before handing them to jsontext, which panics
// on anything that is not a space or a tab.
func indentOptions(prefix, indent string) ([]Options, error) {
	if trimmed := strings.Trim(prefix, " \t"); trimmed != "" {
		return nil, fmt.Errorf("invalid character %q in indent prefix, only spaces and tabs are allowed", trimmed)
	}

	if trimmed := strings.Trim(indent, " \t"); trimmed != "" {
		return nil, fmt.Errorf("invalid character %q in indent, only spaces and tabs are allowed", trimmed)
	}

	return []Options{
		jsontext.Multiline(true),
		jsontext.WithIndentPrefix(prefix),
		jsontext.WithIndent(indent),
	}, nil
}

// marshalOptions restates, as json/v2 options, the encoding decisions the vendored v1 fork
// made. Each one is needed to reproduce its output; TestMarshalJSONRPCGolden fails if any is
// dropped.
//
// String escaping is the deliberate exception, and is left at the json/v2 defaults. The fork
// escaped no HTML character, which json/v2 also does not, but it did escape U+2028 and
// U+2029, which json/v2 leaves as raw UTF-8. That escaping existed so JSON could be embedded
// in a <script> tag before ES2019 made both characters legal in JavaScript string literals.
// A JSON-RPC request body is never embedded that way, both forms are valid JSON that every
// parser reads identically, and neither character can occur in a hex quantity, a hex byte
// string, a method name or a block tag.
func marshalOptions(useNumericID bool) []Options {
	return []Options{
		json.WithMarshalers(Marshalers(useNumericID)),

		// The fork sorted map keys, as `encoding/json` v1 always did. json/v2 iterates maps
		// in a randomized order unless asked otherwise.
		json.Deterministic(true),

		// json/v2 emits `[]` and `{}` for nil slices and maps; `encoding/json` v1, and so the
		// fork, emitted `null`. Ethereum nodes distinguish the two for parameters such as
		// `topics`.
		json.FormatNilSliceAsNull(true),
		json.FormatNilMapAsNull(true),
	}
}

var (
	hexIDMarshalers     = newMarshalers(false)
	numericIDMarshalers = newMarshalers(true)
)

func newMarshalers(useNumericID bool) *json.Marshalers {
	var marshalers []*json.Marshalers

	if useNumericID {
		// Ahead of everything else: eth.Int is the type of a JSON-RPC request `id`, and
		// this mode sends it as a plain JSON number.
		marshalers = append(marshalers, json.MarshalToFunc(marshalDecimal[eth.Int]))
	}

	// The byte-slice types of the eth package. They implement MarshalerRPC, but dispatching
	// on a pointer type lets json/v2 take the address of the value instead of boxing a copy
	// into an interface, which it has to allocate for. The output is identical;
	// TestFastPathsMatchMarshalJSONRPC checks they stay in step.
	marshalers = append(marshalers,
		marshalEthBytesFunc[eth.Bytes](),
		marshalEthBytesFunc[eth.Hex](),
		marshalEthBytesFunc[eth.Hash](),
		marshalEthBytesFunc[eth.Address](),
	)

	// Any other explicit JSON-RPC representation on the type itself wins over everything
	// below.
	marshalers = append(marshalers, json.MarshalToFunc(marshalMarshalerRPC))

	// Fast paths. json/v2 matches these on exact type identity and dispatches to them
	// without any reflection, which covers nearly every value this package encodes.
	marshalers = append(marshalers,
		marshalSignedFunc[int](),
		marshalSignedFunc[int8](),
		marshalSignedFunc[int16](),
		marshalSignedFunc[int32](),
		marshalSignedFunc[int64](),
		marshalSignedFunc[eth.Int](),
		marshalSignedFunc[eth.Int8](),
		marshalSignedFunc[eth.Int16](),
		marshalSignedFunc[eth.Int32](),
		marshalSignedFunc[eth.Int64](),

		marshalUnsignedFunc[uint](),
		marshalUnsignedFunc[uint8](),
		marshalUnsignedFunc[uint16](),
		marshalUnsignedFunc[uint32](),
		marshalUnsignedFunc[uint64](),
		marshalUnsignedFunc[uintptr](),
		marshalUnsignedFunc[eth.Uint8](),
		marshalUnsignedFunc[eth.Uint16](),
		marshalUnsignedFunc[eth.Uint32](),
		marshalUnsignedFunc[eth.Uint64](),
		marshalUnsignedFunc[eth.TransactionType](),

		json.MarshalToFunc(marshalByteSlice),
	)

	// Everything the fast paths miss. Registered against `any`, which json/v2 offers every
	// value, because the rules must also apply to named types declared outside this module,
	// and to big.Int, whose nil pointer form json/v2 would otherwise turn into `null`.
	marshalers = append(marshalers, json.MarshalToFunc(marshalQuantityOrData))

	return json.JoinMarshalers(marshalers...)
}

type signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

func marshalSignedFunc[T signed]() *json.Marshalers {
	return json.MarshalToFunc(func(encoder *jsontext.Encoder, value T) error {
		return writeSignedQuantity(encoder, int64(value))
	})
}

func marshalUnsignedFunc[T unsigned]() *json.Marshalers {
	return json.MarshalToFunc(func(encoder *jsontext.Encoder, value T) error {
		return writeUnsignedQuantity(encoder, uint64(value))
	})
}

func marshalByteSlice(encoder *jsontext.Encoder, value []byte) error {
	return writeData(encoder, value)
}

func marshalEthBytesFunc[T ~[]byte]() *json.Marshalers {
	return json.MarshalToFunc(func(encoder *jsontext.Encoder, value *T) error {
		return writeData(encoder, *value)
	})
}

func marshalMarshalerRPC(encoder *jsontext.Encoder, value MarshalerRPC) error {
	out, err := value.MarshalJSONRPC()
	if err != nil {
		return err
	}

	return encoder.WriteValue(jsontext.Value(out))
}

var (
	bigIntType        = reflect.TypeFor[big.Int]()
	bigIntPtrType     = reflect.TypeFor[*big.Int]()
	marshalerType     = reflect.TypeFor[json.Marshaler]()
	textMarshalerType = reflect.TypeFor[encoding.TextMarshaler]()
)

// marshalQuantityOrData applies the JSON-RPC encoding of the two data types the spec
// defines: QUANTITY, a "0x"-prefixed hex number with no leading zeroes, and DATA, a
// "0x"-prefixed hex byte string.
//
// It returns [errors.ErrUnsupported] for anything it does not recognise, which tells
// json/v2 to fall through to the next marshaler and ultimately to its own default
// behavior.
//
// json/v2 always hands an `any`-typed marshaler a pointer to the value being encoded, so
// the value of interest is one dereference away.
func marshalQuantityOrData(encoder *jsontext.Encoder, value any) error {
	pointer := reflect.ValueOf(value)
	if pointer.Kind() != reflect.Pointer || pointer.IsNil() {
		return errors.ErrUnsupported
	}

	target := pointer.Elem()

	// Before the Marshaler check below, because big.Int has a MarshalJSON method that
	// would emit a JSON number.
	switch target.Type() {
	case bigIntType:
		return writeBigInt(encoder, target.Addr().Interface().(*big.Int))
	case bigIntPtrType:
		// Handled here rather than through a *big.Int marshaler so that a nil pointer
		// still encodes as "0x0"; json/v2 resolves nil pointers to `null` before it
		// reaches the value they point at.
		return writeBigInt(encoder, target.Interface().(*big.Int))
	}

	// A type that spells out its own JSON or text representation keeps it, matching the
	// precedence the v1 fork applied.
	if implementsMarshaler(target.Type()) {
		return errors.ErrUnsupported
	}

	switch target.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return writeSignedQuantity(encoder, target.Int())

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return writeUnsignedQuantity(encoder, target.Uint())

	case reflect.Slice:
		if target.Type().Elem().Kind() != reflect.Uint8 {
			return errors.ErrUnsupported
		}

		return writeData(encoder, target.Bytes())

	case reflect.Array:
		if target.Type().Elem().Kind() != reflect.Uint8 {
			return errors.ErrUnsupported
		}

		// The v1 fork gave byte arrays no special treatment, so each element came out as
		// its own quantity. json/v2 would otherwise encode them as a single base64 string.
		if err := encoder.WriteToken(jsontext.BeginArray); err != nil {
			return err
		}

		for i := range target.Len() {
			if err := writeUnsignedQuantity(encoder, target.Index(i).Uint()); err != nil {
				return err
			}
		}

		return encoder.WriteToken(jsontext.EndArray)

	default:
		return errors.ErrUnsupported
	}
}

func implementsMarshaler(t reflect.Type) bool {
	pointer := reflect.PointerTo(t)

	return t.Implements(marshalerType) || pointer.Implements(marshalerType) ||
		t.Implements(textMarshalerType) || pointer.Implements(textMarshalerType)
}

func writeBigInt(encoder *jsontext.Encoder, value *big.Int) error {
	if value == nil {
		return encoder.WriteToken(jsontext.String("0x0"))
	}

	if value.IsUint64() {
		return writeUnsignedQuantity(encoder, value.Uint64())
	}

	// big.Int.Append writes the sign ahead of the digits, so the prefix goes in after it.
	digits := value.Append(nil, 16)

	buf := make([]byte, 0, len(digits)+5)
	buf = append(buf, '"')
	if digits[0] == '-' {
		buf, digits = append(buf, '-'), digits[1:]
	}
	buf = append(append(buf, '0', 'x'), digits...)

	return encoder.WriteValue(append(buf, '"'))
}

// quantityBuffer holds the longest QUANTITY a fixed-width integer can produce: a quote, a
// minus sign, the "0x" prefix, 16 hex digits and a closing quote.
type quantityBuffer [21]byte

// writeSignedQuantity and writeUnsignedQuantity render into a stack buffer rather than
// concatenating strings, which keeps the common case allocation-free. strconv never emits
// leading zeroes, so the digits need no trimming.
//
// A negative value is written as "-0xff", the form go-ethereum's hexutil.EncodeBig uses.
// Note that the JSON-RPC spec has no negative QUANTITY at all, so no node will accept one;
// this only settles which spelling a caller passing a negative value gets.
func writeSignedQuantity(encoder *jsontext.Encoder, value int64) error {
	var buf quantityBuffer

	out := append(buf[:0], '"')

	// Negate in the unsigned domain so that math.MinInt64, which has no positive
	// counterpart, does not overflow.
	magnitude := uint64(value)
	if value < 0 {
		out, magnitude = append(out, '-'), -magnitude
	}

	out = strconv.AppendUint(append(out, '0', 'x'), magnitude, 16)
	return encoder.WriteValue(append(out, '"'))
}

func writeUnsignedQuantity(encoder *jsontext.Encoder, value uint64) error {
	var buf quantityBuffer

	out := strconv.AppendUint(append(buf[:0], '"', '0', 'x'), value, 16)
	return encoder.WriteValue(append(out, '"'))
}

// writeData renders a JSON-RPC DATA value, a "0x"-prefixed hex byte string, in one
// allocation.
func writeData(encoder *jsontext.Encoder, value []byte) error {
	out := make([]byte, 0, 2*len(value)+4)
	out = append(out, '"', '0', 'x')
	out = hex.AppendEncode(out, value)

	return encoder.WriteValue(append(out, '"'))
}

func marshalDecimal[T ~int](encoder *jsontext.Encoder, value T) error {
	return encoder.WriteToken(jsontext.Int(int64(value)))
}
