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
	"strings"
	"time"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

// Game contains a single Secret Hitler game
type Game struct {
	Players       []*Player
	Cards         Cards
	Discarding    []Card
	FailedGovs    int
	VetoRequested bool
	Started       bool

	PresidentIndex int
	President      *Player
	Chancellor     *Player
}

// Join the given player
func (game *Game) Join(name string, conn Connection) (int, string) {
	if game.Started {
		return -3, ""
	}
	for i, player := range game.Players {
		if player == nil {
			authtoken := game.createAuthToken()
			game.Players[i] = &Player{Name: name, AuthToken: authtoken, Alive: true, Vote: VoteEmpty, Conn: conn, Game: game}
			game.Broadcast(JoinQuit{Type: TypeJoin, Name: name})
			return i, authtoken
		} else if player.Name == name {
			return -2, ""
		}
	}
	return -1, ""
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

func (game *Game) createAuthToken() string {
	cs := make([]byte, 32)
	_, err := crand.Read(cs)
	if err != nil {
		rand.Read(cs)
	}
	return base64.StdEncoding.EncodeToString(cs)
}

// PlayerCount gets the count of players in the game.
func (game *Game) PlayerCount() int {
	var i = 0
	for _, player := range game.Players {
		if player != nil {
			i++
		}
	}
	return i
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

// SpecialAction is a special action that happens when a number of facist policies have been enacted
type SpecialAction int

// All special actions
const (
	Nothing           SpecialAction = iota
	PolicyPeek        SpecialAction = iota
	InvestigatePlayer SpecialAction = iota
	SelectPresident   SpecialAction = iota
	Execution         SpecialAction = iota
)

// GetAction gets the action that should happen now.
func (game *Game) GetAction() SpecialAction {
	switch game.PlayerCount() {
	case 5:
		fallthrough
	case 6:
		switch game.Cards.TableFacist {
		case 3:
			return PolicyPeek
		case 4:
			return Execution
		case 5:
			return Execution
		}
	case 7:
		fallthrough
	case 8:
		switch game.Cards.TableFacist {
		case 2:
			return InvestigatePlayer
		case 3:
			return SelectPresident
		case 4:
			return Execution
		case 5:
			return Execution
		}
	case 9:
		fallthrough
	case 10:
		switch game.Cards.TableFacist {
		case 1:
			return InvestigatePlayer
		case 2:
			return InvestigatePlayer
		case 3:
			return SelectPresident
		case 4:
			return Execution
		case 5:
			return Execution
		}
	}
	return Nothing
}

// Facists returns the recommended amount of facist players
func (game *Game) Facists() int {
	liberals := game.Liberals()
	if liberals == -1 {
		return -1
	}
	return game.PlayerCount() - liberals - 1
}

// CreateGame creates a game with the default cards and max 10 players
func CreateGame() *Game {
	return &Game{Players: make([]*Player, 10), Cards: CreateDeck()}
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

// Cards contains all the cards in the game.
type Cards struct {
	DeckLiberal      int
	DeckFacist       int
	DiscardedLiberal int
	DiscardedFacist  int
	TableLiberal     int
	TableFacist      int
}

// Card is a single card (facist or liberal)
type Card string

// The possible card types
const (
	CardLiberal Card = "liberal"
	CardFacist  Card = "facist"
)

// CreateDeck creates a Cards object with 6 liberal and 11 facist cards in the deck
func CreateDeck() Cards {
	return Cards{DeckLiberal: 6, DeckFacist: 11, DiscardedLiberal: 0, DiscardedFacist: 0, TableLiberal: 0, TableFacist: 0}
}

// PickCards picks `n` random cards from the deck
func (cards Cards) PickCards(n int) []Card {
	picked := make([]Card, n)
	for i := 0; i < n; i++ {
		if cards.TableFacist == 0 && cards.TableLiberal == 0 {
			cards.ResetDiscarded()
		}
		if cards.TableFacist == 0 {
			picked[i] = CardLiberal
			cards.TableLiberal--
		} else if cards.TableLiberal == 0 {
			picked[i] = CardFacist
			cards.TableFacist--
		} else {
			if rand.Int()%2 == 0 {
				picked[i] = CardLiberal
				cards.TableLiberal--
			} else {
				picked[i] = CardFacist
				cards.TableFacist--
			}
		}
	}
	return picked
}

// DeckSize returns the amount of cards in the deck
func (cards Cards) DeckSize() int {
	return cards.DeckLiberal + cards.DeckFacist
}

// DiscardedSize returns the amount of discarded cards
func (cards Cards) DiscardedSize() int {
	return cards.DiscardedLiberal + cards.DiscardedFacist
}

// ResetDiscarded moves all discarded cards back to the deck
func (cards Cards) ResetDiscarded() {
	cards.DeckLiberal += cards.DiscardedLiberal
	cards.DeckFacist += cards.DiscardedFacist
	cards.DiscardedLiberal = 0
	cards.DiscardedFacist = 0
}

// Player is a single player in a single Secret Hitler game
type Player struct {
	Role      Role
	Name      string
	AuthToken string
	Alive     bool
	Vote      Vote
	Conn      Connection
	Game      *Game
}

// ReceiveMessage should be called by the connection when the client sends a message
func (player *Player) ReceiveMessage(msg map[string]string) {
	game := player.Game
	if msg["type"] == TypeChat.String() {
		game.Broadcast(Chat{Type: TypeChat, Sender: player.Name, Message: msg["message"]})
	} else if msg["type"] == TypeVote.String() {
		game.Vote(player, msg["vote"])
	} else if msg["type"] == TypePickChancellor.String() && game.President == player {
		game.PickChancellor(msg["name"])
	} else if msg["type"] == TypeDiscard.String() && (game.President == player || game.Chancellor == player) && len(game.Discarding) == 2 {
		game.DiscardCard(msg["index"])
	} else if msg["type"] == TypeVetoRequest.String() && game.Chancellor == player {
		game.VetoRequest()
	} else if msg["type"] == TypeVetoAccept.String() && game.President == player && game.VetoRequested {
		game.VetoAccept()
	}
}

// Connection is a way to send messages to a player
type Connection interface {
	SendMessage(msg interface{})
}

// Vote is a simple yes/no vote
type Vote string

// ParseVote creates a Vote from the given string
func ParseVote(vote string) Vote {
	switch strings.ToLower(vote) {
	case "ja":
		return VoteJa
	case "nein":
		return VoteNein
	default:
		return VoteEmpty
	}
}

// The possible votes
const (
	VoteEmpty Vote = ""
	VoteJa    Vote = "ja"
	VoteNein  Vote = "nein"
)

// Role is the role of a player
type Role string

// The possible roles
const (
	RoleLiberal Role = "liberal"
	RoleFacist  Role = "facist"
	RoleHitler  Role = "hitler"
)
