// Copyright (c) 2013 CloudFlare, Inc.

// Extensions to "container/list" that allowing reuse of Elements.

package lrucache

func (l *list[T]) PushElementFront(e *element[T]) *element[T] {
	return l.insert(e, &l.root)
}

func (l *list[T]) PushElementBack(e *element[T]) *element[T] {
	return l.insert(e, l.root.prev)
}

func (l *list[T]) PopElementFront() *element[T] {
	el := l.Front()
	l.Remove(el)
	return el
}

func (l *list[T]) PopFront() interface{} {
	el := l.Front()
	l.Remove(el)
	return el.Value
}
