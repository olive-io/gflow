/*
Copyright 2025 The gflow Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package scheduler

import (
	"container/heap"
	"sync"
)

type ScoreFn[T any] func(v T) int64

type INode interface {
	ID() int64
}

type item[T any] struct {
	value T
	fn    ScoreFn[T]
	index int
}

type priorityQueue[T any] struct {
	list []*item[T]
}

func (pq *priorityQueue[T]) Len() int { return len(pq.list) }

func (pq *priorityQueue[T]) Less(i, j int) bool {
	x, y := pq.list[i], pq.list[j]
	return x.fn(x.value) > y.fn(y.value)
}

func (pq *priorityQueue[T]) Swap(i, j int) {
	pq.list[i], pq.list[j] = pq.list[j], pq.list[i]
	pq.list[i].index = i
	pq.list[j].index = j
}

func (pq *priorityQueue[T]) Push(x any) {
	n := len(pq.list)
	value := x.(*item[T])
	value.index = n
	pq.list = append(pq.list, value)
}

func (pq *priorityQueue[T]) Pop() any {
	old := *pq
	n := len(old.list)
	value := old.list[n-1]
	old.list[n-1] = nil
	value.index = -1
	pq.list = old.list[0 : n-1]
	return value
}

func (pq *priorityQueue[T]) update(item *item[T], value T, fn ScoreFn[T]) {
	item.value = value
	if fn != nil {
		item.fn = fn
	}
	heap.Fix(pq, item.index)
}

type PriorityQueue[T INode] struct {
	pq      priorityQueue[T]
	store   map[int64]*item[T]
	scoreFn ScoreFn[T]
}

func New[T INode](fn ScoreFn[T]) *PriorityQueue[T] {
	queue := &PriorityQueue[T]{
		pq:      priorityQueue[T]{},
		store:   map[int64]*item[T]{},
		scoreFn: fn,
	}

	return queue
}

func (q *PriorityQueue[T]) Push(val T) {
	it := &item[T]{
		value: val,
		fn:    q.scoreFn,
	}
	if v, ok := q.store[val.ID()]; ok {
		v.value = val
		q.pq.update(v, val, q.scoreFn)
	} else {
		heap.Push(&q.pq, it)
		q.store[val.ID()] = it
	}
}

func (q *PriorityQueue[T]) Pop() (any, bool) {
	if q.pq.Len() == 0 {
		return nil, false
	}
	x := heap.Pop(&q.pq)
	it := x.(*item[T])
	delete(q.store, it.value.ID())
	return it.value, true
}

func (q *PriorityQueue[T]) Get(id int64) (any, bool) {
	v, ok := q.store[id]
	if !ok {
		return nil, false
	}
	return v.value, true
}

func (q *PriorityQueue[T]) Remove(id int64) (any, bool) {
	v, ok := q.store[id]
	if !ok {
		return nil, false
	}
	idx := v.index
	x := heap.Remove(&q.pq, idx)
	v = x.(*item[T])
	delete(q.store, id)
	return v.value, true
}

func (q *PriorityQueue[T]) Set(val T) {
	v, ok := q.store[val.ID()]
	if !ok {
		q.Push(val)
		return
	}
	v.value = val
	q.pq.update(v, val, q.scoreFn)
	q.store[val.ID()] = v
}

func (q *PriorityQueue[T]) Len() int {
	return q.pq.Len()
}

type SyncPriorityQueue[T INode] struct {
	mu sync.RWMutex
	pq *PriorityQueue[T]
}

func NewSync[T INode](fn ScoreFn[T]) *SyncPriorityQueue[T] {
	pq := New(fn)
	queue := &SyncPriorityQueue[T]{pq: pq}

	return queue
}

func (q *SyncPriorityQueue[T]) Push(val T) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.pq.Push(val)
}

func (q *SyncPriorityQueue[T]) Pop() (any, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	result, ok := q.pq.Pop()
	return result, ok
}

func (q *SyncPriorityQueue[T]) Get(id int64) (any, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	result, ok := q.pq.Get(id)
	return result, ok
}

func (q *SyncPriorityQueue[T]) Remove(id int64) (any, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	result, ok := q.pq.Remove(id)
	return result, ok
}

func (q *SyncPriorityQueue[T]) Set(val T) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.pq.Set(val)
}

func (q *SyncPriorityQueue[T]) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.pq.Len()
}
