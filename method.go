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

package eth

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

//go:generate go-enum -f=$GOFILE --lower --marshal --names

// ENUM(
//
//	Pure
//	View
//	NonPayable
//	Payable
//
// )
type StateMutability int

type MethodParameter struct {
	// Name represents the name of the parameter as defined by the
	// developer in the Solidity code of the contract.
	Name string
	// TypeName represents the type of the parameter, this is standard
	// types known to Solidity. Array have a suffix `[]` (can be nested)
	// and struct type is always `tuple` with an filled up `Components`
	// defining the struct.
	TypeName string
	// Type represents the parsed type of this parameter.
	Type SolidityType
	// TypeMutability is unclear, requires more investigation, I don't recall
	// to which Solidity concept it refers to.
	TypeMutability string
	// Payable determines if the parameter is a payable value so an
	// Ether value.
	Payable bool
	// InternalType is the internal type to the contract, usually equal
	// to `TypeName` but can be different for example if a type `uint`
	// was defined, in this case `TypeName` will be `uint256` and internal
	// type will be `uint`. Tuple which have `TypeName` `tuple` but internal
	// type is `struct <Contract>.<Struct>` is another exceptions.
	InternalType string
	// Components represents that struct fields of a particular tuple. Only
	// filled up when `TypeName` is equal to `tuple` (or array of tuples).
	Components []*StructComponent
}

func newMethodParameter(mStr string) (*MethodParameter, error) {
	mStr = strings.TrimLeft(mStr, " ")
	mStr = strings.TrimRight(mStr, " ")
	if mStr == "" {
		return nil, fmt.Errorf("invalid method parameter")
	}
	chunks := strings.Split(mStr, " ")
	// TODO: we should check the type
	m := &MethodParameter{TypeName: chunks[0]}
	if len(chunks) > 1 {
		m.Name = chunks[len(chunks)-1]
	}
	return m, nil
}

func (p *MethodParameter) Signature() string {
	typeName := p.TypeName
	if strings.HasPrefix(typeName, "tuple") {
		componentSigs := make([]string, len(p.Components))
		for i, component := range p.Components {
			componentSigs[i] = component.Signature()
		}

		typeName = strings.Replace(typeName, "tuple", fmt.Sprintf("(%s)", strings.Join(componentSigs, ",")), 1)
	}

	return typeName
}

type MethodDef struct {
	Name             string
	Parameters       []*MethodParameter
	ReturnParameters []*MethodParameter
	StateMutability  StateMutability
}

// ConstructorDef represents the constructor of a Solidity contract.
// Unlike methods, constructors have no name and no return parameters.
type ConstructorDef struct {
	Parameters      []*MethodParameter
	StateMutability StateMutability
}

// Signature returns the canonical signature of the constructor parameters
// in the format "(type1,type2,...)".
func (c *ConstructorDef) Signature() string {
	var args []string
	for _, parameter := range c.Parameters {
		args = append(args, parameter.Signature())
	}

	return fmt.Sprintf("(%s)", strings.Join(args, ","))
}

// String returns a human-readable representation of the constructor.
func (c *ConstructorDef) String() string {
	var args []string
	for _, parameter := range c.Parameters {
		args = append(args, fmt.Sprintf("%s %s", parameter.TypeName, parameter.Name))
	}

	return fmt.Sprintf("constructor(%s)", strings.Join(args, ","))
}

// NewCall instantiates a new call from the constructor definition and uses
// the received arguments as elements used to resolve the parameters values.
//
// A constructor call can be encoded to produce the data that should be
// appended to the contract bytecode when deploying the contract.
func (c *ConstructorDef) NewCall(args ...interface{}) *ConstructorCall {
	call := &ConstructorCall{ConstructorDef: c}
	if len(args) > 0 {
		call.Data = make([]interface{}, len(args))
	}

	for i, arg := range args {
		call.Data[i] = arg
	}

	return call
}

// NewCallFromString works exactly like NewCall except that it assumes all
// arguments are string representations of the actual Ethereum types defined
// by the constructor and converts them to the correct types.
func (c *ConstructorDef) NewCallFromString(args ...string) *ConstructorCall {
	call := &ConstructorCall{ConstructorDef: c}

	for _, arg := range args {
		call.AppendArgFromString(arg)
	}

	return call
}

