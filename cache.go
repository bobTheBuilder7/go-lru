// Copyright (c) 2013 CloudFlare, Inc.

// Package lrucache implements a last recently used cache data structure.
//
// This code tries to avoid dynamic memory allocations - all required
// memory is allocated on creation.  Access to the data structure is
// O(1). Modification O(log(n)) if expiry is used, O(1)
// otherwise.
//
// This package exports three things:
//
//	LRUCache: is the main implementation. It supports multithreading by
//	    using guarding mutex lock.
//
//	MultiLRUCache: is a sharded implementation. It supports the same
//	    API as LRUCache and uses it internally, but is not limited to
//	    a single CPU as every shard is separately locked. Use this
//	    data structure instead of LRUCache if you have lock
//	    contention issues.
//
//	Cache interface: Both implementations fulfill it.
package lrucache

import (
	"time"
)

// Cache interface is fulfilled by the LRUCache and MultiLRUCache
// implementations.
type Cache[T any] interface {
	// Get Methods not needing to know current time.
	//
	// Get a key from the cache, possibly stale. Update its LRU
	// score.
	Get(key string) (value T, ok bool)
	// GetQuiet Get a key from the cache, possibly stale. Don't modify its LRU score. O(1)
	GetQuiet(key string) (value T, ok bool)
	// Del Get and remove a key from the cache.
	Del(key string) (value T, ok bool)
	// Clear Evict all items from the cache.
	Clear() int
	// Len Number of entries used in the LRU
	Len() int
	// Capacity Get the total capacity of the LRU
	Capacity() int

	// Set Methods use time.Now() when necessary to determine expiry.
	//
	// Add an item to the cache overwriting existing one if it
	// exists.
	Set(key string, T, expire time.Time)
	// GetNotStale get a key from the cache, make sure it's not stale. Update
	// its LRU score.
	GetNotStale(key string) (value T, ok bool)
	// Expire Evict all the expired items.
	Expire() int

	// SetNow Methods allowing to explicitly specify time used to
	// determine if items are expired.
	//
	// Add an item to the cache overwriting existing one if it
	// exists. Allows specifying current time required to expire an
	// item when no more slots are used.
	SetNow(key string, T, expire time.Time, now time.Time)
	// GetNotStaleNow Get a key from the cache, make sure it's not stale. Update
	// its LRU score.
	GetNotStaleNow(key string, now time.Time) (T, ok bool)
	// ExpireNow Evict items that expire before Now.
	ExpireNow(now time.Time) int
}
