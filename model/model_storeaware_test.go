// +build integration

package model

import (
	"github.com/emetsger/negtracker/store/mongo"
	"github.com/stretchr/testify/assert"
	"testing"
)

var mongoStore = &mongo.MongoStore{}

var sampleNeg = Neg{
	ID:          "negId",
	Film:        "Tri-X",
	EI:          200,
	Developer:   "HC-110 (B)",
	FrameNumber: "3",
	Tags:        []string{"druid hill", "daffodil", "spring"},
	Description: "Druid Hill",
	Format:      "120",
}

func Test_NegStoreAndRetrieve(t *testing.T) {
	mongoStore.Configure(mongo.TestConfig)

	var err error
	var id string

	id, err = sampleNeg.Store(mongoStore)
	assert.Nil(t, err)

	retrievedNeg := Neg{}
	assert.Nil(t, retrievedNeg.Retrieve(mongoStore, id))

	assert.Equal(t, sampleNeg, retrievedNeg)
}
