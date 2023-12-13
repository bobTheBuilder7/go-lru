package lrucache

import (
	"github.com/bobTheBuilder7/SLdent/internal/assert"
	"math/rand"
	"runtime"
	"testing"
	"time"
)

func TestBasicExpiry(t *testing.T) {
	t.Parallel()
	b := NewLRUCache[string](3)
	if _, ok := b.Get("a"); ok {
		t.Error("")
	}

	now := time.Now()
	b.Set("b", "vb", now.Add(2*time.Second))
	b.Set("a", "va", now.Add(1*time.Second))
	b.Set("c", "vc", now.Add(3*time.Second))

	if v, _ := b.Get("a"); v != "va" {
		t.Error("")
	}
	if v, _ := b.Get("b"); v != "vb" {
		t.Error("")
	}
	if v, _ := b.Get("c"); v != "vc" {
		t.Error("")
	}

	b.Set("d", "vd", now.Add(4*time.Second))
	if _, ok := b.Get("a"); ok {
		t.Error("Expecting element A to be evicted")
	}

	b.Set("e", "ve", now.Add(-4*time.Second))
	if _, ok := b.Get("b"); ok {
		t.Error("Expecting element B to be evicted")
	}

	b.Set("f", "vf", now.Add(5*time.Second))
	if _, ok := b.Get("e"); ok {
		t.Error("Expecting element E to be evicted")
	}

	if v, _ := b.Get("c"); v != "vc" {
		t.Error("Expecting element C to not be evicted")
	}
	n := now.Add(10 * time.Second)
	b.SetNow("g", "vg", now.Add(5*time.Second), n)
	if _, ok := b.Get("c"); ok {
		t.Error("Expecting element C to be evicted")
	}

	if b.Len() != 3 {
		t.Error("Expecting different length")
	}
	b.Del("miss")
	b.Del("g")
	assert.Equal(t, b.Len(), 2)
	b.Clear()
	if b.Len() != 0 {
		t.Error("Expecting different length")
	}

	now = time.Now()
	b.Set("b", "vb", now.Add(2*time.Second))
	b.Set("a", "va", now.Add(1*time.Second))
	b.Set("d", "vd", now.Add(4*time.Second))
	b.Set("c", "vc", now.Add(3*time.Second))

	if _, ok := b.Get("b"); ok {
		t.Error("Expecting miss")
	}

	b.GetQuiet("miss")
	if v, _ := b.GetQuiet("a"); v != "va" {
		t.Error("Expecting hit")
	}

	b.Set("e", "ve", now.Add(5*time.Second))
	if _, ok := b.Get("a"); ok {
		t.Error("Expecting miss")
	}

	if b.Capacity() != 3 {
		t.Error("Expecting different capacity")
	}
}

func TestBasicNoExpiry(t *testing.T) {
	t.Parallel()
	b := NewLRUCache[string](3)
	if _, ok := b.Get("a"); ok {
		t.Error("")
	}

	b.Set("b", "vb", time.Time{})
	b.Set("a", "va", time.Time{})
	b.Set("c", "vc", time.Time{})
	b.Set("d", "vd", time.Time{})

	if _, ok := b.Get("b"); ok {
		t.Error("expecting miss")
	}

	if v, _ := b.Get("a"); v != "va" {
		t.Error("expecting hit")
	}
	if v, _ := b.Get("c"); v != "vc" {
		t.Error("expecting hit")
	}
	if v, _ := b.Get("d"); v != "vd" {
		t.Error("expecting hit")
	}

	past := time.Now().Add(-10 * time.Second)

	b.Set("e", "ve", past)

	if _, ok := b.Get("a"); ok {
		t.Error("expecting miss")
	}
	if v, _ := b.Get("e"); v != "ve" {
		t.Error("expecting hit")
	}

	// Make sure expired items get evicted before items without expiry
	b.Set("f", "vf", time.Time{})
	if _, ok := b.Get("e"); ok {
		t.Error("expecting miss")
	}

	r := b.Clear()
	if b.Len() != 0 || r != 3 {
		t.Error("Expecting different length")
	}

	b.Set("c", "vc", time.Time{})
	b.Set("d", "vd", time.Time{})
	b.Set("e", "ve", past)

	if b.Len() != 3 {
		t.Error("Expecting different length")
	}
	r = b.Expire()
	if b.Len() != 2 || r != 1 {
		t.Error("Expecting different length")
	}
	r = b.Clear()
	if b.Len() != 0 || r != 2 {
		t.Error("Expecting different length")
	}
}

func TestNil(t *testing.T) {
	t.Parallel()
	b := NewLRUCache[string](3)

	// value nil
	if _, ok := b.Get("a"); ok != false {
		t.Error("expecting miss")
	}

	// value not nil (sanity check)
	if _, ok := b.Get("b"); ok != false {
		t.Error("expecting miss")
	}

	b.Set("b", "vb", time.Time{})

	if v, ok := b.Get("b"); v != "vb" || ok != true {
		t.Error("expecting miss")
	}
}

