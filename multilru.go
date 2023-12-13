// Copyright (c) 2013 CloudFlare, Inc.

package lrucache

import (
	"hash/crc32"
	"time"
)

// MultiLRUCache data structure. Never dereference it or copy it by
// value. Always use it through a pointer.
type MultiLRUCache[T any] struct {
	buckets uint
	cache   []*LRUCache[T]
}

// Using this constructor is almost always wrong. Use NewMultiLRUCache instead.
func (m *MultiLRUCache[T]) init(buckets, bucketCapacity uint) {
	m.buckets = buckets
	m.cache = make([]*LRUCache[T], buckets)
	for i := uint(0); i < buckets; i++ {
		m.cache[i] = NewLRUCache[T](bucketCapacity)
	}
}

// Set the stale expiry grace period for each cache in the multicache instance.
func (m *MultiLRUCache[T]) SetExpireGracePeriod(p time.Duration) {
	for _, c := range m.cache {
		c.ExpireGracePeriod = p
	}
}

func NewMultiLRUCache[T any](buckets, bucketCapacity uint) *MultiLRUCache[T] {
	m := &MultiLRUCache[T]{}
	m.init(buckets, bucketCapacity)
	return m
}

func (m *MultiLRUCache[T]) bucketNo(key string) uint {
	// Arbitrary choice. Any fast hash will do.
	return uint(crc32.ChecksumIEEE([]byte(key))) % m.buckets
}

func (m *MultiLRUCache[T]) Set(key string, value T, expire time.Time) {
	m.cache[m.bucketNo(key)].Set(key, value, expire)
}

func (m *MultiLRUCache[T]) SetNow(key string, value T, expire time.Time, now time.Time) {
	m.cache[m.bucketNo(key)].SetNow(key, value, expire, now)
}

func (m *MultiLRUCache[T]) Get(key string) (value T, ok bool) {
	return m.cache[m.bucketNo(key)].Get(key)
}

func (m *MultiLRUCache[T]) GetQuiet(key string) (value T, ok bool) {
	return m.cache[m.bucketNo(key)].Get(key)
}

func (m *MultiLRUCache[T]) GetNotStale(key string) (value T, ok bool) {
	return m.cache[m.bucketNo(key)].GetNotStale(key)
}

func (m *MultiLRUCache[T]) GetNotStaleNow(key string, now time.Time) (value T, ok bool) {
	return m.cache[m.bucketNo(key)].GetNotStaleNow(key, now)
}

func (m *MultiLRUCache[T]) GetStale(key string) (value T, ok, expired bool) {
	return m.cache[m.bucketNo(key)].GetStale(key)
}

func (m *MultiLRUCache[T]) GetStaleNow(key string, now time.Time) (value T, ok, expired bool) {
	return m.cache[m.bucketNo(key)].GetStaleNow(key, now)
}

func (m *MultiLRUCache[T]) Del(key string) (value T, ok bool) {
	return m.cache[m.bucketNo(key)].Del(key)
}

func (m *MultiLRUCache[T]) Clear() int {
	var s int
	for _, c := range m.cache {
		s += c.Clear()
	}
	return s
}

func (m *MultiLRUCache[T]) Len() int {
	var s int
	for _, c := range m.cache {
		s += c.Len()
	}
	return s
}

func (m *MultiLRUCache[T]) Capacity() int {
	var s int
	for _, c := range m.cache {
		s += c.Capacity()
	}
	return s
}

func (m *MultiLRUCache[T]) Expire() int {
	var s int
	for _, c := range m.cache {
		s += c.Expire()
	}
	return s
}

func (m *MultiLRUCache[T]) ExpireNow(now time.Time) int {
	var s int
	for _, c := range m.cache {
		s += c.ExpireNow(now)
	}
	return s
}
