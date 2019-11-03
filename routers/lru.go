/***********************************************
        File Name: lru.go
        Author: Abby Cin
        Mail: abbytsing@gmail.com
        Created Time: 9/21/19 3:49 PM
***********************************************/

package routers

import (
	"sync"
)

// fuck golang
type ICache interface {
	Write(c *Context)
	Name() string
	Size() int64
	MaxSize() int64
	Update(args ...interface{}) (interface{}, error)
	Data() []byte
}

type lru_node struct {
	Data ICache
	Size int64
	Next *lru_node
	Prev *lru_node
}

type ListIter struct {
	n *lru_node
}

func (l *ListIter) HasNext() bool {
	return l.n != nil
}

func (l *ListIter) Next() {
	l.n = l.n.Next
}

func (l *ListIter) GetData() ICache {
	return l.n.Data
}

type List struct {
	Root *lru_node
	Tail *lru_node
}

func NewList() *List {
	return &List{
		Root: nil,
		Tail: nil,
	}
}

func (l *List) PushBack(c ICache) *lru_node {
	n := &lru_node{c, c.Size(), nil, nil}
	if l.Root == nil {
		l.Tail = n
		l.Root = l.Tail
	} else {
		l.Tail.Next = n
		n.Prev = l.Tail
		l.Tail = n
	}
	return n
}

func (l *List) RemoveNode(n *lru_node) {
	if l.Root == l.Tail && l.Root == n {
		l.Root = nil
		l.Tail = nil
		return
	}
	if l.Root == n {
		l.Root = l.Root.Next
		l.Root.Prev = nil
	} else if l.Tail == n {
		l.Tail = l.Tail.Prev
		l.Tail.Next = nil
	} else {
		n.Prev.Next = n.Next
		n.Next.Prev = n.Prev
	}
}

func (l *List) MovetoBack(n *lru_node) {
	if l.Root == l.Tail || l.Root == nil || l.Tail == n {
		return
	}
	if l.Root == n {
		l.Root.Next.Prev = nil
		l.Root = l.Root.Next
	} else {
		n.Prev.Next = n.Next
		n.Next.Prev = n.Prev
	}
	n.Next = nil
	l.Tail.Next = n
	n.Prev = l.Tail
	l.Tail = n
}

func (l *List) First() *lru_node {
	return l.Root
}

func (l *List) Last() *lru_node {
	return l.Tail
}

func (l *List) Iter() *ListIter {
	return &ListIter{l.Root}
}

type LRU struct {
	mtx          sync.Mutex
	capacity     int // how many item can be cached
	size         int
	queue        *List
	totalBytes   int64 // how many bytes can be cached
	currentBytes int64
	index        map[string]*lru_node
}

func NewLRU(capacity int, limit int64) *LRU {
	return &LRU{
		mtx:          sync.Mutex{},
		capacity:     capacity,
		size:         0,
		queue:        new(List),
		totalBytes:   limit,
		currentBytes: 0,
		index:        make(map[string]*lru_node),
	}
}

func (l *LRU) Add(c ICache) bool {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	if l.size == l.capacity || c.Size()+l.currentBytes > l.totalBytes {
		if c.Size() <= l.queue.First().Data.Size() {
			l.currentBytes -= l.queue.First().Data.Size()
			delete(l.index, l.queue.First().Data.Name())
			l.queue.RemoveNode(l.queue.First()) // least recent used
			r := l.queue.PushBack(c)
			l.index[c.Name()] = r
			l.currentBytes += c.Size()
			return true
		}
		return false
	} else {
		r := l.queue.PushBack(c)
		l.index[c.Name()] = r
		l.currentBytes += c.Size()
		l.size += 1
		return true
	}
}

func (l *LRU) Get(k string) ICache {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	idx, ok := l.index[k]
	if !ok {
		return nil
	}
	l.queue.MovetoBack(idx)
	return idx.Data
}

func (l *LRU) Remove(k string) bool {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	old, ok := l.index[k]
	if !ok {
		return false
	}
	l.currentBytes -= old.Size
	l.queue.RemoveNode(old)
	delete(l.index, k)
	return true
}

func (l *LRU) Update(c ICache) {
	_, err := c.Update()
	if err == nil {
		l.Remove(c.Name())
		l.Add(c)
	}
}

func (l *LRU) UpdateAll() {
	l.mtx.Lock()
	caches := make([]ICache, 0)
	for k, v := range l.index {
		l.queue.RemoveNode(v)
		if _, err := v.Data.Update(); err == nil {
			caches = append(caches, v.Data)
			delete(l.index, k)
		}
	}
	l.mtx.Unlock()
	for _, cache := range caches {
		l.Add(cache)
	}
}
