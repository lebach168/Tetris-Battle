package game

import (
	"fmt"
	"time"
)

const defaultTicks = 30 // 30tick per second ~ 33.33ms
type GameLoop struct {
	tickFrame int
	quit      chan struct{}
	pause     chan struct{}
	resume    chan struct{}
	tick      time.Duration //t per sec
	tickerC   <-chan time.Time
	input     chan Message
	attacked  chan Attack
	//callback
	onUpdate       func(chan Packet)
	recordInputs   func([]Input, int, chan Packet)
	receiveGarbage func(Attack)
}

func NewGameLoop(onUpdate func(chan Packet), recordInputs func([]Input, int, chan Packet),
	receiveGarbage func(attack Attack)) *GameLoop {
	return &GameLoop{
		tickFrame:      0,
		quit:           make(chan struct{}),
		pause:          make(chan struct{}),
		resume:         make(chan struct{}),
		tick:           defaultTicks,
		tickerC:        nil,
		input:          make(chan Message),
		attacked:       make(chan Attack),
		onUpdate:       onUpdate,
		recordInputs:   recordInputs,
		receiveGarbage: receiveGarbage,
	}
}
func (gl *GameLoop) NewTicker() *time.Ticker {
	tickInterval := time.Second / gl.tick
	return time.NewTicker(tickInterval)
}

func (gl *GameLoop) Run(broadcast chan Packet) {

	ticker := gl.NewTicker()
	gl.tickerC = ticker.C
	for {
		select {
		case <-gl.tickerC:
			//compute tickFrame every tick
			gl.onUpdate(broadcast)
			gl.tickFrame++
		case msg := <-gl.input:
			gl.recordInputs(msg.Payload.Inputs, msg.Payload.LatestFrame, broadcast)
		case atk := <-gl.attacked:
			gl.receiveGarbage(atk)

		case <-gl.pause:
			ticker.Stop()
			gl.tickerC = nil
			<-gl.resume
			ticker = gl.NewTicker()
			gl.tickerC = ticker.C
		case <-gl.quit:
			ticker.Stop()
			fmt.Println("end loop") //debug
			gl.tickerC = nil
			return
		}
	}
}
func (g *GameLoop) startLoop() {

}
func (g *GameLoop) restart() {

}
