package eth

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

//go:generate go-enum -f=$GOFILE --noprefix --prefix Type --names

// ENUM(
//
//		Boolean
//		Address
//		SignedInteger
//		UnsignedInteger
//		SignedFixedPoint
//		UsignedFixedPoint
//		FixedSizeBytes
//		Bytes
//		String
//	 FixedSizeArray
//	 Array
//	Struct
//
// Mappings
// )
type TypeKind int

type SolidityType interface {
	Name() string
	Kind() TypeKind
	IsDynamic() bool
}

var arrayRegex = regexp.MustCompile(`^(.+)\[]$`)
var integerRegex = regexp.MustCompile(`^(u)?int([0-9]+)$`)
var fixedBytesRegex = regexp.MustCompile(`^bytes([0-9]+)$`)
var fixedPointRegex = regexp.MustCompile(`^(u)?fixed([0-9]+)x([0-9]+)$`)
var fixedArrayRegex = regexp.MustCompile(`^(.+)\[([0-9]+)\]$`)

func ParseType(raw string) (SolidityType, error) {
	in := strings.ToLower(raw)

	switch in {
	case "bool":
		return BooleanType{}, nil

	case "address":
		return AddressType{}, nil

	case "int":
		return SignedIntegerType{BitsSize: 256, ByteSize: 32}, nil

	case "uint":
		return UnsignedIntegerType{BitsSize: 256, ByteSize: 32}, nil

	case "bytes":
		return BytesType{}, nil

	case "string":
		return StringType{}, nil

	case "fixed":
		return SignedFixedPointType{BitsSize: 128, ByteSize: 16, Decimals: 18}, nil

	case "ufixed":
		return UnsignedFixedPointType{BitsSize: 128, ByteSize: 16, Decimals: 18}, nil

	case "tuple":
		return StructType{}, nil
	}

	if matches := integerRegex.FindAllStringSubmatch(in, 1); len(matches) == 1 {
		return parseIntegerType(raw, matches[0])
	}

	if matches := fixedBytesRegex.FindAllStringSubmatch(in, 1); len(matches) == 1 {
		return parseFixedBytes(raw, matches[0])
	}

	if matches := fixedPointRegex.FindAllStringSubmatch(in, 1); len(matches) == 1 {
		return parseFixedPointType(raw, matches[0])
	}

	if matches := fixedArrayRegex.FindAllStringSubmatch(in, 1); len(matches) == 1 {
		return parseFixedArray(raw, matches[0])
	}

	if matches := arrayRegex.FindAllStringSubmatch(in, 1); len(matches) == 1 {
		return parseArray(raw, matches[0])
	}

	return nil, fmt.Errorf("type %q unknown", raw)
}

func parseFixedBytes(raw string, matchGroups []string) (SolidityType, error) {
	byteSizeRaw := matchGroups[1]
	byteSizeRawU64, err := strconv.ParseUint(byteSizeRaw, 10, strconv.IntSize)
	if err != nil {
		panic(fmt.Errorf("impossible (regex validated) invalid fixed bytes type %q: byte size not a number: %w", raw, err))
	}

	byteSize := uint(byteSizeRawU64)

	if byteSize < 1 {
		return nil, fmt.Errorf("invalid fixed bytes type %q: bits size %d is lower than 1", raw, byteSize)
	}

	if byteSize > 32 {
		return nil, fmt.Errorf("invalid fixed bytes type %q: bits size %d is bigger than 32", raw, byteSize)
	}

	return FixedSizeBytesType{ByteSize: byteSize}, nil
}

func parseIntegerType(raw string, matchGroups []string) (SolidityType, error) {
	bitsSizeRaw := matchGroups[2]
	bitsSizeU64, err := strconv.ParseUint(bitsSizeRaw, 10, strconv.IntSize)
	if err != nil {
		panic(fmt.Errorf("impossible (regex validated) invalid integer type %q: bits size not a number: %w", raw, err))
	}

	bitsSize := uint(bitsSizeU64)

	if bitsSize < 8 {
		return nil, fmt.Errorf("invalid integer type %q: bits size %d is lower than 8", raw, bitsSize)
	}

	if bitsSize > 256 {
		return nil, fmt.Errorf("invalid integer type %q: bits size %d is bigger than 256", raw, bitsSize)
	}

	if bitsSize%8 != 0 {
		return nil, fmt.Errorf("invalid integer type %q: bits size %d must be divisble by 8", raw, bitsSize)
	}

	byteSize := bitsSize / 8

	if matchGroups[1] == "u" {
		return UnsignedIntegerType{BitsSize: bitsSize, ByteSize: byteSize}, nil
	}

	return SignedIntegerType{BitsSize: bitsSize, ByteSize: byteSize}, nil
}

