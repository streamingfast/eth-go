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
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/crypto/sha3"
)

type buffer []byte

func (b buffer) String() string {
	return hex.EncodeToString([]byte(b))
}

type Encoder struct {
	buffer []byte
}

func NewEncoder() *Encoder {
	return &Encoder{}
}

func (e *Encoder) String() string {
	return hex.EncodeToString(e.buffer)
}

func (e *Encoder) Buffer() []byte {
	return e.buffer
}

func (e *Encoder) WriteMethodCall(method *MethodCall) error {
	if len(method.Data) != len(method.MethodDef.Parameters) {
		return fmt.Errorf("method is expecting %d parameters but %d were provided", len(method.MethodDef.Parameters), len(method.Data))
	}

	methodSignature := method.MethodDef.Signature()
	err := e.write("method", nil, methodSignature)
	if err != nil {
		return fmt.Errorf("unable to write method in buffer: %w", err)
	}

	if tracer.Enabled() {
		zlog.Debug("written method name in buffer",
			zap.Stringer("buf", buffer(e.buffer)),
			zap.String("method_name", methodSignature),
		)
	}

	return e.writeParameters(4, method.MethodDef.Parameters, method.Data)
}

// WriteConstructorCall encodes a constructor call to the buffer. Unlike method calls,
// constructor calls do not include a 4-byte method selector - they only contain the
// encoded parameters which get appended to the contract bytecode during deployment.
func (e *Encoder) WriteConstructorCall(constructor *ConstructorCall) error {
	if len(constructor.Data) != len(constructor.ConstructorDef.Parameters) {
		return fmt.Errorf("constructor is expecting %d parameters but %d were provided", len(constructor.ConstructorDef.Parameters), len(constructor.Data))
	}

	if tracer.Enabled() {
		zlog.Debug("encoding constructor call",
			zap.String("signature", constructor.ConstructorDef.Signature()),
			zap.Int("param_count", len(constructor.ConstructorDef.Parameters)),
		)
	}

	// Constructor encoding has no method selector, offset starts at 0
	return e.writeParameters(0, constructor.ConstructorDef.Parameters, constructor.Data)
}

func (e *Encoder) WriteLogData(parameters []*LogParameter, data []interface{}) error {
	asFakeMethodParams := make([]*MethodParameter, len(parameters))
	for i, param := range parameters {
		asFakeMethodParams[i] = &MethodParameter{Name: param.Name, TypeName: param.TypeName, Components: param.Components}
	}

	return e.writeParameters(0, asFakeMethodParams, data)
}

func (e *Encoder) WriteParameters(parameters []*MethodParameter, data []interface{}) error {
	return e.writeParameters(0, parameters, data)
}

