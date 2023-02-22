package httpexpect

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
)

// Number provides methods to inspect attached float64 value
// (Go representation of JSON number).
type Number struct {
	noCopy noCopy
	chain  *chain
	value  big.Float
}

// NewNumber returns a new Number instance.
//
// If reporter is nil, the function panics.
//
// Example:
//
//	number := NewNumber(t, 123.4)
func NewNumber(reporter Reporter, value float64) *Number {
	return newNumber(newChainWithDefaults("Number()", reporter), value)
}

// NewNumberC returns a new Number instance with config.
//
// Requirements for config are same as for WithConfig function.
//
// Example:
//
//	number := NewNumberC(config, 123.4)
func NewNumberC(config Config, value float64) *Number {
	return newNumber(newChainWithConfig("Number()", config.withDefaults()), value)
}

func newNumber(parent *chain, val interface{}) *Number {
	number, _ := canonNumber(parent, val)
	return &Number{chain: parent.clone(), value: number}
}

// Raw returns underlying value attached to Number.
// This is the value originally passed to NewNumber.
//
// Example:
//
//	number := NewNumber(t, 123.4)
//	assert.Equal(t, 123.4, number.Raw())
//
// Deprecated: Use AsFloat() instead.
func (n *Number) Raw() float64 {
	value, _ := n.value.Float64()
	return value
}

// Decode unmarshals the underlying value attached to the Number to a target variable.
// target should be one of these:
//
//   - pointer to an empty interface
//   - pointer to any integer or floating type
//
// Example:
//
//	value := NewNumber(t, 123)
//
//	var target interface{}
//	valude.decode(&target)
//
//	assert.Equal(t, 123, target)
func (n *Number) Decode(target interface{}) *Number {
	opChain := n.chain.enter("Decode()")
	defer opChain.leave()

	if opChain.failed() {
		return n
	}

	canonNumberDecode(opChain, n.value, target)
	return n
}

// Alias is similar to Value.Alias.
func (n *Number) Alias(name string) *Number {
	opChain := n.chain.enter("Alias(%q)", name)
	defer opChain.leave()

	n.chain.setAlias(name)
	return n
}

// Path is similar to Value.Path.
func (n *Number) Path(path string) *Value {
	opChain := n.chain.enter("Path(%q)", path)
	defer opChain.leave()

	return jsonPath(opChain, json.Number(n.value.String()), path)
}

// Schema is similar to Value.Schema.
func (n *Number) Schema(schema interface{}) *Number {
	opChain := n.chain.enter("Schema()")
	defer opChain.leave()

	jsonSchema(opChain, json.Number(n.value.String()), schema)
	return n
}

// IsEqual succeeds if number is equal to given value.
//
// value should have numeric type convertible to float64. Before comparison,
// it is converted to float64.
//
// Example:
//
//	number := NewNumber(t, 123)
//	number.IsEqual(float64(123))
//	number.IsEqual(int32(123))
func (n *Number) IsEqual(value interface{}) *Number {
	opChain := n.chain.enter("IsEqual()")
	defer opChain.leave()

	if opChain.failed() {
		return n
	}

	num, ok := canonNumber(opChain, value)
	if !ok {
		opChain.fail(AssertionFailure{
			Type:     AssertEqual,
			Actual:   &AssertionValue{n.value},
			Expected: &AssertionValue{num},
			Errors: []error{
				errors.New("expected: numbers are equal"),
			},
		})
		return n
	}

	if n.value.Cmp(&num) != 0 {
		opChain.fail(AssertionFailure{
			Type:     AssertEqual,
			Actual:   &AssertionValue{n.value},
			Expected: &AssertionValue{num},
			Errors: []error{
				errors.New("expected: numbers are equal"),
			},
		})
	}

	return n
}

