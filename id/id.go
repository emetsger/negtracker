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
func (bid BusinessId) Lock() {
	bid.l.Lock()
}

// Release the lock of this BusinessId after exiting a critical path
func (bid BusinessId) Unlock() {
	bid.l.Unlock()
}

// Return the (identifier, type) tuple for this BusinessId
func (bid BusinessId) String() string {
	return fmt.Sprintf("Id: %s Type: %s", bid.idTypeTuple.Id, bid.idTypeTuple.Type)
}

// Tuple representing a unique business identifier
type idTypeTuple struct {
	// The string identifier
	Id string
	// The type being identified (as a string)
	Type string
}

// Maintains a least recently used cache and table mapping idTypeTuple to sync.Mutex
type lockTable struct {
	// Guards access to this table
	mu sync.Mutex
	// The capacity of the table and lru
	cap int
	// Maps idTypeTuple to list elements
	table map[idTypeTuple]*list.Element
	// lru of BusinessId instances
	lru *list.List // elements are BusinessId instances
}

// global lock table, with a capacity for 100 business ids
var locks = lockTable{sync.Mutex{}, 100, make(map[idTypeTuple]*list.Element, 100), &list.List{}}

// Obtain a business id for the given identifier string and type.  The returned BusinessId
func GetId(id string, t interface{}) BusinessId {
	locks.mu.Lock()
	defer func() { locks.mu.Unlock() }()

	// check to see if the business id exists in the table, if so, push it to the head of the list and return it
	key := idTypeTuple{id, typeAsString(t)}
	if elem, ok := locks.table[key]; ok {
		locks.lru.MoveToFront(elem)
		return elem.Value.(BusinessId)
	}

	// if the business id does not exist in the table:
	// 	1. create the id
	//	3. push the id onto the head of the list
	//	4. put the resulting element in the table
	//  5. if the list has exceeded capacity, remove the tail and the entry in the table
	//	6. return business id

	bid := BusinessId{key, &sync.Mutex{}}
	elem := locks.lru.PushFront(bid)
	locks.table[key] = elem

	if locks.lru.Len() > locks.cap {
		removed := locks.lru.Remove(locks.lru.Back()).(BusinessId)
		delete(locks.table, removed.idTypeTuple)
	}

	return bid
}

func typeAsString(t interface{}) string {
	return fmt.Sprintf("%T", t)
}
