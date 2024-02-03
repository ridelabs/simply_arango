package orm

import (
	"context"
	"fmt"
	"github.com/houqp/gtest"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"reflect"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
	"os"
	"testing"
)

type RealOrmTests struct {
	collection *Collection
	conn       *Connection
}

type DeepFruit struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type TestDocument struct {
	Id             string `json:"id"`
	OrganizationId string `json:"organization_id"`

	A          string       `json:"a"`
	B          string       `json:"b"`
	C          string       `json:"c"`
	Name       string       `json:"name"`
	Email      string       `json:"email"`
	Counter    int          `json:"counter"`
	Fruits     []string     `json:"fruits"`
	DeepFruits []*DeepFruit `json:"deep_fruits"`
	Birthday1  string       `json:"birthday_1"`
	Birthday2  string       `json:"birthday_2,omitempty"`
}

func (s *RealOrmTests) Setup(t *testing.T) {
	// To make these tests work, you must have a .env file in this directory with the below variables
	repoPath := os.Getenv("PWD") + "/../.env"
	if err := godotenv.Load(repoPath); err != nil {
		t.Fatal("Failed to load env file")
	}
	dbUser := os.Getenv("ARANGODB_USER")
	dbPass := os.Getenv("ARANGODB_PASS")
	dbUrl := os.Getenv("ARANGODB_URL")
	dbName := os.Getenv("TEST_DB_NAME")

	ctx := context.TODO()

	conn, err := NewConnection(ctx, dbName, dbUser, dbPass, dbUrl)
	if err != nil {
		log.Error("Failed to connect to arangodb", log.Fields{"err": err})
		os.Exit(5)
	}

	s.conn = conn
}

func (s *RealOrmTests) Teardown(t *testing.T) {}

func (s *RealOrmTests) BeforeEach(t *testing.T) {
	s.collection = &Collection{
		Connection:        s.conn,
		TableName:         "foo",
		OrganizationIdKey: "organization_id",
		AllocateRecord: func() interface{} {
			return &TestDocument{}
		},
	}
	ctx := context.TODO()

	err := s.collection.Drop(ctx)
	assert.NoError(t, err, "Should either drop or already be gone")

	err = s.collection.Initialize(ctx)
	assert.NoError(t, err, "Should have been created")
}

func (s *RealOrmTests) AfterEach(t *testing.T) {}

func (s *RealOrmTests) createDocuments(t *testing.T) int {
	ctx := context.TODO()
	counter := 0
	// Create 16 documents
	for _, orgId := range []string{"121212", "232323"} {
		for _, fruit := range []string{"apple", "pear"} {
			for _, name := range []string{"Bob", "James", "Jeremy", "Jeren"} {
				id, err := s.collection.Create(ctx, &TestDocument{
					Name:           name,
					A:              "alpaca",
					B:              "bear",
					C:              string(rune(97 + counter)),
					Email:          strings.ToLower(name) + "@abc.com",
					Counter:        0,
					OrganizationId: orgId,
					Fruits: []string{
						"cherry", fruit, "persimmon",
					},
					DeepFruits: []*DeepFruit{
						{
							Id:   "111",
							Name: "cherry",
						},
						{
							Id:   "222",
							Name: fruit,
						},
						{
							Id:   "333",
							Name: "persimmon",
						},
					},
				})
				assert.Nil(t, err)
				assert.NotEmptyf(t, id, "ID shouldn't have been empty")
				counter++
			}
		}
	}
	return counter
}

