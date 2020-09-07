package id

import (
	"fmt"
	"github.com/emetsger/negtracker/model"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"sync"
	"testing"
	"time"
)

// validate initial state and that one entry is added to the table and lru each time a different business id is
// requested.
func Test_GetBusinessId(t *testing.T) {
	// there should be no entries in the table or list
	require.Equal(t, 0, len(locks.table))
	require.Equal(t, 0, locks.lru.Len())

	uuid := Mint()
	bidType := model.Neg{}

	bid := GetId(uuid, bidType)
	require.NotNil(t, bid)

	// there should be one element in the table and list, and they should contain the newly created business id
	expectedBid := BusinessId{idTypeTuple{uuid, typeAsString(bidType)}, &sync.Mutex{}}
	require.Equal(t, 1, len(locks.table))
	require.Equal(t, 1, locks.lru.Len())
	require.Equal(t, expectedBid, locks.lru.Front().Value)
	require.Equal(t, expectedBid, locks.table[idTypeTuple{uuid, typeAsString(bidType)}].Value)

	// Get the same business id, the table and list size shouldn't change
	bid = GetId(uuid, bidType)
	require.NotNil(t, bid)
	require.Equal(t, 1, len(locks.table))
	require.Equal(t, 1, locks.lru.Len())

	// Get another, and the table and list should increment by one
	bid = GetId("foo", bidType)
	require.NotNil(t, bid)
	require.Equal(t, 2, len(locks.table))
	require.Equal(t, 2, locks.lru.Len())
}

// ensure that the backing table and lru caps out
func Test_GetBusinessIdCap(t *testing.T) {
	var i int
	for i = len(locks.table); i <= locks.cap*2; i++ {
		_ = GetId(Mint(), model.Neg{})
	}

	require.Equal(t, locks.cap, len(locks.table))
	require.Equal(t, locks.cap, locks.lru.Len())
	require.True(t, i > locks.cap)
}

// ensure concurrent goroutines won't corrupt the global lock table
func Test_LockTableConcurrency(t *testing.T) {
	routines := 20
	count := locks.cap * 10
	totalcalls := routines * count

	// Mix creation and retrievals by having more calls to GetId than UUIDs being used.
	// Create a table of UUIDs, and have the threads select one at random
	ids := make(map[int]string, totalcalls/4)
	for i := 0; i < totalcalls/4; i++ {
		ids[i] = fmt.Sprintf("%d", i)
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// make 10 go routines that are retrieving business ids, communicating them back through the channel
	f := func(c chan BusinessId) {
		for i := 0; i < count; i++ {
			c <- GetId(ids[r.Intn(totalcalls/4)], model.Neg{})
		}
	}

	ch := make(chan BusinessId)

	for i := 0; i < routines; i++ {
		go f(ch)
	}

	for i := 0; i < totalcalls; i++ {
		<-ch
	}
}

func Test_BusinessIdLockSame(t *testing.T) {
	id := "moo"
	n := model.Neg{}
	bid1 := GetId(id, n)
	bid2 := GetId(id, n)
	require.Same(t, bid1.l, bid2.l)

	bid3 := GetId("foo", n)
	require.NotSame(t, bid1.l, bid3.l)
}

// ensure that BusinessIds returned by GetBusinessId will use the same mutex
func Test_BusinessIdLock(t *testing.T) {
	id := "moo"
	n := model.Neg{}
	ch := make(chan time.Time)

	f := func(chan time.Time) {
		bid := GetId(id, n)
		bid.Lock()
		defer func() {
			time.Sleep(1 * time.Second)
			bid.Unlock()
		}()
		ch <- time.Now()
	}

	go f(ch)
	go f(ch)

	one := <-ch
	two := <-ch

	require.True(t, math.Abs(one.Sub(two).Seconds()) >= 1)
}

func Test_BusinessIdLockFail(t *testing.T) {
	// 1. obtain a business id
	// 2. generate exactly enough business ids to expire the id obtained by 1
	// 3. obtain a business id using the same tuple as 1
	// 4. verify they do not share a lock.

	// 1
	id := "moo"
	n := model.Neg{}
	bid1 := GetId(id, n)

	// 2
	for i := 0; i < locks.cap; i++ {
		_ = GetId(Mint(), n)
	}
	require.Nil(t, locks.table[idTypeTuple{id, typeAsString(n)}])

	// 3
	bid2 := GetId(id, n)

	// 4
	require.Same(t, bid1.l, bid2.l)
}
