package game

import (
	"fmt"
	"time"
)

const BOARD_WIDTH = 10
const BOARD_HEIGHT = 22
const SOFT_DROP = 100
const LOCKDELAY float64 = 300
const DROPSPEED float64 = 800
const QUEUE_SIZE = 100
const TICK = 30
const INCOMING = 45

var INTERVAL = float64(time.Second.Milliseconds()) / float64(time.Duration(TICK))

var Tetromino = map[int]Block{
	1: { //I
		shape: [][]int{
			{0, 0, 0, 0},
			{1, 1, 1, 1},
			{0, 0, 0, 0},
			{0, 0, 0, 0},
		},
		form: 0,
	},
	2: { //O
		shape: [][]int{
			{2, 2},
			{2, 2},
		},
		form: 0,
	},
	3: { //T
		shape: [][]int{
			{0, 3, 0},
			{3, 3, 3},
			{0, 0, 0},
		},
		form: 0,
	},

	4: { //Z
		shape: [][]int{
			{4, 4, 0},
			{0, 4, 4},
			{0, 0, 0},
		},
		form: 0,
	},

	5: { //L
		shape: [][]int{
			{0, 0, 5},
			{5, 5, 5},
			{0, 0, 0},
		},
		form: 0,
	},
	6: { //S
		shape: [][]int{
			{0, 6, 6},
			{6, 6, 0},
			{0, 0, 0},
		},
		form: 0,
	},
	7: { //J
		shape: [][]int{
			{7, 0, 0},
			{7, 7, 7},
			{0, 0, 0},
		},
		form: 0,
	},
}

var ReverseTetrominoMap = map[int]string{
	1: "I", 2: "O", 3: "T", 4: "Z", 5: "L", 6: "S", 7: "J",
}

type Block struct {
	shape [][]int
	form  int //0,1,2,3,
}
type KickOffsets [][2]int
type WallKickTable map[string]KickOffsets

var WallKickJLSTZ = WallKickTable{
	"0->1": {{0, 0}, {-1, 0}, {-1, 1}, {0, -2}, {-1, -2}}, // 0->R
	"1->2": {{0, 0}, {1, 0}, {1, -1}, {0, 2}, {1, 2}},     // R->2
	"2->3": {{0, 0}, {1, 0}, {1, 1}, {0, -2}, {1, -2}},    // 2->L
	"3->0": {{0, 0}, {-1, 0}, {-1, -1}, {0, 2}, {-1, 2}},  // L->0
	"1->0": {{0, 0}, {1, 0}, {1, -1}, {0, 2}, {1, 2}},     // R->0
	"2->1": {{0, 0}, {-1, 0}, {-1, 1}, {0, -2}, {-1, -2}}, // 2->R
	"3->2": {{0, 0}, {-1, 0}, {-1, -1}, {0, 2}, {-1, 2}},  // L->2
	"0->3": {{0, 0}, {1, 0}, {1, 1}, {0, -2}, {1, -2}},    // 0->L
}
var WallKickI = WallKickTable{
	"0->1": {{0, 0}, {-2, 0}, {1, 0}, {-2, -1}, {1, 2}}, // 0->R
	"1->2": {{0, 0}, {-1, 0}, {2, 0}, {-1, 2}, {2, -1}}, // R->2
	"2->3": {{0, 0}, {2, 0}, {-1, 0}, {2, 1}, {-1, -2}}, // 2->L
	"3->0": {{0, 0}, {1, 0}, {-2, 0}, {1, -2}, {-2, 1}}, // L->0
	"1->0": {{0, 0}, {2, 0}, {-1, 0}, {2, 1}, {-1, -2}}, // R->0
	"2->1": {{0, 0}, {1, 0}, {-2, 0}, {1, -2}, {-2, 1}}, // 2->R
	"3->2": {{0, 0}, {-2, 0}, {1, 0}, {-2, -1}, {1, 2}}, // L->2
	"0->3": {{0, 0}, {-1, 0}, {2, 0}, {-1, 2}, {2, -1}}, // 0->L
}
var WallKickIAkira = WallKickTable{
	"0->1": {{0, 0}, {-2, 0}, {1, 0}, {-2, -1}, {1, 2}}, // 0->R
	"1->2": {{0, 0}, {-1, 0}, {2, 0}, {-1, 2}, {2, -1}}, // R->2
	"2->3": {{0, 0}, {2, 0}, {-1, 0}, {2, 1}, {-1, -1}}, // 2->L
	"3->0": {{0, 0}, {1, 0}, {-2, 0}, {1, -2}, {-2, 1}}, // L->0
	"1->0": {{0, 0}, {2, 0}, {-1, 0}, {2, 1}, {-1, -2}}, // R->0
	"2->1": {{0, 0}, {1, 0}, {-2, 0}, {1, -2}, {1, -1}}, // 2->R
	"3->2": {{0, 0}, {-2, 0}, {1, 0}, {-2, -1}, {1, 2}}, // L->2
	"0->3": {{0, 0}, {-1, 0}, {2, 0}, {-1, 2}, {2, -1}}, // 0->L
}

func GetWallKickData(size int, fromShape, toShape int) KickOffsets {
	//if size ==2 -> tetromino O -> no need wallkick for this
	//if size == 4 -> tetromino I -> game for only I
	if size == 2 {
		return KickOffsets{{0, 0}}
	}
	k := fmt.Sprintf("%d->%d", fromShape, toShape)
	if size == 4 {
		return WallKickI[k]
	}
	return WallKickJLSTZ[k]
}
func CalculateGarbageRows(lines int, hasSpin bool, combo int, b2b bool, perfect bool) int {
	base := lines - 1
	bonus := 0
	if lines == 0 {
		return 0
	}

	if combo > 0 {
		bonus += 1
	}
	if hasSpin {
		bonus += 2
	}
	if b2b {
		if hasSpin {
			bonus += 3
		} else {
			bonus += 2
		}
	}
	if perfect {
		base += 4
	}
	return base + bonus
}
