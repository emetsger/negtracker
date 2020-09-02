// +build integration

package mongo

import (
	"fmt"
	"github.com/emetsger/negtracker/model"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"
)

const ENV_DB_KEEP_RESULTS = "DB_KEEP_RESULTS"

var r = strconv.Itoa(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(1000))

// Uses a unique database for this test instance
var config = &MongoConfig{
	DbUri:         os.Getenv(ENV_DB_URI),
	DbName:        genRandString(ENV_DB_NAME, r),
	NegCollection: os.Getenv(ENV_DB_NEG_COLLECTION),
	Opts:          options.Client().SetAppName("mongostore_test").SetServerSelectionTimeout(5 * time.Second),
}
var underTest = &MongoStore{}

var sampleNeg = model.Neg{
	ID:          "negId",
	Film:        "Tri-X",
	EI:          200,
	Developer:   "HC-110 (B)",
	FrameNumber: "3",
	Tags:        []string{"druid hill", "daffodil", "spring"},
	Description: "Druid Hill",
	Format:      "120",
}

func TestMongoStore_StoreAndRetrieve(t *testing.T) {
	id, err := underTest.Store(sampleNeg)
	assert.Nil(t, err)
	assert.NotEqual(t, "", id)

	neg, err := underTest.Retrieve(id)
	assert.Nil(t, err)
	assert.NotNil(t, neg)

	assert.Equal(t, sampleNeg, neg)
}

func TestMain(m *testing.M) {
	// Configure the store
	underTest.Configure(config)

	// Wait for the database to be available
	times := 5
	startTime := time.Now()
	var err error
	for err = underTest.client.Ping(underTest.ctx, nil); err != nil && times > 0; times-- {
		err = underTest.client.Ping(underTest.ctx, nil)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		panic(fmt.Sprintf("Server has not started after %ds, %s", int(time.Since(startTime).Seconds()), err.Error()))
	}

	if run := m.Run(); run > 0 {
		log.Printf("Errors encountered in test, leaving database %s (%s) for postmortem",
			config.DbName, config.DbUri)
		os.Exit(run)
	} else {
		if keep, _ := strconv.ParseBool(os.Getenv(ENV_DB_KEEP_RESULTS)); !keep {
			if err := underTest.negCol.Drop(underTest.ctx); err != nil {
				log.Printf("Error dropping collection %s: %s", config.NegCollection, err.Error())
			}
			if err := underTest.db.Drop(underTest.ctx); err != nil {
				log.Printf("Error dropping database %s: %s", config.DbName, err.Error())
			}
		} else {
			log.Printf("Leaving database %s (%s) for postmortem per %s",
				config.DbName, config.DbUri, ENV_DB_KEEP_RESULTS)
		}

		os.Exit(run)

	}

}

func genRandString(env, r string) string {
	return fmt.Sprintf("%s_%s_%s", os.Getenv(env), "mongostore_test", r)
}
