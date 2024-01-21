package orm

import "fmt"

type Operator struct {
	variableFactory *VariableFactory
}

func (c *Operator) MakeVariableIfNative(input interface{}) Expression {
	switch v := input.(type) {
	case float32, float64, int, int8, int16, int32, int64, byte, bool, string:
		return c.variableFactory.MakeVariable(v)
	default:
		return &ExpressionWrapper{item: v}
	}
}

// ----------------------
// Boolean operators
// ----------------------

func (c *Operator) And(left, right Expression) Expression {
	return &AndExpression{
		left:  left,
		right: right,
	}
}

func (c *Operator) Or(left, right Expression) Expression {
	return &OrExpression{left: left, right: right}
}

func (c *Operator) Not(expression interface{}) Expression {
	return &NotExpression{expression: c.MakeVariableIfNative(expression)}
}

// ----------------------
// Equality operators
// ----------------------

func (c *Operator) Equal(attribute string, right interface{}) Expression {
	return &EqualityExpression{left: NewAttribute(attribute), operator: EqualityExpressionEqual, right: c.MakeVariableIfNative(right)}
}

func (c *Operator) LessThan(attribute string, right Expression) Expression {
	return &EqualityExpression{left: NewAttribute(attribute), operator: EqualityExpressionLessThan, right: c.MakeVariableIfNative(right)}
}

func (c *Operator) LessThanOrEqual(attribute string, right Expression) Expression {
	return &EqualityExpression{left: NewAttribute(attribute), operator: EqualityExpressionLessThanOrEqualTo, right: c.MakeVariableIfNative(right)}
}

func (c *Operator) GreaterThan(attribute string, right Expression) Expression {
	return &EqualityExpression{left: NewAttribute(attribute), operator: EqualityExpressionGreaterThan, right: c.MakeVariableIfNative(right)}
}

func (c *Operator) GreaterThanOrEqual(attribute string, right Expression) Expression {
	return &EqualityExpression{left: NewAttribute(attribute), operator: EqualityExpressionGreaterThanOrEqualTo, right: c.MakeVariableIfNative(right)}
}

// ----------------------
// String operators
// ----------------------

func (c *Operator) EndsWith(attribute string, pattern string) Expression {
	return &LikeExpression{
		left:  NewAttribute(attribute),
		right: c.variableFactory.MakeVariable(fmt.Sprintf("%%%s", pattern)),
	}
}

func (c *Operator) StartsWith(attribute string, pattern string) Expression {
	return &LikeExpression{
		left:  NewAttribute(attribute),
		right: c.variableFactory.MakeVariable(fmt.Sprintf("%s%%", pattern)),
	}
}

func (c *Operator) Contains(attribute string, pattern string) Expression {
	return &LikeExpression{
		left:  NewAttribute(attribute),
		right: c.variableFactory.MakeVariable(fmt.Sprintf("%%%s%%", pattern)),
	}
}
