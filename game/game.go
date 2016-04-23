// shitlerd - A manager for online Secret Hitler games
// Copyright (C) 2016 Tulir Asokan

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// Package game contains the game management code
package game

import (
	crand "crypto/rand"
	"encoding/base64"
	"math/rand"
	"time"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

// Game contains a single Secret Hitler game
type Game struct {
	Name       string
	Players    []*Player
	Cards      Cards
	Discarding []Card
	Started    bool
	Ended      bool

	FailedGovs    int
	VetoRequested bool
	State         Action

	PresidentIndex int
	President      *Player
	Chancellor     *Player
}

// CreateGame creates a game with the default cards and max 10 players
func CreateGame(name string) *Game {
	return &Game{Name: RandomName(), Players: make([]*Player, 10), Cards: CreateDeck()}
}

// Join the given player
func (game *Game) Join(name string, conn Connection) (int, *Player) {
	if game.Started {
		return -3, nil
	}
	for i, player := range game.Players {
		if player == nil {
			game.Players[i] = &Player{Name: name, AuthToken: game.createAuthToken(), Connected: false, Alive: true, Vote: VoteEmpty, Conn: conn, Game: game}
			game.Broadcast(JoinQuit{Type: TypeJoin, Name: name})
			return i, game.Players[i]
		} else if player.Name == name {
			return -2, nil
		}
	}
	return -1, nil
}

// Leave the given player
func (game *Game) Leave(name string) {
	for i, player := range game.Players {
		if player != nil && player.Name == name {
			if !game.Started {
				game.Players[i] = nil
			} else {
				game.Players[i].Alive = false
			}
			game.Broadcast(JoinQuit{Type: TypeQuit, Name: name})
		}
	}
}

// GetPlayer gets the given player in this game
func (game *Game) GetPlayer(name string) *Player {
	for _, player := range game.Players {
		if player != nil && player.Name == name {
			return player
		}
	}
	return nil
}

func (game *Game) createAuthToken() string {
	cs := make([]byte, 32)
	_, err := crand.Read(cs)
	if err != nil {
		rand.Read(cs)
	}
	return base64.StdEncoding.EncodeToString(cs)
}

// PlayerCount gets the count of players in the game.
func (game *Game) PlayerCount() (i int) {
	for _, player := range game.Players {
		if player != nil {
			i++
		}
	}
	return
}

// ConnectedPlayers gets the amount of connected players
func (game *Game) ConnectedPlayers() (i int) {
	for _, player := range game.Players {
		if player != nil && player.Connected {
			i++
		}
	}
	return
}

// Liberals returns the recommended amount of liberal players
func (game *Game) Liberals() int {
	switch game.PlayerCount() {
	case 5:
		return 3
	case 6:
		return 4
	case 7:
		return 4
	case 8:
		return 5
	case 9:
		return 5
	case 10:
		return 6
	}
	return -1
}

// Facists returns the recommended amount of facist players
func (game *Game) Facists() int {
	liberals := game.Liberals()
	if liberals == -1 {
		return -1
	}
	return game.PlayerCount() - liberals - 1
}

// Broadcast a message to all players
func (game *Game) Broadcast(msg interface{}) {
	for _, player := range game.Players {
		player.Conn.SendMessage(msg)
	}
}

// BroadcastTable broadcasts the current status of the table to everyone
func (game *Game) BroadcastTable() {
	game.Broadcast(Table{
		Type:         TypeTable,
		Deck:         game.Cards.DeckLiberal + game.Cards.DeckFacist,
		Discarded:    game.Cards.DiscardedLiberal + game.Cards.DiscardedFacist,
		TableLiberal: game.Cards.TableLiberal,
		TableFacist:  game.Cards.TableFacist,
	})
}

// Player is a single player in a single Secret Hitler game
type Player struct {
	Role      Role
	Name      string
	AuthToken string
	Connected bool
	Alive     bool
	Vote      Vote
	Conn      Connection
	Game      *Game
}

// Connect is called when a player connects
func (player *Player) Connect(conn Connection) {
	player.Conn = conn
	player.Connected = true
	player.Game.Broadcast(JoinQuit{Type: TypeConnected, Name: player.Name})
}

// Disconnect is called when a player disconnects
func (player *Player) Disconnect() {
	player.Connected = false
	player.Game.Broadcast(JoinQuit{Type: TypeDisconnected, Name: player.Name})
}

// ReceiveMessage should be called by the connection when the client sends a message
func (player *Player) ReceiveMessage(msg map[string]string) {
	if !player.Alive {
		return
	}
	game := player.Game
	if msg["type"] == TypeChat.String() {
		game.Broadcast(Chat{Type: TypeChat, Sender: player.Name, Message: msg["message"]})
	} else if msg["type"] == TypeVote.String() && game.State == ActVote {
		game.Vote(player, msg["vote"])
	} else if msg["type"] == TypePickChancellor.String() && game.President == player && game.State == ActPickChancellor {
		game.PickChancellor(msg["name"])
	} else if msg["type"] == TypeDiscard.String() &&
		((game.President == player && game.State == ActDiscardPresident) ||
			(game.Chancellor == player && game.State == ActDiscardChancellor)) {
		game.DiscardCard(msg["index"])
	} else if msg["type"] == TypeVetoRequest.String() && game.Chancellor == player && game.State == ActDiscardChancellor && game.Cards.TableFacist >= 5 {
		game.VetoRequest()
	} else if msg["type"] == TypeVetoAccept.String() && game.President == player && game.VetoRequested {
		game.VetoAccept()
	} else if msg["type"] == TypeStart.String() && !game.Started && game.PlayerCount() >= 5 {
		game.Start()
	}
}

// Connection is a way to send messages to a player
type Connection interface {
	SendMessage(msg interface{})
}
