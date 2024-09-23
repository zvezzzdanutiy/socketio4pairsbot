package models

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type SocketResponse struct {
	SID          string   `json:"sid"`
	Upgrades     []string `json:"upgrades"`
	PingInterval int      `json:"pingInterval"`
	PingTimeout  int      `json:"pingTimeout"`
	MaxPayload   int      `json:"maxPayload"`
}
type GameManager struct {
	KnownCards map[int][]int
	KnownLens  map[int]int
	Random     *rand.Rand
	HttpClient *http.Client
	Conn       *websocket.Conn
}
type GameResponse struct {
	State       string `json:"state"`
	GameSession struct {
		ID                string        `json:"id"`
		Participants      []Participant `json:"participants"`
		MaxParticipants   int           `json:"maxParticipants"`
		State             string        `json:"state"`
		LastTurnTimestamp int64         `json:"lastTurnTimestamp"`
		TurnTimeout       int           `json:"turnTimeout"`
		Game              Game          `json:"game"`
	} `json:"gameSession"`
}

type Participant struct {
	UserID   int    `json:"userId"`
	Username string `json:"username"`
	Credits  int    `json:"credits"`
	IsActive bool   `json:"isActive"`
	Score    int    `json:"score"`
}

type Game struct {
	Properties struct {
		Columns   int `json:"columns"`
		Pairs     int `json:"pairs"`
		MaxCardID int `json:"maxCardId"`
	} `json:"properties"`
	Board  [][]int `json:"board,omitempty"`
	Scores []int   `json:"scores,omitempty"`
	Opened []int   `json:"opened,omitempty"`
}

func NewGameManager(httpClient *http.Client) *GameManager {
	return &GameManager{
		KnownCards: make(map[int][]int),
		KnownLens:  make(map[int]int),
		Random:     rand.New(rand.NewSource(time.Now().UnixNano())),
		HttpClient: httpClient,
	}
}