func (e *Encoder) writeParameters(methodSelectorOffset int, parameters []*MethodParameter, data []interface{}) error {
	type arrayToInsert struct {
		buffOffset uint64
		typeName   string
		components []*StructComponent
		value      interface{}
	}

	slicesToInsert := []arrayToInsert{}
	for idx, param := range parameters {
		if isDynamicType(param.TypeName, param.Components) {
			slicesToInsert = append(slicesToInsert, arrayToInsert{
				buffOffset: uint64(len(e.buffer)),
				typeName:   param.TypeName,
				components: param.Components,
				value:      data[idx],
			})

			if tracer.Enabled() {
				zlog.Debug("writting placeholder offset in buffer", zap.String("input_type", param.TypeName), zap.Int("input_idx", idx))
			}

			if err := e.write("uint64", nil, uint64(0)); err != nil {
				return fmt.Errorf("unable to write slice placeholder: %w", err)
			}

			if tracer.Enabled() {
				zlog.Debug("written slice placeholder in buffer",
					zap.String("input_type", param.TypeName),
					zap.Int("input_idx", idx),
				)
			}

			continue
		}

		if err := e.write(param.TypeName, param.Components, data[idx]); err != nil {
			return fmt.Errorf("unable to write input.%d %q in buffer: %w", idx, param.TypeName, err)
		}

		if tracer.Enabled() {
			zlog.Debug("written input data in buffer",
				zap.Stringer("buf", buffer(e.buffer)),
				zap.String("input_type", param.TypeName),
				zap.Int("input_idx", idx),
			)
		}
	}

	for sidx, slc := range slicesToInsert {
		// Offset should not include the signatures' bytes if present, the `methodSelectorOffset` argument represents that
		dataLength := uint64(len(e.buffer)) - uint64(methodSelectorOffset)
		d, err := e.encodeUint(dataLength, 64)
		if err != nil {
			return fmt.Errorf("unable to encode slice offset: %w", err)
		}

		err = e.override(slc.buffOffset, d)
		if err != nil {
			return fmt.Errorf("unable to insert slice offset in buffer: %w", err)
		}

		if tracer.Enabled() {
			zlog.Debug("inserted slice offset in buffer",
				zap.String("input_type", slc.typeName),
				zap.Int("slice_idx", sidx),
			)
		}

		err = e.write(slc.typeName, slc.components, slc.value)
		if err != nil {
			return fmt.Errorf("unable to write slice in buffer: %w", err)
		}

		if tracer.Enabled() {
			zlog.Debug("inserted slice in buffer",
				zap.Stringer("buf", buffer(e.buffer)),
				zap.String("input_tyewpe", slc.typeName),
				zap.Int("slice_idx", sidx),
			)
		}
	}

	return nil
}

func (e *Encoder) Write(parameter *MethodParameter, in interface{}) error {
	return e.write(parameter.TypeName, parameter.Components, in)
}

func (e *Encoder) WriteLogParameter(parameter *LogParameter, in interface{}) error {
	return e.write(parameter.TypeName, parameter.Components, in)
}

func (e *Encoder) write(typeName string, components []*StructComponent, in interface{}) error {
	var isAnArray bool
	isAnArray, resolvedTypeName := isArray(typeName)
	if !isAnArray {
		return e.writeElement(resolvedTypeName, components, in)
	}

	s := reflect.ValueOf(in)
	switch s.Kind() {
	case reflect.Slice:
		if tracer.Enabled() {
			zlog.Debug("writing length of array", zap.String("typeName", typeName), zap.Int("length", s.Len()))
		}

		err := e.writeElement("uint64", nil, uint64(s.Len()))
		if err != nil {
			return fmt.Errorf("cannot write slice %s size: %w", typeName, err)
		}

		if tracer.Enabled() {
			zlog.Debug("writing elements of array", zap.String("typeName", typeName))
		}

		for i := 0; i < s.Len(); i++ {
			err := e.writeElement(resolvedTypeName, components, s.Index(i).Interface())
			if err != nil {
				return fmt.Errorf("cannot write item from slice %s.%d: %w", typeName, i, err)
			}
		}

		if tracer.Enabled() {
			zlog.Debug("ended writing elements of array", zap.String("typeName", typeName))
		}

		return nil
	}
	return fmt.Errorf("writing type %q is not handled right now", typeName)
}

