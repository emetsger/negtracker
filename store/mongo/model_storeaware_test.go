// +build integration

package mongo

import (
	"github.com/emetsger/negtracker/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

var mongoStore = &MongoStore{}

func Test_NegStoreAndRetrieve(t *testing.T) {
	mongoStore.Configure(TestConfig)

	var err error
	_, err = sampleNeg.Store(mongoStore)
	assert.Nil(t, err)

	retrievedNeg := &model.Neg{}
	assert.Nil(t, retrievedNeg.Retrieve(mongoStore, sampleNeg.Id))

	assert.Equal(t, sampleNeg, *retrievedNeg)
}