// NotEqual succeeds if number is not equal to given value.
//
// value should have numeric type convertible to float64. Before comparison,
// it is converted to float64.
//
// Example:
//
//	number := NewNumber(t, 123)
//	number.NotEqual(float64(321))
//	number.NotEqual(int32(321))
func (n *Number) NotEqual(value interface{}) *Number {
	opChain := n.chain.enter("NotEqual()")
	defer opChain.leave()

	if opChain.failed() {
		return n
	}

	num, ok := canonNumber(opChain, value)

	if n.value.Cmp(&num) == 0 || !ok {
		opChain.fail(AssertionFailure{
			Type:     AssertNotEqual,
			Actual:   &AssertionValue{n.value},
			Expected: &AssertionValue{num},
			Errors: []error{
				errors.New("expected: numbers are non-equal"),
			},
		})
	}

	return n
}

// Deprecated: use IsEqual instead.
func (n *Number) Equal(value interface{}) *Number {
	return n.IsEqual(value)
}

// InDelta succeeds if two numerals are within delta of each other.
//
// Example:
//
//	number := NewNumber(t, 123.0)
//	number.InDelta(123.2, 0.3)
func (n *Number) InDelta(value, delta interface{}) *Number {
	opChain := n.chain.enter("InDelta()")
	defer opChain.leave()

	if opChain.failed() {
		return n
	}

	num, ok := canonNumber(opChain, value)
	del, okDel := canonNumber(opChain, delta)

	if !ok || !okDel || delta != delta || value != value {
		opChain.fail(AssertionFailure{
			Type:     AssertEqual,
			Actual:   &AssertionValue{n.value},
			Expected: &AssertionValue{value},
			Delta:    &AssertionValue{delta},
			Errors: []error{
				errors.New("expected: numbers are comparable"),
			},
		})
		return n
	}

	diff := big.NewFloat(0).Sub(&n.value, &num)

	if diff.Cmp(del.Neg(&del)) < 0 || diff.Cmp(&del) > 0 {
		opChain.fail(AssertionFailure{
			Type:     AssertEqual,
			Actual:   &AssertionValue{n.value},
			Expected: &AssertionValue{value},
			Delta:    &AssertionValue{delta},
			Errors: []error{
				errors.New("expected: numbers lie within delta"),
			},
		})
		return n
	}

	return n
}

// NotInDelta succeeds if two numerals are not within delta of each other.
//
// Example:
//
//	number := NewNumber(t, 123.0)
//	number.NotInDelta(123.2, 0.1)
func (n *Number) NotInDelta(value, delta interface{}) *Number {
	opChain := n.chain.enter("NotInDelta()")
	defer opChain.leave()

	if opChain.failed() {
		return n
	}

	num, ok := canonNumber(opChain, value)
	del, okDel := canonNumber(opChain, delta)

	if !ok || !okDel || delta != delta || value != value {
		opChain.fail(AssertionFailure{
			Type:     AssertNotEqual,
			Actual:   &AssertionValue{n.value},
			Expected: &AssertionValue{value},
			Delta:    &AssertionValue{delta},
			Errors: []error{
				errors.New("expected: numbers are comparable"),
			},
		})
		return n
	}

	diff := n.value.Sub(&n.value, &num)
	fmt.Println(diff.Cmp(del.Neg(&del)) < 0, diff.Cmp(&del) > 0)
	fmt.Println(diff, del.Neg(&del))

	if diff.Cmp(del.Neg(&del)) >= 0 && diff.Cmp(&del) <= 0 {
		opChain.fail(AssertionFailure{
			Type:     AssertNotEqual,
			Actual:   &AssertionValue{n.value},
			Expected: &AssertionValue{value},
			Delta:    &AssertionValue{delta},
			Errors: []error{
				errors.New("expected: numbers do not lie within delta"),
			},
		})
		return n
	}
	return n
}

// Deprecated: use InDelta instead.
func (n *Number) EqualDelta(value, delta float64) *Number {
	return n.InDelta(value, delta)
}

// Deprecated: use NotInDelta instead.
func (n *Number) NotEqualDelta(value, delta float64) *Number {
	return n.NotInDelta(value, delta)
}