func parseFixedPointType(raw string, matchGroups []string) (SolidityType, error) {
	// Bits Size

	bitsSizeRaw := matchGroups[2]
	bitsSizeU64, err := strconv.ParseUint(bitsSizeRaw, 10, strconv.IntSize)
	if err != nil {
		panic(fmt.Errorf("impossible (regex validated) invalid integer type %q: bits size not a number: %w", raw, err))
	}

	bitsSize := uint(bitsSizeU64)

	if bitsSize < 8 {
		return nil, fmt.Errorf("invalid fixed point type %q: bits size %d is lower than 8", raw, bitsSize)
	}

	if bitsSize > 256 {
		return nil, fmt.Errorf("invalid fixed point type %q: bits size %d is bigger than 256", raw, bitsSize)
	}

	if bitsSize%8 != 0 {
		return nil, fmt.Errorf("invalid fixed point type %q: bits size %d must be divisble by 8", raw, bitsSize)
	}

	byteSize := bitsSize / 8

	// Decimals

	decimalsRaw := matchGroups[3]
	decimalsU64, err := strconv.ParseUint(decimalsRaw, 10, strconv.IntSize)
	if err != nil {
		panic(fmt.Errorf("impossible (regex validated) invalid fixed point type %q: decimals not a number: %w", raw, err))
	}

	decimals := uint(decimalsU64)

	if decimals > 80 {
		return nil, fmt.Errorf("invalid fixed point type %q: decimals %d is bigger than 80", raw, decimals)
	}

	if matchGroups[1] == "u" {
		return UnsignedFixedPointType{BitsSize: bitsSize, ByteSize: byteSize, Decimals: decimals}, nil
	}

	return SignedFixedPointType{BitsSize: bitsSize, ByteSize: byteSize, Decimals: decimals}, nil
}

func parseArray(raw string, matchGroups []string) (SolidityType, error) {
	// TODO: Handle recursion limit

	elementTypeRaw := matchGroups[1]
	elementType, err := ParseType(elementTypeRaw)
	if err != nil {
		return nil, fmt.Errorf("invalid array type %q: element type invalid: %w", raw, err)
	}

	return ArrayType{ElementType: elementType}, nil
}

func parseFixedArray(raw string, matchGroups []string) (SolidityType, error) {
	// TODO: Handle recursion limit

	elementTypeRaw := matchGroups[1]
	elementType, err := ParseType(elementTypeRaw)
	if err != nil {
		return nil, fmt.Errorf("invalid fixed array type %q: element type invalid: %w", raw, err)
	}

	lenghtRaw := matchGroups[2]
	lenghtRawU64, err := strconv.ParseUint(lenghtRaw, 10, strconv.IntSize)
	if err != nil {
		panic(fmt.Errorf("impossible (regex validated) invalid fixed array type %q: length not a number: %w", raw, err))
	}

	return FixedSizeArrayType{ElementType: elementType, Length: uint(lenghtRawU64)}, nil
}

var _ SolidityType = BooleanType{}
var _ SolidityType = AddressType{}
var _ SolidityType = SignedIntegerType{}
var _ SolidityType = UnsignedIntegerType{}
var _ SolidityType = SignedFixedPointType{}
var _ SolidityType = UnsignedFixedPointType{}
var _ SolidityType = FixedSizeBytesType{}
var _ SolidityType = BytesType{}
var _ SolidityType = StringType{}
var _ SolidityType = FixedSizeArrayType{}
var _ SolidityType = ArrayType{}
var _ SolidityType = StructType{}

type BooleanType struct {
}

func (t BooleanType) Name() string {
	return t.Kind().String()
}

func (a BooleanType) Kind() TypeKind {
	return TypeBoolean
}

func (t BooleanType) IsDynamic() bool {
	return false
}

type AddressType struct {
}

func (t AddressType) Name() string {
	return t.Kind().String()
}

func (a AddressType) Kind() TypeKind {
	return TypeAddress
}

func (t AddressType) IsDynamic() bool {
	return false
}

type SignedIntegerType struct {
	// BitsSize is the number of bites taken by this type, BitsSize / 8 = ByteSize, range of values is between 8 and 256 (inclusive) by step of 8
	BitsSize uint
	// ByteSize is the number of bytes taken by this type, ByteSize * 8 = BitsSize, range of values is between 1 and 32 (inclusive)
	ByteSize uint
}

func (t SignedIntegerType) Name() string {
	return fmt.Sprintf("int%d", t.BitsSize)
}

func (t SignedIntegerType) Kind() TypeKind {
	return TypeSignedInteger
}

func (t SignedIntegerType) IsDynamic() bool {
	return false
}

type UnsignedIntegerType struct {
	// BitsSize is the number of bites taken by this type, BitsSize / 8 = ByteSize, range of values is between 8 and 256 (inclusive) by step of 8
	BitsSize uint
	// ByteSize is the number of bytes taken by this type, ByteSize * 8 = BitsSize, range of values is between 1 and 32 (inclusive)
	ByteSize uint
}

