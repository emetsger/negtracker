package mongo

import (
	"context"
	"fmt"
	"github.com/emetsger/negtracker/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
)

const (
	ENV_DB_URI            = "DB_URI"
	ENV_DB_NAME           = "DB_NAME"
	ENV_DB_NEG_COLLECTION = "DB_NEG_COLLECTION"
)

// Represents the configuration used for the MongoDB driver
type MongoConfig struct {
	// env var DB_URI
	DbUri string
	// env var DB_NAME
	DbName string
	// env var DB_NEG_COLLECTION
	NegCollection string
	//// *no* env var, for unit testing
	//initOnConnect bool
	Opts *options.ClientOptions
}

type MongoStore struct {
	ctx    context.Context
	client *mongo.Client
	db     *mongo.Database
	negCol *mongo.Collection
}

func (m *MongoStore) Retrieve(id string) (neg model.Neg, err error) {
	var objid primitive.ObjectID
	if objid, err = primitive.ObjectIDFromHex(id); err != nil {
		panic(fmt.Sprintf("Error creating ObjectId from id '%s'", id))
	}

	res := m.negCol.FindOne(m.ctx, bson.M{"_id": objid})
	err = res.Decode(&neg)

	return
}

func (m *MongoStore) Store(n model.Neg) (id string, err error) {
	var data []byte
	var res *mongo.InsertOneResult

	if data, err = bson.Marshal(n); err == nil {
		if res, err = m.negCol.InsertOne(m.ctx, data); err == nil {
			id = res.InsertedID.(primitive.ObjectID).Hex()
		}
	}

	return
}

func (m *MongoStore) Configure(c interface{}) {
	var config MongoConfig
	var err error

	config = verifyConfig(c)

	// Create an initial client, configure it with the DbUri supplied in the configuration
	opts := options.Client().ApplyURI(config.DbUri)

	// Allow the *ClientOptions on the config override anything in the initial client
	if config.Opts != nil {
		m.client, err = mongo.NewClient(opts, config.Opts)
	} else {
		m.client, err = mongo.NewClient(opts)
	}

	if err != nil {
		panic("store/mongo: error creating Mongo Client: " + err.Error())
	}

	m.ctx = context.Background()

	if err = m.client.Connect(m.ctx); err != nil {
		panic(fmt.Sprintf("store/mongo: error connecting to %s, %s", config.DbUri, err.Error()))
	}

	m.db = m.client.Database(config.DbName)
	m.negCol = m.db.Collection(config.NegCollection)
}

func verifyConfig(c interface{}) MongoConfig {
	if c == nil {
		panic("store/mongo: config must not be nil")
	}

	var config *MongoConfig
	var ok bool

	if config, ok = c.(*MongoConfig); !ok {
		panic(fmt.Sprintf("store/mongo: config must be a *MongoConfig (was: %v, %T)", config, config))
	}

	checkLen("MongoConfig.DbUri", config.DbUri)
	checkLen("MongoConfig.DbName", config.DbName)
	checkLen("MongoConfig.NegCollection", config.NegCollection)

	return *config
}

func checkLen(fieldName, fieldValue string) {
	if len(strings.TrimSpace(fieldValue)) == 0 {
		panic(fmt.Sprintf("store/mongo: %s is required", fieldName))
	}
}
