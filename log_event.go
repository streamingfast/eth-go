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
	"fmt"
	"strings"
)

type Log struct {
	Address     []byte   `json:"address,omitempty"`
	Topics      [][]byte `json:"topics,omitempty"`
	Data        []byte   `json:"data,omitempty"`
	IsAnonymous bool     `json:"anonymous,omitempty"`

	// supplement
	Index      uint32 `json:"index,omitempty"`
	BlockIndex uint32 `json:"blockIndex,omitempty"`
}

type LogEventDef struct {
	Name       string
	Parameters []*LogParameter
}

type LogParameter struct {
	Name     string
	TypeName string
	Type     SolidityType
	Indexed  bool

	// Components represents that struct fields of a particular tuple. Only
	// filled up when `TypeName` is equal to `tuple` (or array of tuples).
	//
	// TODO: This is a blind copy from Method as I think the same concept applies
	// to event, needs to validate using a test in `eth-go`.
	Components []*StructComponent
}

// Signature returns the canonical type signature for this parameter,
// recursively expanding nested tuples.
func (p *LogParameter) Signature() string {
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

// returned instantiate a new event from the log definition and uses
// the received arguments as elements used to resolve the parameters
// values (indexed and non-indexed).
//
// A call is a particular instance of a Method where ultimately the
// parameters's value will be resolved. A method call in opposition
// to a method definition can be encoded according to Ethereum rules
// or decode data returned for this call against the definition.
func (f *LogEventDef) NewEvent(args ...interface{}) *LogEvent {
	event := &LogEvent{Def: f}
	if len(args) > 0 {
		event.Data = make([]interface{}, len(args))
	}

	for i, arg := range args {
		event.Data[i] = arg
	}

	return event
}

// NewCallFromString works exactly like `NewCall“ except that it
// actually assumes all arguments are string version of the actual
// Ethereum types defined by the method and append them to the Data
// slice by calling `AppendArgFromString` which converts the string
// representation to the correct type.
func (f *LogEventDef) NewEventFromString(args ...string) *LogEvent {
	event := &LogEvent{Def: f}

	for _, arg := range args {
		event.AppendArgFromString(arg)
	}

	return event
}

type LogEvent struct {
	Def  *LogEventDef
	Data []interface{}

	err []error
}

func (f *LogEvent) AppendArgFromString(v string) {
	i := len(f.Data)
	if i >= len(f.Def.Parameters) {
		f.err = append(f.err, fmt.Errorf("args exceeds method definition parameter count %d", len(f.Def.Parameters)))
		return
	}

	param := f.Def.Parameters[i]
	out, err := argToDataType(v, param.TypeName, param.Components)
	if err != nil {
		f.err = append(f.err, fmt.Errorf("invalid string argument %q for parameter %s: %w", param.Name, v, err))
		return
	}

	f.Data = append(f.Data, out)
}

func (f *LogEvent) MustEncode() (topics []Topic, data []byte) {
	topics, data, err := f.Encode()
	if err != nil {
		panic(fmt.Errorf("unable to encode log event: %w", err))
	}

	return topics, data
}

func (f *LogEvent) Encode() (topics []Topic, data []byte, err error) {
	if len(f.err) > 0 {
		return nil, nil, fmt.Errorf("%s", f.err)
	}

	if len(f.Data) != len(f.Def.Parameters) {
		return nil, nil, fmt.Errorf("event has %d parameters but %d were provided", len(f.Def.Parameters), len(f.Data))
	}

	var indexedParameters []*LogParameter
	var indexesFieldValues []interface{}

	var unindexedParameters []*LogParameter
	var unindexesDataValues []interface{}

	for i, param := range f.Def.Parameters {
		if param.Indexed {
			indexedParameters = append(indexedParameters, param)
			indexesFieldValues = append(indexesFieldValues, f.Data[i])
		} else {
			unindexedParameters = append(unindexedParameters, param)
			unindexesDataValues = append(unindexesDataValues, f.Data[i])
		}
	}

	topics = make([]Topic, len(indexedParameters)+1)
	copy(topics[0][:], f.Def.LogID())

	if len(indexedParameters) > 0 {
		for i, param := range indexedParameters {
			enc := NewEncoder()
			if err := enc.WriteLogParameter(param, indexesFieldValues[i]); err != nil {
				return nil, nil, fmt.Errorf("unable to encode %q (topic%d): %w", param.Name, i+1, err)
			}

			copy(topics[i+1][:], enc.Buffer())
		}
	}

	enc := NewEncoder()
	err = enc.WriteLogData(unindexedParameters, unindexesDataValues)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to encode log data: %w", err)
	}

	return topics, enc.Buffer(), err
}

func (p *LogParameter) GetName(index int) string {
	name := p.Name
	if name == "" {
		name = fmt.Sprintf("unamed%d", index+1)
	}

	return name
}

func (l *LogEventDef) LogID() []byte {
	return Keccak256([]byte(l.Signature()))
}

func (l *LogEventDef) Signature() string {
	var args []string
	for _, parameter := range l.Parameters {
		args = append(args, parameter.Signature())
	}

	// It's important that no spaces is introduced
	return fmt.Sprintf("%s(%s)", l.Name, strings.Join(args, ","))
}

func (l *LogEventDef) String() string {
	var args []string
	for i, parameter := range l.Parameters {
		indexed := ""
		if parameter.Indexed {
			indexed = " indexed"
		}

		args = append(args, fmt.Sprintf("%s%s %s", parameter.TypeName, indexed, parameter.GetName(i)))
	}

	return fmt.Sprintf("%s(%s)", l.Name, strings.Join(args, ", "))
}