func (s *RealOrmTests) SubTestNullOrEmptyCheck(t *testing.T) {
	// Note: these tests rely on TestDocument's Birthday1 and Birthday2 with the appropriate non-omitempty and omitempty encoding modifiers
	ctx := context.TODO()
	myOrg := "777"
	_, err := s.collection.Create(ctx, &TestDocument{
		Name:           "alfred",
		A:              "alpaca",
		B:              "bear",
		OrganizationId: myOrg,
		Birthday1:      "2024-02-03T11:27:31−07:00",
		Birthday2:      "2024-02-03T11:27:31−07:00",
	})
	assert.Nil(t, err)

	_, err = s.collection.Create(ctx, &TestDocument{
		Name:           "alfred",
		A:              "alpaca",
		B:              "bear",
		OrganizationId: myOrg,
		Birthday1:      "",
		Birthday2:      "",
	})
	assert.Nil(t, err)

	_, err = s.collection.Create(ctx, &TestDocument{
		Name:           "alfred",
		A:              "alpaca",
		B:              "bear",
		OrganizationId: myOrg,
	})
	assert.Nil(t, err)

	q := s.collection.Query()
	o := q.Operator()

	// all documents in our org
	matches, err := q.WithinOrg(myOrg).List().All(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(matches), "expected 3 matches")

	// filter by empty and null values

	// all 3 should be not null (not omitempty)
	q = s.collection.Query()
	o = q.Operator()
	matches, err = q.WithinOrg(myOrg).Where(o.IsNotNull("birthday_1")).List().All(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(matches))

	// 1 should be not null (omitempty causes the other 2 to be null)
	q = s.collection.Query()
	o = q.Operator()
	matches, err = q.WithinOrg(myOrg).Where(o.IsNotNull("birthday_2")).List().All(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(matches))

	// no nulls for the non omitempty
	q = s.collection.Query()
	o = q.Operator()
	matches, err = q.WithinOrg(myOrg).Where(o.IsNull("birthday_1")).List().All(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(matches))

	// 2 nulls for the omitempty
	q = s.collection.Query()
	o = q.Operator()
	matches, err = q.WithinOrg(myOrg).Where(o.IsNull("birthday_2")).List().All(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(matches))

	// for non omitempty, we should find 1, the others were converted to ""
	q = s.collection.Query()
	o = q.Operator()
	matches, err = q.WithinOrg(myOrg).Where(o.IsNotEmpty("birthday_1")).List().All(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(matches))

	// for omitempty, we should find that 0 are ""
	q = s.collection.Query()
	o = q.Operator()
	matches, err = q.WithinOrg(myOrg).Where(o.IsNotEmpty("birthday_2")).List().All(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(matches))

	// for non omitempty, we should find 2 that are empty
	q = s.collection.Query()
	o = q.Operator()
	matches, err = q.WithinOrg(myOrg).Where(o.IsEmpty("birthday_1")).List().All(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(matches))

	// for omitempty, we should find 0 that are empty
	q = s.collection.Query()
	o = q.Operator()
	matches, err = q.WithinOrg(myOrg).Where(o.IsEmpty("birthday_2")).List().All(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(matches))

	// for both omitempty and non-omitempty, we should only get 1
	q = s.collection.Query()
	o = q.Operator()
	matches, err = q.WithinOrg(myOrg).Where(
		o.And(
			o.Not(o.IsEmpty("birthday_1")),
			o.Not(o.IsNull("birthday_1")),
		),
	).List().All(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(matches))
	//
	q = s.collection.Query()
	o = q.Operator()
	matches, err = q.WithinOrg(myOrg).Where(
		o.And(
			o.Not(o.IsEmpty("birthday_2")),
			o.Not(o.IsNull("birthday_2")),
		),
	).List().All(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(matches))

}