func (e *Encoder) writeElement(typeName string, components []*StructComponent, in interface{}) error {
	if tracer.Enabled() {
		zlog.Debug("writing element", zap.String("typeName", typeName), zap.Bool("has_components", len(components) > 0))
	}

	var d []byte
	var err error
	switch typeName {
	case "bool":
		if v, ok := in.(bool); !ok {
			err = fmt.Errorf("type %q input should be bool, got %T", typeName, v)
		} else {
			d, err = e.encodeBool(v)
		}
	case "uint8":
		d, err = e.encodeUintFromInterface(in, 8)
	case "uint16":
		d, err = e.encodeUintFromInterface(in, 16)
	case "uint24":
		d, err = e.encodeUintFromInterface(in, 24)
	case "uint32":
		d, err = e.encodeUintFromInterface(in, 32)
	case "uint40":
		d, err = e.encodeUintFromInterface(in, 40)
	case "uint48":
		d, err = e.encodeUintFromInterface(in, 48)
	case "uint56":
		d, err = e.encodeUintFromInterface(in, 56)
	case "uint64":
		d, err = e.encodeUintFromInterface(in, 64)
	case "uint72", "uint80", "uint88", "uint96", "uint104", "uint112", "uint120", "uint128", "uint136", "uint144", "uint152", "uint160", "uint168", "uint176", "uint184", "uint192", "uint200", "uint208", "uint216", "uint224", "uint232", "uint240", "uint248", "uint256":
		switch v := in.(type) {
		case big.Int:
			d, err = e.encodeBigInt(&v)
		case *big.Int:
			d, err = e.encodeBigInt(v)
		default:
			err = fmt.Errorf("type %q input should be big.Int or *big.Int, got %T", typeName, v)
		}

	case "method":
		if v, ok := in.(string); !ok {
			err = fmt.Errorf("type %q input should be string, got %T", typeName, v)
		} else {
			d, err = e.encodeMethod(v)
		}
	case "address":
		if v, ok := in.(Address); !ok {
			err = fmt.Errorf("type %q input should be eth.Address, got %T", typeName, v)
		} else {
			d, err = e.encodeAddress(v)
		}
	case "string":
		if v, ok := in.(string); !ok {
			err = fmt.Errorf("type %q input should be string, got %T", typeName, v)
		} else {
			d, err = e.encodeString(v)
		}
	case "bytes":
		d, err = e.encodeBytesFromInterface(in)
	case "bytes1", "bytes2", "bytes3", "bytes4", "bytes5", "bytes6", "bytes7", "bytes8",
		"bytes9", "bytes10", "bytes11", "bytes12", "bytes13", "bytes14", "bytes15", "bytes16",
		"bytes17", "bytes18", "bytes19", "bytes20", "bytes21", "bytes22", "bytes23", "bytes24",
		"bytes25", "bytes26", "bytes27", "bytes28", "bytes29", "bytes30", "bytes31", "bytes32":
		// no need to catch error, will never fail as it will be bytes1...bytes32
		var input []byte

		switch v := in.(type) {
		case [4]byte:
			fourBytes := in.([4]byte)
			input = fourBytes[:]
		case []byte:
			input = v
		default:
			return fmt.Errorf("unsupported input type %T", in)
		}

		data, _ := strconv.ParseUint(strings.TrimPrefix(typeName, "bytes"), 10, 64)
		d, err = e.encodeFixedBytes(input, data)
	case "event":
		if v, ok := in.(string); !ok {
			err = fmt.Errorf("type %q input should be string, got %T", typeName, v)
		} else {
			d, err = e.encodeEvent(v)
		}
	case "tuple":
		return e.writeTuple("<unknown>", components, in)

	default:
		return fmt.Errorf("writing element type %q is not handled right now", typeName)
	}

	if err != nil {
		return err
	}

	if tracer.Enabled() {
		zlog.Debug("appending to buffer", zap.String("typeName", typeName), zap.Int("actual_offset", len(e.buffer)), zap.Int("new_offset", len(e.buffer)+len(d)), zap.String("bytes", hex.EncodeToString(d)))
	}
	e.buffer = append(e.buffer, d...)
	return nil
}

