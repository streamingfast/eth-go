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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func ParseABI(abiFilePath string) (*ABI, error) {
	file, err := os.Open(abiFilePath)
	if err != nil {
		return nil, fmt.Errorf("open abi file: %w", err)
	}
	defer file.Close()

	return parseABIFromReader(file)
}

func ParseABIFromBytes(content []byte) (*ABI, error) {
	return parseABIFromReader(bytes.NewBuffer(content))
}

func parseDeclarations(content []byte) ([]*declaration, error) {
	// Find first non-whitespace character to determine format
	content = bytes.TrimSpace(content)
	if len(content) == 0 {
		return nil, fmt.Errorf("empty abi content")
	}

	switch content[0] {
	case '[':
		// Direct array format: [...]
		var declarations []*declaration
		if err := json.Unmarshal(content, &declarations); err != nil {
			return nil, fmt.Errorf("read abi array: %w", err)
		}
		return declarations, nil

	case '{':
		// Object format: { "abi": [...], ... }
		var wrapper struct {
			ABI []*declaration `json:"abi"`
		}
		if err := json.Unmarshal(content, &wrapper); err != nil {
			return nil, fmt.Errorf("read abi object: %w", err)
		}
		if wrapper.ABI == nil {
			return nil, fmt.Errorf("abi object missing 'abi' key")
		}
		return wrapper.ABI, nil

	default:
		return nil, fmt.Errorf("invalid abi format: expected '[' or '{', got %q", content[0])
	}
}

func parseABIFromReader(reader io.Reader) (*ABI, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read content: %w", err)
	}

	declarations, err := parseDeclarations(content)
	if err != nil {
		return nil, err
	}

	abi := &ABI{
		LogEventsMap:    map[string][]*LogEventDef{},
		FunctionsMap:    map[string][]*MethodDef{},
		ConstructorsMap: map[string][]*ConstructorDef{},

		LogEventsByNameMap:    map[string][]*LogEventDef{},
		FunctionsByNameMap:    map[string][]*MethodDef{},
		ConstructorsByNameMap: map[string][]*ConstructorDef{},
	}

	for _, decl := range declarations {
		if decl.Type == DeclarationTypeFunction {
			methodDef, err := decl.toFunctionDef()
			if err != nil {
				return nil, fmt.Errorf("invalid method %q: %w", decl.Name, err)
			}

			methodID := string(methodDef.MethodID())

			abi.FunctionsMap[methodID] = append(abi.FunctionsMap[methodID], methodDef)
			abi.FunctionsByNameMap[methodDef.Name] = append(abi.FunctionsByNameMap[methodDef.Name], methodDef)
		}

		if decl.Type == DeclarationTypeConstructor {
			constructorDef, err := decl.toConstructorDef()
			if err != nil {
				return nil, fmt.Errorf("invalid constructor: %w", err)
			}

			// Constructors are indexed by their signature (parameter types)
			signature := constructorDef.Signature()
			abi.ConstructorsMap[signature] = append(abi.ConstructorsMap[signature], constructorDef)
			// Also store under empty name for easy lookup (most contracts have one constructor)
			abi.ConstructorsByNameMap[""] = append(abi.ConstructorsByNameMap[""], constructorDef)
		}

		if decl.Type == DeclarationTypeEvent {
			logEventDef, err := decl.toLogEventDef()
			if err != nil {
				return nil, fmt.Errorf("invalid log event %q: %w", decl.Name, err)
			}

			logID := string(logEventDef.LogID())

			abi.LogEventsMap[logID] = append(abi.LogEventsMap[logID], logEventDef)
			abi.LogEventsByNameMap[logEventDef.Name] = append(abi.LogEventsByNameMap[logEventDef.Name], logEventDef)
		}
	}

	return abi, nil
}

//go:generate go-enum -f=$GOFILE --lower --marshal --names

// ENUM(
//
//	Function
//	Constructor
//	Receive
//	Fallback
//	Event
//	Error
//
// )
type DeclarationType int

// declaration is a generic struct output for each ABI element of an Ethereum contact
// compiled through solidity. It's a fairly generic structure encompassing multiple
// elements like function, events, constructors and others.
//
// See https://docs.soliditylang.org/en/v0.8.11/abi-spec.html#json
type declaration struct {
	// Common to functions and events
	Name   string          `json:"name,omitempty"`
	Type   DeclarationType `json:"type"`
	Inputs []*typeInfo     `json:"inputs"`

	// Functions only
	Outputs         []*typeInfo     `json:"outputs,omitempty"`
	StateMutability StateMutability `json:"stateMutability,omitempty"`
	// Functions only but removed in `solc` >= 0.5, now in `StateMutability` directly
	Payable  bool `json:"payable,omitempty"`
	Constant bool `json:"constant,omitempty"`

	// Events only
	Anonymous bool `json:"anonymous,omitempty"`
}

