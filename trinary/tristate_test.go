package trinary

import (
	"log/slog"
	"testing"

	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/stretchr/testify/suite"
)

type TristateTestSuite struct {
	suite.Suite
}

func (suite *TristateTestSuite) SetupSuite() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(suite.T().Output(), nil)))
}

func (suite *TristateTestSuite) BeforeTest(suiteName, testName string) {
	slog.InfoContext(suite.T().Context(), "BeforeTest start", slog.String("TestSuite", suiteName), slog.String("TestName", testName))
	defer slog.InfoContext(suite.T().Context(), "BeforeTest end", slog.String("TestSuite", suiteName), slog.String("TestName", testName))
}

func (suite *TristateTestSuite) AfterTest(suiteName, testName string) {
	slog.InfoContext(suite.T().Context(), "AfterTest start", slog.String("TestSuite", suiteName), slog.String("TestName", testName))
	defer slog.InfoContext(suite.T().Context(), "AfterTest end", slog.String("TestSuite", suiteName), slog.String("TestName", testName))
}

func (suite *TristateTestSuite) TearDownSuite() {
	slog.InfoContext(suite.T().Context(), "TearDownSuite")
	defer slog.InfoContext(suite.T().Context(), "TearDownSuite end")
}

// TestValueConstants tests the Value constants
func (s *TristateTestSuite) TestValueConstants() {
	s.Equal(Value(-1), False)
	s.Equal(Value(0), Unknown)
	s.Equal(Value(1), True)
}

// TestString tests the String() method
func (s *TristateTestSuite) TestString() {
	s.Equal("true", True.String())
	s.Equal("false", False.String())
	s.Equal("unknown", Unknown.String())

	// Test invalid value
	invalidValue := Value(999)
	s.Equal("unknown", invalidValue.String())
}

// TestMarshalJSON tests JSON marshaling
func (s *TristateTestSuite) TestMarshalJSON() {
	trueJSON, err := True.MarshalJSON()
	s.NoError(err)
	s.Equal(`"true"`, string(trueJSON))

	falseJSON, err := False.MarshalJSON()
	s.NoError(err)
	s.Equal(`"false"`, string(falseJSON))

	unknownJSON, err := Unknown.MarshalJSON()
	s.NoError(err)
	s.Equal(`"unknown"`, string(unknownJSON))
}

// TestNot tests the Not() method
func (s *TristateTestSuite) TestNot() {
	s.Equal(False, True.Not())
	s.Equal(True, False.Not())
	s.Equal(Unknown, Unknown.Not())

	// Test invalid value
	invalidValue := Value(999)
	s.Equal(Unknown, invalidValue.Not())
}

// TestAnd tests the And() method with comprehensive truth table
func (s *TristateTestSuite) TestAnd() {
	// True AND cases
	s.Equal(True, True.And(True))
	s.Equal(False, True.And(False))
	s.Equal(Unknown, True.And(Unknown))

	// False AND cases
	s.Equal(False, False.And(True))
	s.Equal(False, False.And(False))
	s.Equal(False, False.And(Unknown))

	// Unknown AND cases
	s.Equal(Unknown, Unknown.And(True))
	s.Equal(False, Unknown.And(False))
	s.Equal(Unknown, Unknown.And(Unknown))

	// Test invalid values
	invalidValue := Value(999)
	s.Equal(Unknown, invalidValue.And(True))      // invalid AND True = Unknown (default case)
	s.Equal(Unknown, invalidValue.And(False))     // invalid AND False = Unknown (default case)
	s.Equal(Unknown, invalidValue.And(Unknown))   // invalid AND Unknown = Unknown (default case)
	s.Equal(invalidValue, True.And(invalidValue)) // True AND invalid = invalid (True ∧ x = x)
	s.Equal(False, False.And(invalidValue))       // False AND anything = False
	s.Equal(Unknown, Unknown.And(invalidValue))   // Unknown AND invalid = Unknown (default case)
}

// TestOr tests the Or() method with comprehensive truth table
func (s *TristateTestSuite) TestOr() {
	// True OR cases
	s.Equal(True, True.Or(True))
	s.Equal(True, True.Or(False))
	s.Equal(True, True.Or(Unknown))

	// False OR cases
	s.Equal(True, False.Or(True))
	s.Equal(False, False.Or(False))
	s.Equal(Unknown, False.Or(Unknown))

	// Unknown OR cases
	s.Equal(True, Unknown.Or(True))
	s.Equal(Unknown, Unknown.Or(False))
	s.Equal(Unknown, Unknown.Or(Unknown))

	// Test invalid values
	invalidValue := Value(999)
	s.Equal(Unknown, invalidValue.Or(True))       // invalid OR True = Unknown (default case)
	s.Equal(Unknown, invalidValue.Or(False))      // invalid OR False = Unknown (default case)
	s.Equal(Unknown, invalidValue.Or(Unknown))    // invalid OR Unknown = Unknown (default case)
	s.Equal(True, True.Or(invalidValue))          // True OR anything = True
	s.Equal(invalidValue, False.Or(invalidValue)) // False OR invalid = invalid (False ∨ x = x)
	s.Equal(Unknown, Unknown.Or(invalidValue))    // Unknown OR invalid = Unknown (default case)
}

