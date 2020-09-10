package id

import (
	"container/list"
	"fmt"
	"github.com/google/uuid"
	"sync"
)

// Responsible for minting unique identifiers for business model objects.
// Business model objects receive exactly one, unique, immutable identifier
// which is used to retrieve, update or link to an object.
//
// Identifiers are guaranteed to be unique within an instance of negtracker.
// Specifically, they are not guaranteed to be globally unique.
type Minter interface {
	Mint(interface{}) string
}

// Mints a UUID
func Mint() string {
	return uuid.New().String()
}

// Represents an identifier for business model objects that is also a sync.Locker
// Critical operations on business objects (e.g. processing PATCH or POST requests that contain self-supplied
// identifiers) are expected to be guarded by the lock on their BusinessId.
//
// BusinessIds are a (identifier, type) tuple which is guaranteed to be unique within an instance of negtracker.
type BusinessId struct {
	idTypeTuple
	l *sync.Mutex
}

// Obtain the lock of this BusinessId prior to entering a critical path
func (bid *BusinessId) Lock() {
	// Contact the lockTable for the current *sync.Mutex and lock it.
	bid.l = getLock(bid.idTypeTuple)
	bid.l.Lock()
}

// Release the lock of this BusinessId after exiting a critical path
func (bid *BusinessId) Unlock() {
	// Unlock the *sync.Mutex and Contact the lockTable via a chan and return it.
	bid.l.Unlock()

	// We don't nil this out, primarily so that unit tests can compare the value of the lock between two different
	// BusinessId structs.
	//bid.l = nil
}

// Obtain a business id for the given identifier string and type.
func GetId(id string, t interface{}) BusinessId {
	tuple := idTypeTuple{id, typeAsString(t)}
	bid := BusinessId{idTypeTuple: tuple}
	return bid
}

// Return the (identifier, type) tuple for this BusinessId as a string
func (bid BusinessId) String() string {
	return fmt.Sprintf("Id: %s Type: %s", bid.idTypeTuple.bid, bid.idTypeTuple.bidType)
}

// Tuple representing a unique business identifier
type idTypeTuple struct {
	// The string identifier
	bid string
	// The type being identified (as a string)
	bidType string
}

// Maintains a least recently used cache and table mapping idTypeTuple to sync.Mutex
type lockTable struct {
	// Guards access to this table
	mu sync.Mutex
	// The capacity of the table and lru
	cap int
	// Maps idTypeTuple to list elements (whose Value() returns a pair)
	table map[idTypeTuple]*list.Element
	// lru of BusinessId instances
	lru *list.List // elements are BusinessId instances
}

type pair struct {
	tuple idTypeTuple
	l     *sync.Mutex
}

// Global lock table, with a capacity for 100 business ids.  When the capacity is exceeded, the table is scanned and
// the least recently used business id is removed.  Not super efficient when the table gets full.
var locks = lockTable{sync.Mutex{}, 100, make(map[idTypeTuple]*list.Element, 100), &list.List{}}

func typeAsString(t interface{}) string {
	return fmt.Sprintf("%T", t)
}

// Obtain a lock for an (id, type) tuple.  Multiple requests for the same (id, type) tuple should return the same
// *sync.Mutex.
func getLock(tuple idTypeTuple) *sync.Mutex {
	// Guards access to the map and list
	locks.mu.Lock()
	defer func() { locks.mu.Unlock() }()

	// check to see if the lock for the type tuple exists in the table,
	// if so, push it to the head of the list and return it
	if elem, ok := locks.table[tuple]; ok {
		locks.lru.MoveToFront(elem)
		return elem.Value.(pair).l
	}

	// if the type tuple does not exist in the table:
	// 	1. create a new pair{tuple, *sync.Mutex}
	//	2. push the pair onto the head of the list, resulting in a new element
	//	3. put the element in the table for the tuple
	//  4a. if the list has exceeded capacity, 4b. remove the list tail and its corresponding entry in the table
	//	5. return the lock for the id

	p := pair{tuple, &sync.Mutex{}} // 1
	elem := locks.lru.PushFront(p)  // 2
	locks.table[tuple] = elem       // 3

	// TODO: this is quite sub-optimal.  Once the table fills to capacity, it is scanned each time for values to
	//  evict.  It would be nice to avoid this scan.
	if locks.lru.Len() > locks.cap { // 4a
		eValue := locks.lru.Remove(locks.lru.Back()) // 4b
		eValue.(pair).l.Lock()                       // FIXME: locked entries can't be removed so this may block,
		//  preventing other ids from being generated.  This will
		//  happen when > cap business ids have been Locked(), yet
		//  to be Unlock()ed
		defer func() { eValue.(pair).l.Unlock() }()

		// remove the Element with the pair from the map
		delete(locks.table, eValue.(pair).tuple)
	}

	return p.l
}