// writeTuple writes a tuple (defined as a `struct` in Solidity code) to the buffer. The components are the
// ordered definition of fields that form that structure. The `in` is the actual Go type that we should use
// to resolve the struct components. Here all the future types that we should handle:
//
// - Go struct and reflection to resolve the Go fields against the components
// - `map[string]interface{}“ to resolve the Go fields against the components
// - `[]interface{}“ to resolve the element against the components
//
// **Important** Right now, only []interface{} is supported.
func (e *Encoder) writeTuple(structName string, components []*StructComponent, in interface{}) error {
	switch v := in.(type) {
	case []interface{}:
		return e.writeTupleFromSlice(structName, components, v)

	case map[string]interface{}:
		return e.writeTupleFromMap(structName, components, v)

	default:
		if in != nil {
			rv := reflect.Indirect(reflect.ValueOf(in))
			if rv.Kind() == reflect.Struct {
				return e.writeTupleFromStruct(structName, components, rv)
			}
		}

		return fmt.Errorf("invalid input type %T when encoding struct %s, only `[]interface{}` and `map[string]interface{}` are supported", v, structName)
	}
}

func (e *Encoder) writeTupleFromSlice(structName string, components []*StructComponent, in []interface{}) error {
	if len(in) != len(components) {
		return fmt.Errorf(`input "[]interface{}" value has %d elements, but struct %q has %d fields`, len(in), structName, len(components))
	}

	// Check if any component is dynamic - if so, we need offset-based encoding
	hasDynamic := false
	for _, component := range components {
		if isDynamicType(component.TypeName, component.Components) {
			hasDynamic = true
			break
		}
	}

	// If no dynamic components, just write inline
	if !hasDynamic {
		for i, fieldIn := range in {
			if err := e.writeComponent(structName, components[i], fieldIn); err != nil {
				return err
			}
		}
		return nil
	}

	// For tuples with dynamic components, use offset-based encoding
	type dynamicToInsert struct {
		buffOffset uint64
		component  *StructComponent
		value      interface{}
	}

	dynamicItems := []dynamicToInsert{}
	tupleStartOffset := uint64(len(e.buffer))

	for i, component := range components {
		if isDynamicType(component.TypeName, component.Components) {
			// Write placeholder offset
			dynamicItems = append(dynamicItems, dynamicToInsert{
				buffOffset: uint64(len(e.buffer)),
				component:  component,
				value:      in[i],
			})

			if err := e.write("uint64", nil, uint64(0)); err != nil {
				return fmt.Errorf("unable to write dynamic component placeholder: %w", err)
			}
		} else {
			// Write static component inline
			if err := e.writeComponent(structName, component, in[i]); err != nil {
				return err
			}
		}
	}

	// Now write the dynamic data and fill in offsets
	for _, item := range dynamicItems {
		// Calculate offset relative to tuple start
		dataOffset := uint64(len(e.buffer)) - tupleStartOffset
		offsetData, err := e.encodeUint(dataOffset, 64)
		if err != nil {
			return fmt.Errorf("unable to encode dynamic component offset: %w", err)
		}

		if err := e.override(item.buffOffset, offsetData); err != nil {
			return fmt.Errorf("unable to insert dynamic component offset in buffer: %w", err)
		}

		// Write the actual dynamic data
		if err := e.write(item.component.TypeName, item.component.Components, item.value); err != nil {
			return fmt.Errorf("unable to write dynamic component %q: %w", item.component.Name, err)
		}
	}

	return nil
}

