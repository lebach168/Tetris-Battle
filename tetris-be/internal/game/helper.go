package game

import (
	"errors"
)

// This is fixed size circular queue stores cap/2 latest data state
type FrameQueue struct {
	data []*BoardState

	garbage  []int // size > 45*2 (delay 45 frame +- head and tail
	cancel   []int
	simFrame int //simulation frame index
	cap      int
	size     int
}
type Attack struct {
	lines   int
	atFrame int
}

func NewQueue(cap int) *FrameQueue {
	this := &FrameQueue{
		data:     make([]*BoardState, cap),
		garbage:  make([]int, cap),
		cancel:   make([]int, cap), // The line sweep algorithm is applied to simulate the expiration of canceled upcoming garbage lines.
		simFrame: 0,
		cap:      cap,
		size:     0,
	}
	for i := 0; i < cap; i++ {
		this.data[i] = NewDefaultBoardState()
	}
	return this
}

func (q *FrameQueue) Get(frame int) (*BoardState, error) {
	diff := frame - q.simFrame
	if diff > q.cap/2 || diff < -q.cap/2 {
		return nil, errors.New("out of range")
	}
	return q.data[frame%q.cap], nil
}
func (q *FrameQueue) GetGarbage(frame int) int {
	diff := frame - q.simFrame
	if diff > q.cap/2 || diff < -q.cap/2 {
		return -1
	}
	lines := q.garbage[frame%q.cap]

	q.garbage[frame%q.cap] = 0 // clear old state after apply
	return lines
}
func (q *FrameQueue) GetCancel(frame int) int {
	lines := q.garbage[frame%q.cap]
	q.cancel[frame%q.cap] = 0 // clear old state after apply
	return lines

}
func (q *FrameQueue) GarbageUpcoming(frame int, qty int) error {
	diff := frame - q.simFrame
	if diff > q.cap/2 || diff < -q.cap/2 {
		return errors.New("out of range")
	}
	q.garbage[(frame+INCOMING)%q.cap] += qty
	return nil
}
func (q *FrameQueue) CancelGarbage(frame int, qty int) error {
	diff := frame - q.simFrame
	if diff > q.cap/2 || diff < -q.cap/2 {
		return errors.New("out of range")
	}
	q.cancel[frame%q.cap] -= qty
	q.cancel[(frame+INCOMING)%q.cap] += qty //expire
	return nil
}

func (q *FrameQueue) Forward() {
	q.simFrame++
}
func (q *FrameQueue) Size() int {
	return q.size
}
