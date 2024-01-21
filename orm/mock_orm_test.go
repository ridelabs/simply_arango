package orm

import (
	"context"
	"github.com/houqp/gtest"
	"testing"

	"github.com/ridelabs/simply_arango/utils"
	"github.com/stretchr/testify/assert"
)

type MyDoc struct {
	Name           string `json:"name"`
	B              string `json:"b"`
	C              string `json:"c"`
	D              string `json:"d"`
	Id             string `json:"id"`
	Counter        int    `json:"counter"`
	OrganizationId string `json:"organization_id"`
}

type OrmTests struct {
	collection *Collection
	database   *utils.MockDatabase
}

func (s *OrmTests) Setup(t *testing.T) {}

func (s *OrmTests) Teardown(t *testing.T) {}

func (s *OrmTests) BeforeEach(t *testing.T) {
	myCursor := &utils.MockCursor{
		Items: []string{
			`{"name": "Suzie Q", "b":"Bank", "c": "Cash", "d": "dollars", "_key": "11"}`,
			`{"name": "Bill", "b":"Blank", "c": "Check", "d": "dollars", "_key": "22" }`,
			`{"name": "Hank", "b":"blanket", "c": "cleaners", "d": "delighted", "_key": "33"}`,
		},
		Index: int64(0),
	}
	s.database = &utils.MockDatabase{
		MyCursor: myCursor,
	}

	s.collection = &Collection{
		Connection: &Connection{
			Database: s.database,
			Client:   nil,
		},

		TableName:         "foo",
		OrganizationIdKey: "organization_id",
		AllocateRecord: func() interface{} {
			return &MyDoc{}
		},
	}
}

func (s *OrmTests) AfterEach(t *testing.T) {}

func (s *OrmTests) assertBasicMockRecords(t *testing.T, objects []interface{}) {
	// check out the decoding of the docs
	assert.Equal(t, 3, len(objects))

	o0, ok := objects[0].(*MyDoc)
	assert.True(t, ok)
	assert.Equal(t, "Suzie Q", o0.Name)
	assert.Equal(t, "Bank", o0.B)
	assert.Equal(t, "Cash", o0.C)
	assert.Equal(t, "dollars", o0.D)
	assert.Equal(t, "11", o0.Id)

	o1, ok := objects[1].(*MyDoc)
	assert.True(t, ok)
	assert.Equal(t, "Bill", o1.Name)
	assert.Equal(t, "Blank", o1.B)
	assert.Equal(t, "Check", o1.C)
	assert.Equal(t, "dollars", o1.D)
	assert.Equal(t, "22", o1.Id)

	o2, ok := objects[2].(*MyDoc)
	assert.True(t, ok)
	assert.Equal(t, "Hank", o2.Name)
	assert.Equal(t, "blanket", o2.B)
	assert.Equal(t, "cleaners", o2.C)
	assert.Equal(t, "delighted", o2.D)
	assert.Equal(t, "33", o2.Id)
}

func (s *OrmTests) SubTestQueryMultipleFilter(t *testing.T) {
	objects, err := s.collection.Query().WithinOrg("8675309").ById("33333").Filter("a", "apple").
		Filter("b", "bravo").Filter("c", "charlie").Filter("xxx", "charlie").List().OrderBy("name").Desc().
		Paging(40, 4).All(context.TODO())

	assert.Nil(t, err)

	s.assertBasicMockRecords(t, objects)

	// check that the arango query was built correctly
	q := utils.StripExtraWS(s.database.LastQuery)
	assert.Equal(t, "FOR doc IN @@collection "+
		"FILTER (doc.organization_id == @var_0) "+
		"FILTER (doc._key == @var_1) "+
		"FILTER (doc.a == @var_2) "+
		"FILTER (doc.b == @var_3) "+
		"FILTER (doc.c == @var_4) "+
		"FILTER (doc.xxx == @var_4) "+
		"SORT doc.name DESC "+
		"LIMIT @var_5, @var_6 "+
		"RETURN doc", q)

	// check that the query variables were built correctly
	assert.Equal(t, map[string]interface{}{
		"@collection": "foo",
		"var_0":       "8675309",
		"var_1":       "33333",
		"var_2":       "apple",
		"var_3":       "bravo",
		"var_4":       "charlie",
		"var_5":       160,
		"var_6":       40,
	}, s.database.LastBindVars)

}

