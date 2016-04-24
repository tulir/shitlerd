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
	if game.PlayerCount() < 5 {
		return
	}
	game.Started = true

	game.PresidentIndex = r.Intn(len(game.Players))

	facistsAvailable := game.Facists()
	liberalsAvailable := game.Liberals()
	hitlerAvailable := true

	for _, player := range game.Players {
		if player == nil {
			continue
		}
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

		switch player.Role {
		case RoleLiberal:
			liberalsAvailable--
		case RoleFacist:
			facistsAvailable--
		case RoleHitler:
			hitlerAvailable = false
		}
	}

	var playersToLiberal = make(map[string]Role)
	var playersToFacists = make(map[string]Role)
	pc := game.PlayerCount()
	for _, player := range game.Players {
		if player == nil {
			continue
		}
		playersToLiberal[player.Name] = "unknown"
		playersToFacists[player.Name] = player.Role
	}

	for _, player := range game.Players {
		if player == nil {
			continue
		}
		if player.Role == RoleLiberal || (pc > 6 && player.Role == RoleHitler) {
			player.SendMessage(Start{Type: TypeStart, Role: player.Role, Players: playersToLiberal})
		} else if player.Role == RoleFacist || (pc < 7 && player.Role == RoleHitler) {
			player.SendMessage(Start{Type: TypeStart, Role: player.Role, Players: playersToFacists})
		}
	}
	game.NextPresident()
}

// NextPresident moves the game to the next president
func (game *Game) NextPresident() {
	if game.PlayersInGame() < 4 {
		game.Error("Not enough players left")
	}
	game.PresidentIndex++
	if game.PresidentIndex >= 10 {
		game.PresidentIndex = 0
	}
	if game.Players[game.PresidentIndex] == nil || !game.Players[game.PresidentIndex].Alive {
		game.NextPresident()
		return
	}
	game.State = ActSelectPresident
	game.SetPresident(game.Players[game.PresidentIndex])
}

// SetPresident sets the new president
func (game *Game) SetPresident(player *Player) {
	game.State = ActPickChancellor
	game.President = player
	game.Broadcast(President{Type: TypePresident, Name: game.President.Name})
}

// PickChancellor is called when the president picks his/her chancellor
func (game *Game) PickChancellor(name string) {
	p := game.GetPlayer(name)
	if p != nil {
		game.Chancellor = p
		game.State = ActVote
		game.Broadcast(StartVote{Type: TypeStartVote, President: game.President.Name, Chancellor: game.Chancellor.Name})
	}
}

// Vote is called when the player sends a vote command
func (game *Game) Vote(player *Player, vote string) {
	player.Vote = ParseVote(vote)
	player.SendMessage(VoteMessage{Type: TypeVote, Vote: player.Vote})

	var ja, nein = 0, 0
	for _, player := range game.Players {
		if player == nil || !player.Alive {
			continue
		}
		switch player.Vote {
		case VoteEmpty:
			if player.Connected {
				return
			}
		case VoteJa:
			ja++
		case VoteNein:
			nein++
		}
	}
	for _, player := range game.Players {
		if player == nil {
			continue
		}
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
	card := game.Cards.PickCard()
	game.Broadcast(EnactForce{Type: TypeEnactForce, Policy: card})
	game.Enact(card, true)
}

// StartDiscard is executed when everyone has voted and accepted the president
func (game *Game) StartDiscard() {
	if game.Cards.TableFacist >= 3 {
		if game.Chancellor.Role == RoleHitler {
			game.End(CardFacist)
			return
		}
	}
	game.State = ActDiscardPresident

	game.FailedGovs = 0
	game.Broadcast(Discard{Type: TypePresidentDiscard, Name: game.President.Name})
	game.Discarding = game.Cards.PickCards()
	game.BroadcastTable()
	game.President.SendMessage(CardsMessage{Type: TypeCards, Cards: game.Discarding})
}

// DiscardCard is called when the chancellor or president discards a card
func (game *Game) DiscardCard(c string) {
	game.VetoRequested = false
	card, err := strconv.Atoi(c)
	if err != nil || card >= len(game.Discarding) || card < 0 {
		return
	}
	game.Cards.Discarded = append(game.Cards.Discarded, game.Discarding[card])
	game.Discarding[card] = game.Discarding[len(game.Discarding)-1]
	game.Discarding = game.Discarding[:len(game.Discarding)-1]

	if len(game.Discarding) == 2 {
		game.BroadcastTable()
		game.Broadcast(Discard{Type: TypeChancellorDiscard, Name: game.Chancellor.Name})
		game.Chancellor.SendMessage(CardsMessage{Type: TypeCards, Cards: game.Discarding})
		game.State = ActDiscardChancellor
	} else if len(game.Discarding) == 1 {
		game.Broadcast(Enact{Type: TypeEnact, President: game.President.Name, Chancellor: game.Chancellor.Name, Policy: game.Discarding[0]})
		game.Enact(game.Discarding[0], false)
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
		game.Cards.Discarded = append(game.Cards.Discarded, card)
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
func (game *Game) Enact(card Card, force bool) {
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

	if force || card == CardLiberal {
		game.NextPresident()
		return
	}
	act := game.GetSpecialAction()
	switch act {
	case ActPolicyPeek:
		game.Broadcast(PresidentAction{Type: TypePeekBroadcast, President: game.President.Name})
		game.President.SendMessage(CardsMessage{Type: TypePeek, Cards: game.Cards.Peek()})
		game.NextPresident()
	case ActInvestigatePlayer:
		game.Broadcast(PresidentAction{Type: TypeInvestigate, President: game.President.Name})
	case ActSelectPresident:
		game.Broadcast(PresidentAction{Type: TypePresidentSelect, President: game.President.Name})
	case ActExecution:
		game.Broadcast(PresidentAction{Type: TypeExecute, President: game.President.Name})
	case ActNothing:
		game.NextPresident()
		return
	}
	game.State = act
}

// Investigated is called when the president has investigated a player
func (game *Game) Investigated(name string) {
	p := game.GetPlayer(name)
	if p != nil {
		game.Broadcast(PresidentActionFinished{Type: TypeInvestigated, President: game.President.Name, Name: p.Name})
		game.President.SendMessage(InvestigateResult{Type: TypeInvestigateResult, Name: p.Name, Result: p.Role.Card()})
		game.NextPresident()
	}
}

// SelectedPresident is called when the president selects the next president
func (game *Game) SelectedPresident(name string) {
	p := game.GetPlayer(name)
	if p != nil {
		game.Broadcast(PresidentActionFinished{Type: TypePresidentSelected, President: game.President.Name, Name: p.Name})
		game.SetPresident(p)
	}
}

// ExecutedPlayer is called when the president executes a player
func (game *Game) ExecutedPlayer(name string) {
	p := game.GetPlayer(name)
	if p != nil {
		game.Broadcast(PresidentActionFinished{Type: TypeExecuted, President: game.President.Name, Name: p.Name})
		p.Alive = false
		if p.Role == RoleHitler {
			game.End(CardLiberal)
		} else {
			game.NextPresident()
		}
	}
}

func (game *Game) Error(msg string) {
	game.Broadcast(Error{Type: TypeError, Message: msg})
	game.Ended = true
	Remove(game.Name)
}

// End the game with the given winner
func (game *Game) End(winner Card) {
	var end = End{Type: TypeEnd, Winner: winner, Roles: make(map[string]Role)}
	for _, player := range game.Players {
		if player == nil {
			continue
		}
		end.Roles[player.Name] = player.Role
	}
	game.Broadcast(end)
	game.Ended = true
	Remove(game.Name)
}
