package models

import (
	"time"
)

type Participant struct {
	UserID   int    `json:"userId"`
	Username string `json:"username"`
	Credits  int    `json:"credits"`
	IsActive bool   `json:"isActive"`
	Score    int    `json:"score"`
}

type GameProperties struct {
	Columns  int `json:"columns"`
	MaxChips int `json:"maxChips"`
}

type GameDetails struct {
	Properties GameProperties `json:"properties"`
	Board      [][]int        `json:"board"`
	Stroke     [][]int        `json:"stroke"`
	Type       string         `json:"type"`
}

type Game struct {
	ID                string        `json:"id"`
	Participants      []Participant `json:"participants"`
	MaxParticipants   int           `json:"maxParticipants"`
	State             string        `json:"state"`
	LastTurnTimestamp int64         `json:"lastTurnTimestamp"`
	TurnTimeout       int           `json:"turnTimeout"`
	Game              GameDetails   `json:"game"`
}

type Turn struct {
	ColumnID      int    `json:"columnId"`
	GameSessionID string `json:"gameSessionId"`
}

func NewGameManager() *Game {
	return &Game{
		ID:                "",
		Participants:      []Participant{},
		MaxParticipants:   0,
		State:             "",
		LastTurnTimestamp: time.Now().Unix(),
		TurnTimeout:       0,
		Game: GameDetails{
			Properties: GameProperties{
				Columns:  0,
				MaxChips: 0,
			},
			Board:  [][]int{},
			Stroke: [][]int{},
			Type:   "",
		},
	}
}
