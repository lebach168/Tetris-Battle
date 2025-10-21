package game

import (
	"fmt"
	"time"
)

const defaultTicks = 30 // 30tick per second ~ 33.33ms
type GameLoop struct {
	serverFrame int
	playerId    string
	quit        chan struct{}
	pause       chan struct{}
	resume      chan struct{}
	tick        time.Duration //t per sec
	tickerC     <-chan time.Time
	onUpdate    func(string, chan Packet)
}

func NewGameLoop(playerId string, onUpdate func(string, chan Packet)) *GameLoop {
	return &GameLoop{
		serverFrame: 0,
		playerId:    playerId,
		quit:        make(chan struct{}),
		pause:       make(chan struct{}),
		resume:      make(chan struct{}),
		tick:        defaultTicks,
		tickerC:     nil,
		onUpdate:    onUpdate,
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
			//compute serverFrame every tick

			gl.onUpdate(gl.playerId, broadcast)
			gl.serverFrame++
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
