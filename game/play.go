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
	if game.ConnectedPlayers() < 5 {
		return
	}
	for _, p := range game.Players {
		if p != nil && !p.Connected {
			game.Leave(p.Name)
		}
	}
	game.debugln("Starting...")
	game.Started = true

	game.PresidentIndex = r.Intn(len(game.Players))
	game.debugln("  President index:", game.PresidentIndex)

	game.GiveRoles()
	game.MapAndSendRoles()

	game.BroadcastTable()
	game.NextPresident()
}

// GiveRoles gives everyone roles (but doesn't send them yet)
func (game *Game) GiveRoles() {
	game.debugln("  Players:", game.PlayerCount())
	fascistsAvailable := game.Fascists()
	game.debugln("  Fascists:", fascistsAvailable)
	liberalsAvailable := game.Liberals()
	game.debugln("  Liberals:", liberalsAvailable)
	hitlerAvailable := true

	for _, player := range game.Players {
		if player == nil {
			continue
		}
		var availableRoles []Role
		if !hitlerAvailable && fascistsAvailable == 0 {
			availableRoles = []Role{RoleLiberal}
		} else if !hitlerAvailable {
			availableRoles = []Role{RoleLiberal, RoleFascist}
		} else if fascistsAvailable == 0 {
			availableRoles = []Role{RoleLiberal, RoleHitler}
		} else {
			availableRoles = []Role{RoleLiberal, RoleFascist, RoleHitler}
		}

		player.Role = availableRoles[r.Intn(len(availableRoles))]
		game.debugln("   ", player.Name, "is a", player.Role)

		switch player.Role {
		case RoleLiberal:
			liberalsAvailable--
		case RoleFascist:
			fascistsAvailable--
		case RoleHitler:
			hitlerAvailable = false
		}
	}
}

// MapRoles maps the info about others' roles that must be sent to players with specific roles
func (game *Game) MapRoles() (toLiberals map[string]Role, toFascists map[string]Role) {
	toLiberals = make(map[string]Role)
	toFascists = make(map[string]Role)
	for _, player := range game.Players {
		if player == nil {
			continue
		}
		toLiberals[player.Name] = "unknown"
		toFascists[player.Name] = player.Role
	}
	return
}

// MapAndSendRoles sends a start message to players containing their roles and possibly other players' roles
func (game *Game) MapAndSendRoles() {
	toLiberals, toFascists := game.MapRoles()
	pc := game.PlayerCount()
	for _, player := range game.Players {
		if player == nil {
			continue
		}
		if player.Role == RoleLiberal || (pc > 6 && player.Role == RoleHitler) {
			player.SendMessage(Start{Type: TypeStart, Role: player.Role, Players: toLiberals})
		} else if player.Role == RoleFascist || (pc < 7 && player.Role == RoleHitler) {
			player.SendMessage(Start{Type: TypeStart, Role: player.Role, Players: toFascists})
		}
	}
}

// NextPresident moves the game to the next president
func (game *Game) NextPresident() {
	game.debugln("Moving to next president...")
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
	game.PreviousPresident = game.President
	game.President = player
	game.debugln(game.President.Name, "is now the president")
	game.Broadcast(President{Type: TypePresident, Name: game.President.Name})
}

// PickChancellor is called when the president picks his/her chancellor
func (game *Game) PickChancellor(name string) {
	p := game.GetPlayer(name)
	if p != nil && p.Alive && p != game.President && p != game.PreviousChancellor && (game.PlayerCount() == 5 || p != game.PreviousPresident) {
		game.PreviousChancellor = game.Chancellor
		game.Chancellor = p
		game.State = ActVote
		game.debugln(game.President.Name, "picked", game.Chancellor.Name, "as the chancellor")
		game.Broadcast(StartVote{Type: TypeStartVote, President: game.President.Name, Chancellor: game.Chancellor.Name})
	}
}

// Vote is called when the player sends a vote command
func (game *Game) Vote(player *Player, vote string) {
	player.Vote = ParseVote(vote)
	player.SendMessage(VoteMessage{Type: TypeVote, Vote: player.Vote})
	game.debugln(player.Name, "voted", player.Vote)

	var ja, nein = game.CalculateVotes()
	if ja == -1 || nein == -1 {
		return
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
		game.GovernmentFailed(false)
	}
}

// CalculateVotes gets the amount of votes
func (game *Game) CalculateVotes() (ja, nein int) {
	for _, player := range game.Players {
		if player == nil || !player.Alive {
			continue
		}
		switch player.Vote {
		case VoteEmpty:
			if player.Connected {
				return -1, -1
			}
		case VoteJa:
			ja++
		case VoteNein:
			nein++
		}
	}
	return
}

// GovernmentFailed is called when the government fails.
func (game *Game) GovernmentFailed(veto bool) {
	game.FailedGovs++
	game.debugfln("The government has failed (#%d)", game.FailedGovs)
	if game.FailedGovs >= 3 {
		game.ThreeGovsFailed()
	} else {
		game.Broadcast(GovernmentFailed{Type: TypeGovernmentFailed, Times: game.FailedGovs, Veto: veto})
		game.NextPresident()
	}
}

// ThreeGovsFailed is called when three consequent government attempts have been downvoted
func (game *Game) ThreeGovsFailed() {
	card := game.Cards.PickCard()
	game.debugln("Three governments failed")
	game.Broadcast(EnactForce{Type: TypeEnactForce, Policy: card})
	game.Enact(card, true)
}

