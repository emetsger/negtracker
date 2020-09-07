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

// Used to identify the field that mongo will use for storing business ids on persisted documents
//
// Note: the value of this constant must align with the entity structs in the `model` package; the structs must use a
// field named `Id`, or have a bson tag that maps the entity's business id field to "id", e.g.:
//   type struct Foo {
//   	FooId string `bson:"id"`
//   }
const idField = "id"

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

func (m *MongoStore) Retrieve(id string, t interface{}) (err error) {
	var res *mongo.SingleResult

	// If t is an WebResource, then treat the supplied id as a business identifier,
	// otherwise treat it as a persistence identifier.

	// Granted, this implementation does not need to perform this check; any document can be retrieved from Mongo as
	// long as its key can be found in the `idField`.

	// The only reason this check is here is to prevent a programmer error, where the caller may be attempting to
	// retrieve something that is not a business object.

	if _, ok := t.(model.WebResource); !ok {
		panic(fmt.Sprintf("store/mongo: can only retrieve objects of type model.WebResource, not %T", t))
	} else {
		res = m.negCol.FindOne(m.ctx, bson.M{idField: id})
	}

	err = res.Decode(t)

	return
}

func (m *MongoStore) Store(obj interface{}) (id string, err error) {
	var data []byte
	var res *mongo.InsertOneResult

	if data, err = bson.Marshal(obj); err == nil {
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
