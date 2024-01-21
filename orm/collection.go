package orm

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/ridelabs/simply_arango/encoding"
	log "github.com/sirupsen/logrus"

	"github.com/arangodb/go-driver"
)

// -------------------------------------
// Simple Object Relational Mapper (ORM)
// For the golang arangodb driver
// -------------------------------------

const DocumentName = "doc"

type ObjectFactory func() interface{}

type Collection struct {
	Connection        *Connection
	TableName         string
	OrganizationIdKey string
	AllocateRecord    ObjectFactory
}

func (c *Collection) Initialize(ctx context.Context) error {
	exists, err := c.Connection.Database.CollectionExists(ctx, c.TableName)
	if err != nil {
		return err
	}
	if !exists {
		_, err := c.Connection.Database.CreateCollection(ctx, c.TableName, nil)
		if err != nil {
			return err
		}
	}

	if c.OrganizationIdKey == "" {
		c.OrganizationIdKey = "organization_id"
	}

	return nil
}

func (c *Collection) Drop(ctx context.Context) error {
	col, err := c.Connection.Database.Collection(ctx, c.TableName)
	if err != nil {
		if IsNotFound(err) {
			return nil
		}
		return err
	}

	return col.Remove(ctx)
}

// users of this api must supply an id with their objects
func getId(doc map[string]interface{}) (string, error) {
	if id, exists := doc["id"]; !exists {
		return "", errors.New("document doesn't have an id")
	} else if convertedId, ok := id.(string); !ok {
		return "", errors.New("document must have a string id")
	} else if convertedId == "" {
		return "", errors.New("document must have an actual id")
	} else {
		return convertedId, nil
	}
}

func (c *Collection) Increment(ctx context.Context, obj interface{}, varName string) error {
	// get the object details
	doc, err := encoding.ObjectToMap(obj)
	if err != nil {
		return err
	}

	id, err := getId(doc)
	if err != nil {
		return err
	}

	organizationId, ok := doc[c.OrganizationIdKey]
	if !ok {
		return errors.New("must have organization_id in record")
	}

	// build query
	query := `FOR d IN @@collection
  FILTER d._key == @key && d.organization_id == @org_id
  UPDATE d WITH { ` + varName + `: d.` + varName + ` + 1 } IN @@collection
	`
	cursor, err := c.Connection.Database.Query(ctx, query, map[string]interface{}{
		"@collection": c.TableName,
		"key":         id,
		"org_id":      organizationId,
	})

	if err != nil {
		return err
	}

	defer cursor.Close()

	return nil
}

func (c *Collection) Get(ctx context.Context, id string) (interface{}, error) {
	// get the collection info
	collection, err := c.Connection.Database.Collection(ctx, c.TableName)
	if err != nil {
		return nil, err
	}

	// read the doc
	return ReadDoc(c.AllocateRecord, func(doc map[string]interface{}) error {
		_, err := collection.ReadDocument(ctx, id, &doc)
		return err
	})
}

func (c *Collection) Update(ctx context.Context, obj interface{}) error {
	// get the object ready to update
	doc, err := encoding.ObjectToMap(obj)
	if err != nil {
		return err
	}

	id, err := getId(doc)
	if err != nil {
		return err
	}

	delete(doc, "id") // don't store the id in the database record

	collection, err := c.Connection.Database.Collection(ctx, c.TableName)
	if err != nil {
		return err
	}

	// store it
	fmt.Printf("sending in update for database=%s, table=%s, id=%s, doc=%+v\n", c.Connection.Database.Name(), c.TableName, id, doc)
	meta, err := collection.UpdateDocument(ctx, id, doc)
	fmt.Printf("meta=%+v, err=%s\n", meta, err)
	if err != nil {
		return err
	}

	return nil
}

func (c *Collection) Create(ctx context.Context, obj interface{}) (string, error) {
	// get the object ready to update
	doc, err := encoding.ObjectToMap(obj)
	if err != nil {
		return "", err
	}

	doc["_key"] = uuid.NewString() // convert id to a key for arango's meta key
	delete(doc, "id")              // don't store the id in the database record

	collection, err := c.Connection.Database.Collection(ctx, c.TableName)
	if err != nil {
		return "", err
	}

	// store it
	meta, err := collection.CreateDocument(ctx, doc)
	if err != nil {
		return "", nil
	}

	return meta.Key, nil
}

func (c *Collection) Delete(ctx context.Context, obj interface{}) error {
	// get the object ready to update
	doc, err := encoding.ObjectToMap(obj)
	if err != nil {
		return err
	}

	id, err := getId(doc)
	if err != nil {
		return err
	}

	collection, err := c.Connection.Database.Collection(ctx, c.TableName)
	if err != nil {
		return err
	}

	// un-store it
	k, err := collection.RemoveDocument(ctx, id)
	if err != nil {
		return err
	}

	if k.Key != id {
		return fmt.Errorf("while attempting to remove %s with id=%s, key=%s was returned", c.TableName, id, k.Key)
	}

	return nil
}

func (c *Collection) Query() *CollectionFilter {
	f := CollectionFilter{
		expressions:     make([]interface{}, 0),
		collection:      c,
		variableFactory: NewVariableFactory(),
	}

	return &f
}

// ----------------
// CollectionFilter
// ----------------

type CollectionFilter struct {
	collection      *Collection
	expressions     []interface{}
	variableFactory *VariableFactory
}

func (c *CollectionFilter) Operator() *Operator {
	return &Operator{
		variableFactory: c.variableFactory,
	}
}

type ExpressionWrapper struct {
	item interface{}
}

