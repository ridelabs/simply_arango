package orm

import "fmt"

// ---------------------
// Manage variables
// ---------------------

type Variable interface {
	String() string
}

func NewVariableFactory() *VariableFactory {
	return &VariableFactory{
		keyTracker:  make(map[string]string),
		symbolTable: make(map[string]interface{}),
	}
}

const VarPrefix = "var"

type VariableFactory struct {
	variableCounter int
	keyTracker      map[string]string
	symbolTable     map[string]interface{}
}

func (c *VariableFactory) MakeVariable(value interface{}) Variable {
	valHash := fmt.Sprint(value)
	var varName string
	if v, ok := c.keyTracker[valHash]; ok {
		varName = v // reuse the variable name for this value
	} else {
		varName = fmt.Sprintf("%s_%d", VarPrefix, c.variableCounter)
		c.variableCounter++
		c.keyTracker[valHash] = varName
		c.symbolTable[varName] = value
	}

	return &QueryVariable{
		name: varName,
	}
}

func (c *VariableFactory) SymbolTable() map[string]interface{} {
	return c.symbolTable
}

type QueryVariable struct {
	name string
}

func (c *QueryVariable) String() string {
	return fmt.Sprintf("@%s", c.name)
}
