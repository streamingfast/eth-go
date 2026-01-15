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

import "go.uber.org/zap"

// ABI is our custom internal definition of a contract's ABI that bridges the information
// between the two ABI like formats for contract's, i.e. `.abi` file and AST file as output
// by `solc` compiler.
type ABI struct {
	LogEventsMap    map[string][]*LogEventDef
	FunctionsMap    map[string][]*MethodDef
	ConstructorsMap map[string][]*MethodDef

	LogEventsByNameMap    map[string][]*LogEventDef
	FunctionsByNameMap    map[string][]*MethodDef
	ConstructorsByNameMap map[string][]*MethodDef
}

// Deprecated Use FindLogByTopic
func (a *ABI) FindLog(topic []byte) *LogEventDef {
	return a.FindLogByTopic(topic)
}

// FindLogByTopic finds the first log with the give `topic`. Multiple events could have the
// same topic, use `FindLogByTopic` to get them all.
func (a *ABI) FindLogByTopic(topic []byte) *LogEventDef {
	zlog.Info("looking for first log by topic", zap.Stringer("topic", Hash(topic)))

	return firstOr(a.LogEventsMap[string(topic)], nil)
}

// FindLogByTopic returns **all**  logs with the give `topic`.
func (a *ABI) FindLogsByTopic(topic []byte) []*LogEventDef {
	zlog.Info("looking for all logs by topic", zap.Stringer("topic", Hash(topic)))

	return a.LogEventsMap[string(topic)]
}

// FindLogByName finds the first log with the give `topic`. Multiple events could have the
// same topic, use `FindLogsByName` to get them all.
func (a *ABI) FindLogByName(name string) *LogEventDef {
	zlog.Info("looking for first log by name", zap.String("event_name", name))

	return firstOr(a.LogEventsByNameMap[name], nil)
}

// FindLogsByName returns **all** logs with the give `topic`.
func (a *ABI) FindLogsByName(name string) []*LogEventDef {
	zlog.Info("looking for all logs by name", zap.String("event_name", name))

	return a.LogEventsByNameMap[name]
}

// Deprecated: Use FindFunctionByHash
func (a *ABI) FindFunction(functionHash []byte) *MethodDef {
	return a.FindFunctionByHash(functionHash)
}

// FindFunctionByHash finds the function with the give `hash`.
func (a *ABI) FindFunctionByHash(functionHash []byte) *MethodDef {
	zlog.Info("looking for first function by hash", zap.Stringer("method_hash", Hash(functionHash)))

	return firstOr(a.FunctionsMap[string(functionHash)], nil)
}

// FindFunctionByName finds the first function with the give `name`. Multiple functions could have the
// same name, use `FindFunctionsByName` to get them all.
func (a *ABI) FindFunctionByName(name string) *MethodDef {
	zlog.Info("looking for function by name", zap.Stringer("method_name", Hash(name)))

	return firstOr(a.FunctionsByNameMap[name], nil)
}

// FindFunctionsByName returns **all** functions with the give `name`.
func (a *ABI) FindFunctionsByName(name string) []*MethodDef {
	zlog.Info("looking for function by name", zap.Stringer("method_name", Hash(name)))

	return a.FunctionsByNameMap[name]
}

func firstOr[T any](elements []T, defaultValue T) T {
	if len(elements) == 0 {
		return defaultValue
	}

	return elements[0]
}
