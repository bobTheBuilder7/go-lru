// Copyright (c) 2013 CloudFlare, Inc.

// This code is based on golang example from "container/heap" package.

package lrucache

type priorityQueue[T any] []*entry[T]

func (pq priorityQueue[T]) Len() int {
	return len(pq)
}

func (pq priorityQueue[T]) Less(i, j int) bool {
	return pq[i].expire.Before(pq[j].expire)
}

func (pq priorityQueue[T]) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue[T]) Push(x *entry[T]) {
	n := len(*pq)
	x.index = n
	*pq = append(*pq, x)
}

func (pq *priorityQueue[T]) Pop() *entry[T] {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1
	*pq = old[0 : n-1]
	return item
}