func (t UnsignedIntegerType) Name() string {
	return fmt.Sprintf("uint%d", t.BitsSize)
}

func (t UnsignedIntegerType) Kind() TypeKind {
	return TypeUnsignedInteger
}

func (t UnsignedIntegerType) IsDynamic() bool {
	return false
}

type SignedFixedPointType struct {
	// BitsSize is the number of bites taken by this type, BitsSize / 8 = ByteSize, range of values is between 8 and 256 (inclusive) by step of 8
	BitsSize uint
	// ByteSize is the number of bytes taken by this type, ByteSize * 8 = BitsSize, range of values is between 1 and 32 (inclusive)
	ByteSize uint
	// Decimals represents the precision in decimals of this fixed type and will be between 0 and 80 (inclusive)
	Decimals uint
}

func (t SignedFixedPointType) Name() string {
	return fmt.Sprintf("fixed%d", t.BitsSize)

}

func (t SignedFixedPointType) Kind() TypeKind {
	return TypeSignedFixedPoint
}

func (t SignedFixedPointType) IsDynamic() bool {
	return false
}

type UnsignedFixedPointType struct {
	// BitsSize is the number of bites taken by this type, BitsSize / 8 = ByteSize, range of values is between 8 and 256 (inclusive) by step of 8
	BitsSize uint
	// ByteSize is the number of bytes taken by this type, ByteSize * 8 = BitsSize, range of values is between 1 and 32 (inclusive)
	ByteSize uint
	// Decimals represents the precision in decimals of this fixed type and will be between 0 and 80 (inclusive)
	Decimals uint
}

func (t UnsignedFixedPointType) Name() string {
	return t.Kind().String()
}

func (t UnsignedFixedPointType) Kind() TypeKind {
	return TypeUsignedFixedPoint
}

func (t UnsignedFixedPointType) IsDynamic() bool {
	return false
}

type FixedSizeBytesType struct {
	// ByteSize is the number of bytes taken by this type, range of values is between 1 and 32 (inclusive)
	ByteSize uint
}

func (t FixedSizeBytesType) Name() string {
	return t.Kind().String()
}

func (t FixedSizeBytesType) Kind() TypeKind {
	return TypeFixedSizeBytes
}

func (t FixedSizeBytesType) IsDynamic() bool {
	return false
}

type BytesType struct {
}

func (t BytesType) Name() string {
	return t.Kind().String()
}

func (t BytesType) Kind() TypeKind {
	return TypeBytes
}

func (t BytesType) IsDynamic() bool {
	return true
}

type StringType struct {
}

func (t StringType) Name() string {
	return t.Kind().String()
}

func (t StringType) Kind() TypeKind {
	return TypeString
}

func (t StringType) IsDynamic() bool {
	return true
}

type FixedSizeArrayType struct {
	// ElementType represents the underlying stored element
	ElementType SolidityType
	// Length is the number of fixed element this array contains
	Length uint
}

func (t FixedSizeArrayType) Name() string {
	return fmt.Sprintf("%s[%d]", t.ElementType, t.Length)
}

func (t FixedSizeArrayType) Kind() TypeKind {
	return TypeFixedSizeArray
}

func (t FixedSizeArrayType) IsDynamic() bool {
	return t.ElementType.IsDynamic()
}

type ArrayType struct {
	// ElementType represents the underlying stored element
	ElementType SolidityType
}

func (t ArrayType) Name() string {
	return fmt.Sprintf("%s[]", t.ElementType)
}

func (t ArrayType) Kind() TypeKind {
	return TypeArray
}

func (t ArrayType) IsDynamic() bool {
	return true
}

type StructType struct {
}

func (t StructType) Name() string {
	return t.Kind().String()
}

func (t StructType) Kind() TypeKind {
	return TypeStruct
}

func (t StructType) IsDynamic() bool {
	return false
}

type StructComponent struct {
	InternalType string
	Name         string
	TypeName     string
	Type         SolidityType
	// Components holds nested struct components when TypeName is "tuple" or "tuple[]"
	Components []*StructComponent
}

func (c *StructComponent) String() string {
	return fmt.Sprintf("%s %s (%s)", c.TypeName, c.Name, c.InternalType)
}

// Signature returns the canonical type signature for this struct component,
// recursively expanding nested tuples. For example, a tuple containing
// (address, (uint256, string), uint64) returns "(address,(uint256,string),uint64)".
func (c *StructComponent) Signature() string {
	typeName := c.TypeName
	if strings.HasPrefix(typeName, "tuple") {
		componentSigs := make([]string, len(c.Components))
		for i, component := range c.Components {
			componentSigs[i] = component.Signature()
		}

		// Replace "tuple" with "(sig1,sig2,...)" preserving any array suffix like [] or [N]
		typeName = strings.Replace(typeName, "tuple", "("+strings.Join(componentSigs, ",")+")", 1)
	}

	return typeName
}