// TestEquals tests the Equals() method
func (s *TristateTestSuite) TestEquals() {
	s.True(True.Equals(True))
	s.False(True.Equals(False))
	s.False(True.Equals(Unknown))

	s.False(False.Equals(True))
	s.True(False.Equals(False))
	s.False(False.Equals(Unknown))

	s.False(Unknown.Equals(True))
	s.False(Unknown.Equals(False))
	s.True(Unknown.Equals(Unknown))

	// Test invalid values
	invalidValue := Value(999)
	s.False(True.Equals(invalidValue))
	s.False(False.Equals(invalidValue))
	s.False(Unknown.Equals(invalidValue))
	s.True(invalidValue.Equals(invalidValue))
}

// TestIsTrue tests the IsTrue() method
func (s *TristateTestSuite) TestIsTrue() {
	s.True(True.IsTrue())
	s.False(False.IsTrue())
	s.False(Unknown.IsTrue())

	// Test invalid value
	invalidValue := Value(999)
	s.False(invalidValue.IsTrue())
}

// TestFromToken tests the FromToken() function
func (s *TristateTestSuite) TestFromToken() {
	// Test with valid token kinds
	trueToken := tokens.New(tokens.KeywordTrue, "true", tokens.Range{})
	s.Equal(True, FromToken(trueToken))

	falseToken := tokens.New(tokens.KeywordFalse, "false", tokens.Range{})
	s.Equal(False, FromToken(falseToken))

	unknownToken := tokens.New(tokens.KeywordUnknown, "unknown", tokens.Range{})
	s.Equal(Unknown, FromToken(unknownToken))

	// Test with invalid token kinds
	invalidToken := tokens.New(tokens.KeywordNull, "null", tokens.Range{})
	s.Equal(False, FromToken(invalidToken))

	identToken := tokens.New(tokens.Ident, "someIdentifier", tokens.Range{})
	s.Equal(False, FromToken(identToken))

	stringToken := tokens.New(tokens.String, "someString", tokens.Range{})
	s.Equal(False, FromToken(stringToken))
}

// TestFrom tests the From() function with various Go types
func (s *TristateTestSuite) TestFrom() {
	// Test nil
	s.Equal(Unknown, From(nil))

	// Test trinary values
	s.Equal(True, From(True))
	s.Equal(False, From(False))
	s.Equal(Unknown, From(Unknown))

	// Test bool
	s.Equal(True, From(true))
	s.Equal(False, From(false))

	// Test *bool
	truePtr := &[]bool{true}[0]
	falsePtr := &[]bool{false}[0]
	s.Equal(True, From(truePtr))
	s.Equal(False, From(falsePtr))
	s.Equal(Unknown, From((*bool)(nil)))

	// Test HasTrinary interface
	trinaryValue := &testTrinaryValue{value: True}
	s.Equal(True, From(trinaryValue))

	trinaryValue.value = False
	s.Equal(False, From(trinaryValue))

	trinaryValue.value = Unknown
	s.Equal(Unknown, From(trinaryValue))

	// Test various Go types for truthiness
	s.Equal(True, From("non-empty string"))
	s.Equal(False, From(""))
	s.Equal(True, From(42))
	s.Equal(False, From(0))
	s.Equal(True, From(3.14))
	s.Equal(False, From(0.0))
	s.Equal(True, From([]int{1, 2, 3}))
	s.Equal(False, From([]int{}))
	s.Equal(True, From(map[string]int{"key": 1}))
	s.Equal(False, From(map[string]int{}))

	// Test pointers
	var nilPtr *int
	s.Equal(False, From(nilPtr))

	intPtr := &[]int{42}[0]
	s.Equal(True, From(intPtr))

	// Test struct
	type testStruct struct {
		Field string
	}
	s.Equal(True, From(testStruct{Field: "value"}))
	s.Equal(True, From(testStruct{})) // Non-nil struct is truthy
}