func (s *RealOrmTests) SubTestBasicORM(t *testing.T) {
	ctx := context.TODO()

	obj1 := &TestDocument{
		Id:             "111111",
		A:              "1apple",
		B:              "1bacon",
		C:              "1chewy",
		Name:           "1no name",
		Counter:        0,
		OrganizationId: "90210",
	}
	obj2 := &TestDocument{
		Id:             "211111",
		A:              "2apple",
		B:              "2bacon",
		C:              "2chewy",
		Name:           "2no name",
		Counter:        0,
		OrganizationId: "90210",
	}
	obj3 := &TestDocument{
		Id:             "3111111",
		A:              "3apple",
		B:              "3bacon",
		C:              "3chewy",
		Name:           "3no name",
		Counter:        0,
		OrganizationId: "90210",
	}

	// This orm can Create
	id1, err := s.collection.Create(ctx, obj1)
	assert.Nil(t, err)
	assert.NotNil(t, id1)

	id2, err := s.collection.Create(ctx, obj2)
	assert.Nil(t, err)
	assert.NotNil(t, id2)

	id3, err := s.collection.Create(ctx, obj3)
	assert.Nil(t, err)
	assert.NotNil(t, id3)

	// We can Get by id only
	newObj1, err := s.collection.Get(ctx, id1)
	assert.Nil(t, err)
	foundObj1, ok := newObj1.(*TestDocument)
	assert.True(t, ok)
	assert.Equal(t, id1, foundObj1.Id)
	assert.Equal(t, obj1.A, foundObj1.A)
	assert.Equal(t, obj1.B, foundObj1.B)
	assert.Equal(t, obj1.C, foundObj1.C)
	assert.Equal(t, obj1.Name, foundObj1.Name)
	assert.Equal(t, 0, foundObj1.Counter)

	newObj2, err := s.collection.Get(ctx, id2)
	assert.Nil(t, err)
	foundObj2, ok := newObj2.(*TestDocument)
	assert.True(t, ok)
	assert.Equal(t, id2, foundObj2.Id)
	assert.Equal(t, obj2.A, foundObj2.A)
	assert.Equal(t, obj2.B, foundObj2.B)
	assert.Equal(t, obj2.C, foundObj2.C)
	assert.Equal(t, obj2.Name, foundObj2.Name)
	assert.Equal(t, 0, foundObj2.Counter)

	newObj3, err := s.collection.Get(ctx, id3)
	assert.Nil(t, err)
	foundObj3, ok := newObj3.(*TestDocument)
	assert.True(t, ok)
	assert.Equal(t, id3, foundObj3.Id)
	assert.Equal(t, obj3.A, foundObj3.A)
	assert.Equal(t, obj3.B, foundObj3.B)
	assert.Equal(t, obj3.C, foundObj3.C)
	assert.Equal(t, obj3.Name, foundObj3.Name)
	assert.Equal(t, 0, foundObj3.Counter)

	// You can update an object
	foundObj3.Name = "Fred Flintstone"
	err = s.collection.Update(ctx, foundObj3)
	assert.Nil(t, err)
	newObj3, err = s.collection.Get(ctx, id3)
	assert.Nil(t, err)
	foundObj3v2, ok := newObj3.(*TestDocument)
	assert.True(t, ok)
	assert.Equal(t, id3, foundObj3v2.Id)
	assert.Equal(t, "Fred Flintstone", foundObj3v2.Name)

	// You can increment an attribute in a record
	for i := 0; i < 10; i++ {
		err = s.collection.Increment(ctx, foundObj3, "counter")
		newObj3, err = s.collection.Get(ctx, id3)
		assert.Nil(t, err)
		foundObj3v3, ok := newObj3.(*TestDocument)
		assert.True(t, ok)
		assert.Equal(t, i+1, foundObj3v3.Counter)
	}

	// You can delete them
	err = s.collection.Delete(ctx, foundObj3v2)
	assert.Nil(t, err)
	theRecordThatShouldBeGone, err := s.collection.Get(ctx, id3)
	assert.True(t, IsNotFound(err))
	assert.Nil(t, theRecordThatShouldBeGone)
}

func (s *RealOrmTests) SubTestBasicQuery(t *testing.T) {
	ctx := context.TODO()

	q := s.collection.Query()
	count, err := q.Count(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 0, count)

	counter := s.createDocuments(t)
	assert.Equal(t, 16, counter)

	// Test counting them
	count, err = q.Count(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 16, count)

	// further restrict the query to just one org
	q.WithinOrg("232323")
	count, err = q.Count(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 8, count)

	// You can chain create query filters and then perform operations on them
	o := q.Operator()
	q.InArray("apple", "fruits").
		Where(
			o.Or(
				o.Equal("email", "bob@abc.com"),
				o.And(
					o.EndsWith("name", "emy"),
					o.StartsWith("email", "jer"),
				),
			),
		)
	count, err = q.Count(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 2, count)

	// now get it down to just bob
	q.Where(o.Not(o.Equal("name", "Jeremy")))
	count, err = q.Count(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, count)

	q.InArrayOfDocuments("apple", "deep_fruits", "name")

	// now check out the record
	objects, err := q.List().All(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(objects))
	jObject, ok := objects[0].(*TestDocument)
	assert.True(t, ok)
	assert.Equal(t, "bob@abc.com", jObject.Email)
}

