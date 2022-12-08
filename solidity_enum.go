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
	// TypeBoolean is a TypeKind of type Boolean.
	TypeBoolean TypeKind = iota
	// TypeAddress is a TypeKind of type Address.
	TypeAddress
	// TypeSignedInteger is a TypeKind of type SignedInteger.
	TypeSignedInteger
	// TypeUnsignedInteger is a TypeKind of type UnsignedInteger.
	TypeUnsignedInteger
	// TypeSignedFixedPoint is a TypeKind of type SignedFixedPoint.
	TypeSignedFixedPoint
	// TypeUsignedFixedPoint is a TypeKind of type UsignedFixedPoint.
	TypeUsignedFixedPoint
	// TypeFixedSizeBytes is a TypeKind of type FixedSizeBytes.
	TypeFixedSizeBytes
	// TypeBytes is a TypeKind of type Bytes.
	TypeBytes
	// TypeString is a TypeKind of type String.
	TypeString
	// TypeFixedSizeArray is a TypeKind of type FixedSizeArray.
	TypeFixedSizeArray
	// TypeArray is a TypeKind of type Array.
	TypeArray
	// TypeStruct is a TypeKind of type Struct.
	TypeStruct
	// TypeMappings is a TypeKind of type Mappings.
	TypeMappings
)

const _TypeKindName = "BooleanAddressSignedIntegerUnsignedIntegerSignedFixedPointUsignedFixedPointFixedSizeBytesBytesStringFixedSizeArrayArrayStructMappings"

var _TypeKindNames = []string{
	_TypeKindName[0:7],
	_TypeKindName[7:14],
	_TypeKindName[14:27],
	_TypeKindName[27:42],
	_TypeKindName[42:58],
	_TypeKindName[58:75],
	_TypeKindName[75:89],
	_TypeKindName[89:94],
	_TypeKindName[94:100],
	_TypeKindName[100:114],
	_TypeKindName[114:119],
	_TypeKindName[119:125],
	_TypeKindName[125:133],
}

// TypeKindNames returns a list of possible string values of TypeKind.
func TypeKindNames() []string {
	tmp := make([]string, len(_TypeKindNames))
	copy(tmp, _TypeKindNames)
	return tmp
}

var _TypeKindMap = map[TypeKind]string{
	TypeBoolean:           _TypeKindName[0:7],
	TypeAddress:           _TypeKindName[7:14],
	TypeSignedInteger:     _TypeKindName[14:27],
	TypeUnsignedInteger:   _TypeKindName[27:42],
	TypeSignedFixedPoint:  _TypeKindName[42:58],
	TypeUsignedFixedPoint: _TypeKindName[58:75],
	TypeFixedSizeBytes:    _TypeKindName[75:89],
	TypeBytes:             _TypeKindName[89:94],
	TypeString:            _TypeKindName[94:100],
	TypeFixedSizeArray:    _TypeKindName[100:114],
	TypeArray:             _TypeKindName[114:119],
	TypeStruct:            _TypeKindName[119:125],
	TypeMappings:          _TypeKindName[125:133],
}

// String implements the Stringer interface.
func (x TypeKind) String() string {
	if str, ok := _TypeKindMap[x]; ok {
		return str
	}
	return fmt.Sprintf("TypeKind(%d)", x)
}

var _TypeKindValue = map[string]TypeKind{
	_TypeKindName[0:7]:     TypeBoolean,
	_TypeKindName[7:14]:    TypeAddress,
	_TypeKindName[14:27]:   TypeSignedInteger,
	_TypeKindName[27:42]:   TypeUnsignedInteger,
	_TypeKindName[42:58]:   TypeSignedFixedPoint,
	_TypeKindName[58:75]:   TypeUsignedFixedPoint,
	_TypeKindName[75:89]:   TypeFixedSizeBytes,
	_TypeKindName[89:94]:   TypeBytes,
	_TypeKindName[94:100]:  TypeString,
	_TypeKindName[100:114]: TypeFixedSizeArray,
	_TypeKindName[114:119]: TypeArray,
	_TypeKindName[119:125]: TypeStruct,
	_TypeKindName[125:133]: TypeMappings,
}

// ParseTypeKind attempts to convert a string to a TypeKind
func ParseTypeKind(name string) (TypeKind, error) {
	if x, ok := _TypeKindValue[name]; ok {
		return x, nil
	}
	return TypeKind(0), fmt.Errorf("%s is not a valid TypeKind, try [%s]", name, strings.Join(_TypeKindNames, ", "))
}
