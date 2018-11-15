package main

import (
	"log"
	"os"
	"strconv"

	"github.com/TerrexTech/go-commonutils/commonutil"
	"github.com/TerrexTech/go-eventspoll/poll"
	"github.com/TerrexTech/go-mongoutils/mongo"
	"github.com/pkg/errors"
)

func loadMongoConfig(collectionName string, schema interface{}) (*poll.MongoConfig, error) {
	hosts := *commonutil.ParseHosts(
		os.Getenv("MONGO_HOSTS"),
	)

	database := os.Getenv("MONGO_DATABASE")
	metaCollection := os.Getenv("MONGO_META_COLLECTION")

	username := os.Getenv("MONGO_USERNAME")
	password := os.Getenv("MONGO_PASSWORD")
	connTimeoutStr := os.Getenv("MONGO_CONNECTION_TIMEOUT_MS")
	connTimeout, err := strconv.Atoi(connTimeoutStr)
	if err != nil {
		err = errors.Wrap(err, "Error converting MONGO_CONNECTION_TIMEOUT_MS to integer")
		log.Println(err)
		log.Println("A defalt value of 3000 will be used for MONGO_CONNECTION_TIMEOUT_MS")
		connTimeout = 3000
	}

	mongoConfig := mongo.ClientConfig{
		Hosts:               hosts,
		Username:            username,
		Password:            password,
		TimeoutMilliseconds: uint32(connTimeout),
	}

	// MongoDB Client
	client, err := mongo.NewClient(mongoConfig)
	if err != nil {
		err = errors.Wrap(err, "Error creating MongoClient")
		log.Println(err)
	}

	resTimeoutStr := os.Getenv("MONGO_CONNECTION_TIMEOUT_MS")
	resTimeout, err := strconv.Atoi(resTimeoutStr)
	if err != nil {
		err = errors.Wrap(err, "Error converting MONGO_RESOURCE_TIMEOUT_MS to integer")
		log.Println(err)
		log.Println("A defalt value of 5000 will be used for MONGO_RESOURCE_TIMEOUT_MS")
		connTimeout = 5000
	}
	conn := &mongo.ConnectionConfig{
		Client:  client,
		Timeout: uint32(resTimeout),
	}

	aggMongoCollection, err := createMongoCollection(conn, database, collectionName, schema)
	if err != nil {
		err = errors.Wrap(err, "Error creating MongoCollection")
		return nil, err
	}

	return &poll.MongoConfig{
		AggregateID:        aggregateID,
		AggCollection:      aggMongoCollection,
		Connection:         conn,
		MetaDatabaseName:   database,
		MetaCollectionName: metaCollection,
	}, nil
}

func createMongoCollection(conn *mongo.ConnectionConfig, db string, coll string, schema interface{}) (*mongo.Collection, error) {
	// Index Configuration
	indexConfigs := []mongo.IndexConfig{
		mongo.IndexConfig{
			ColumnConfig: []mongo.IndexColumnConfig{
				mongo.IndexColumnConfig{
					Name: "reportID",
				},
			},
			IsUnique: true,
			Name:     "reportID_index",
		},
	}

	// Create New Collection
	c := &mongo.Collection{
		Connection:   conn,
		Database:     db,
		Name:         coll,
		SchemaStruct: schema,
		Indexes:      indexConfigs,
	}
	collection, err := mongo.EnsureCollection(c)
	if err != nil {
		err = errors.Wrap(err, "Error creating MongoCollection")
		return nil, err
	}
	return collection, nil
}

// createClient creates a MongoDB-Client.
func CreateClient() (*mongo.Client, error) {
	// Would ideally set these config-params as environment vars
	config := mongo.ClientConfig{
		Hosts:               []string{"mongo:27017"},
		Username:            "root",
		Password:            "root",
		TimeoutMilliseconds: 3000,
	}

	// ====> MongoDB Client
	client, err := mongo.NewClient(config)
	// Let the parent functions handle error, always -.-
	// (Even though in these examples, we won't, for simplicity)
	return client, err
}

// createCollection demonstrates creating the collection and the associated database.
func CreateCollection(client *mongo.Client, collName string, schema interface{}) (*mongo.Collection, error) {

	database := os.Getenv("MONGO_DATABASE")

	// ====> Collection Configuration
	conn := &mongo.ConnectionConfig{
		Client:  client,
		Timeout: 5000,
	}
	// Index Configuration
	indexConfigs := []mongo.IndexConfig{
		mongo.IndexConfig{
			ColumnConfig: []mongo.IndexColumnConfig{
				mongo.IndexColumnConfig{
					Name: "itemID",
				},
				mongo.IndexColumnConfig{
					Name: "saleID",
				},
			},
			Name: "itemID_saleID_index",
		},
	}

	// ====> Create New Collection
	c := &mongo.Collection{
		Connection:   conn,
		Name:         collName,
		Database:     database,
		SchemaStruct: schema,
		Indexes:      indexConfigs,
	}
	return mongo.EnsureCollection(c)
}
