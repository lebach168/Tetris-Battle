package game

import (
	"errors"
	"fmt"
	"sync"
)

type GameState struct {
	frames map[string]*FrameQueue // map key : player id
	/*
		If the gap between curFrame and gl.serverFrame becomes large because the client hasn't sent periodic
		updates—  the server will auto simulate the frames state for that duration.
	*/
	curFrame     int //lastest confirmed serverFrame, if diff from curFrame and gl.serverFrame
	listBlock    []int
	mu           sync.Mutex
	inputCounter int
}

func NewGameState() *GameState {
	return &GameState{
		frames:    make(map[string]*FrameQueue),
		curFrame:  0,
		listBlock: make([]int, 0),
	}
}

// This is fixed size circular queue stores 30 latest data state
type FrameQueue struct {
	data      []*PlayerState
	headFrame int //headframe is a pointer on latest frame has confirmed
	head      int //head index
	tail      int //tail index
	cap       int
	size      int
}

// NewQueue creates and returns a new generic Queue
func NewQueue(cap int) *FrameQueue {
	this := &FrameQueue{
		data:      make([]*PlayerState, cap),
		headFrame: 0,
		head:      0,
		cap:       cap,
		size:      0,
	}
	for i := 0; i < cap; i++ {
		this.data[i] = NewDefaultPlayerState()
	}
	return this
}

func (q *FrameQueue) PlayerStateOfFrame(frame int) (*PlayerState, error) {
	diff := q.headFrame - frame
	if diff > 30 || diff < -30 {
		return nil, errors.New("out of range")
	}
	return q.data[(q.head-diff+q.cap)%q.cap], nil
}

// Dequeue removes and returns the front element of the queue
func (q *FrameQueue) Poll() (*PlayerState, error) {
	var state *PlayerState
	if q.IsEmpty() {
		return state, fmt.Errorf("queue is empty")
	}
	q.tail = (q.tail + 1) % q.cap
	q.size--
	return q.data[q.tail], nil
}
func (q *FrameQueue) Pop() (*PlayerState, error) {
	var state *PlayerState
	if q.IsEmpty() {
		return state, fmt.Errorf("queue is empty")
	}
	q.size--
	q.head = (q.head - 1 + q.cap) % q.cap
	return q.data[q.head], nil
}

// Front returns the front element without removing it
func (q *FrameQueue) Peek() (*PlayerState, error) {
	var zero *PlayerState
	if q.IsEmpty() {
		return zero, fmt.Errorf("queue is empty")
	}
	return q.data[q.head], nil
}

// IsEmpty checks if the queue is empty
func (q *FrameQueue) IsEmpty() bool {
	return q.size == 0
}

// Size returns the number of elements in the queue
func (q *FrameQueue) Size() int {
	return q.size
}
