// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a permission.

package manage

import (
	"sync"

	"github.com/repofuel/repofuel/ingest/pkg/identifier"
)

const (
	DefaultQueueWorkersCount   = 4
	DefaultManagerWorkersCount = 8
)

type QueueID uint8

const (
	QueueNotSpecify QueueID = iota
	QueueNewRepos
	QueuePullRequests
	QueueNewCommits
	QueueRecovered
)

var queues = map[QueueID]string{
	QueueNewRepos:     "New Repositories Queue",
	QueuePullRequests: "Fetch Requests Queue",
	QueueNewCommits:   "New Commits Queue",
	QueueRecovered:    "Recovered Repositories",
}

func (id QueueID) String() string {
	return queues[id]
}

type Queue struct {
	NumWorkers int
	processing int
	items      []identifier.RepositoryID

	mu sync.Mutex
}

type QueueRegistry struct {
	items map[QueueID]*Queue
	mu    sync.Mutex
}

func newQueueRegistry() *QueueRegistry {
	return &QueueRegistry{
		items: make(map[QueueID]*Queue),
	}
}

func (set *QueueRegistry) GetOrCreate(id QueueID) *Queue {
	set.mu.Lock()
	defer set.mu.Unlock()

	item, ok := set.items[id]
	if !ok {
		item = &Queue{NumWorkers: DefaultQueueWorkersCount}
		set.items[id] = item
	}
	return item
}

//var ErrQueueIsEmpty = errors.New("cannot pup from an empty queue")

func (q *Queue) pup() identifier.RepositoryID {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		return identifier.RepositoryID{}
	}

	i := q.items[0]
	q.items = q.items[1:]
	return i
}

func (q *Queue) NumWaiting() int {
	return len(q.items)
}

func (q *Queue) NumProcessing() int {
	return q.processing
}

func (q *Queue) HasFreeWorkers() bool {
	return q.NumWorkers > q.processing
}

func (q *Queue) processStarted() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.processing++
}

func (q *Queue) processDone() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.processing--
}

func (q *Queue) push(i identifier.RepositoryID) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.items = append(q.items, i)
}

func (q *Queue) removeRepository(id identifier.RepositoryID) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i := range q.items {
		if q.items[i] == id {
			// remove and free up the memory
			if i < len(q.items)-1 {
				copy(q.items[i:], q.items[i+1:])
			}
			q.items = q.items[:len(q.items)-1]
		}
	}
}
