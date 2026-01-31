package execution

import (
	"container/heap"
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

/* PriorityQueue provides priority-based job queue functionality */
type PriorityQueue struct {
	mu    sync.RWMutex
	items *priorityQueueHeap
}

/* NewPriorityQueue creates a new priority queue */
func NewPriorityQueue() *PriorityQueue {
	pq := &PriorityQueue{
		items: &priorityQueueHeap{},
	}
	heap.Init(pq.items)
	return pq
}

/* JobItem represents a job in the priority queue */
type JobItem struct {
	ID          uuid.UUID
	Priority    int       // Higher number = higher priority
	CreatedAt   time.Time
	ExecuteAt   *time.Time // Optional scheduled execution time
	JobType     string
	JobData     interface{}
	index       int // Internal heap index
}

/* Push adds a job to the priority queue */
func (pq *PriorityQueue) Push(job *JobItem) {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	
	heap.Push(pq.items, job)
}

/* Pop removes and returns the highest priority job */
func (pq *PriorityQueue) Pop() *JobItem {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	
	if pq.items.Len() == 0 {
		return nil
	}
	
	item := heap.Pop(pq.items).(*JobItem)
	return item
}

/* Peek returns the highest priority job without removing it */
func (pq *PriorityQueue) Peek() *JobItem {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	
	if pq.items.Len() == 0 {
		return nil
	}
	
	return (*pq.items)[0]
}

/* Len returns the number of items in the queue */
func (pq *PriorityQueue) Len() int {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	
	return pq.items.Len()
}

/* Remove removes a specific job from the queue */
func (pq *PriorityQueue) Remove(jobID uuid.UUID) bool {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	
	for i, item := range *pq.items {
		if item.ID == jobID {
			heap.Remove(pq.items, i)
			return true
		}
	}
	return false
}

/* priorityQueueHeap implements heap.Interface */
type priorityQueueHeap []*JobItem

func (pq priorityQueueHeap) Len() int { return len(pq) }

func (pq priorityQueueHeap) Less(i, j int) bool {
	// Higher priority first, then by creation time
	if pq[i].Priority != pq[j].Priority {
		return pq[i].Priority > pq[j].Priority
	}
	
	// If priorities are equal, earlier created items first
	return pq[i].CreatedAt.Before(pq[j].CreatedAt)
}

func (pq priorityQueueHeap) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueueHeap) Push(x interface{}) {
	n := len(*pq)
	item := x.(*JobItem)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueueHeap) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

/* PriorityQueueManager manages multiple priority queues */
type PriorityQueueManager struct {
	queues map[string]*PriorityQueue
	mu     sync.RWMutex
}

/* NewPriorityQueueManager creates a new priority queue manager */
func NewPriorityQueueManager() *PriorityQueueManager {
	return &PriorityQueueManager{
		queues: make(map[string]*PriorityQueue),
	}
}

/* GetQueue gets or creates a priority queue for a tenant/resource */
func (pqm *PriorityQueueManager) GetQueue(name string) *PriorityQueue {
	pqm.mu.Lock()
	defer pqm.mu.Unlock()
	
	if queue, exists := pqm.queues[name]; exists {
		return queue
	}
	
	queue := NewPriorityQueue()
	pqm.queues[name] = queue
	return queue
}

/* RemoveQueue removes a queue */
func (pqm *PriorityQueueManager) RemoveQueue(name string) {
	pqm.mu.Lock()
	defer pqm.mu.Unlock()
	
	delete(pqm.queues, name)
}

/* GetQueueStats returns statistics for all queues */
func (pqm *PriorityQueueManager) GetQueueStats(ctx context.Context) map[string]interface{} {
	pqm.mu.RLock()
	defer pqm.mu.RUnlock()
	
	stats := make(map[string]interface{})
	for name, queue := range pqm.queues {
		stats[name] = map[string]interface{}{
			"length": queue.Len(),
		}
	}
	return stats
}
