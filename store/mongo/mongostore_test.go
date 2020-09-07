// +build integration

package mongo

import (
	"errors"
	"fmt"
	"github.com/emetsger/negtracker/id"
	"github.com/emetsger/negtracker/model"
	"github.com/emetsger/negtracker/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"os"
	"strconv"
	"testing"
	"time"
)

var underTest = &MongoStore{}

var sampleNeg = model.Neg{
	Id:          "negId",
	Film:        "Tri-X",
	EI:          200,
	Developer:   "HC-110 (B)",
	FrameNumber: "3",
	Tags:        []string{"druid hill", "daffodil", "spring"},
	Description: "Druid Hill",
	Format:      "120",
}

func TestMongoStore_StoreAndRetrieve(t *testing.T) {
	businessId := id.Mint()
	sampleNeg.Id = businessId
	persistenceId, err := underTest.Store(sampleNeg)
	assert.Nil(t, err)
	assert.NotEqual(t, "", persistenceId)

	neg := model.Neg{}
	err = underTest.Retrieve(businessId, &neg)
	assert.Nil(t, err)
	assert.NotNil(t, neg)

	assert.Equal(t, sampleNeg, neg)
}

func TestMongoStore_dupKeyCause(t *testing.T) {
	var err error

	_, err = underTest.negCol.InsertOne(underTest.ctx, bson.M{idField: "1"})
	require.Nil(t, err)

	_, err = underTest.negCol.InsertOne(underTest.ctx, bson.M{idField: "1"})
	require.NotNil(t, err)

	// errors.Is and errors.As are broken for mongo.WriteException, I believe
	wex, ok := err.(mongo.WriteException)
	require.True(t, ok)

	for i := range wex.WriteErrors {
		we := err.(mongo.WriteException).WriteErrors[i]
		require.Equal(t, errCodeDupKey, we.Code)
	}

	// the logic above is used in dupKeyCause(error) function
	require.True(t, dupKeyCause(err))
}

func TestMongoStore_DuplicateBusinessIds(t *testing.T) {
	obj := sampleNeg
	obj.Id = "TestMongoStore_DuplicateBusinessIds"

	_, err := underTest.Store(obj)
	require.Nil(t, err)

	_, err = underTest.Store(obj)

	require.True(t, errors.Is(err, store.DuplicateKeyErr))
	log.Print(err.Error())
}

func TestMain(m *testing.M) {
	// Configure the store
	underTest.Configure(TestConfig)

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
			TestConfig.DbName, TestConfig.DbUri)
		os.Exit(run)
	} else {
		if keep, _ := strconv.ParseBool(os.Getenv(store.EnvDbKeepResults)); !keep {
			if err := underTest.negCol.Drop(underTest.ctx); err != nil {
				log.Printf("Error dropping collection %s: %s", TestConfig.NegCollection, err.Error())
			}
			if err := underTest.db.Drop(underTest.ctx); err != nil {
				log.Printf("Error dropping database %s: %s", TestConfig.DbName, err.Error())
			}
		} else {
			log.Printf("Leaving database %s (%s) for postmortem per %s",
				TestConfig.DbName, TestConfig.DbUri, store.EnvDbKeepResults)
		}

		os.Exit(run)
	}
}
