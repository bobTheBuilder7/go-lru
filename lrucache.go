// Copyright (c) 2013 CloudFlare, Inc.

package lrucache

import (
	"sync"
	"time"
)

// Every element in the cache is linked to three data structures:
// Table map, PriorityQueue heap ordered by expiry and a LruList list
// ordered by decreasing popularity.
type entry[T any] struct {
	element element[T] // list element. value is a pointer to this entry
	key     string     // key is a key!
	value   T          //
	expire  time.Time  // time when the item is expired. it's okay to be stale.
	index   int        // index for priority queue needs. -1 if entry is free
}

// LRUCache data structure. Never dereference it or copy it by
// value. Always use it through a pointer.
type LRUCache[T any] struct {
	lock          sync.Mutex
	table         map[string]*entry[T] // all entries in table must be in lruList
	priorityQueue priorityQueue[T]     // some elements from table may be in priorityQueue
	lruList       list[T]              // every entry is either used and resides in lruList
	freeList      list[T]              // or free and is linked to freeList

	ExpireGracePeriod time.Duration // time after an expired entry is purged from cache (unless pushed out of LRU)
}

// Initialize the LRU cache instance. O(capacity)
func (b *LRUCache[T]) init(capacity uint) {
	b.table = make(map[string]*entry[T], capacity)
	b.priorityQueue = make([]*entry[T], 0, capacity)
	b.lruList.Init()
	b.freeList.Init()
	HeapInit[T](&b.priorityQueue)

	// Reserve all the entries in one giant continous block of memory
	arrayOfEntries := make([]entry[T], capacity)
	for i := uint(0); i < capacity; i++ {
		e := &arrayOfEntries[i]
		e.element.Value = e
		e.index = -1
		b.freeList.PushElementBack(&e.element)
	}
}

// Create new LRU cache instance. Allocate all the needed memory. O(capacity)
func NewLRUCache[T any](capacity uint) *LRUCache[T] {
	b := &LRUCache[T]{}
	b.init(capacity)
	return b
}

// Give me the entry with lowest expiry field if it's before now.
func (b *LRUCache[T]) expiredEntry(now time.Time) *entry[T] {
	if len(b.priorityQueue) == 0 {
		return nil
	}

	if now.IsZero() {
		// Fill it only when actually used.
		now = time.Now()
	}

	if e := b.priorityQueue[0]; e.expire.Before(now) {
		return e
	}
	return nil
}

// Give me the least used entry.
func (b *LRUCache[T]) leastUsedEntry() *entry[T] {
	return b.lruList.Back().Value
}

func (b *LRUCache[T]) freeSomeEntry(now time.Time) (e *entry[T], used bool) {
	if b.freeList.Len() > 0 {
		return b.freeList.Front().Value, false
	}

	e = b.expiredEntry(now)
	if e != nil {
		return e, true
	}

	if b.lruList.Len() == 0 {
		return nil, false
	}

	return b.leastUsedEntry(), true
}

// Move entry from used/lru list to a free list. Clear the entry as well.
func (b *LRUCache[T]) removeEntry(e *entry[T]) {
	if e.element.list != &b.lruList {
		panic("list lruList")
	}

	if e.index != -1 {
		HeapRemove(&b.priorityQueue, e.index)
	}
	b.lruList.Remove(&e.element)
	b.freeList.PushElementFront(&e.element)
	delete(b.table, e.key)
	e.key = ""
	var t T
	e.value = t
}

func (b *LRUCache[T]) insertEntry(e *entry[T]) {
	if e.element.list != &b.freeList {
		panic("list freeList")
	}

	if !e.expire.IsZero() {
		HeapPush(&b.priorityQueue, e)
	}
	b.freeList.Remove(&e.element)
	b.lruList.PushElementFront(&e.element)
	b.table[e.key] = e
}

func (b *LRUCache[T]) touchEntry(e *entry[T]) {
	b.lruList.MoveToFront(&e.element)
}

// SetNow adds an item to the cache overwriting existing one if it
// exists. Allows specifing current time required to expire an item
// when no more slots are used. O(log(n)) if expiry is set, O(1) when
// clear.
func (b *LRUCache[T]) SetNow(key string, value T, expire time.Time, now time.Time) {
	b.lock.Lock()
	defer b.lock.Unlock()

	var used bool

	e := b.table[key]
	if e != nil {
		used = true
	} else {
		e, used = b.freeSomeEntry(now)
		if e == nil {
			return
		}
	}
	if used {
		b.removeEntry(e)
	}

	e.key = key
	e.value = value
	e.expire = expire
	b.insertEntry(e)
}