func rec(foo func()) (recovered int) {
	recovered = 0
	defer func() {
		if r := recover(); r != nil {
			recovered += 1
		}
	}()
	foo()
	return recovered
}

func TestPanicByValue(t *testing.T) {
	t.Parallel()
	b := NewLRUCache[string](3)

	b.Set("a", "a", time.Time{})

	c := *b
	r := rec(func() {
		c.Del("a")
	})
	if r != 1 {
		t.Error("Expecting panic")
	}

	b.Del("a")

	r = rec(func() {
		c.Set("a", "A", time.Time{})
	})
	if r != 1 {
		t.Error("Expecting panic")
	}
}

func TestZeroLength(t *testing.T) {
	t.Parallel()
	b := NewLRUCache[string](0)

	if _, ok := b.Get("a"); ok {
		t.Error("Expected miss")
	}

	b.Set("a", "va", time.Time{})
	if _, ok := b.Get("a"); ok {
		t.Error("Expected miss")
	}

	b.Clear()
}

func TestExtra(t *testing.T) {
	t.Parallel()
	b := NewLRUCache[string](3)
	if _, ok := b.Get("a"); ok {
		t.Error("")
	}

	now := time.Now()
	b.Set("b", "vb", now.Add(-2*time.Second))
	b.Set("a", "va", now.Add(-1*time.Second))
	b.Set("c", "vc", now.Add(3*time.Second))

	if v, _ := b.Get("a"); v != "va" {
		t.Error("expecting value")
	}

	if _, ok := b.GetNotStale("a"); ok {
		t.Error("not expecting value")
	}
	if _, ok := b.GetNotStale("miss"); ok {
		t.Error("not expecting value")
	}
	if v, _ := b.GetNotStale("c"); v != "vc" {
		t.Error("expecting hit")
	}

	if b.Len() != 2 {
		t.Error("Expecting different length")
	}
	if b.Expire() != 1 {
		t.Error("Expecting different length")
	}
}

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(65 + rand.Intn(90-65))
	}
	return string(bytes)
}

func createFilledBucket(expire time.Time) *LRUCache[string] {
	b := NewLRUCache[string](1000)
	for i := 0; i < 1000; i++ {
		b.Set(randomString(2), "value", expire)
	}
	return b
}

func TestConcurrentGet(t *testing.T) {
	t.Parallel()
	b := createFilledBucket(time.Now().Add(time.Duration(4)))

	done := make(chan bool)
	worker := func() {
		for i := 0; i < 10000; i++ {
			b.Get(randomString(2))
		}
		done <- true
	}
	workers := 4
	for i := 0; i < workers; i++ {
		go worker()
	}
	for i := 0; i < workers; i++ {
		_ = <-done
	}
}

func TestConcurrentSet(t *testing.T) {
	t.Parallel()
	b := createFilledBucket(time.Now().Add(time.Duration(4)))

	done := make(chan bool)
	worker := func() {
		expire := time.Now().Add(4 * time.Second)
		for i := 0; i < 10000; i++ {
			b.Set(randomString(2), "value", expire)
		}
		done <- true
	}
	workers := 4
	for i := 0; i < workers; i++ {
		go worker()
	}
	for i := 0; i < workers; i++ {
		_ = <-done
	}
}

func BenchmarkConcurrentGetLRUCache(bb *testing.B) {
	bb.ReportAllocs()
	b := createFilledBucket(time.Now().Add(time.Duration(4)))

	cpu := runtime.GOMAXPROCS(0)
	ch := make(chan bool)
	worker := func() {
		for i := 0; i < bb.N/cpu; i++ {
			b.Get(randomString(2))
		}
		ch <- true
	}
	for i := 0; i < cpu; i++ {
		go worker()
	}
	for i := 0; i < cpu; i++ {
		_ = <-ch
	}
}

func BenchmarkConcurrentSetLRUCache(bb *testing.B) {
	bb.ReportAllocs()
	b := createFilledBucket(time.Now().Add(time.Duration(4)))

	cpu := runtime.GOMAXPROCS(0)
	ch := make(chan bool)
	worker := func() {
		for i := 0; i < bb.N/cpu; i++ {
			expire := time.Now().Add(4 * time.Second)
			b.Set(randomString(2), "v", expire)
		}
		ch <- true
	}
	for i := 0; i < cpu; i++ {
		go worker()
	}
	for i := 0; i < cpu; i++ {
		_ = <-ch
	}
}

// No expiry
func BenchmarkConcurrentSetNXLRUCache(bb *testing.B) {
	bb.ReportAllocs()
	b := createFilledBucket(time.Time{})

	cpu := runtime.GOMAXPROCS(0)
	ch := make(chan bool)
	worker := func() {
		for i := 0; i < bb.N/cpu; i++ {
			b.Set(randomString(2), "v", time.Time{})
		}
		ch <- true
	}
	for i := 0; i < cpu; i++ {
		go worker()
	}
	for i := 0; i < cpu; i++ {
		_ = <-ch
	}
}
