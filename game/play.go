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
	"strconv"
)

// Start the game already!
func (game *Game) Start() {
	game.Started = true

	game.PresidentIndex = r.Intn(game.PlayerCount())

	facistsAvailable := game.Facists()
	liberalsAvailable := game.Liberals()
	hitlerAvailable := true
	for _, player := range game.Players {
		var availableRoles []Role
		if !hitlerAvailable && facistsAvailable == 0 {
			availableRoles = []Role{RoleLiberal}
		} else if !hitlerAvailable {
			availableRoles = []Role{RoleLiberal, RoleFacist}
		} else if facistsAvailable == 0 {
			availableRoles = []Role{RoleLiberal, RoleHitler}
		} else {
			availableRoles = []Role{RoleLiberal, RoleFacist, RoleHitler}
		}

		player.Role = availableRoles[r.Intn(len(availableRoles))]
		player.Conn.SendMessage(Start{Type: TypeStart, Role: player.Role})

		switch player.Role {
		case RoleLiberal:
			liberalsAvailable--
		case RoleFacist:
			facistsAvailable--
		case RoleHitler:
			hitlerAvailable = false
		}
	}
}

// NextPresident moves the game to the next president
func (game *Game) NextPresident() {
	if game.PlayerCount() < 4 {
		game.Error("Not enough players left")
	}
	game.PresidentIndex++
	if game.PresidentIndex >= 10 {
		game.PresidentIndex = 0
	}
	game.President = game.Players[game.PresidentIndex]
	if game.President == nil {
		game.NextPresident()
		return
	}
	game.Broadcast(President{Type: TypePresident, Name: game.President.Name})
}

// PickChancellor is called when the president picks his/her chancellor
func (game *Game) PickChancellor(name string) {
	for _, player := range game.Players {
		if player.Name == name {
			game.Chancellor = player
			game.Broadcast(StartVote{Type: TypeStartVote, President: game.President.Name, Chancellor: game.Chancellor.Name})
			return
		}
	}
}

// Vote is called when the player sends a vote command
func (game *Game) Vote(player *Player, vote string) {
	player.Vote = ParseVote(vote)
	player.Conn.SendMessage(VoteMessage{Type: TypeVote, Vote: player.Vote})

	var ja, nein = 0, 0
	for _, player := range game.Players {
		switch player.Vote {
		case VoteEmpty:
			return
		case VoteJa:
			ja++
		case VoteNein:
			nein++
		}
	}
	for _, player := range game.Players {
		player.Vote = VoteEmpty
	}
	if ja > nein {
		game.StartDiscard()
	} else {
		game.FailedGovs++
		if game.FailedGovs >= 3 {
			game.ThreeGovsFailed()
		} else {
			game.NextPresident()
		}
	}
}

// ThreeGovsFailed is called when three consequent government attempts have been downvoted
func (game *Game) ThreeGovsFailed() {
	card := game.Cards.PickCards(1)[0]
	game.Broadcast(EnactForce{Type: TypeEnactForce, Policy: card})
	game.Enact(card)
}

// StartDiscard is executed when everyone has voted and accepted the president
func (game *Game) StartDiscard() {
	if game.Cards.TableFacist >= 3 {
		if game.Chancellor.Role == RoleHitler {
			game.End(CardFacist)
			return
		}
	}

	game.FailedGovs = 0
	game.Broadcast(Discard{Type: TypePresidentDiscard, Name: game.President.Name})
	game.BroadcastTable()
	game.Discarding = game.Cards.PickCards(3)
	game.President.Conn.SendMessage(CardsMessage{Type: TypeCards, Cards: game.Discarding})
}

// DiscardCard is called when the chancellor or president discards a card
func (game *Game) DiscardCard(c string) {
	game.VetoRequested = false
	card, err := strconv.Atoi(c)
	if err != nil || card >= len(game.Discarding) || card < 0 {
		return
	}
	switch game.Discarding[card] {
	case CardFacist:
		game.Cards.DiscardedFacist++
	case CardLiberal:
		game.Cards.DiscardedLiberal++
	}
	game.Discarding[card] = game.Discarding[len(game.Discarding)-1]
	game.Discarding = game.Discarding[:len(game.Discarding)-1]

	game.BroadcastTable()
	if len(game.Discarding) == 2 {
		game.Broadcast(Discard{Type: TypeChancellorDiscard, Name: game.Chancellor.Name})
		game.Chancellor.Conn.SendMessage(CardsMessage{Type: TypeCards, Cards: game.Discarding})
	} else if len(game.Discarding) == 1 {
		game.Broadcast(Enact{Type: TypeEnact, President: game.President.Name, Chancellor: game.Chancellor.Name, Policy: game.Discarding[0]})
		game.Enact(game.Discarding[0])
	} else {
		game.Error("Invalid amount of cards to discard")
	}
}

// VetoRequest is called when the chancellor wants to veto the current discard
func (game *Game) VetoRequest() {
	game.VetoRequested = true
	game.Broadcast(Veto{Type: TypeVetoRequest, President: game.President.Name, Chancellor: game.Chancellor.Name})
}

// VetoAccept is called when the president accepts the chancellors veto request
func (game *Game) VetoAccept() {
	game.VetoRequested = false
	game.Broadcast(Veto{Type: TypeVetoAccept, President: game.President.Name, Chancellor: game.Chancellor.Name})

	for _, card := range game.Discarding {
		switch card {
		case CardLiberal:
			game.Cards.DiscardedLiberal++
		case CardFacist:
			game.Cards.DiscardedFacist++
		}
	}
	game.BroadcastTable()
	game.Discarding = []Card{}

	game.FailedGovs++
	if game.FailedGovs >= 3 {
		game.ThreeGovsFailed()
	} else {
		game.NextPresident()
	}
}

// Enact is called when a card is enacted
func (game *Game) Enact(card Card) {
	switch card {
	case CardFacist:
		game.Cards.TableFacist++
	case CardLiberal:
		game.Cards.TableLiberal++
	}
	game.BroadcastTable()
	if game.Cards.TableFacist >= 6 {
		game.End(CardFacist)
		return
	} else if game.Cards.TableLiberal >= 5 {
		game.End(CardLiberal)
		return
	}

	// TODO special actions
	switch game.GetAction() {
	case PolicyPeek:
	case InvestigatePlayer:
	case SelectPresident:
	case Execution:
	case Nothing:
		game.NextPresident()
	}
}

func (game *Game) Error(msg string) {
	game.Broadcast(Error{Type: TypeError, Message: msg})
}

// End the game with the given winner
func (game *Game) End(winner Card) {
	var end = End{Type: TypeEnd, Winner: winner, Roles: make(map[string]Role)}
	for _, player := range game.Players {
		end.Roles[player.Name] = player.Role
	}
	game.Broadcast(end)
}