// ConstructorCall represents a specific invocation of a constructor with
// concrete parameter values. It can be encoded to produce the ABI-encoded
// constructor arguments.
type ConstructorCall struct {
	ConstructorDef *ConstructorDef
	Data           []interface{}

	err []error
}

// AppendArgFromString appends an argument by converting its string representation
// to the appropriate type based on the constructor parameter definition.
func (c *ConstructorCall) AppendArgFromString(v string) {
	i := len(c.Data)
	if i >= len(c.ConstructorDef.Parameters) {
		c.err = append(c.err, fmt.Errorf("args exceeds constructor parameter count %d", len(c.ConstructorDef.Parameters)))
		return
	}

	param := c.ConstructorDef.Parameters[i]
	out, err := argToDataType(v, param.TypeName, param.Components)
	if err != nil {
		c.err = append(c.err, fmt.Errorf("invalid string argument %q for parameter %s: %w", param.Name, v, err))
		return
	}

	c.Data = append(c.Data, out)
}

// AppendArg appends a raw argument value to the constructor call data.
func (c *ConstructorCall) AppendArg(v interface{}) {
	c.Data = append(c.Data, v)
}

// MustEncode encodes the constructor call and panics if encoding fails.
func (c *ConstructorCall) MustEncode() []byte {
	out, err := c.Encode()
	if err != nil {
		panic(fmt.Errorf("unable to encode constructor call: %w", err))
	}

	return out
}

// Encode encodes the constructor call arguments according to the Ethereum ABI
// specification. Unlike method calls, constructor encoding does not include
// a 4-byte method selector - it's just the encoded parameters.
func (c *ConstructorCall) Encode() ([]byte, error) {
	if len(c.err) > 0 {
		return nil, fmt.Errorf("%s", c.err)
	}

	enc := NewEncoder()
	err := enc.WriteConstructorCall(c)
	if err != nil {
		return nil, err
	}
	return enc.Buffer(), nil
}

func MustNewMethodDef(signature string) *MethodDef {
	def, err := NewMethodDef(signature)
	if err != nil {
		panic(fmt.Errorf("invalid method definition %q: %w", signature, err))
	}

	return def
}

func NewMethodDef(signature string) (*MethodDef, error) {
	method, inputs, outputs, err := parseSignature(signature)
	if err != nil {
		return nil, fmt.Errorf("invalid signature %q: %w", signature, err)
	}

	return &MethodDef{
		Name:             method,
		Parameters:       inputs,
		ReturnParameters: outputs,
	}, nil
}

// NewCall instantiate a new call from the method definition and uses
// the received arguments as elements used to resolve the parameters
// values.
//
// A call is a particular instance of a Method where ultimately the
// parameters's value will be resolved. A method call in opposition
// to a method definition can be encoded according to Ethereum rules
// or decode data returned for this call against the definition.
func (f *MethodDef) NewCall(args ...interface{}) *MethodCall {
	call := &MethodCall{MethodDef: f}
	if len(args) > 0 {
		call.Data = make([]interface{}, len(args))
	}

	for i, arg := range args {
		call.Data[i] = arg
	}

	return call
}

// NewCallFromString works exactly like `NewCall“ except that it
// actually assumes all arguments are string version of the actual
// Ethereum types defined by the method and append them to the Data
// slice by calling `AppendArgFromString` which converts the string
// representation to the correct type.
func (f *MethodDef) NewCallFromString(args ...string) *MethodCall {
	call := &MethodCall{MethodDef: f}

	for _, arg := range args {
		call.AppendArgFromString(arg)
	}

	return call
}

func (f *MethodDef) MethodID() []byte {
	return Keccak256([]byte(f.Signature()))[0:4]
}

func (f *MethodDef) Signature() string {
	var args []string
	for _, parameter := range f.Parameters {
		args = append(args, parameter.Signature())
	}

	return fmt.Sprintf("%s(%s)", f.Name, strings.Join(args, ","))
}

func (f *MethodDef) String() string {
	var args []string
	for _, parameter := range f.Parameters {
		args = append(args, fmt.Sprintf("%s %s", parameter.TypeName, parameter.Name))
	}

	return fmt.Sprintf("%s(%s)", f.Name, strings.Join(args, ","))
}

