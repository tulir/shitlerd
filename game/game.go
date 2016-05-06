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
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"
)

var debug = flag.Bool("debug", false, "Print gameplay debug/log messages")
var r = rand.New(rand.NewSource(time.Now().UnixNano()))

// Game contains a single Secret Hitler game
type Game struct {
	Name       string
	Players    []*Player
	Cards      *Cards
	Discarding []Card
	Started    bool
	Ended      bool

	VetoRequested bool
	State         Action
	FailedGovs    int

	PresidentIndex     int
	PreviousPresident  *Player
	PreviousChancellor *Player
	President          *Player
	Chancellor         *Player
}

// CreateGame creates a game with the default cards and max 10 players
func CreateGame(name string) *Game {
	return &Game{Name: name, Players: make([]*Player, 10), Cards: CreateDeck()}
}

// Join the given player
func (game *Game) Join(name, authtoken string, conn Connection) (int, *Player) {
	if game.Started && len(authtoken) == 0 {
		return -3, nil
	} else if !validName(name) {
		return -4, nil
	}
	for i, player := range game.Players {
		if player == nil {
			game.Broadcast(JoinQuit{Type: TypeJoin, Name: name})
			game.Players[i] = &Player{Name: name, AuthToken: game.createAuthToken(), Connected: true, Alive: true, Vote: VoteEmpty, Conn: conn, Game: game}
			return i, game.Players[i]
		} else if player.Name == name {
			if player.AuthToken == authtoken {
				player.Game.Broadcast(JoinQuit{Type: TypeConnected, Name: player.Name})
				if player.Conn != nil {
					player.SendMessage("connected-other")
					player.Conn.Close()
				}
				player.Conn = conn
				player.Connected = true
				return i, player
			}
			return -2, nil
		}
	}
	return -1, nil
}

func validName(name string) bool {
	return validNameLength(name) && validNameChars(name)
}

func validNameLength(name string) bool {
	if len(name) < 3 || len(name) > 16 {
		return false
	}
	return true
}

func validNameChars(name string) bool {
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			continue
		}
		return false
	}
	return true
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

// PlayersInGame gets the amount of alive players in the game
func (game *Game) PlayersInGame() (i int) {
	for _, player := range game.Players {
		if player != nil && player.Alive {
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

// Fascists returns the recommended amount of fascist players
func (game *Game) Fascists() int {
	liberals := game.Liberals()
	if liberals == -1 {
		return -1
	}
	return game.PlayerCount() - liberals - 1
}

// Broadcast a message to all players
func (game *Game) Broadcast(msg interface{}) {
	for _, player := range game.Players {
		if player != nil {
			player.SendMessage(msg)
		}
	}
}

// BroadcastTable broadcasts the current status of the table to everyone
func (game *Game) BroadcastTable() {
	game.Broadcast(Table{
		Type:         TypeTable,
		Deck:         len(game.Cards.Deck),
		Discarded:    len(game.Cards.Discarded),
		TableLiberal: game.Cards.TableLiberal,
		TableFascist: game.Cards.TableFascist,
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

// Disconnect is called when a player disconnects
func (player *Player) Disconnect() {
	player.Connected = false
	player.Conn = nil
	player.Game.Broadcast(JoinQuit{Type: TypeDisconnected, Name: player.Name})
	outln(player.Name, "has disconnected from", player.Game.Name)
}

// SendMessage sends a message to the client
func (player *Player) SendMessage(msg interface{}) {
	if player.Conn != nil {
		player.Conn.SendMessage(msg)
	}
}

// ReceiveMessage should be called by the connection when the client sends a message
func (player *Player) ReceiveMessage(msg map[string]string) {
	game := player.Game
	if msg["type"] == TypeChat.String() && player.Alive {
		game.Broadcast(Chat{Type: TypeChat, Sender: player.Name, Message: msg["message"]})
	} else if msg["type"] == TypeStart.String() && !game.Started && game.PlayerCount() >= 5 {
		game.Start()
	} else if msg["type"] == TypeQuit.String() {
		game.Leave(player.Name)
	} else if !game.Started || game.Ended || !player.Alive {
		errln(player.Name, "tried to send a", msg["type"], "message! Started:", game.Started, "Ended:", game.Ended, "Alive:", player.Alive)
		return
	} else {
		player.ReceiveGameMessage(msg)
	}
}

// ReceiveGameMessage is called from ReceiveMessage when the received message is directly related to the ongoing game.
func (player *Player) ReceiveGameMessage(msg map[string]string) {
	game := player.Game
	if msg["type"] == TypeVote.String() && TypeVote.ReceiveRequirements(player) {
		game.Vote(player, msg["vote"])
	} else if msg["type"] == TypePickChancellor.String() && TypePickChancellor.ReceiveRequirements(player) {
		game.PickChancellor(msg["name"])
	} else if msg["type"] == TypeDiscard.String() && TypeDiscard.ReceiveRequirements(player) {
		game.DiscardCard(msg["index"])
	} else if msg["type"] == TypeVetoRequest.String() && TypeVetoRequest.ReceiveRequirements(player) {
		game.VetoRequest()
	} else if msg["type"] == TypeVetoAccept.String() && TypeVetoAccept.ReceiveRequirements(player) {
		game.VetoAccept()
	} else if msg["type"] == TypePresidentSelect.String() && TypePresidentSelect.ReceiveRequirements(player) {
		game.SelectedPresident(msg["name"])
	} else if msg["type"] == TypeExecute.String() && TypeExecute.ReceiveRequirements(player) {
		game.ExecutedPlayer(msg["name"])
	} else if msg["type"] == TypeInvestigate.String() && TypeInvestigate.ReceiveRequirements(player) {
		game.Investigated(msg["name"])
	}
}

// Connection is a way to send messages to a player
type Connection interface {
	SendMessage(msg interface{})
	Close()
}

func outln(msg ...interface{}) {
	os.Stderr.WriteString("[Game] ")
	fmt.Fprintln(os.Stdout, msg...)
}

func outf(msg string, args ...interface{}) {
	os.Stderr.WriteString("[Game] ")
	fmt.Fprintf(os.Stdout, msg, args...)
}

func outfp(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, msg, args...)
}

func errln(msg ...interface{}) {
	os.Stderr.WriteString("[Game] ")
	fmt.Fprintln(os.Stderr, msg...)
}

func errf(msg string, args ...interface{}) {
	os.Stderr.WriteString("[Game] ")
	fmt.Fprintf(os.Stderr, msg, args...)
}

func errfp(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg, args...)
}
