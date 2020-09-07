package mongo

import (
	"context"
	"errors"
	"fmt"
	"github.com/emetsger/negtracker/model"
	"github.com/emetsger/negtracker/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"strings"
)

// Used to identify the field that mongo will use for storing business ids on persisted documents
//
// Note: the value of this constant must align with the entity structs in the `model` package; the structs must use a
// field named `Id`, or have a bson tag that maps the entity's business id field to "id", e.g.:
//   type struct Foo {
//   	FooId string `bson:"id"`
//   }

const (
	idField       = "id"
	errCodeDupKey = 11000
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

func (m *MongoStore) Retrieve(id string, t interface{}) error {
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

	err := res.Decode(t)

	if err != nil {
		return store.SentinelErr(store.DecodingErr, fmt.Sprintf("type:  %T", t), fmt.Sprintf("%v", err))
	}

	return nil
}

func (m *MongoStore) Store(obj interface{}) (string, error) {
	var data []byte
	var res *mongo.InsertOneResult
	var id string
	var err error

	if data, err = bson.Marshal(obj); err == nil {
		if res, err = m.negCol.InsertOne(m.ctx, data); err == nil {
			id = res.InsertedID.(primitive.ObjectID).Hex()
		}
	}

	if err != nil {
		if dupKeyCause(err) {
			return id, store.SentinelErr(store.DuplicateKeyErr, "underlying error", fmt.Sprintf("%v", err))
		} else {
			return id, store.GenericErr(fmt.Sprintf("attempt to insert document with key %s failed", id),
				fmt.Sprintf("%v", err))
		}
	}

	return id, nil
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

	// create unique index on business id for the NegCollection
	idxKeys := bson.D{{"id", 1}}
	idxBool := true
	idxName := "Negative Business Id"
	idxOpts := options.IndexOptions{Unique: &idxBool, Name: &idxName}
	if idxName, idxErr := m.negCol.Indexes().CreateOne(m.ctx, mongo.IndexModel{idxKeys, &idxOpts}); idxErr != nil {
		panic("Unable to create unique business id index on NegCollection, " + idxErr.Error())
	} else {
		log.Printf("Created unique business id index on %s, %s", config.NegCollection, idxName)
	}
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

// Returns true if the error is a mongo.WriteException caused by insertion of a duplicate key into a unique index
func dupKeyCause(err error) bool {
	wex := mongo.WriteException{}

	if !errors.As(err, &wex) {
		return false
	}

	for i := range wex.WriteErrors {
		werr := wex.WriteErrors[i]
		if werr.Code == errCodeDupKey {
			return true
		}
	}

	return false
}
