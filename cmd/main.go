package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/ridelabs/simply_arango/orm"
	"os"
)

type TestDocument struct {
	Id             string `json:"id"`
	OrganizationId string `json:"organization_id"`

	A       string   `json:"a"`
	B       string   `json:"b"`
	C       string   `json:"c"`
	Name    string   `json:"name"`
	Email   string   `json:"email"`
	Counter int      `json:"counter"`
	Fruits  []string `json:"fruits"`
}

func main() {
	fmt.Printf("Hey, pwd=%s\n", os.Getenv("PWD"))
	repoPath := os.Getenv("PWD") + "/.env"
	if err := godotenv.Load(repoPath); err != nil {
		fmt.Printf("Failed to load .env, please create one by copying/editing the dot.env.example")
		os.Exit(1)
	}

	dbUser := os.Getenv("ARANGODB_USER")
	dbPass := os.Getenv("ARANGODB_PASS")
	dbUrl := os.Getenv("ARANGODB_URL")
	dbName := os.Getenv("TEST_DB_NAME")

	ctx := context.TODO()
	conn, err := orm.NewConnection(ctx, dbName, dbUser, dbPass, dbUrl)
	if err != nil {
		fmt.Printf("Failed to connect to arangodb %s", err)
		os.Exit(5)
	}

	// Make a collection
	collection := &orm.Collection{
		Connection: conn,
		TableName:  "foo",
		AllocateRecord: func() interface{} {
			// This is for allocating a record of the correct type
			return &TestDocument{}
		},
		OrganizationIdKey: "organization_id",
	}

	if err := collection.Drop(ctx); err != nil {
		fmt.Printf("Failed to drop collection!\n")
		os.Exit(7)
	}

	if err := collection.Initialize(ctx); err != nil {
		fmt.Printf("Failed to create collection!\n")
		os.Exit(7)
	}

	id, err := collection.Create(ctx, &TestDocument{
		Name:           "Fred",
		Email:          "freddy@mycorp.com",
		Counter:        0,
		OrganizationId: "3434",
		Fruits: []string{
			"cherry", "apple", "persimmon",
		},
	})

	fmt.Printf("created doc %s", id)

	// simple query within an org where an email equals
	obj, err := collection.Query().WithinOrg("3434").Filter("email", "freddy@mycorp.com").First(ctx)
	if err != nil {
		fmt.Printf("Failure! %s\n", err)
		os.Exit(4)
	}

	// convert the obj to our type
	doc, ok := obj.(*TestDocument)
	if !ok {
		fmt.Printf("Failure2 couldn't convert doc\n")
		os.Exit(5)
	}

	fmt.Printf("Found user with id = %s\n", doc.Id)

	q := collection.Query()
	o := q.Operator()
	count, err := q.Where(
		o.And(
			o.EndsWith("email", "mycorp.com"),
			o.LessThan("counter", o.MakeVariableIfNative(5)),
		),
	).InArray("persimmon", "fruits").Count(ctx)

	if err != nil {
		fmt.Printf("Failed to count it! %s\n", err)
		os.Exit(6)
	}

	fmt.Printf("Fount a total of %d docs\n", count)

	obj, err = q.First(ctx)
	if err != nil {
		fmt.Printf("Failed to get the doc, what? %s\n", err)
		os.Exit(8)
	}

	doc, ok = obj.(*TestDocument)
	if !ok {
		fmt.Printf("We couldn't convert to doc... weird\n")
		os.Exit(9)
	}

	fmt.Printf("Ok, the doc is %+v\n", doc)

}