func (f *MethodDef) DecodeOutput(data []byte) ([]interface{}, error) {
	if len(f.ReturnParameters) == 0 {
		return nil, fmt.Errorf("no return parameters defined for method")
	}

	return NewDecoder(data).ReadOutput(f.ReturnParameters)
}

func (f *MethodDef) DecodeOutputFromString(data string) ([]interface{}, error) {
	if len(f.ReturnParameters) == 0 {
		return nil, fmt.Errorf("no return parameters defined for method")
	}

	decoder, err := NewDecoderFromString(data)
	if err != nil {
		return nil, fmt.Errorf("data is not a valid hexadecimal value")
	}

	return decoder.ReadOutput(f.ReturnParameters)
}

func (f *MethodDef) DecodeToObjectFromBytes(data []byte) (out map[string]interface{}, err error) {
	return f.DecodeToObjectFromDecoder(NewDecoder(data))
}

func (f *MethodDef) DecodeToObjectFromString(data string) (out map[string]interface{}, err error) {
	decoder, err := NewDecoderFromString(data)
	if err != nil {
		return nil, fmt.Errorf("data is not a valid hexadecimal value")
	}

	return f.DecodeToObjectFromDecoder(decoder)
}

func (f *MethodDef) DecodeToObjectFromDecoder(decoder *Decoder) (out map[string]interface{}, err error) {
	if len(f.ReturnParameters) == 0 {
		return nil, fmt.Errorf("no return parameters defined for method")
	}

	values, err := decoder.ReadOutput(f.ReturnParameters)
	if err != nil {
		return nil, fmt.Errorf("unable to read output")
	}

	out = make(map[string]interface{})
	for i, returnParm := range f.ReturnParameters {
		fieldName := returnParm.Name
		if fieldName == "" {
			if len(f.ReturnParameters) == 1 {
				fieldName = f.Name
			} else {
				fieldName = fmt.Sprintf("%s%d", f.Name, i)
			}
		}

		zlog.Debug("object decoding assigning new field", zap.String("field_name", fieldName), zap.Reflect("value", values[i]))
		out[fieldName] = values[i]
	}

	return out, nil
}

type MethodCall struct {
	MethodDef *MethodDef
	Data      []interface{}

	err []error
}

func (f *MethodCall) AppendArgFromString(v string) {
	i := len(f.Data)
	if i >= len(f.MethodDef.Parameters) {
		f.err = append(f.err, fmt.Errorf("args exceeds method definition parameter count %d", len(f.MethodDef.Parameters)))
		return
	}

	param := f.MethodDef.Parameters[i]
	out, err := argToDataType(v, param.TypeName, param.Components)
	if err != nil {
		f.err = append(f.err, fmt.Errorf("invalid string argument %q for parameter %s: %w", param.Name, v, err))
		return
	}

	f.Data = append(f.Data, out)
}

