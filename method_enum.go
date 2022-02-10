// Code generated by go-enum DO NOT EDIT.
// Version:
// Revision:
// Build Date:
// Built By:

package eth

import (
	"fmt"
	"strings"
)

const (
	// StateMutabilityPure is a StateMutability of type Pure.
	StateMutabilityPure StateMutability = iota
	// StateMutabilityView is a StateMutability of type View.
	StateMutabilityView
	// StateMutabilityNonPayable is a StateMutability of type NonPayable.
	StateMutabilityNonPayable
	// StateMutabilityPayable is a StateMutability of type Payable.
	StateMutabilityPayable
)

const _StateMutabilityName = "PureViewNonPayablePayable"

var _StateMutabilityNames = []string{
	_StateMutabilityName[0:4],
	_StateMutabilityName[4:8],
	_StateMutabilityName[8:18],
	_StateMutabilityName[18:25],
}

// StateMutabilityNames returns a list of possible string values of StateMutability.
func StateMutabilityNames() []string {
	tmp := make([]string, len(_StateMutabilityNames))
	copy(tmp, _StateMutabilityNames)
	return tmp
}

var _StateMutabilityMap = map[StateMutability]string{
	StateMutabilityPure:       _StateMutabilityName[0:4],
	StateMutabilityView:       _StateMutabilityName[4:8],
	StateMutabilityNonPayable: _StateMutabilityName[8:18],
	StateMutabilityPayable:    _StateMutabilityName[18:25],
}

// String implements the Stringer interface.
func (x StateMutability) String() string {
	if str, ok := _StateMutabilityMap[x]; ok {
		return str
	}
	return fmt.Sprintf("StateMutability(%d)", x)
}

var _StateMutabilityValue = map[string]StateMutability{
	_StateMutabilityName[0:4]:                    StateMutabilityPure,
	strings.ToLower(_StateMutabilityName[0:4]):   StateMutabilityPure,
	_StateMutabilityName[4:8]:                    StateMutabilityView,
	strings.ToLower(_StateMutabilityName[4:8]):   StateMutabilityView,
	_StateMutabilityName[8:18]:                   StateMutabilityNonPayable,
	strings.ToLower(_StateMutabilityName[8:18]):  StateMutabilityNonPayable,
	_StateMutabilityName[18:25]:                  StateMutabilityPayable,
	strings.ToLower(_StateMutabilityName[18:25]): StateMutabilityPayable,
}

// ParseStateMutability attempts to convert a string to a StateMutability
func ParseStateMutability(name string) (StateMutability, error) {
	if x, ok := _StateMutabilityValue[name]; ok {
		return x, nil
	}
	return StateMutability(0), fmt.Errorf("%s is not a valid StateMutability, try [%s]", name, strings.Join(_StateMutabilityNames, ", "))
}

// MarshalText implements the text marshaller method
func (x StateMutability) MarshalText() ([]byte, error) {
	return []byte(x.String()), nil
}

// UnmarshalText implements the text unmarshaller method
func (x *StateMutability) UnmarshalText(text []byte) error {
	name := string(text)
	tmp, err := ParseStateMutability(name)
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}