// TestIsTruthy tests the IsTruthy() function
func (s *TristateTestSuite) TestIsTruthy() {
	// Test nil
	s.False(IsTruthy(nil))

	// Test bool
	s.True(IsTruthy(true))
	s.False(IsTruthy(false))

	// Test string
	s.True(IsTruthy("non-empty"))
	s.False(IsTruthy(""))

	// Test integers
	s.True(IsTruthy(42))
	s.False(IsTruthy(0))
	s.True(IsTruthy(int8(1)))
	s.False(IsTruthy(int8(0)))
	s.True(IsTruthy(int16(1)))
	s.False(IsTruthy(int16(0)))
	s.True(IsTruthy(int32(1)))
	s.False(IsTruthy(int32(0)))
	s.True(IsTruthy(int64(1)))
	s.False(IsTruthy(int64(0)))

	// Test unsigned integers
	s.True(IsTruthy(uint(1)))
	s.False(IsTruthy(uint(0)))
	s.True(IsTruthy(uint8(1)))
	s.False(IsTruthy(uint8(0)))
	s.True(IsTruthy(uint16(1)))
	s.False(IsTruthy(uint16(0)))
	s.True(IsTruthy(uint32(1)))
	s.False(IsTruthy(uint32(0)))
	s.True(IsTruthy(uint64(1)))
	s.False(IsTruthy(uint64(0)))
	s.True(IsTruthy(uintptr(1)))
	s.False(IsTruthy(uintptr(0)))

	// Test floats
	s.True(IsTruthy(3.14))
	s.False(IsTruthy(0.0))
	s.True(IsTruthy(float32(1.0)))
	s.False(IsTruthy(float32(0.0)))
	s.True(IsTruthy(float64(1.0)))
	s.False(IsTruthy(float64(0.0)))

	// Test slices and arrays
	s.True(IsTruthy([]int{1, 2, 3}))
	s.False(IsTruthy([]int{}))
	s.True(IsTruthy([3]int{1, 2, 3}))
	s.False(IsTruthy([0]int{}))

	// Test maps
	s.True(IsTruthy(map[string]int{"key": 1}))
	s.False(IsTruthy(map[string]int{}))

	// Test pointers
	var nilPtr *int
	s.False(IsTruthy(nilPtr))

	intPtr := &[]int{42}[0]
	s.True(IsTruthy(intPtr))

	// Test interfaces
	var nilInterface interface{}
	s.False(IsTruthy(nilInterface))

	var interfaceValue interface{} = 42
	s.True(IsTruthy(interfaceValue))

	// Test nested pointers
	var nilPtrPtr **int
	s.False(IsTruthy(nilPtrPtr))

	ptrPtr := &intPtr
	s.True(IsTruthy(ptrPtr))

	// Test struct
	type testStruct struct {
		Field string
	}
	s.True(IsTruthy(testStruct{Field: "value"}))
	s.True(IsTruthy(testStruct{})) // Non-nil struct is truthy

	// Test default case (non-nil values are truthy)
	s.True(IsTruthy(struct{}{}))
}

// testTrinaryValue implements HasTrinary interface for testing
type testTrinaryValue struct {
	value Value
}

func (t *testTrinaryValue) ToTrinary() Value {
	return t.value
}

// TestEdgeCases tests various edge cases and error conditions
func (s *TristateTestSuite) TestEdgeCases() {
	// Test very large invalid values
	largeValue := Value(999999)
	s.Equal("unknown", largeValue.String())
	s.Equal(Unknown, largeValue.Not())
	s.False(largeValue.IsTrue())

	// Test negative values
	negativeValue := Value(-999)
	s.Equal("unknown", negativeValue.String())
	s.Equal(Unknown, negativeValue.Not())
	s.False(negativeValue.IsTrue())

	// Test zero value (should be Unknown)
	zeroValue := Value(0)
	s.Equal(Unknown, zeroValue)
	s.True(zeroValue.Equals(Unknown))

	// Test complex nested structures
	type nestedStruct struct {
		Inner *nestedStruct
		Value int
	}

	nilNested := (*nestedStruct)(nil)
	s.False(IsTruthy(nilNested))
	s.Equal(False, From(nilNested))

	emptyNested := &nestedStruct{}
	s.True(IsTruthy(emptyNested))
	s.Equal(True, From(emptyNested))

	// Test channels
	var nilChan chan int
	s.True(IsTruthy(nilChan)) // Zero value of channel is truthy by default
	s.Equal(True, From(nilChan))

	ch := make(chan int, 1)
	s.True(IsTruthy(ch)) // Non-nil channels are truthy by default
	s.Equal(True, From(ch))
	close(ch)

	// Test functions
	var nilFunc func()
	s.True(IsTruthy(nilFunc)) // Zero value of function is truthy by default
	s.Equal(True, From(nilFunc))

	funcValue := func() {}
	s.True(IsTruthy(funcValue)) // Non-nil functions are truthy by default
	s.Equal(True, From(funcValue))

	// Test complex slices
	var nilSlice []int
	s.False(IsTruthy(nilSlice))
	s.Equal(False, From(nilSlice))

	emptySlice := []int{}
	s.False(IsTruthy(emptySlice))
	s.Equal(False, From(emptySlice))

	// Test complex maps
	var nilMap map[string]int
	s.False(IsTruthy(nilMap))
	s.Equal(False, From(nilMap))

	emptyMap := map[string]int{}
	s.False(IsTruthy(emptyMap))
	s.Equal(False, From(emptyMap))
}