func (c *ExpressionWrapper) String() string {
	return fmt.Sprintf("%s", c.item)
}

// ----------------------
// Expressions
// ----------------------

type Expression interface {
	String() string
}

func (c *CollectionFilter) Where(expression Expression) *CollectionFilter {
	c.expressions = append(c.expressions, expression)

	return c
}

func (c *CollectionFilter) Filter(key string, value interface{}) *CollectionFilter {
	return c.Where(&EqualityExpression{
		left:     &DocumentAttribute{name: key},
		operator: EqualityExpressionEqual,
		right:    c.variableFactory.MakeVariable(value),
	})
}

func (c *CollectionFilter) formatExpressions() string {
	var buffer bytes.Buffer
	for _, expression := range c.expressions {
		buffer.WriteString(fmt.Sprintf("FILTER %s\n", expression))
	}

	return buffer.String()
}

func (c *CollectionFilter) InArrayOfDocuments(value string, arrayName string, documentKey string) *CollectionFilter {
	return c.InArray(value, fmt.Sprintf("%s[*].%s", arrayName, documentKey))
}

func (c *CollectionFilter) InArray(value string, arrayName string) *CollectionFilter {
	return c.Where(&InArrayExpression{
		value:     c.variableFactory.MakeVariable(value),
		arrayName: NewAttribute(arrayName),
	})
}

// ---------------------
// Commonly used filters/expressions
// ---------------------

func (c *CollectionFilter) WithinOrg(orgId string) *CollectionFilter {
	return c.Where(&EqualityExpression{
		left:     &DocumentAttribute{name: c.collection.OrganizationIdKey},
		operator: EqualityExpressionEqual,
		right:    c.variableFactory.MakeVariable(orgId),
	})
}

func (c *CollectionFilter) ById(docId string) *CollectionFilter {
	return c.Where(&EqualityExpression{
		left:     &DocumentAttribute{name: "_key"},
		operator: EqualityExpressionEqual,
		right:    c.variableFactory.MakeVariable(docId),
	})
}

// ----------------
// These functions end the chaining with a result
// ----------------

func (c *CollectionFilter) First(ctx context.Context) (interface{}, error) {
	matches, err := c.List().Limit(1).All(ctx)
	if err != nil {
		return nil, err
	}

	if len(matches) < 1 {
		return nil, nil
	}

	return matches[0], nil
}

func (c *CollectionFilter) Count(ctx context.Context) (int, error) {
	query := fmt.Sprintf(`
FOR doc IN @@collection
 %s
 COLLECT WITH COUNT INTO length
    RETURN length`, c.formatExpressions())

	variables := c.variableFactory.SymbolTable()
	variables["@collection"] = c.collection.TableName
	log.Info("ORM Count ", log.Fields{"query": query, "filters": variables})

	cursor, err := c.collection.Connection.Database.Query(ctx, query, variables)

	if err != nil {
		return -1, err
	}

	defer cursor.Close()

	// Declare a variable to hold the document
	var length int

	_, err = cursor.ReadDocument(ctx, &length)
	if err != nil {
		return -1, err
	}

	return length, nil
}

// ----------------
// These functions end the chaining with a modifying operation
// ----------------

func (c *CollectionFilter) DeleteAll(ctx context.Context) ([]string, error) {
	query := fmt.Sprintf(`
FOR doc IN @@collection
 %s
REMOVE doc IN @@collection
LET removed = OLD
 RETURN removed._key`, c.formatExpressions())

	variables := c.variableFactory.SymbolTable()
	variables["@collection"] = c.collection.TableName

	log.Info("ORM ", log.Fields{"query": query, "filters": variables})

	cursor, err := c.collection.Connection.Database.Query(ctx, query, variables)

	if err != nil {
		return nil, err
	}

	defer cursor.Close()

	// read the ids
	ids := make([]string, 0)
	for cursor.HasMore() {
		var removedId string
		_, err := cursor.ReadDocument(ctx, &removedId)
		if err != nil {
			return nil, err
		}
		ids = append(ids, removedId)
	}

	return ids, nil
}

func (c *CollectionFilter) formatUpdates(updates map[string]interface{}) string {
	var buffer bytes.Buffer
	buffer.WriteString("{")
	for k, v := range updates {
		buffer.WriteString(fmt.Sprintf("%s:%s", k, c.variableFactory.MakeVariable(v)))
	}
	buffer.WriteString("}")

	return buffer.String()
}

func (c *CollectionFilter) UpdateAll(ctx context.Context, updates map[string]interface{}) ([]string, error) {
	query := fmt.Sprintf(`
FOR doc IN @@collection
 %s
 UPDATE doc with %s in @@collection
 RETURN doc._key`, c.formatExpressions(), c.formatUpdates(updates))

	variables := c.variableFactory.SymbolTable()
	variables["@collection"] = c.collection.TableName

	log.Info("ORM ", log.Fields{"query": query, "filters": variables})

	cursor, err := c.collection.Connection.Database.Query(ctx, query, variables)

	if err != nil {
		return nil, err
	}

	defer cursor.Close()

	// read the ids
	ids := make([]string, 0)
	for cursor.HasMore() {
		var modifiedId string
		_, err := cursor.ReadDocument(ctx, &modifiedId)
		if err != nil {
			return nil, err
		}
		ids = append(ids, modifiedId)
	}

	return ids, nil
}

// ----------------
// These function end the chain with read only types of options
// ----------------

func (c *CollectionFilter) List() *ItemsOperator {
	return &ItemsOperator{
		collectionFilter: c,
	}
}

func IsNotFound(err error) bool {
	return driver.IsNotFound(err) || driver.IsNoMoreDocuments(err)
}