func argToDataType(in interface{}, typeName string, components []*StructComponent) (interface{}, error) {
	switch typeName {
	case "string":
		if v, ok := in.(string); ok {
			return v, nil
		}

	case "bytes":
		switch v := in.(type) {
		case string:
			data, err := NewHex(v)
			if err != nil {
				return nil, fmt.Errorf("invalid hex: %w", err)
			}

			return data, nil
		case []byte:
			return v, nil
		}

	case "bytes32":
		switch v := in.(type) {
		case string:
			data, err := NewHash(v)
			if err != nil {
				return nil, fmt.Errorf("invalid bytes32: %w", err)
			}

			return data, nil
		case []byte:
			return v, nil
		}

	case "address[]":
		if v, ok := in.(string); ok {
			var addrs []Address
			err := json.Unmarshal([]byte(v), &addrs)
			if err != nil {
				return nil, fmt.Errorf("invalid JSON address array: %w", err)
			}

			return addrs, nil
		}

	case "address":
		switch v := in.(type) {
		case string:
			addr, err := NewAddress(v)
			if err != nil {
				return nil, fmt.Errorf("invalid address: %w", err)
			}

			return addr, nil
		case []byte:
			return Address(v), nil
		}

	case "uint8":
		switch v := in.(type) {
		case string:
			var value Uint8
			if err := value.UnmarshalText([]byte(v)); err != nil {
				return nil, fmt.Errorf("invalid uint8: %w", err)
			}

			return value, nil

		// Type float64 arise when parsing JSON numbers
		case float64:
			return Uint8(v), nil
		}

	case "uint16":
		switch v := in.(type) {
		case string:
			var value Uint16
			if err := value.UnmarshalText([]byte(v)); err != nil {
				return nil, fmt.Errorf("invalid uint16: %w", err)
			}

			return value, nil

		// Type float64 arise when parsing JSON numbers
		case float64:
			return Uint16(v), nil
		}

	case "uint24", "uint32":
		switch v := in.(type) {
		case string:
			var value Uint32
			if err := value.UnmarshalText([]byte(v)); err != nil {
				return nil, fmt.Errorf("invalid %s: %w", typeName, err)
			}

			return value, nil

		// Type float64 arise when parsing JSON numbers
		case float64:
			return Uint32(v), nil
		}

	case "uint40", "uint48", "uint56", "uint64":
		switch v := in.(type) {
		case string:
			var value Uint64
			if err := value.UnmarshalText([]byte(v)); err != nil {
				return nil, fmt.Errorf("invalid %s: %w", typeName, err)
			}

			return value, nil

		// Type float64 arise when parsing JSON numbers
		case float64:
			return Uint64(v), nil
		}

	case "uint72", "uint80", "uint88", "uint96", "uint104", "uint112", "uint120", "uint128", "uint136", "uint144", "uint152", "uint160", "uint168", "uint176", "uint184", "uint192", "uint200", "uint208", "uint216", "uint224", "uint232", "uint240", "uint248", "uint256":
		switch v := in.(type) {
		case string:
			out, ok := new(big.Int).SetString(v, 0)
			if !ok {
				return nil, fmt.Errorf("invalid %s", typeName)
			}

			return out, nil

		// Type float64 arise when parsing JSON numbers
		case float64:
			return new(big.Int).SetUint64(uint64(v)), nil
		}

	case "tuple":
		switch v := in.(type) {
		case string:
			var t interface{}
			if err := json.Unmarshal([]byte(v), &t); err != nil {
				return nil, fmt.Errorf("invalid JSON %w", err)
			}

			switch vt := t.(type) {
			case []interface{}:
				return tupleInterfaceSliceToDataType(vt, components)

			case map[string]interface{}:
				return tupleMapSliceToDataType(vt, components)

			default:
				return nil, fmt.Errorf("accepting only JSON array or JSON object, got %T", vt)
			}

		case []interface{}:
			return tupleInterfaceSliceToDataType(v, components)

		case map[string]interface{}:
			return tupleMapSliceToDataType(v, components)
		}

	case "tuple[]":
		switch v := in.(type) {
		case string:
			var t interface{}
			if err := json.Unmarshal([]byte(v), &t); err != nil {
				return nil, fmt.Errorf("invalid JSON %w", err)
			}

			switch vt := t.(type) {
			case []interface{}:
				return tupleArrayInterfaceSliceToDataType(vt, components)

			default:
				return nil, fmt.Errorf("accepting only JSON array, got %T", vt)
			}

		case []interface{}:
			return tupleArrayInterfaceSliceToDataType(v, components)

		}

	case "bool":
		return in == "true", nil

	default:
		return nil, fmt.Errorf("unsupported type %s", typeName)
	}

	return nil, fmt.Errorf("converting %T to type %s is unsupported", in, typeName)
}

func tupleInterfaceSliceToDataType(in []interface{}, components []*StructComponent) (out interface{}, err error) {
	if len(in) != len(components) {
		return nil, fmt.Errorf(`input "[]interface{}" value has %d elements, but there is %d struct components`, len(in), len(components))
	}

	elements := make([]interface{}, len(components))
	for i, component := range components {
		elements[i], err = argToDataType(in[i], component.TypeName, nil)
		if err != nil {
			return nil, fmt.Errorf("unable to transfrom struct component %s from input type %T: %w", component.Name, in[i], err)
		}
	}

	return elements, nil
}

func tupleArrayInterfaceSliceToDataType(in []interface{}, components []*StructComponent) (out interface{}, err error) {
	elements := make([]interface{}, len(in))
	for i := range in {
		elements[i], err = argToDataType(in[i], "tuple", components)
		if err != nil {
			return nil, fmt.Errorf("unable to transfrom index %d of tuple array from input type %T: %w", i, in[i], err)
		}
	}

	return elements, nil
}