// TestLogicalOperatorCommutativity tests that logical operators are commutative where expected
func (s *TristateTestSuite) TestLogicalOperatorCommutativity() {
	// AND should be commutative for True and False
	s.Equal(True.And(False), False.And(True))
	s.Equal(True.And(True), True.And(True))
	s.Equal(False.And(False), False.And(False))

	// OR should be commutative for True and False
	s.Equal(True.Or(False), False.Or(True))
	s.Equal(True.Or(True), True.Or(True))
	s.Equal(False.Or(False), False.Or(False))

	// Unknown cases may not be commutative due to Kleene logic
	// True AND Unknown = Unknown, but Unknown AND True = Unknown (commutative)
	s.Equal(True.And(Unknown), Unknown.And(True))
	// False AND Unknown = False, but Unknown AND False = False (commutative)
	s.Equal(False.And(Unknown), Unknown.And(False))

	// True OR Unknown = True, but Unknown OR True = True (commutative)
	s.Equal(True.Or(Unknown), Unknown.Or(True))
	// False OR Unknown = Unknown, but Unknown OR False = Unknown (commutative)
	s.Equal(False.Or(Unknown), Unknown.Or(False))
}

// TestLogicalOperatorAssociativity tests associativity where applicable
func (s *TristateTestSuite) TestLogicalOperatorAssociativity() {
	// Test that (A AND B) AND C = A AND (B AND C) for various combinations
	values := []Value{True, False, Unknown}

	for _, a := range values {
		for _, b := range values {
			for _, c := range values {
				leftAssoc := a.And(b).And(c)
				rightAssoc := a.And(b.And(c))
				s.Equal(leftAssoc, rightAssoc, "AND should be associative for %v, %v, %v", a, b, c)

				leftOrAssoc := a.Or(b).Or(c)
				rightOrAssoc := a.Or(b.Or(c))
				s.Equal(leftOrAssoc, rightOrAssoc, "OR should be associative for %v, %v, %v", a, b, c)
			}
		}
	}
}

// TestDeMorganLaws tests De Morgan's laws for trinary logic
func (s *TristateTestSuite) TestDeMorganLaws() {
	// De Morgan's laws: NOT(A AND B) = NOT(A) OR NOT(B)
	// and NOT(A OR B) = NOT(A) AND NOT(B)
	values := []Value{True, False, Unknown}

	for _, a := range values {
		for _, b := range values {
			// NOT(A AND B) = NOT(A) OR NOT(B)
			left := a.And(b).Not()
			right := a.Not().Or(b.Not())
			s.Equal(left, right, "De Morgan's law 1 failed for %v AND %v", a, b)

			// NOT(A OR B) = NOT(A) AND NOT(B)
			left2 := a.Or(b).Not()
			right2 := a.Not().And(b.Not())
			s.Equal(left2, right2, "De Morgan's law 2 failed for %v OR %v", a, b)
		}
	}
}

// TestDoubleNegation tests that NOT(NOT(x)) = x
func (s *TristateTestSuite) TestDoubleNegation() {
	values := []Value{True, False, Unknown}

	for _, v := range values {
		doubleNeg := v.Not().Not()
		s.Equal(v, doubleNeg, "Double negation should return original value for %v", v)
	}

	// Test with invalid values
	invalidValue := Value(999)
	doubleNeg := invalidValue.Not().Not()
	s.Equal(Unknown, doubleNeg) // Invalid values should become Unknown
}

// TestTruthTableCompleteness verifies that all combinations are tested
func (s *TristateTestSuite) TestTruthTableCompleteness() {
	values := []Value{True, False, Unknown}

	// Test all AND combinations
	for _, a := range values {
		for _, b := range values {
			result := a.And(b)
			// Verify result is one of the three valid values
			s.Contains([]Value{True, False, Unknown}, result, "AND result should be valid for %v AND %v", a, b)
		}
	}

	// Test all OR combinations
	for _, a := range values {
		for _, b := range values {
			result := a.Or(b)
			// Verify result is one of the three valid values
			s.Contains([]Value{True, False, Unknown}, result, "OR result should be valid for %v OR %v", a, b)
		}
	}
}

// TestTristateTestSuite runs the test suite
func TestTristateTestSuite(t *testing.T) {
	suite.Run(t, new(TristateTestSuite))
}