// InRange succeeds if number is within given range [min; max].
//
// min and max should have numeric type convertible to float64. Before comparison,
// they are converted to float64.
//
// Example:
//
//	number := NewNumber(t, 123)
//	number.InRange(float32(100), int32(200))  // success
//	number.InRange(100, 200)                  // success
//	number.InRange(123, 123)                  // success
func (n *Number) InRange(min, max interface{}) *Number {
	opChain := n.chain.enter("InRange()")
	defer opChain.leave()

	if opChain.failed() {
		return n
	}

	a, ok := canonNumber(opChain, min)
	if !ok {
		return n
	}

	b, ok := canonNumber(opChain, max)
	if !ok {
		return n
	}

	if n.value.Cmp(&a) < 0 || n.value.Cmp(&b) > 0 {
		opChain.fail(AssertionFailure{
			Type:     AssertInRange,
			Actual:   &AssertionValue{n.value},
			Expected: &AssertionValue{AssertionRange{a, b}},
			Errors: []error{
				errors.New("expected: number is within given range"),
			},
		})
	}

	return n
}

// NotInRange succeeds if number is not within given range [min; max].
//
// min and max should have numeric type convertible to float64. Before comparison,
// they are converted to float64.
//
// Example:
//
//	number := NewNumber(t, 100)
//	number.NotInRange(0, 99)
//	number.NotInRange(101, 200)
func (n *Number) NotInRange(min, max interface{}) *Number {
	opChain := n.chain.enter("NotInRange()")
	defer opChain.leave()

	if opChain.failed() {
		return n
	}

	a, ok := canonNumber(opChain, min)
	if !ok {
		return n
	}

	b, ok := canonNumber(opChain, max)
	if !ok {
		return n
	}

	if n.value.Cmp(&a) >= 0 && n.value.Cmp(&b) <= 0 {
		opChain.fail(AssertionFailure{
			Type:     AssertNotInRange,
			Actual:   &AssertionValue{n.value},
			Expected: &AssertionValue{AssertionRange{a, b}},
			Errors: []error{
				errors.New("expected: number is not within given range"),
			},
		})
	}

	return n
}

// InList succeeds if the number is equal to one of the values from given list
// of numbers. Before comparison, each value is converted to canonical form.
//
// Each value should be numeric type convertible to float64. If at least one
// value has wrong type, failure is reported.
//
// Example:
//
//	number := NewNumber(t, 123)
//	number.InList(float64(123), int32(123))
func (n *Number) InList(values ...interface{}) *Number {
	opChain := n.chain.enter("IsList()")
	defer opChain.leave()

	if opChain.failed() {
		return n
	}

	if len(values) == 0 {
		opChain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected empty list argument"),
			},
		})
		return n
	}

	var isListed bool
	for _, v := range values {
		num, ok := canonNumber(opChain, v)
		if !ok {
			return n
		}

		if n.value.Cmp(&num) == 0 {
			isListed = true
			// continue loop to check that all values are correct
		}
	}

	if !isListed {
		opChain.fail(AssertionFailure{
			Type:     AssertBelongs,
			Actual:   &AssertionValue{n.value},
			Expected: &AssertionValue{AssertionList(values)},
			Errors: []error{
				errors.New("expected: number is equal to one of the values"),
			},
		})
	}

	return n
}

// NotInList succeeds if the number is not equal to any of the values from given
// list of numbers. Before comparison, each value is converted to canonical form.
//
// Each value should be numeric type convertible to float64. If at least one
// value has wrong type, failure is reported.
//
// Example:
//
//	number := NewNumber(t, 123)
//	number.NotInList(float64(456), int32(456))
func (n *Number) NotInList(values ...interface{}) *Number {
	opChain := n.chain.enter("NotInList()")
	defer opChain.leave()

	if opChain.failed() {
		return n
	}

	if len(values) == 0 {
		opChain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected empty list argument"),
			},
		})
		return n
	}

	for _, v := range values {
		num, ok := canonNumber(opChain, v)
		if !ok {
			return n
		}

		if n.value.Cmp(&num) == 0 {
			opChain.fail(AssertionFailure{
				Type:     AssertNotBelongs,
				Actual:   &AssertionValue{n.value},
				Expected: &AssertionValue{AssertionList(values)},
				Errors: []error{
					errors.New("expected: number is not equal to any of the values"),
				},
			})
			return n
		}
	}

	return n
}