func (s *RealOrmTests) SubTestDelete(t *testing.T) {
	ctx := context.TODO()

	// make sure there are no docs
	count, err := s.collection.Query().Count(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 0, count)

	// create the docs
	counter := s.createDocuments(t)
	assert.Equal(t, 16, counter)

	// double check the count
	count, err = s.collection.Query().Count(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 16, count)

	// count the apple ones
	apples, err := s.collection.Query().WithinOrg("121212").InArray("apple", "fruits").List().All(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(apples))

	// gather the ids
	appleIds := s.extractAttributes(apples, "Id")
	sort.Strings(appleIds)

	// delete the apple ones
	deletedAppleIds, err := s.collection.Query().WithinOrg("121212").InArray("apple", "fruits").DeleteAll(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(deletedAppleIds))
	sort.Strings(deletedAppleIds)
	assert.Equal(t, appleIds, deletedAppleIds)

	// count the apple ones
	apples2, err := s.collection.Query().WithinOrg("121212").InArray("apple", "fruits").List().All(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(apples2))

	// double check the total count
	count, err = s.collection.Query().Count(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 12, count)

	// delete the other org's docs
	deletedOtherOrgIds, err := s.collection.Query().WithinOrg("232323").DeleteAll(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 8, len(deletedOtherOrgIds))

	// check all other docs
	allDocsLeft, err := s.collection.Query().List().All(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(allDocsLeft))

	for _, doc := range allDocsLeft {
		obj, ok := doc.(*TestDocument)
		assert.True(t, ok)
		assert.Equal(t, "121212", obj.OrganizationId)
	}
}

func (s *RealOrmTests) SubTestUpdate(t *testing.T) {
	ctx := context.TODO()

	// make sure there are no docs
	q := s.collection.Query()
	count, err := q.Count(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 0, count)

	// create the docs
	counter := s.createDocuments(t)
	assert.Equal(t, 16, counter)

	// count the apple ones
	apples, err := s.collection.Query().WithinOrg("121212").InArray("apple", "fruits").List().All(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(apples))

	// gather the ids
	appleIds := s.extractAttributes(apples, "Id")
	sort.Strings(appleIds)

	// update all apple based ones to have island fruits instead
	matchedIds, err := s.collection.Query().WithinOrg("121212").InArray("apple", "fruits").UpdateAll(ctx, map[string]interface{}{
		"fruits": []string{"mango", "coconut", "kiwi"},
	})

	assert.Nil(t, err)
	assert.Equal(t, 4, len(matchedIds))
	sort.Strings(matchedIds)
	assert.Equal(t, appleIds, matchedIds)

	// count the apple ones
	apples2, err := s.collection.Query().WithinOrg("121212").InArray("apple", "fruits").List().All(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(apples2))

	// count the coconut ones
	coconuts, err := s.collection.Query().WithinOrg("121212").InArray("coconut", "fruits").List().All(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(coconuts))
	coconutIds := s.extractAttributes(coconuts, "Id")
	sort.Strings(coconutIds)

	assert.Equal(t, coconutIds, matchedIds)
}

