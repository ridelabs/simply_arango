package orm

import "fmt"

// ---------------------
// logical operators
// ---------------------

type AndExpression struct {
	left  interface{}
	right interface{}
}

func (c *AndExpression) String() string {
	return fmt.Sprintf("(%s && %s)", c.left, c.right)
}

type OrExpression struct {
	left  interface{}
	right interface{}
}

func (c *OrExpression) String() string {
	return fmt.Sprintf("(%s || %s)", c.left, c.right)
}

type NotExpression struct {
	expression interface{}
}

func (c *NotExpression) String() string {
	return fmt.Sprintf("(NOT %s) ", c.expression)
}

// ---------------------
// equiv operators
// ---------------------

type EqualityOperator string

const EqualityExpressionEqual = EqualityOperator("==")
const EqualityExpressionNotEqual = EqualityOperator("!=")
const EqualityExpressionLessThan = EqualityOperator("<")
const EqualityExpressionGreaterThan = EqualityOperator(">")
const EqualityExpressionLessThanOrEqualTo = EqualityOperator("<=")
const EqualityExpressionGreaterThanOrEqualTo = EqualityOperator(">=")

type EqualityExpression struct {
	left     interface{}
	operator EqualityOperator
	right    interface{}
}

func (c *EqualityExpression) String() string {
	return fmt.Sprintf("(%s %s %s) ", c.left, c.operator, c.right)
}

// ---------------------
// string operators
// ---------------------

type LikeExpression struct {
	left  interface{}
	right interface{}
}

func (c *LikeExpression) String() string {
	return fmt.Sprintf("%s LIKE %s ", c.left, c.right)
}

// ---------------------
// Arrays
// ---------------------

type InArrayExpression struct {
	value     interface{}
	arrayName interface{}
}

func (c *InArrayExpression) String() string {
	return fmt.Sprintf("%s IN %s", c.value, c.arrayName)
}

type EmptyArrayExpression struct {
	arrayName interface{}
}

func (c *EmptyArrayExpression) String() string {
	return fmt.Sprintf("%s == null OR LENGTH(%s) == 0", c.arrayName, c.arrayName)
}

// ---------------------
// native types
// ---------------------

type NativeExpression struct {
	value interface{}
}

func (c *NativeExpression) String() string {
	switch v := c.value.(type) {
	case string:
		return fmt.Sprintf("\"%s\"", v)
	default:
		return fmt.Sprintf("%s", v)
	}
}

func RawValue(value interface{}) Expression {
	return &NativeExpression{value: value}
}
