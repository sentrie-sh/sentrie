// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package box

import (
	"encoding/json"
	"math"
)

const (
	twoTo53     = int64(1) << 53
	twoTo53Plus = twoTo53 + 1 // 9007199254740993, the first integer float64 cannot hold
)

func (s *BoxTestSuite) TestNumberConstructorRepresentation() {
	s.True(Number(int64(twoTo53Plus)).IsInt())
	s.True(Number(7).IsInt())
	s.True(Number(uint8(7)).IsInt())
	s.True(Number(uint64(math.MaxInt64)).IsInt())
	s.False(Number(uint64(math.MaxUint64)).IsInt(), "unsigned beyond int64 falls back to float")
	s.False(Number(2.0).IsInt(), "a float stays a float even when integral")
	s.False(Number(float32(1.5)).IsInt())
	s.Equal(ValueNumber, Number(7).Kind())
	s.Equal(ValueNumber, Number(7.5).Kind())

	i, ok := Number(int64(twoTo53Plus)).IntValue()
	s.True(ok)
	s.Equal(twoTo53Plus, i)
	_, ok = Number(2.0).IntValue()
	s.False(ok)
	_, ok = String("1").IntValue()
	s.False(ok)
}

func (s *BoxTestSuite) TestIntegralValue() {
	i, ok := Float(4.0).IntegralValue()
	s.True(ok)
	s.Equal(int64(4), i)
	_, ok = Float(4.5).IntegralValue()
	s.False(ok)
	_, ok = Float(math.Inf(1)).IntegralValue()
	s.False(ok)
	_, ok = Float(two63).IntegralValue()
	s.False(ok, "2^63 does not fit int64")
	i, ok = Float(-two63).IntegralValue()
	s.True(ok)
	s.Equal(int64(math.MinInt64), i)
}

func (s *BoxTestSuite) TestExactIntegerRendering() {
	v := Int(twoTo53Plus)
	s.Equal("9007199254740993", v.String())
	b, err := json.Marshal(v)
	s.Require().NoError(err)
	s.Equal("9007199254740993", string(b))
	b, err = json.Marshal(List([]Value{Int(math.MaxInt64), Int(math.MinInt64)}))
	s.Require().NoError(err)
	s.Equal("[9223372036854775807,-9223372036854775808]", string(b))
	s.Equal(int64(twoTo53Plus), v.Any())
}

func (s *BoxTestSuite) TestFromAnyJSONNumber() {
	v := FromAny(json.Number("9007199254740993"))
	i, ok := v.IntValue()
	s.True(ok)
	s.Equal(twoTo53Plus, i)

	v = FromAny(json.Number("1.5"))
	s.False(v.IsInt())
	f, _ := v.NumberValue()
	s.Equal(1.5, f)

	v = FromAny(json.Number("1.0"))
	s.False(v.IsInt(), "a JSON number written with a fraction is a float")

	v = FromAny(json.Number("18446744073709551616"))
	s.False(v.IsInt(), "beyond int64 falls back to float")
	s.Equal(ValueNumber, v.Kind())

	v = FromAny(json.Number("1e400"))
	f, _ = v.NumberValue()
	s.True(math.IsInf(f, 1))

	v = FromBoundaryAny(map[string]any{"id": json.Number("9007199254740993")})
	m, _ := v.DictValue()
	i, ok = m["id"].IntValue()
	s.True(ok)
	s.Equal(twoTo53Plus, i)
}

func (s *BoxTestSuite) TestParseNumber() {
	v, err := ParseNumber("42")
	s.Require().NoError(err)
	s.True(v.IsInt())
	v, err = ParseNumber("4.5")
	s.Require().NoError(err)
	s.False(v.IsInt())
	_, err = ParseNumber("abc")
	s.Error(err)
	v, err = ParseNumber("1e400")
	s.Error(err, "a magnitude float64 cannot hold is an error, never an infinity")
	s.True(v.IsUndefined())
}

func (s *BoxTestSuite) TestCheckJSONNumbers() {
	s.NoError(CheckJSONNumbers(map[string]any{"a": json.Number("9007199254740993"), "b": []any{json.Number("1.5"), "x", nil}}))
	s.ErrorContains(CheckJSONNumbers(json.Number("1e400")), "number 1e400 is out of range")
	s.Error(CheckJSONNumbers(map[string]any{"a": []any{json.Number("-1e500")}}))
	s.NoError(CheckJSONNumbers(json.Number("1e308")))
}

func (s *BoxTestSuite) TestIntegerArithmeticIsExact() {
	v, err := Add(Int(twoTo53), Int(1))
	s.Require().NoError(err)
	s.Equal(Int(twoTo53Plus), v)

	v, err = Sub(Int(twoTo53Plus), Int(1))
	s.Require().NoError(err)
	s.Equal(Int(twoTo53), v)

	v, err = Mul(Int(3), Int(3002399751580331))
	s.Require().NoError(err)
	s.Equal(Int(9007199254740993), v)

	v, err = Div(Int(8), Int(2))
	s.Require().NoError(err)
	s.Equal(Int(4), v)

	v, err = Div(Int(7), Int(2))
	s.Require().NoError(err)
	s.Equal(Float(3.5), v, "inexact integer division gives a float, as before")

	v, err = Mod(Int(twoTo53Plus), Int(2))
	s.Require().NoError(err)
	s.Equal(Int(1), v)

	v, err = Mod(Int(-7), Int(3))
	s.Require().NoError(err)
	s.Equal(Int(-1), v, "remainder takes the sign of the dividend, like math.Mod")

	v, err = Mod(Int(math.MinInt64), Int(-1))
	s.Require().NoError(err)
	s.Equal(Int(0), v)
}

