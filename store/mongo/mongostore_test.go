// +build integration

package mongo

import (
	"fmt"
	"github.com/emetsger/negtracker/model"
	"github.com/emetsger/negtracker/store"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"strconv"
	"testing"
	"time"
)

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

	neg := model.Neg{}
	err = underTest.Retrieve(id, &neg)
	assert.Nil(t, err)
	assert.NotNil(t, neg)

	assert.Equal(t, sampleNeg, neg)
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
