package model

type Room struct {
	ID string   `json:"id"`
	P  []Player `json:"p"`
}

func (r *Room) AddPlayer(playerID string) {
	r.P = append(r.P, Player{
		ID: playerID,
	})
}
