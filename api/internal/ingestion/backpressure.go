package ingestion

import (
	"context"
	"fmt"
	"sync"
	"time"
)

/* BackpressureMonitor monitors and handles backpressure */
type BackpressureMonitor struct {
	maxConcurrentJobs int
	currentJobs       int
	queueSize         int
	maxQueueSize      int
	mu                sync.Mutex
	throttleDelay     time.Duration
}

/* NewBackpressureMonitor creates a new backpressure monitor */
func NewBackpressureMonitor(maxConcurrentJobs, maxQueueSize int) *BackpressureMonitor {
	return &BackpressureMonitor{
		maxConcurrentJobs: maxConcurrentJobs,
		maxQueueSize:      maxQueueSize,
		throttleDelay:     100 * time.Millisecond,
	}
}

/* AcquireSlot acquires a slot for job execution */
func (b *BackpressureMonitor) AcquireSlot(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	for b.currentJobs >= b.maxConcurrentJobs {
		b.mu.Unlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(b.throttleDelay):
			b.mu.Lock()
		}
	}

	b.currentJobs++
	return nil
}

/* ReleaseSlot releases a slot after job completion */
func (b *BackpressureMonitor) ReleaseSlot() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.currentJobs > 0 {
		b.currentJobs--
	}
}

/* CheckQueueSize checks if queue is at capacity */
func (b *BackpressureMonitor) CheckQueueSize() (bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.queueSize >= b.maxQueueSize {
		return false, fmt.Errorf("queue at capacity: %d/%d", b.queueSize, b.maxQueueSize)
	}
	return true, nil
}

/* IncrementQueue increments queue size */
func (b *BackpressureMonitor) IncrementQueue() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.queueSize++
}

/* DecrementQueue decrements queue size */
func (b *BackpressureMonitor) DecrementQueue() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.queueSize > 0 {
		b.queueSize--
	}
}

/* GetMetrics returns current backpressure metrics */
func (b *BackpressureMonitor) GetMetrics() (currentJobs, queueSize int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.currentJobs, b.queueSize
}

/* SetThrottleDelay sets the throttle delay */
func (b *BackpressureMonitor) SetThrottleDelay(delay time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.throttleDelay = delay
}
