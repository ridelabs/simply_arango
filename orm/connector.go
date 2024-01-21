package orm

import (
	"context"
	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	log "github.com/sirupsen/logrus"
)

type Connection struct {
	Database driver.Database
	Client   driver.Client
}

func NewConnection(ctx context.Context, databaseName, dbUser, dbPass, dbUrl string) (*Connection, error) {
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{dbUrl},
	})
	if err != nil {
		return nil, err
	}

	arangoClient, err := driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication(dbUser, dbPass),
	})
	if err != nil {
		return nil, err
	}

	if err != nil {
		log.Error("Failed to connect to arangodb", log.Fields{"err": err})
		return nil, err
		//os.Exit(3)
	}

	exists, err := arangoClient.DatabaseExists(ctx, databaseName)
	if err != nil {
		log.Error("Failed query for database!", log.Fields{"err": err})
		return nil, err
	}

	var db driver.Database
	if !exists {
		db, err = arangoClient.CreateDatabase(ctx, databaseName, nil)
		if err != nil {
			return nil, err
		}
	} else {
		db, err = arangoClient.Database(ctx, databaseName)
		if err != nil {
			return nil, err
		}
	}

	return &Connection{
		Database: db,
		Client:   arangoClient,
	}, nil
}
