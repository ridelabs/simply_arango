# simply_arango

simply_arango is a Go library for talking to an arangodb database.

## What this is
simply_arango is a very simple ORM for ArangoDB in golang.

The beauty of this library is how easy it is to make a query to filter records in a collection and then perform an
operation on those records.

We are using the builder pattern to make this library easy to use and not clutter up your code even when building complex
AQL queries.


Be aware, Arangodb has a lot of cool stuff that I haven't needed in my project, so those
things aren't implemented yet, especially graphs. Feel free to submit PRs if you have things you'd like to see in here.


## Installation
Standard `go get`:
```
$ go get github.com/ridelabs/simply_arango
```

## Usage & Example
The unit real_orm_test.go has a lot of great examples that you can look at and run. To run them, you should copy the
dot.env.example to .env and change the values to your particular database.

Initialize things
```go
// Make a connection to the database
orm.NewConnection(ctx, dbName, dbUser, dbPass, dbUrl)

// define a collection
collection := &orm.Collection{
    Connection: conn,
    TableName:  "foo",
    AllocateRecord: func() interface{} {
        // This is for allocating a record of the correct type
        return &TestDocument{}
    },
    OrganizationIdKey: "organization_id",
}

// Basic CRUD

// create a doc
id, err := collection.Create(ctx, &TestDocument{
    Name:           "Fred",
    Email:          "freddy@mycorp.com",
    Counter:        0,
    OrganizationId: "3434",
    Fruits: []string{
        "cherry", "apple", "persimmon",
    },
})

// read the doc back by id
obj, err := collection.Get(ctx, id)
if err != nil {
    return nil, err
}

// read the doc back but be sure it's in the same domain (dipping into our Query abilities [see below])
object, err := collection.Query().WithinOrg("3434").ById(id).First(ctx)
if err != nil {
    return nil, err
}

doc, ok := object.(*TestDocument)
if !ok {
    return errors.New("shouldn't happen!")
}

// update the doc
if err := collection.Update(ctx, doc); err != nil {
    return nil, err
}

// delete the doc
if err := collection.Delete(ctx, doc); err != nil {
    return nil, err
}
 
```

Simple chaining query (chain in as many filters as you want)
```go
obj, err := collection.Query().WithinOrg("3434").Filter("email", "freddy@mycorp.com").First(ctx)

```

More complex query
```go
o := q.Operator() // here you have to get the object used to create operators for this more complex Where query
q := q.Where(
    o.And(
        o.EndsWith("email", "mycorp.com"),
        o.LessThan("counter", o.MakeVariableIfNative(5)),
    ),
).InArray("persimmon", "fruits")
```

Take a query then you can perform multi-record operations on it like count, get the first, get all, page through them, update all or delete all
```go
	q.Count(ctx)
	q.First(ctx)
	q.List().All(ctx)
	q.List().OrderBy("c").Asc().Paging(4, 0).All(ctx)
	q.UpdateAll(ctx, map[string]interface{}{
		"fruits": []string{"mango", "coconut", "kiwi"},
	})
	q.DeleteAll(ctx)
```

We also support ordering and paging etc. Editing with an autocompleting editor makes it really easy to see what functions are available each step of the way. Chain things as deep as you want.

Check out https://github.com/ridelabs/simply_arango/blob/main/orm/real_orm_test.go for the best example of what this golang arangodb orm wrapper usage looks like.


Author: Eric Harrison (mailplum.com)