// Gt succeeds if number is greater than given value.
//
// value should have numeric type convertible to float64. Before comparison,
// it is converted to float64.
//
// Example:
//
//	number := NewNumber(t, 123)
//	number.Gt(float64(122))
//	number.Gt(int32(122))
func (n *Number) Gt(value interface{}) *Number {
	opChain := n.chain.enter("Gt()")
	defer opChain.leave()

	if opChain.failed() {
		return n
	}

	num, ok := canonNumber(opChain, value)
	if !ok {
		return n
	}

	if n.value.Cmp(&num) <= 0 {
		opChain.fail(AssertionFailure{
			Type:     AssertGt,
			Actual:   &AssertionValue{n.value},
			Expected: &AssertionValue{num},
			Errors: []error{
				errors.New("expected: number is larger than given value"),
			},
		})
	}

	return n
}

// Ge succeeds if number is greater than or equal to given value.
//
// value should have numeric type convertible to float64. Before comparison,
// it is converted to float64.
//
// Example:
//
//	number := NewNumber(t, 123)
//	number.Ge(float64(122))
//	number.Ge(int32(122))
func (n *Number) Ge(value interface{}) *Number {
	opChain := n.chain.enter("Ge()")
	defer opChain.leave()

	if opChain.failed() {
		return n
	}

	num, ok := canonNumber(opChain, value)
	if !ok {
		return n
	}

	if n.value.Cmp(&num) < 0 {
		opChain.fail(AssertionFailure{
			Type:     AssertGe,
			Actual:   &AssertionValue{n.value},
			Expected: &AssertionValue{num},
			Errors: []error{
				errors.New("expected: number is larger than or equal to given value"),
			},
		})
	}

	return n
}

// Lt succeeds if number is lesser than given value.
//
// value should have numeric type convertible to float64. Before comparison,
// it is converted to float64.
//
// Example:
//
//	number := NewNumber(t, 123)
//	number.Lt(float64(124))
//	number.Lt(int32(124))
func (n *Number) Lt(value interface{}) *Number {
	opChain := n.chain.enter("Lt()")
	defer opChain.leave()

	if opChain.failed() {
		return n
	}

	num, ok := canonNumber(opChain, value)
	if !ok {
		return n
	}

	if n.value.Cmp(&num) >= 0 {
		opChain.fail(AssertionFailure{
			Type:     AssertLt,
			Actual:   &AssertionValue{n.value},
			Expected: &AssertionValue{num},
			Errors: []error{
				errors.New("expected: number is less than given value"),
			},
		})
	}

	return n
}

// Le succeeds if number is lesser than or equal to given value.
//
// value should have numeric type convertible to float64. Before comparison,
// it is converted to float64.
//
// Example:
//
//	number := NewNumber(t, 123)
//	number.Le(float64(124))
//	number.Le(int32(124))
func (n *Number) Le(value interface{}) *Number {
	opChain := n.chain.enter("Le()")
	defer opChain.leave()

	if opChain.failed() {
		return n
	}

	num, ok := canonNumber(opChain, value)
	if !ok {
		return n
	}

	if n.value.Cmp(&num) > 0 {
		opChain.fail(AssertionFailure{
			Type:     AssertLe,
			Actual:   &AssertionValue{n.value},
			Expected: &AssertionValue{num},
			Errors: []error{
				errors.New("expected: number is less than or equal to given value"),
			},
		})
	}

	return n
}