func (s *BoxTestSuite) TestMixedArithmeticPromotesToFloat() {
	v, err := Add(Int(1), Float(0.5))
	s.Require().NoError(err)
	s.Equal(Float(1.5), v)
	v, err = Mul(Float(2), Int(3))
	s.Require().NoError(err)
	s.Equal(Float(6), v)
	v, err = Mod(Float(7.5), Int(2))
	s.Require().NoError(err)
	s.Equal(Float(1.5), v)
}

func (s *BoxTestSuite) TestIntegerOverflowIsAnError() {
	cases := []struct {
		name string
		fn   func(Value, Value) (Value, error)
		l, r int64
	}{
		{"add", Add, math.MaxInt64, 1},
		{"add negative", Add, math.MinInt64, -1},
		{"sub", Sub, math.MinInt64, 1},
		{"sub negative", Sub, math.MaxInt64, -1},
		{"mul", Mul, math.MaxInt64, 2},
		{"mul min by minus one", Mul, math.MinInt64, -1},
		{"mul minus one by min", Mul, -1, math.MinInt64},
		{"mul large", Mul, 1 << 32, 1 << 31},
		{"div min by minus one", Div, math.MinInt64, -1},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			_, err := tc.fn(Int(tc.l), Int(tc.r))
			s.ErrorIs(err, ErrIntegerOverflow)
		})
	}

	v, err := Mul(Int(1<<31), Int(1<<31))
	s.Require().NoError(err, "2^62 fits")
	s.Equal(Int(1<<62), v)

	_, err = Negate(Int(math.MinInt64))
	s.ErrorIs(err, ErrIntegerOverflow)
	v, err = Negate(Int(math.MaxInt64))
	s.Require().NoError(err)
	s.Equal(Int(-math.MaxInt64), v)
	v, err = Negate(Float(1.5))
	s.Require().NoError(err)
	s.Equal(Float(-1.5), v)
}

func (s *BoxTestSuite) TestArithmeticOperandErrors() {
	_, err := Add(String("1"), Int(1))
	s.EqualError(err, "left operand is not a number")
	_, err = Sub(Int(1), Bool(true))
	s.EqualError(err, "right operand is not a number")
	_, err = Div(Int(1), Int(0))
	s.EqualError(err, "divide by zero")
	_, err = Div(Float(1), Int(0))
	s.EqualError(err, "divide by zero")
	_, err = Mod(Int(1), Int(0))
	s.EqualError(err, "divide by zero")
	_, err = Mod(Float(1), Float(0))
	s.EqualError(err, "divide by zero")
	_, err = Negate(String("x"))
	s.Error(err)
}

func (s *BoxTestSuite) TestCompareNumbersExact() {
	cmp := func(a, b Value) int {
		c, ordered, err := CompareNumbers(a, b)
		s.Require().NoError(err)
		s.Require().True(ordered)
		return c
	}
	s.Equal(1, cmp(Int(twoTo53Plus), Int(twoTo53)))
	s.Equal(1, cmp(Int(twoTo53Plus), Float(float64(twoTo53))), "int above a float that rounds to it")
	s.Equal(-1, cmp(Float(float64(twoTo53)), Int(twoTo53Plus)))
	s.Equal(0, cmp(Int(5), Float(5)))
	s.Equal(-1, cmp(Int(5), Float(5.5)))
	s.Equal(1, cmp(Int(-5), Float(-5.5)))
	s.Equal(-1, cmp(Int(math.MaxInt64), Float(two63)))
	s.Equal(1, cmp(Int(math.MinInt64), Float(math.Inf(-1))))
	s.Equal(-1, cmp(Int(math.MaxInt64), Float(math.Inf(1))))
	s.Equal(0, cmp(Int(math.MinInt64), Float(-two63)))
	s.Equal(-1, cmp(Float(1.5), Float(2.5)))

	_, ordered, err := CompareNumbers(Int(1), Float(math.NaN()))
	s.Require().NoError(err)
	s.False(ordered)
	_, ordered, err = CompareNumbers(Float(math.NaN()), Int(1))
	s.Require().NoError(err)
	s.False(ordered)
	_, ordered, err = CompareNumbers(Float(math.NaN()), Float(1))
	s.Require().NoError(err)
	s.False(ordered)

	_, _, err = CompareNumbers(Int(1), String("1"))
	s.Error(err)
}

func (s *BoxTestSuite) TestEqualValuesAcrossNumberRepresentations() {
	s.True(EqualValues(Int(5), Float(5)))
	s.True(EqualValues(Float(5), Int(5)))
	s.False(EqualValues(Int(twoTo53Plus), Int(twoTo53)))
	s.False(EqualValues(Int(twoTo53Plus), Float(float64(twoTo53Plus))), "the float rounded down to 2^53")
	s.True(EqualValues(List([]Value{Int(1), Float(2)}), List([]Value{Float(1), Int(2)})))
	s.False(EqualValues(Float(math.NaN()), Float(math.NaN())))
	s.True(ContainsValue(List([]Value{Int(twoTo53Plus)}), Int(twoTo53Plus)))
	s.False(ContainsValue(List([]Value{Int(twoTo53)}), Int(twoTo53Plus)))
}

func (s *BoxTestSuite) TestTrinaryFromExactInteger() {
	s.Equal(TrinaryFrom(Int(0)), TrinaryFrom(Float(0)))
	s.Equal(TrinaryFrom(Int(3)), TrinaryFrom(Float(3)))
}