func (e *Encoder) writeTupleFromMap(structName string, components []*StructComponent, in map[string]interface{}) error {
	if len(in) != len(components) {
		return fmt.Errorf(`input "map[string]interface{}" value has %d elements, but struct %q has %d fields`, len(in), structName, len(components))
	}

	// Check if any component is dynamic - if so, we need offset-based encoding
	hasDynamic := false
	for _, component := range components {
		if isDynamicType(component.TypeName, component.Components) {
			hasDynamic = true
			break
		}
	}

	// Collect all values first to check they exist
	values := make([]interface{}, len(components))
	for i, component := range components {
		fieldIn, found := in[component.Name]
		if !found {
			return fmt.Errorf(`struct %q has a field %q but it was not found in input "map[string]interface{}" (keys %q)`, structName, component.Name, strings.Join(mapStringInterfaceKeys(in), ", "))
		}
		values[i] = fieldIn
	}

	// If no dynamic components, just write inline
	if !hasDynamic {
		for i, component := range components {
			if err := e.writeComponent(structName, component, values[i]); err != nil {
				return err
			}
		}
		return nil
	}

	// For tuples with dynamic components, use offset-based encoding
	type dynamicToInsert struct {
		buffOffset uint64
		component  *StructComponent
		value      interface{}
	}

	dynamicItems := []dynamicToInsert{}
	tupleStartOffset := uint64(len(e.buffer))

	for i, component := range components {
		if isDynamicType(component.TypeName, component.Components) {
			// Write placeholder offset
			dynamicItems = append(dynamicItems, dynamicToInsert{
				buffOffset: uint64(len(e.buffer)),
				component:  component,
				value:      values[i],
			})

			if err := e.write("uint64", nil, uint64(0)); err != nil {
				return fmt.Errorf("unable to write dynamic component placeholder: %w", err)
			}
		} else {
			// Write static component inline
			if err := e.writeComponent(structName, component, values[i]); err != nil {
				return err
			}
		}
	}

	// Now write the dynamic data and fill in offsets
	for _, item := range dynamicItems {
		// Calculate offset relative to tuple start
		dataOffset := uint64(len(e.buffer)) - tupleStartOffset
		offsetData, err := e.encodeUint(dataOffset, 64)
		if err != nil {
			return fmt.Errorf("unable to encode dynamic component offset: %w", err)
		}

		if err := e.override(item.buffOffset, offsetData); err != nil {
			return fmt.Errorf("unable to insert dynamic component offset in buffer: %w", err)
		}

		// Write the actual dynamic data
		if err := e.write(item.component.TypeName, item.component.Components, item.value); err != nil {
			return fmt.Errorf("unable to write dynamic component %q: %w", item.component.Name, err)
		}
	}

	return nil
}

func (e *Encoder) writeTupleFromStruct(structName string, components []*StructComponent, in reflect.Value) error {
	fieldCount := in.NumField()
	if fieldCount < len(components) {
		return fmt.Errorf(`input %q value has only %d fields, but struct %q has %d fields`, in.Type().String(), fieldCount, structName, len(components))
	}

	// Collect all field values first
	values := make([]interface{}, 0, len(components))
	for i := 0; i < fieldCount; i++ {
		field := in.Field(i)
		if !field.CanInterface() {
			if tracer.Enabled() {
				zlog.Debug("skipping struct field", zap.String("field", in.Type().Field(i).Name))
			}
			continue
		}

		if len(values) >= len(components) {
			return fmt.Errorf(`input %q value with %d fields has more field to write than struct %q which has %d fields`, in.Type().String(), fieldCount, structName, len(components))
		}

		values = append(values, field.Interface())
	}

	if len(values) != len(components) {
		return fmt.Errorf(`input %q value has %d exportable fields, but struct %q has %d fields`, in.Type().String(), len(values), structName, len(components))
	}

	// Check if any component is dynamic - if so, we need offset-based encoding
	hasDynamic := false
	for _, component := range components {
		if isDynamicType(component.TypeName, component.Components) {
			hasDynamic = true
			break
		}
	}

	// If no dynamic components, just write inline
	if !hasDynamic {
		for i, component := range components {
			if err := e.writeComponent(structName, component, values[i]); err != nil {
				return err
			}
		}
		return nil
	}

	// For tuples with dynamic components, use offset-based encoding
	type dynamicToInsert struct {
		buffOffset uint64
		component  *StructComponent
		value      interface{}
	}

	dynamicItems := []dynamicToInsert{}
	tupleStartOffset := uint64(len(e.buffer))

	for i, component := range components {
		if isDynamicType(component.TypeName, component.Components) {
			// Write placeholder offset
			dynamicItems = append(dynamicItems, dynamicToInsert{
				buffOffset: uint64(len(e.buffer)),
				component:  component,
				value:      values[i],
			})

			if err := e.write("uint64", nil, uint64(0)); err != nil {
				return fmt.Errorf("unable to write dynamic component placeholder: %w", err)
			}
		} else {
			// Write static component inline
			if err := e.writeComponent(structName, component, values[i]); err != nil {
				return err
			}
		}
	}

	// Now write the dynamic data and fill in offsets
	for _, item := range dynamicItems {
		// Calculate offset relative to tuple start
		dataOffset := uint64(len(e.buffer)) - tupleStartOffset
		offsetData, err := e.encodeUint(dataOffset, 64)
		if err != nil {
			return fmt.Errorf("unable to encode dynamic component offset: %w", err)
		}

		if err := e.override(item.buffOffset, offsetData); err != nil {
			return fmt.Errorf("unable to insert dynamic component offset in buffer: %w", err)
		}

		// Write the actual dynamic data
		if err := e.write(item.component.TypeName, item.component.Components, item.value); err != nil {
			return fmt.Errorf("unable to write dynamic component %q: %w", item.component.Name, err)
		}
	}

	return nil
}