func (d *declaration) toFunctionDef() (*MethodDef, error) {
	out := &MethodDef{}
	out.Name = d.Name
	out.StateMutability = d.StateMutability

	// Those were removed in `solc` >= 0.5, there are most probably exclusive to each other
	if d.Payable {
		out.StateMutability = StateMutabilityNonPayable
	}
	if d.Constant {
		out.StateMutability = StateMutabilityPure
	}

	if len(d.Inputs) > 0 {
		out.Parameters = make([]*MethodParameter, len(d.Inputs))
		for i, input := range d.Inputs {
			parsedType, err := ParseType(input.Type)
			if err != nil {
				return nil, fmt.Errorf("invalid input parameter %q type: %w", input.Name, err)
			}

			structComponents, err := toStructComponents(input.Components)
			if err != nil {
				return nil, fmt.Errorf("invalid input %q struct component: %w", input.Name, err)
			}

			out.Parameters[i] = &MethodParameter{
				Name:         input.Name,
				TypeName:     input.Type,
				Type:         parsedType,
				InternalType: input.InternalType,
				Components:   structComponents,
			}
		}
	}

	if len(d.Outputs) > 0 {
		out.ReturnParameters = make([]*MethodParameter, len(d.Outputs))
		for i, output := range d.Outputs {
			parsedType, err := ParseType(output.Type)
			if err != nil {
				return nil, fmt.Errorf("invalid output parameter %q type: %w", output.Name, err)
			}

			structComponents, err := toStructComponents(output.Components)
			if err != nil {
				return nil, fmt.Errorf("invalid output %q struct component: %w", output.Name, err)
			}

			out.ReturnParameters[i] = &MethodParameter{
				Name:         output.Name,
				TypeName:     output.Type,
				Type:         parsedType,
				InternalType: output.InternalType,
				Components:   structComponents,
			}
		}
	}

	return out, nil
}

func (d *declaration) toConstructorDef() (*ConstructorDef, error) {
	out := &ConstructorDef{}
	out.StateMutability = d.StateMutability

	// Handle old solc versions
	if d.Payable {
		out.StateMutability = StateMutabilityPayable
	}

	if len(d.Inputs) > 0 {
		out.Parameters = make([]*MethodParameter, len(d.Inputs))
		for i, input := range d.Inputs {
			parsedType, err := ParseType(input.Type)
			if err != nil {
				return nil, fmt.Errorf("invalid input parameter %q type: %w", input.Name, err)
			}

			structComponents, err := toStructComponents(input.Components)
			if err != nil {
				return nil, fmt.Errorf("invalid input %q struct component: %w", input.Name, err)
			}

			out.Parameters[i] = &MethodParameter{
				Name:         input.Name,
				TypeName:     input.Type,
				Type:         parsedType,
				InternalType: input.InternalType,
				Components:   structComponents,
			}
		}
	}

	return out, nil
}

func toStructComponents(in []*structComponent) (out []*StructComponent, err error) {
	if len(in) == 0 {
		return
	}

	out = make([]*StructComponent, len(in))
	for i, component := range in {
		parsedType, err := ParseType(component.Type)
		if err != nil {
			return nil, fmt.Errorf("invalid component %q type: %w", component.Name, err)
		}

		// Recursively process nested components for tuple types
		nestedComponents, err := toStructComponents(component.Components)
		if err != nil {
			return nil, fmt.Errorf("invalid nested component in %q: %w", component.Name, err)
		}

		out[i] = &StructComponent{
			InternalType: component.InternalType,
			Name:         component.Name,
			TypeName:     component.Type,
			Type:         parsedType,
			Components:   nestedComponents,
		}
	}
	return out, nil
}

func (d *declaration) toLogEventDef() (*LogEventDef, error) {
	out := &LogEventDef{}
	out.Name = d.Name

	out.Parameters = make([]*LogParameter, len(d.Inputs))
	for i, input := range d.Inputs {
		parsedType, err := ParseType(input.Type)
		if err != nil {
			return nil, fmt.Errorf("invalid parameter %q type: %w", input.Name, err)
		}

		out.Parameters[i] = &LogParameter{
			Name:     input.Name,
			TypeName: input.Type,
			Type:     parsedType,
			Indexed:  input.Indexed,
		}
	}

	return out, nil
}

type typeInfo struct {
	InternalType string             `json:"internalType"`
	Name         string             `json:"name"`
	Type         string             `json:"type"`
	Indexed      bool               `json:"indexed"`
	Components   []*structComponent `json:"components,omitempty"`
}

type structComponent struct {
	InternalType string             `json:"internalType"`
	Name         string             `json:"name"`
	Type         string             `json:"type"`
	Components   []*structComponent `json:"components"`
}

func (c *structComponent) String() string {
	return fmt.Sprintf("%s %s (%s)", c.Type, c.Name, c.InternalType)
}