func (s *RealOrmTests) SubTestQueryOrderBy(t *testing.T) {
	ctx := context.TODO()
	_ = s.createDocuments(t)
	// Ascending
	objects, err := s.collection.Query().List().OrderBy("c").Asc().All(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 16, len(objects))
	letters := s.extractAttributes(objects, "C")
	assert.Equal(t, []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p"}, letters)

	// Descending
	objects, err = s.collection.Query().List().OrderBy("c").Desc().All(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 16, len(objects))
	letters = s.extractAttributes(objects, "C")
	assert.Equal(t, []string{"p", "o", "n", "m", "l", "k", "j", "i", "h", "g", "f", "e", "d", "c", "b", "a"}, letters)
}

func (s *RealOrmTests) SubTestQueryRandomAll(t *testing.T) {
	ctx := context.TODO()
	_ = s.createDocuments(t)

	// Random
	objects, err := s.collection.Query().List().RandomOrder().All(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 16, len(objects))
	letters1 := s.extractAttributes(objects, "C")

	differenceFound := false
out:
	for i := 0; i < 10; i++ {
		objects2, err := s.collection.Query().List().RandomOrder().All(ctx)
		assert.Nil(t, err)
		assert.Equal(t, 16, len(objects2))
		letters2 := s.extractAttributes(objects2, "C")
		for j := 0; j < 16; j++ {
			if letters1[j] != letters2[j] {
				// They are different, we are done checking for "randomness"
				differenceFound = true
				break out
			}
		}
	}
	assert.True(t, differenceFound, "Failed to find a difference in 10 tries!")
}

func (s *RealOrmTests) SubTestQueryRandomFirst(t *testing.T) {
	ctx := context.TODO()
	_ = s.createDocuments(t)

	// First random one of the 2 matched
	q := s.collection.Query()
	o := q.Operator()
	q.WithinOrg("121212").
		InArray("apple", "fruits").
		Where(
			o.StartsWith("email", "jer"),
		)

	found := make(map[string]int)
	for i := 0; i < 10; i++ {
		object, err := q.List().RandomOrder().First(ctx)
		assert.NoError(t, err, "expected no error")
		doc, ok := object.(*TestDocument)
		assert.True(t, ok)
		if _, ok := found[doc.Email]; !ok {
			found[doc.Email] = 1
		} else {
			found[doc.Email]++
		}

		if len(found) == 2 {
			break
		}
	}

	assert.Equal(t, len(found), 2, "Expected we'd get at least 2 matches for random order / first")
}

func (s *RealOrmTests) SubTestQueryOrderFirst(t *testing.T) {
	ctx := context.TODO()
	_ = s.createDocuments(t)

	// First random one of the 2 matched
	q := s.collection.Query()
	o := q.Operator()
	q.WithinOrg("121212").
		InArray("apple", "fruits").
		Where(
			o.StartsWith("email", "jer"),
		)

	// order asc
	object, err := q.List().OrderBy("email").Asc().First(ctx)
	assert.NoError(t, err, "should be no err")
	doc1, ok := object.(*TestDocument)
	assert.True(t, ok)
	assert.Equal(t, "jeremy@abc.com", doc1.Email)

	// order the other way
	object, err = q.List().OrderBy("email").Desc().First(ctx)
	assert.NoError(t, err, "should be no err")
	doc2, ok := object.(*TestDocument)
	assert.True(t, ok)
	assert.Equal(t, "jeren@abc.com", doc2.Email)
}

func (s *RealOrmTests) SubTestQueryLimitPaging(t *testing.T) {
	ctx := context.TODO()
	_ = s.createDocuments(t)
	// Check limit
	objects, err := s.collection.Query().List().OrderBy("c").Asc().Limit(4).All(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(objects))
	assert.Equal(t, []string{"a", "b", "c", "d"}, s.extractAttributes(objects, "C"))

	// Check paging, page 0
	objects, err = s.collection.Query().List().OrderBy("c").Asc().Paging(4, 0).All(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(objects))
	assert.Equal(t, []string{"a", "b", "c", "d"}, s.extractAttributes(objects, "C"))
	// Check paging, page 1
	objects, err = s.collection.Query().List().OrderBy("c").Asc().Paging(4, 1).All(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(objects))
	assert.Equal(t, []string{"e", "f", "g", "h"}, s.extractAttributes(objects, "C"))
	// Check paging, page 2
	objects, err = s.collection.Query().List().OrderBy("c").Asc().Paging(4, 2).All(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(objects))
	assert.Equal(t, []string{"i", "j", "k", "l"}, s.extractAttributes(objects, "C"))
	// Check paging, page 3
	objects, err = s.collection.Query().List().OrderBy("c").Asc().Paging(4, 3).All(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(objects))
	assert.Equal(t, []string{"m", "n", "o", "p"}, s.extractAttributes(objects, "C"))
	// Check paging, page 4
	objects, err = s.collection.Query().List().OrderBy("c").Asc().Paging(4, 4).All(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(objects))
	// Check paging, page 5
	objects, err = s.collection.Query().List().OrderBy("c").Asc().Paging(4, 5).All(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(objects))
}

func (s *RealOrmTests) extractAttributes(documents []interface{}, attributeName string) []string {
	extractedAttributes := make([]string, 0)

	for _, genericRecord := range documents {
		objValue := reflect.ValueOf(genericRecord)

		// Make sure the object is a pointer to a struct
		if objValue.Kind() == reflect.Ptr && objValue.Elem().Kind() == reflect.Struct {
			// Get the field value by name
			fieldValue := objValue.Elem().FieldByName(attributeName)

			// Check if the field exists
			if fieldValue.IsValid() {
				extractedAttributes = append(extractedAttributes, fieldValue.String())
			} else {
				// Handle unknown attribute
				panic(fmt.Sprintf("Unknown attribute: %s", attributeName))
			}
		} else {
			// Handle invalid object type
			panic("Invalid object type")
		}
	}

	return extractedAttributes
}

// ------------------------------
// Entry point for test suite
// ------------------------------

func TestOrmByReal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	} else {
		gtest.RunSubTests(t, &RealOrmTests{})
	}
}

// ------------------------------