func (e *Encoder) writeComponent(structName string, component *StructComponent, in interface{}) error {
	if tracer.Enabled() {
		zlog.Debug("about to write struct component", zap.Stringer("component", component), zap.String("input_type", fmt.Sprintf("%T", in)))
	}

	if err := e.writeElement(component.TypeName, component.Components, in); err != nil {
		return fmt.Errorf(`unable to write "%s#%s: %w"`, structName, component.Name, err)
	}

	return nil
}

func (e *Encoder) encodeBytesFromInterface(input interface{}) ([]byte, error) {
	var bytes []byte
	switch v := input.(type) {
	case []byte:
		bytes = v
	case Hex:
		bytes = []byte(v)
	case Hash:
		bytes = []byte(v)
	}

	return e.encodeBytes(bytes)
}

func (e *Encoder) encodeUintFromInterface(input interface{}, size uint64) ([]byte, error) {
	switch v := input.(type) {
	case uint8:
		return e.encodeUint(uint64(v), size)
	case uint16:
		return e.encodeUint(uint64(v), size)
	case uint32:
		return e.encodeUint(uint64(v), size)
	case uint64:
		return e.encodeUint(uint64(v), size)
	case Uint8:
		return e.encodeUint(uint64(v), size)
	case Uint16:
		return e.encodeUint(uint64(v), size)
	case Uint32:
		return e.encodeUint(uint64(v), size)
	case Uint64:
		return e.encodeUint(uint64(v), size)
	case *big.Int:
		return e.encodeUint(v.Uint64(), size)
	default:
		return nil, fmt.Errorf("unsupported uint from type %T", input)
	}
}

func (e *Encoder) encodeUint(input uint64, size uint64) ([]byte, error) {
	byteCount := size / 8
	buf := make([]byte, byteCount)
	_ = buf[byteCount-1] // early bounds check to guarantee safety of writes below
	for i := uint64(0); i < byteCount; i++ {
		shift := (byteCount - 1 - i) * 8
		buf[i] = byte(input >> shift)
	}

	return pad(buf), nil
}

func (e *Encoder) encodeBigInt(input *big.Int) ([]byte, error) {
	return pad(input.Bytes()), nil
}

func (e *Encoder) encodeBool(input bool) ([]byte, error) {
	var v *big.Int
	if input {
		v = big.NewInt(1)
	} else {
		v = big.NewInt(0)
	}
	return pad(v.Bytes()), nil
}

func (e *Encoder) encodeAddress(input Address) ([]byte, error) {
	return pad(input), nil
}

