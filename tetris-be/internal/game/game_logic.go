package game

import (
	"errors"
	"fmt"
)

type GameState struct {
	frames    map[string]*FrameQueue // map key : player id
	curF      int                    //pointer on lastest confirmed frame
	listBlock []int
}

func NewGameState() *GameState {
	return &GameState{
		frames:    make(map[string]*FrameQueue),
		curF:      1,
		listBlock: make([]int, 0),
	}
}

// This is fixed size circular queue stores 30 latest data state
type FrameQueue struct {
	data        []*PlayerState
	latestFrame int
	head        int //head index
	tail        int //tail index
	cap         int
	size        int
}

// NewQueue creates and returns a new generic Queue
func NewQueue(cap int) *FrameQueue {
	return &FrameQueue{
		data:        make([]*PlayerState, cap),
		latestFrame: 0,
		head:        0,
		cap:         cap,
		size:        0,
	}
}
func (q *FrameQueue) PlayerStateOfFrame(frame int) (*PlayerState, error) {
	dif := q.latestFrame - frame
	if dif > 30 {
		return nil, errors.New("outdated")
	}
	return q.data[(q.head-dif+q.cap)%q.cap], nil
}

// Enqueue adds an element to the back of the queue
func (q *FrameQueue) AddNew(value *PlayerState) {

	q.latestFrame++
	if q.size < q.cap {
		q.size++
	}

	if q.data[q.head] == nil {
		q.data[q.head] = value
	}
	*q.data[q.head] = *value
	q.head = (q.head + 1) % q.cap
	if q.tail == q.head {
		q.tail = (q.tail + 1) % q.cap
	}
}

// Add basically increase frame counter and re calculate head and tail index
func (q *FrameQueue) Add() {
	q.latestFrame++
	q.head = (q.head + 1) % q.cap
	if q.tail == q.head {
		q.tail = (q.tail + 1) % q.cap
	}
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
