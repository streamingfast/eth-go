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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogEventDef_Signature(t *testing.T) {
	type fields struct {
		Name       string
		Parameters []*LogParameter
	}
	tests := []struct {
		want   string
		fields fields
	}{
		{
			"EventAddressIdxString(address,string)",
			fields{Name: "EventAddressIdxString", Parameters: []*LogParameter{
				{Name: "any", TypeName: "address", Indexed: true},
				{Name: "any", TypeName: "string", Indexed: false},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			l := &LogEventDef{
				Name:       tt.fields.Name,
				Parameters: tt.fields.Parameters,
			}
			assert.Equal(t, tt.want, l.Signature())
		})
	}
}

func TestLogEventDef_Signature_NestedTuples(t *testing.T) {
	tests := []struct {
		name        string
		event       *LogEventDef
		expectedSig string
	}{
		{
			name: "single-level tuple",
			event: &LogEventDef{
				Name: "ItemCreated",
				Parameters: []*LogParameter{
					{
						Name:     "item",
						TypeName: "tuple",
						Components: []*StructComponent{
							{Name: "id", TypeName: "uint256"},
							{Name: "owner", TypeName: "address"},
						},
					},
				},
			},
			expectedSig: "ItemCreated((uint256,address))",
		},
		{
			name: "nested tuple",
			event: &LogEventDef{
				Name: "ComplexEvent",
				Parameters: []*LogParameter{
					{
						Name:     "data",
						TypeName: "tuple",
						Components: []*StructComponent{
							{
								Name:     "inner",
								TypeName: "tuple",
								Components: []*StructComponent{
									{Name: "value", TypeName: "uint256"},
									{Name: "name", TypeName: "string"},
								},
							},
							{Name: "flag", TypeName: "bool"},
						},
					},
				},
			},
			expectedSig: "ComplexEvent(((uint256,string),bool))",
		},
		{
			name: "deeply nested tuple",
			event: &LogEventDef{
				Name: "DeepEvent",
				Parameters: []*LogParameter{
					{
						Name:     "data",
						TypeName: "tuple",
						Components: []*StructComponent{
							{
								Name:     "level1",
								TypeName: "tuple",
								Components: []*StructComponent{
									{
										Name:     "level2",
										TypeName: "tuple",
										Components: []*StructComponent{
											{Name: "value", TypeName: "uint256"},
										},
									},
									{Name: "addr", TypeName: "address"},
								},
							},
							{Name: "data", TypeName: "bytes"},
						},
					},
				},
			},
			expectedSig: "DeepEvent((((uint256),address),bytes))",
		},
		{
			name: "array of tuples",
			event: &LogEventDef{
				Name: "BatchEvent",
				Parameters: []*LogParameter{
					{
						Name:     "items",
						TypeName: "tuple[]",
						Components: []*StructComponent{
							{Name: "id", TypeName: "uint256"},
							{Name: "owner", TypeName: "address"},
						},
					},
				},
			},
			expectedSig: "BatchEvent((uint256,address)[])",
		},
		{
			name: "array of nested tuples",
			event: &LogEventDef{
				Name: "BatchNestedEvent",
				Parameters: []*LogParameter{
					{
						Name:     "items",
						TypeName: "tuple[]",
						Components: []*StructComponent{
							{
								Name:     "inner",
								TypeName: "tuple",
								Components: []*StructComponent{
									{Name: "value", TypeName: "uint256"},
									{Name: "name", TypeName: "string"},
								},
							},
							{Name: "flag", TypeName: "bool"},
						},
					},
				},
			},
			expectedSig: "BatchNestedEvent(((uint256,string),bool)[])",
		},
		{
			name: "mixed parameters with nested tuple",
			event: &LogEventDef{
				Name: "MixedEvent",
				Parameters: []*LogParameter{
					{Name: "sender", TypeName: "address", Indexed: true},
					{
						Name:     "data",
						TypeName: "tuple",
						Components: []*StructComponent{
							{
								Name:     "nested",
								TypeName: "tuple",
								Components: []*StructComponent{
									{Name: "a", TypeName: "address"},
									{Name: "b", TypeName: "bytes32"},
								},
							},
							{Name: "value", TypeName: "bytes"},
						},
					},
					{Name: "amount", TypeName: "uint256"},
				},
			},
			expectedSig: "MixedEvent(address,((address,bytes32),bytes),uint256)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig := tt.event.Signature()
			assert.Equal(t, tt.expectedSig, sig, "signature mismatch")
		})
	}
}