func (e *Encoder) encodeMethod(input string) ([]byte, error) {
	kec := sha3.NewLegacyKeccak256()
	_, err := kec.Write([]byte(input))
	if err != nil {
		return nil, err
	}
	return kec.Sum(nil)[0:4], nil
}

func (e *Encoder) encodeBytes(input []byte) ([]byte, error) {
	// The number of 32 bytes row of data aside the actu
	bankCount := int(math.Ceil(float64(len(input)) / 32.0))

	buf := make([]byte, 32+(bankCount*32))
	l, err := e.encodeUint(uint64(len(input)), 64)
	if err != nil {
		return nil, fmt.Errorf("unable to encode bytes length: %w", err)
	}
	for i := 0; i < 32; i++ {
		buf[i] = l[i]
	}

	i := 0
	for ; i < len(input); i++ {
		buf[32+i] = input[i]
	}

	return buf, nil
}

func (e *Encoder) encodeFixedBytes(input []byte, size uint64) ([]byte, error) {
	if uint64(len(input)) < size {
		return nil, fmt.Errorf("not enough bytes %d, expected size %d", len(input), size)
	}
	buf := make([]byte, 32)
	copy(buf, input[0:size])
	return buf, nil
}

func (e *Encoder) encodeString(input string) ([]byte, error) {
	// size: 32 bytes[length of the string] +  num_char[1 char is 1 byte] + x
	// where x  pads the the number to fill the last 32 bytes
	buf := make([]byte, (32 + len(input) + (32 - len(input)%32)))
	l, err := e.encodeUint(uint64(len(input)), 64)
	if err != nil {
		return nil, fmt.Errorf("unable to encode string size: %w", err)
	}
	for i := 0; i < 32; i++ {
		buf[i] = l[i]
	}
	for i := 0; i < len(input); i++ {
		buf[32+i] = byte(input[i])
	}
	return buf, nil
}

func (e *Encoder) encodeEvent(input string) ([]byte, error) {
	kec := sha3.NewLegacyKeccak256()
	_, err := kec.Write([]byte(input))
	if err != nil {
		return nil, err
	}
	return kec.Sum(nil), nil
}

func (e *Encoder) override(offset uint64, data []byte) error {
	if uint64(len(e.buffer)) < offset+uint64(len(data)) {
		return fmt.Errorf("insuficient room in buffer with length %d to insert data with length %d at offset %d", len(e.buffer), len(data), offset)
	}

	for i := 0; i < len(data); i++ {
		e.buffer[uint64(i)+offset] = data[i]
	}
	return nil
}

func pad(in []byte) []byte {
	d := make([]byte, 32)
	offset := 32 - len(in)
	for i := 0; i < len(in); i++ {
		d[i+offset] = in[i]
	}
	return d
}

// isDynamicType returns true if the given type is considered dynamic according to the
// Solidity ABI specification. A type is dynamic if:
// - It's `bytes` or `string`
// - It's a dynamic array (e.g., `uint256[]`)
// - It's a `tuple` where at least one component is dynamic
func isDynamicType(typeName string, components []*StructComponent) bool {
	// First as they are probably the most probable type
	if typeName == "bytes" || typeName == "string" {
		return true
	}

	arr, resolvedTypeName := isArray(typeName)
	if arr {
		return true
	}

	// A tuple is dynamic if any of its components is dynamic
	if resolvedTypeName == "tuple" || typeName == "tuple" {
		for _, component := range components {
			if isDynamicType(component.TypeName, component.Components) {
				return true
			}
		}
	}

	return false
}

func isArray(typeName string) (bool, string) {
	check := strings.HasSuffix(typeName, "[]")
	if check {
		return true, strings.TrimRight(typeName, "[]")
	}
	return false, typeName
}

func mapStringInterfaceKeys(in map[string]interface{}) (out []string) {
	if len(in) <= 0 {
		return nil
	}

	i := 0
	out = make([]string, len(in))
	for k := range in {
		out[i] = k
		i++
	}

	return out
}
