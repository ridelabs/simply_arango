package orm

import (
	"context"
	"errors"
	"fmt"
	"github.com/ridelabs/simply_arango/encoding"
	log "github.com/sirupsen/logrus"
)

// ------------------
// Order By
// ------------------

type Order interface {
	OrderFormat() string
}

type OrderBy struct {
	items     *ItemsOperator
	key       string
	direction string
}

func (c *OrderBy) Desc() *ItemsOperator {
	c.direction = "DESC"
	return c.items
}

func (c *OrderBy) Asc() *ItemsOperator {
	c.direction = "ASC"
	return c.items
}

func (c *OrderBy) OrderFormat() string {
	return fmt.Sprintf("SORT %s.%s %s ", DocumentName, c.key, c.direction)
}

type Rand struct{}

func (c *Rand) OrderFormat() string {
	return "SORT RAND()"
}

// ------------------
// Items
// ------------------

type ItemsOperator struct {
	collectionFilter *CollectionFilter

	limit   Variable
	orderBy Order
	paging  *Paging
}

func (c *ItemsOperator) OrderBy(key string) *OrderBy {
	if c.orderBy != nil {
		log.Warn("RandomOrder: dropping old order for these items")
	}
	o := &OrderBy{items: c, key: key}
	c.orderBy = o
	return o
}

func (c *ItemsOperator) RandomOrder() *ItemsOperator {
	if c.orderBy != nil {
		log.Warn("RandomOrder: dropping old order for these items")
	}
	c.orderBy = &Rand{}
	return c
}

func (c *ItemsOperator) formatOrder() string {
	if c.orderBy != nil {
		return c.orderBy.OrderFormat()
	}
	return ""
}

// --------------------
// paging and limit
// --------------------

func (c *ItemsOperator) Paging(pageSize, page int) *ItemsOperator {
	if pageSize > 0 && page > -1 {
		c.paging = &Paging{
			OffsetVar: c.collectionFilter.variableFactory.MakeVariable(page * pageSize),
			CountVar:  c.collectionFilter.variableFactory.MakeVariable(pageSize),
		}
	}

	return c
}

type Paging struct {
	OffsetVar Variable
	CountVar  Variable
}

func (c *Paging) FormatPaging() string {
	return fmt.Sprintf("LIMIT %s, %s", c.OffsetVar, c.CountVar)
}

func (c *ItemsOperator) Limit(count int) *ItemsOperator {
	c.limit = c.collectionFilter.variableFactory.MakeVariable(count)
	return c
}

func (c *ItemsOperator) formatLimitOrPaging() string {
	if c.paging != nil {
		return c.paging.FormatPaging()
	} else if c.limit != nil {
		return fmt.Sprintf("LIMIT %s", c.limit)
	}

	return ""
}

func (c *ItemsOperator) First(ctx context.Context) (interface{}, error) {
	matches, err := c.Limit(1).All(ctx)
	if err != nil {
		return nil, err
	}

	if len(matches) < 1 {
		return nil, nil
	}

	return matches[0], nil
}

func (c *ItemsOperator) All(ctx context.Context) ([]interface{}, error) {
	query := fmt.Sprintf(`
FOR doc IN @@collection
 %s
 %s
 %s
 RETURN doc`, c.collectionFilter.formatExpressions(), c.formatOrder(), c.formatLimitOrPaging())

	variables := c.collectionFilter.variableFactory.SymbolTable()
	variables["@collection"] = c.collectionFilter.collection.TableName

	log.Info("ORM ", log.Fields{"query": query, "filters": variables})

	cursor, err := c.collectionFilter.collection.Connection.Database.Query(ctx, query, variables)

	if err != nil {
		return nil, err
	}

	defer cursor.Close()

	// read data
	items := make([]interface{}, 0)
	for cursor.HasMore() {
		obj, err := ReadDoc(c.collectionFilter.collection.AllocateRecord, func(doc map[string]interface{}) error {
			_, err := cursor.ReadDocument(ctx, &doc)
			return err
		})
		if err != nil {
			return nil, err
		}
		items = append(items, obj)
	}

	return items, nil
}

type Reader func(map[string]interface{}) error

func ReadDoc(objFactory ObjectFactory, reader Reader) (interface{}, error) {
	doc := make(map[string]interface{})

	if err := reader(doc); err != nil {
		return nil, err
	}

	if id, exists := doc["_key"]; !exists {
		return nil, errors.New("arango db record should have had an _key attribute, but didn't")
	} else {
		doc["id"] = id // convert _key to id
		delete(doc, "_key")
		delete(doc, "_id")
	}

	obj := objFactory()
	if err := encoding.MapToObject(doc, obj); err != nil {
		return nil, err
	}

	return obj, nil
}
