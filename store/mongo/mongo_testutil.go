// build +integration

package mongo

import (
	"github.com/emetsger/negtracker/store"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"time"
)

// Uses a unique database for this test instance
var TestConfig = &MongoConfig{
	DbUri:         os.Getenv(store.EnvDbUri),
	DbName:        store.GenTestDbName(store.EnvDbName, store.GenInt()),
	NegCollection: os.Getenv(store.EnvDbNegCollection),
	Opts:          options.Client().SetServerSelectionTimeout(5 * time.Second),
}
