package game

import (
	"errors"
	"sync"
)

// This is fixed size circular queue stores cap/2 latest data state
type FrameQueue struct {
	data   []*PlayerState
	pFrame int //head index
	cap    int
	size   int
	mu     sync.Mutex
}

func NewQueue(cap int) *FrameQueue {
	this := &FrameQueue{
		data:   make([]*PlayerState, cap),
		pFrame: 0,
		cap:    cap,
		size:   0,
	}
	for i := 0; i < cap; i++ {
		this.data[i] = NewDefaultPlayerState()
	}
	return this
}

func (q *FrameQueue) Get(frame int) (*PlayerState, error) {
	diff := frame - q.pFrame
	if diff > q.cap/2 || diff < -q.cap/2 {
		return nil, errors.New("out of range")
	}
	return q.data[frame%q.cap], nil
}
func (q *FrameQueue) Forward() {
	q.pFrame++
}
func (q *FrameQueue) Size() int {
	return q.size
}