// Set adds an item to the cache overwriting existing one if it
// exists. O(log(n)) if expiry is set, O(1) when clear.
func (b *LRUCache[T]) Set(key string, value T, expire time.Time) {
	b.SetNow(key, value, expire, time.Time{})
}

// Get a key from the cache, possibly stale. Update its LRU score. O(1)
func (b *LRUCache[T]) Get(key string) (v T, ok bool) {
	b.lock.Lock()
	defer b.lock.Unlock()

	e := b.table[key]
	if e == nil {
		var t T
		return t, false
	}

	b.touchEntry(e)
	return e.value, true
}

// GetQuiet gets a key from the cache, possibly stale. Don't modify its LRU score. O(1)
func (b *LRUCache[T]) GetQuiet(key string) (v T, ok bool) {
	b.lock.Lock()
	defer b.lock.Unlock()

	e := b.table[key]

	if e == nil {
		var t T
		return t, false
	}

	return e.value, true
}

// GetNotStale gets a key from the cache, make sure it's not stale. Update its
// LRU score. O(log(n)) if the item is expired.
func (b *LRUCache[T]) GetNotStale(key string) (value T, ok bool) {
	return b.GetNotStaleNow(key, time.Now())
}

// GetNotStaleNow gets a key from the cache, make sure it's not stale. Update its
// LRU score. O(log(n)) if the item is expired.
func (b *LRUCache[T]) GetNotStaleNow(key string, now time.Time) (value T, ok bool) {
	b.lock.Lock()
	defer b.lock.Unlock()

	e := b.table[key]
	var t T
	if e == nil {
		return t, false
	}

	if e.expire.Before(now) {
		// Remove entries expired for more than a graceful period
		if b.ExpireGracePeriod == 0 || e.expire.Sub(now) > b.ExpireGracePeriod {
			b.removeEntry(e)
		}
		return t, false
	}

	b.touchEntry(e)
	return e.value, true
}

// GetStale gets a key from the cache, possibly stale. Update its LRU
// score. O(1) always.
func (b *LRUCache[T]) GetStale(key string) (value T, ok, expired bool) {
	return b.GetStaleNow(key, time.Now())
}

// GetStaleNow gets a key from the cache, possibly stale. Update its LRU
// score. O(1) always.
func (b *LRUCache[T]) GetStaleNow(key string, now time.Time) (value T, ok, expired bool) {
	b.lock.Lock()
	defer b.lock.Unlock()

	e := b.table[key]

	if e == nil {
		var t T
		return t, false, false
	}

	b.touchEntry(e)
	return e.value, true, e.expire.Before(now)
}

// Del gets and remove a key from the cache. O(log(n)) if the item is using expiry, O(1) otherwise.
func (b *LRUCache[T]) Del(key string) (v T, ok bool) {
	b.lock.Lock()
	defer b.lock.Unlock()

	e := b.table[key]

	if e == nil {
		var t T
		return t, false
	}

	value := e.value
	b.removeEntry(e)
	return value, true
}

// Evict all items from the cache. O(n*log(n))
func (b *LRUCache[T]) Clear() int {
	b.lock.Lock()
	defer b.lock.Unlock()

	// First, remove entries that have expiry set
	l := len(b.priorityQueue)
	for i := 0; i < l; i++ {
		// This could be reduced to O(n).
		b.removeEntry(b.priorityQueue[0])
	}

	// Second, remove all remaining entries
	r := b.lruList.Len()
	for i := 0; i < r; i++ {
		b.removeEntry(b.leastUsedEntry())
	}
	return l + r
}

// Evict all the expired items. O(n*log(n))
func (b *LRUCache[T]) Expire() int {
	return b.ExpireNow(time.Now())
}

// Evict items that expire before `now`. O(n*log(n))
func (b *LRUCache[T]) ExpireNow(now time.Time) int {
	b.lock.Lock()
	defer b.lock.Unlock()

	i := 0
	for {
		e := b.expiredEntry(now)
		if e == nil {
			break
		}
		b.removeEntry(e)
		i += 1
	}
	return i
}

// Number of entries used in the LRU
func (b *LRUCache[T]) Len() int {
	// yes. this stupid thing requires locking
	b.lock.Lock()
	defer b.lock.Unlock()

	return b.lruList.Len()
}

// Capacity gets the total capacity of the LRU
func (b *LRUCache[T]) Capacity() int {
	// yes. this stupid thing requires locking
	b.lock.Lock()
	defer b.lock.Unlock()

	return b.lruList.Len() + b.freeList.Len()
}
