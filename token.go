package eth

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"strings"
)

var ETHToken = &Token{
	Name:     "Ethereum",
	Symbol:   "ETH",
	Address:  nil, // not sure if this works
	Decimals: 18,
}

type Token struct {
	Name     string  `json:"name"`
	Symbol   string  `json:"symbol"`
	Address  Address `json:"address"`
	Decimals uint    `json:"decimals"`
}

func (t *Token) ID() uint64 {
	return binary.LittleEndian.Uint64(t.Address)
}

func (t *Token) String() string {
	return fmt.Sprintf("%s ([%d,%s] @ %x)", t.Name, t.Decimals, t.Symbol, []byte(t.Address))
}

func (t *Token) AmountBig(value *big.Int) TokenAmount {
	return TokenAmount{Amount: value, Token: t}
}

func (t *Token) Amount(value int64) TokenAmount {
	if t.Decimals == 0 {
		return TokenAmount{Amount: big.NewInt(value), Token: t}
	}

	valueBig := big.NewInt(value)
	return TokenAmount{Amount: valueBig.Mul(valueBig, t.decimalsBig()), Token: t}
}

func (t *Token) decimalsBig() *big.Int {
	if t.Decimals <= uint(len(decimalsBigInt)) {
		return decimalsBigInt[t.Decimals-1]
	}

	return new(big.Int).Exp(_10b, big.NewInt(int64(t.Decimals)), nil)
}

type TokenAmount struct {
	Amount *big.Int
	Token  *Token
}

func (t TokenAmount) Bytes() []byte {
	return t.Amount.Bytes()
}

func (t TokenAmount) Format(truncateDecimalCount uint) string {
	v := prettifyBigIntWithDecimals(t.Amount, t.Token.Decimals, truncateDecimalCount)
	return fmt.Sprintf("%s %s", v, t.Token.Symbol)
}

func (t TokenAmount) String() string {
	return t.Format(4)
}

func prettifyBigIntWithDecimals(in *big.Int, precision, truncateDecimalCount uint) string {
	bigDecimals := DecimalsInBigInt(uint32(precision))
	whole := new(big.Int).Div(in, bigDecimals)

	reminder := new(big.Int).Rem(in, bigDecimals).String()
	missingLeadingZeros := int(precision) - len(reminder)
	fractional := strings.Repeat("0", missingLeadingZeros) + reminder
	if len(fractional) > int(truncateDecimalCount) {
		fractional = fractional[0:truncateDecimalCount]
	}

	return fmt.Sprintf("%s.%s", whole, fractional)
}