func tupleMapSliceToDataType(in map[string]interface{}, components []*StructComponent) (out interface{}, err error) {
	if len(in) != len(components) {
		return nil, fmt.Errorf(`input "map[string]interface{}" value has %d elements, but there is %d struct components`, len(in), len(components))
	}

	i := 0
	elements := make([]interface{}, len(components))
	for _, component := range components {
		fieldIn, found := in[component.Name]
		if !found {
			return fmt.Errorf(`struct component %s was not found in input "map[string]interface{}" (keys %q)`, component.Name, strings.Join(mapStringInterfaceKeys(in), ", ")), nil
		}

		elements[i], err = argToDataType(fieldIn, component.TypeName, nil)
		if err != nil {
			return nil, fmt.Errorf("unable to transfrom struct component %s from input type %T: %w", component.Name, fieldIn, err)
		}

		i++
	}

	return elements, nil
}

func (f *MethodCall) AppendArg(v interface{}) {
	f.Data = append(f.Data, v)
}

func (f *MethodCall) MustEncode() []byte {
	out, err := f.Encode()
	if err != nil {
		panic(fmt.Errorf("unable to encode method call: %w", err))
	}

	return out
}

func (f *MethodCall) Encode() ([]byte, error) {
	if len(f.err) > 0 {
		return nil, fmt.Errorf("%s", f.err)
	}
	enc := NewEncoder()
	err := enc.WriteMethodCall(f)
	if err != nil {
		return nil, err
	}
	return enc.Buffer(), nil
}

func (f *MethodCall) MarshalJSONRPC() ([]byte, error) {
	if len(f.err) > 0 {
		return nil, fmt.Errorf("%s", f.err)
	}

	enc := Encoder{}
	err := enc.WriteMethodCall(f)
	if err != nil {
		return nil, err
	}

	return []byte(`"0x` + enc.String() + `"`), nil
}

var identifierPart = `([a-zA-Z$_][a-zA-Z0-9$_]*)`
var methodRegex = regexp.MustCompile(identifierPart + `\(` + `([^\)]*)` + `\)` + `\s*(returns)?\s*` + `(\(` + `([^\)]*)` + `\))?`)
var methodRegexGroupCount = 6

func parseSignature(signature string) (method string, inputs []*MethodParameter, outputs []*MethodParameter, err error) {
	matches := methodRegex.FindAllStringSubmatch(signature, 1)
	if len(matches) == 0 {
		return "", nil, nil, fmt.Errorf("invalid signature: %s", signature)
	}

	match := matches[0]
	if tracer.Enabled() {
		zlog.Debug("got a match for signature", zap.Int("count", len(match)), zap.Strings("groups", match))
	}

	if len(match) != methodRegexGroupCount {
		panic(fmt.Errorf("method regex was modified without updating code, expected %d groups, got %d", methodRegexGroupCount, len(match)))
	}

	method = match[1]

	inputList := match[2]
	if inputList != "" {
		inputs = parseParameterList(inputList)
	}

	returnsList := match[5]
	if returnsList != "" {
		outputs = parseParameterList(returnsList)
	}

	return
}

var typeNamePart = `(([a-z0-9]+)(\s+(payable|calldata|memory|storage))?(\[\])?)`
var parameterRegex = regexp.MustCompile(typeNamePart + `(\s+` + identifierPart + `)?`)
var parameterRegexGroupCount = 8

func parseParameterList(list string) (out []*MethodParameter) {
	matches := parameterRegex.FindAllStringSubmatch(list, math.MaxInt64)
	if len(matches) <= 0 {
		return nil
	}

	out = make([]*MethodParameter, len(matches))
	for i, match := range matches {
		if tracer.Enabled() {
			zlog.Debug("got a match for parameter", zap.Int("count", len(match)), zap.Strings("groups", match))
		}

		if len(match) != parameterRegexGroupCount {
			panic(fmt.Errorf("parameter regex was modified without updating code, expected %d groups, got %d", parameterRegexGroupCount, len(match)))
		}

		parameter := &MethodParameter{TypeName: match[2], Payable: match[4] == "payable"}
		if match[5] != "" {
			parameter.TypeName += "[]"
		}

		if match[7] != "" {
			parameter.Name = match[7]
		}

		out[i] = parameter
	}
	return
}
