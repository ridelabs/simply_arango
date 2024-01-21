package orm

import "fmt"

// ---------------------
// Wrap document attributes
// ---------------------

type DocumentAttribute struct {
	name string
}

func (c *DocumentAttribute) Name() string {
	return c.name
}

func (c *DocumentAttribute) String() string {
	return fmt.Sprintf("%s.%s", DocumentName, c.name)
}

func NewAttribute(name string) interface{} {
	return &DocumentAttribute{name: name}
}