// StartDiscard is executed when everyone has voted and accepted the president
func (game *Game) StartDiscard() {
	if game.Cards.TableFascist >= 3 {
		if game.Chancellor.Role == RoleHitler {
			game.End(CardFascist)
			return
		}
	}
	game.State = ActDiscardPresident
	game.debugln("Started card discarding with", game.President.Name, "and", game.Chancellor.Name)
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
	game.debugf("A %s card was discarded by the ", game.Discarding[card])
	game.Cards.Discarded = append(game.Cards.Discarded, game.Discarding[card])
	game.Discarding[card] = game.Discarding[len(game.Discarding)-1]
	game.Discarding = game.Discarding[:len(game.Discarding)-1]

	if len(game.Discarding) == 2 {
		game.BroadcastTable()
		game.debugNoPrefix("president\n")
		game.Broadcast(Discard{Type: TypeChancellorDiscard, Name: game.Chancellor.Name})
		game.Chancellor.SendMessage(CardsMessage{Type: TypeCards, Cards: game.Discarding})
		game.State = ActDiscardChancellor
	} else if len(game.Discarding) == 1 {
		game.debugNoPrefix("chancellor\n")
		game.Broadcast(Enact{Type: TypeEnact, President: game.President.Name, Chancellor: game.Chancellor.Name, Policy: game.Discarding[0]})
		game.Enact(game.Discarding[0], false)
	} else {
		game.Error("Invalid amount of cards to discard")
	}
}

// VetoRequest is called when the chancellor wants to veto the current discard
func (game *Game) VetoRequest() {
	game.debugln(game.Chancellor.Name, "has made a veto request")
	game.VetoRequested = true
	game.Broadcast(Veto{Type: TypeVetoRequest, President: game.President.Name, Chancellor: game.Chancellor.Name})
}

// VetoAccept is called when the president accepts the chancellors veto request
func (game *Game) VetoAccept() {
	game.debugln(game.President.Name, "has accepted the veto request")
	game.VetoRequested = false
	game.Broadcast(Veto{Type: TypeVetoAccept, President: game.President.Name, Chancellor: game.Chancellor.Name})

	for _, card := range game.Discarding {
		game.Cards.Discarded = append(game.Cards.Discarded, card)
	}
	game.BroadcastTable()
	game.Discarding = []Card{}

	game.GovernmentFailed(true)
}

// Enact is called when a card is enacted
func (game *Game) Enact(card Card, force bool) {
	switch card {
	case CardFascist:
		game.Cards.TableFascist++
	case CardLiberal:
		game.Cards.TableLiberal++
	}
	if force {
		game.debugln("Enacting", card, "by force")
	} else {
		game.debugln("Enacting", card, "by", game.President.Name, "and", game.Chancellor.Name)
	}
	game.BroadcastTable()
	if game.Cards.TableFascist >= 6 {
		game.End(CardFascist)
		return
	} else if game.Cards.TableLiberal >= 5 {
		game.End(CardLiberal)
		return
	}

	if force || card == CardLiberal {
		game.NextPresident()
		return
	}
	game.RunSpecialAction()
}

// RunSpecialAction runs the next special action
func (game *Game) RunSpecialAction() {
	act := game.GetSpecialAction()
	switch act {
	case ActPolicyPeek:
		game.debugln(game.President.Name, "will now peek on the next three cards")
		game.Broadcast(PresidentAction{Type: TypePeekBroadcast, President: game.President.Name})
		game.President.SendMessage(CardsMessage{Type: TypePeek, Cards: game.Cards.Peek()})
		game.NextPresident()
	case ActInvestigatePlayer:
		game.debugln(game.President.Name, "will now investigate a player")
		game.Broadcast(PresidentAction{Type: TypeInvestigate, President: game.President.Name})
	case ActSelectPresident:
		game.debugln(game.President.Name, "will now select a president")
		game.Broadcast(PresidentAction{Type: TypePresidentSelect, President: game.President.Name})
	case ActExecution:
		game.debugln(game.President.Name, "will now execute a player")
		game.Broadcast(PresidentAction{Type: TypeExecute, President: game.President.Name})
	case ActNothing:
		game.debugln(game.President.Name, "will now do nothing")
		game.NextPresident()
		return
	}
	game.State = act
}

// Investigated is called when the president has investigated a player
func (game *Game) Investigated(name string) {
	p := game.GetPlayer(name)
	if p != nil {
		game.debugln(game.President.Name, "investigated", p.Name)
		game.Broadcast(PresidentActionFinished{Type: TypeInvestigated, President: game.President.Name, Name: p.Name})
		game.President.SendMessage(InvestigateResult{Type: TypeInvestigateResult, Name: p.Name, Result: p.Role.Card()})
		game.NextPresident()
	}
}

// SelectedPresident is called when the president selects the next president
func (game *Game) SelectedPresident(name string) {
	p := game.GetPlayer(name)
	if p != nil && p.Alive && p != game.President {
		game.debugln(game.President.Name, "selected", p.Name, "as the next president")
		game.Broadcast(PresidentActionFinished{Type: TypePresidentSelected, President: game.President.Name, Name: p.Name})
		game.SetPresident(p)
	}
}

// ExecutedPlayer is called when the president executes a player
func (game *Game) ExecutedPlayer(name string) {
	p := game.GetPlayer(name)
	if p != nil && p.Alive {
		game.debugln(game.President.Name, "executed", p.Name)
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
	game.debugln("Error:", msg)
	game.Broadcast(Error{Type: TypeError, Message: msg})
	game.Ended = true
	Remove(game.Name)
}

// End the game with the given winner
func (game *Game) End(winner Card) {
	game.debugln(winner, "won")
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