func (s *OrmTests) SubTestQueryRand(t *testing.T) {
	objects, err := s.collection.Query().WithinOrg("8675309").Filter("a", "apple").
		Filter("a", "dice").Filter("b", "bravo").Filter("c", "charlie").List().RandomOrder().
		Limit(2).All(context.TODO())

	assert.Nil(t, err)

	s.assertBasicMockRecords(t, objects)

	// check that the arango query was built correctly
	q := utils.StripExtraWS(s.database.LastQuery)
	assert.Equal(t, "FOR doc IN @@collection "+
		"FILTER (doc.organization_id == @var_0) "+
		"FILTER (doc.a == @var_1) "+
		"FILTER (doc.a == @var_2) "+
		"FILTER (doc.b == @var_3) "+
		"FILTER (doc.c == @var_4) "+
		"SORT RAND() "+
		"LIMIT @var_5 "+
		"RETURN doc", q)

	// check that the query variables were built correctly
	assert.Equal(t, map[string]interface{}{
		"@collection": "foo",
		"var_0":       "8675309",
		"var_1":       "apple",
		"var_2":       "dice",
		"var_3":       "bravo",
		"var_4":       "charlie",
		"var_5":       2,
	}, s.database.LastBindVars)
}

func (s *OrmTests) SubTestQueryLimit(t *testing.T) {
	objects, err := s.collection.Query().WithinOrg("8675309").Filter("a", "apple").
		Filter("a", "dice").Filter("b", "bravo").Filter("c", "charlie").List().OrderBy("name").Asc().
		Limit(2).All(context.TODO())

	assert.Nil(t, err)

	s.assertBasicMockRecords(t, objects)

	// check that the arango query was built correctly
	q := utils.StripExtraWS(s.database.LastQuery)
	assert.Equal(t, "FOR doc IN @@collection "+
		"FILTER (doc.organization_id == @var_0) "+
		"FILTER (doc.a == @var_1) "+
		"FILTER (doc.a == @var_2) "+
		"FILTER (doc.b == @var_3) "+
		"FILTER (doc.c == @var_4) "+
		"SORT doc.name ASC "+
		"LIMIT @var_5 "+
		"RETURN doc", q)

	// check that the query variables were built correctly
	assert.Equal(t, map[string]interface{}{
		"@collection": "foo",
		"var_0":       "8675309",
		"var_1":       "apple",
		"var_2":       "dice",
		"var_3":       "bravo",
		"var_4":       "charlie",
		"var_5":       2,
	}, s.database.LastBindVars)
}

func (s *OrmTests) SubTestIncrement(t *testing.T) {
	err := s.collection.Increment(context.TODO(), &MyDoc{Name: "obiwan", Counter: 0, OrganizationId: "1138", Id: "1112"}, "counter")
	assert.Nil(t, err)

	q := utils.StripExtraWS(s.database.LastQuery)
	assert.Equal(t, "FOR d IN @@collection FILTER d._key == @key && d.organization_id == @org_id "+
		"UPDATE d WITH { counter: d.counter + 1 } IN @@collection", q)
}

func (s *OrmTests) SubTestQueryFirst(t *testing.T) {
	object, err := s.collection.Query().WithinOrg("90210").Filter("alpha", "A").
		Filter("B", "beta").First(context.TODO())

	assert.Nil(t, err, "Should not have gotten an error")

	q := utils.StripExtraWS(s.database.LastQuery)
	assert.Equal(t, "FOR doc IN @@collection "+
		"FILTER (doc.organization_id == @var_0) "+
		"FILTER (doc.alpha == @var_1) "+
		"FILTER (doc.B == @var_2) "+
		"LIMIT @var_3 "+
		"RETURN doc", q)

	assert.Equal(t, map[string]interface{}{
		"@collection": "foo",
		"var_0":       "90210",
		"var_1":       "A",
		"var_2":       "beta",
		"var_3":       1,
	}, s.database.LastBindVars)

	record, ok := object.(*MyDoc)
	assert.True(t, ok)

	assert.Equal(t, "Suzie Q", record.Name)
	assert.Equal(t, "11", record.Id)
}

