package utils

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/arangodb/go-driver"
	"io"
)

type MockDatabase struct {
	LastQuery    string
	LastBindVars map[string]interface{}
	MyCursor     *MockCursor
}

func (c *MockDatabase) Collection(ctx context.Context, name string) (driver.Collection, error) {
	//TODO implement me
	panic("implement me1")
}

func (c *MockDatabase) CollectionExists(ctx context.Context, name string) (bool, error) {
	//TODO implement me
	panic("implement me2")
}

func (c *MockDatabase) Collections(ctx context.Context) ([]driver.Collection, error) {
	//TODO implement me
	panic("implement me3")
}

func (c *MockDatabase) CreateCollection(ctx context.Context, name string, options *driver.CreateCollectionOptions) (driver.Collection, error) {
	//TODO implement me
	panic("implement me4")
}

func (c *MockDatabase) View(ctx context.Context, name string) (driver.View, error) {
	//TODO implement me
	panic("implement me5")
}

func (c *MockDatabase) ViewExists(ctx context.Context, name string) (bool, error) {
	//TODO implement me
	panic("implement me6")
}

func (c *MockDatabase) Views(ctx context.Context) ([]driver.View, error) {
	//TODO implement me
	panic("implement me7")
}

func (c *MockDatabase) CreateArangoSearchView(ctx context.Context, name string, options *driver.ArangoSearchViewProperties) (driver.ArangoSearchView, error) {
	//TODO implement me
	panic("implement me8")
}

func (c *MockDatabase) Graph(ctx context.Context, name string) (driver.Graph, error) {
	//TODO implement me
	panic("implement me9")
}

func (c *MockDatabase) GraphExists(ctx context.Context, name string) (bool, error) {
	//TODO implement me
	panic("implement me10")
}

func (c *MockDatabase) Graphs(ctx context.Context) ([]driver.Graph, error) {
	//TODO implement me
	panic("implement me11")
}

func (c *MockDatabase) CreateGraph(ctx context.Context, name string, options *driver.CreateGraphOptions) (driver.Graph, error) {
	//TODO implement me
	panic("implement me12")
}

func (c *MockDatabase) CreateGraphV2(ctx context.Context, name string, options *driver.CreateGraphOptions) (driver.Graph, error) {
	//TODO implement me
	panic("implement me13")
}

func (c *MockDatabase) BeginTransaction(ctx context.Context, cols driver.TransactionCollections, opts *driver.BeginTransactionOptions) (driver.TransactionID, error) {
	//TODO implement me
	panic("implement me14")
}

func (c *MockDatabase) CommitTransaction(ctx context.Context, tid driver.TransactionID, opts *driver.CommitTransactionOptions) error {
	//TODO implement me
	panic("implement me15")
}

func (c *MockDatabase) AbortTransaction(ctx context.Context, tid driver.TransactionID, opts *driver.AbortTransactionOptions) error {
	//TODO implement me
	panic("implement me16")
}

func (c *MockDatabase) TransactionStatus(ctx context.Context, tid driver.TransactionID) (driver.TransactionStatusRecord, error) {
	//TODO implement me
	panic("implement me17")
}

func (c *MockDatabase) EnsureAnalyzer(ctx context.Context, analyzer driver.ArangoSearchAnalyzerDefinition) (bool, driver.ArangoSearchAnalyzer, error) {
	//TODO implement me
	panic("implement me18")
}

func (c *MockDatabase) Analyzer(ctx context.Context, name string) (driver.ArangoSearchAnalyzer, error) {
	//TODO implement me
	panic("implement me19")
}

func (c *MockDatabase) Analyzers(ctx context.Context) ([]driver.ArangoSearchAnalyzer, error) {
	//TODO implement me
	panic("implement me20")
}

func (c *MockDatabase) ValidateQuery(ctx context.Context, query string) error {
	//TODO implement me
	panic("implement me21")
}

func (c *MockDatabase) Transaction(ctx context.Context, action string, options *driver.TransactionOptions) (interface{}, error) {
	//TODO implement me
	panic("implement me22")
}

func (c *MockDatabase) Name() string {
	return "MockItyo"
}

func (c *MockDatabase) Info(ctx context.Context) (driver.DatabaseInfo, error) {
	return driver.DatabaseInfo{}, errors.New("Nope")
}

func (c *MockDatabase) EngineInfo(ctx context.Context) (driver.EngineInfo, error) {
	return driver.EngineInfo{}, errors.New("No way!")
}

func (c *MockDatabase) Remove(ctx context.Context) error {
	return nil
}

func (c *MockDatabase) Query(ctx context.Context, query string, bindVars map[string]interface{}) (driver.Cursor, error) {
	c.LastQuery = query
	c.LastBindVars = bindVars
	if c.MyCursor != nil {
		return c.MyCursor, nil
	}

	return &MockCursor{}, nil
}

type MockCursor struct {
	io.Closer
	Items []string
	Index int64
}

func (c *MockCursor) Close() error {
	return nil
}

func (c *MockCursor) Count() int64 {
	return int64(len(c.Items))
}

func (c *MockCursor) Statistics() driver.QueryStatistics {
	//TODO implement me
	panic("implement me23")
}

func (c *MockCursor) Extra() driver.QueryExtra {
	//TODO implement me
	panic("implement me24")
}

func (c *MockCursor) HasMore() bool {
	return c.Index < c.Count()
}

func (c *MockCursor) ReadDocument(ctx context.Context, result interface{}) (driver.DocumentMeta, error) {
	// Check if there are more items in the slice
	if c.Index >= int64(len(c.Items)) {
		return driver.DocumentMeta{}, io.EOF
	}
	item := c.Items[c.Index]
	c.Index++
	if err := json.Unmarshal([]byte(item), result); err != nil {
		return driver.DocumentMeta{}, err
	}

	return driver.DocumentMeta{Key: "1", ID: "1", Rev: "1"}, nil
}

type MockClient struct {
	MockDatabase *MockDatabase
}

func (c *MockClient) Database(ctx context.Context, name string) (driver.Database, error) {
	return c.MockDatabase, nil
}
