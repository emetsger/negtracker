package id

import (
	"fmt"
	"github.com/emetsger/negtracker/model"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"testing"
	"time"
)

// validate initial state and that one entry is added to the table and lru each time a lock for a unique business id is
// requested.
func Test_GetBusinessId(t *testing.T) {
	// there should be no entries in the table or list
	require.Equal(t, 0, len(locks.table))
	require.Equal(t, 0, locks.lru.Len())

	uuid := Mint()
	bidType := model.Neg{}

	bid := GetId(uuid, bidType)
	require.NotNil(t, bid)

	// Lock the business id, which should create one element in the lock table and list for the idTypeTuple
	bid.Lock()

	expectedLock := bid.l
	require.Equal(t, 1, len(locks.table))
	require.Equal(t, 1, locks.lru.Len())
	require.Same(t, expectedLock, locks.lru.Front().Value)
	require.Same(t, expectedLock, locks.table[idTypeTuple{uuid, typeAsString(bidType)}].Value)

	// Unlock the business id later, which shouldn't change the size of the table or its values.
	bid.Unlock()

	require.Equal(t, 1, len(locks.table))
	require.Equal(t, 1, locks.lru.Len())
	require.Same(t, expectedLock, locks.lru.Front().Value)
	require.Same(t, expectedLock, locks.table[idTypeTuple{uuid, typeAsString(bidType)}].Value)

	// Get the same business id but assign to a different variable.  The table and list size shouldn't change.
	bid2 := GetId(uuid, bidType)
	require.Equal(t, 1, len(locks.table))
	require.Equal(t, 1, locks.lru.Len())

	// Lock bid2 (which populates BusinessId.l).  Both bid and bid2 should share the same lock.
	bid2.Lock()
	require.Same(t, bid.l, bid2.l)
	defer func() { bid2.Unlock() }()

	// Get a different business id.  Obtain its lock.  The table and list size should increment by 1.
	bid3 := GetId("foo", bidType)
	bid3.Lock()
	defer func() { bid3.Unlock() }()
	require.Equal(t, 2, len(locks.table))
	require.Equal(t, 2, locks.lru.Len())
}

// ensure that the backing table and lru caps out
func Test_GetBusinessIdCap(t *testing.T) {
	var i int
	for i = len(locks.table); i <= locks.cap*2; i++ {
		// Generate the id
		id := GetId(Mint(), model.Neg{})
		// Add an entry in the locking table
		id.Lock()
		// Immediately release the lock so the lru table scan can complete
		id.Unlock()
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
	require.Nil(t, bid1.l) // idTypeTuple.l is initially nil, only initialized when Lock() is called.
	bid1.Lock()
	bid1.Unlock() // Because Lock() initializes  idTypeTuple.l, it must be called.  Because bid1 and
	// bid2 are using the same lock, Unlock() must be called otherwise bid2.Lock() will
	// hang.
	bid2 := GetId(id, n)
	require.Nil(t, bid2.l)
	bid2.Lock()
	bid2.Unlock()
	require.NotNil(t, bid2.l)

	require.Same(t, bid1.l, bid2.l)

	bid3 := GetId("foo", n)
	bid3.Lock()
	bid3.Unlock()
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
	// 1. obtain a business id, then lock it to create an entry in the lock table, then unlock it.
	// 2. generate exactly enough business ids to expire the id obtained by 1; verify that bid1 has been expired from the table
	// 3. obtain a business id using the same tuple as 1
	// 4. attempt to lock the bid from 1, verify they share a lock.

	// 1
	id := "moo"
	n := model.Neg{}
	bid1 := GetId(id, n)
	bid1.Lock()
	bid1.Unlock() // insure an entry is created in the table

	// 2
	for i := 0; i < locks.cap; i++ {
		j := GetId(Mint(), n)
		j.Lock()
		j.Unlock() // insure an entry is created in the table
	}
	require.Nil(t, locks.table[idTypeTuple{id, typeAsString(n)}]) // the idTypeTuple for bid1 has been expired.

	// 3
	bid2 := GetId(id, n)
	bid2.Lock()
	bid2.Unlock() // initialize the idType.l

	// 4
	bid1.Lock()
	bid1.Unlock() // should get the latest lock from the table, the same lock used by bid2.
	require.Same(t, bid1.l, bid2.l)
}