func (s *OrmTests) SubTestQueryMultipleFilterSubstr(t *testing.T) {
	objects, err := s.collection.Query().WithinOrg("8675309").
		Filter("a", "apple").
		Filter("b", "bravo").
		Filter("c", "charlie").
		List().OrderBy("name").Asc().
		Paging(10, 3).All(context.TODO())

	assert.Nil(t, err)
	s.assertBasicMockRecords(t, objects)

	// check that the arango query was built correctly
	q := utils.StripExtraWS(s.database.LastQuery)
	assert.Equal(t, "FOR doc IN @@collection "+
		"FILTER (doc.organization_id == @var_0) "+
		"FILTER (doc.a == @var_1) "+
		"FILTER (doc.b == @var_2) "+
		"FILTER (doc.c == @var_3) "+
		"SORT doc.name ASC "+
		"LIMIT @var_4, @var_5 "+
		"RETURN doc", q)

	// check that the query variables were built correctly
	assert.Equal(t, map[string]interface{}{
		"@collection": "foo",
		"var_0":       "8675309",
		"var_1":       "apple",
		"var_2":       "bravo",
		"var_3":       "charlie",
		"var_4":       30,
		"var_5":       10,
	}, s.database.LastBindVars)
}

func (s *OrmTests) SubTestQueryExpressions1(t *testing.T) {
	q := s.collection.Query()
	o := q.Operator()
	objects, err := q.WithinOrg("8675309").
		Where(o.Not("apple")).
		Where(o.Not(RawValue("foobar"))).
		Where(
			o.And(
				o.Or(
					o.StartsWith("a", "apple"),
					o.Equal("a", "APPLE"),
				),
				o.Equal("a", "apple"),
			),
		).List().Paging(20, 0).All(context.TODO())

	assert.Nil(t, err)
	s.assertBasicMockRecords(t, objects)

	// check that the arango query was built correctly
	query := utils.StripExtraWS(s.database.LastQuery)
	assert.Equal(t, "FOR doc IN @@collection FILTER (doc.organization_id == @var_0) "+
		"FILTER (NOT @var_1) "+
		"FILTER (NOT \"foobar\") "+
		"FILTER ((doc.a LIKE @var_2 || (doc.a == @var_3) ) && (doc.a == @var_1) ) "+
		"LIMIT @var_4, @var_5 "+
		"RETURN doc", query)

	// check that the query variables were built correctly
	assert.Equal(t, map[string]interface{}{
		"@collection": "foo",
		"var_0":       "8675309",
		"var_1":       "apple",
		"var_2":       "apple%",
		"var_3":       "APPLE",
		"var_4":       0,
		"var_5":       20,
	}, s.database.LastBindVars)
}

func (s *OrmTests) SubTestQuery_InArray(t *testing.T) {
	q := s.collection.Query()
	objects, err := q.WithinOrg("8675309").InArrayOfDocuments("apple", "fruits", "my_fruit_name").List().All(context.TODO())

	assert.Nil(t, err)
	s.assertBasicMockRecords(t, objects)

	query := utils.StripExtraWS(s.database.LastQuery)

	// check the query
	assert.Equal(t, "FOR doc IN @@collection FILTER (doc.organization_id == @var_0) FILTER @var_1 IN doc.fruits[*].my_fruit_name RETURN doc", query)

	// check that the query variables were built correctly
	assert.Equal(t, map[string]interface{}{
		"@collection": "foo",
		"var_0":       "8675309",
		"var_1":       "apple",
	}, s.database.LastBindVars)
}

func (s *OrmTests) SubTestQueryInDocArray(t *testing.T) {
	objects, err := s.collection.Query().WithinOrg("8675309").InArray("apple", "fruits").List().All(context.TODO())
	assert.Nil(t, err)
	s.assertBasicMockRecords(t, objects)

	query := utils.StripExtraWS(s.database.LastQuery)

	// check the query
	assert.Equal(t, "FOR doc IN @@collection FILTER (doc.organization_id == @var_0) FILTER @var_1 IN doc.fruits RETURN doc", query)

	// check that the query variables were built correctly
	assert.Equal(t, map[string]interface{}{
		"@collection": "foo",
		"var_0":       "8675309",
		"var_1":       "apple",
	}, s.database.LastBindVars)
}

// ------------------------------
// Entry point for test suite
// ------------------------------

func TestOrmByMocks(t *testing.T) {
	gtest.RunSubTests(t, &OrmTests{})
}

// ------------------------------
