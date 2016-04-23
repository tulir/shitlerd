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
	"strings"
)

// Action is an action
type Action int

// All actions
const (
	ActNothing           Action = iota
	ActPickChancellor    Action = iota
	ActVote              Action = iota
	ActDiscardPresident  Action = iota
	ActDiscardChancellor Action = iota
	ActPolicyPeek        Action = iota
	ActInvestigatePlayer Action = iota
	ActSelectPresident   Action = iota
	ActExecution         Action = iota
)

// GetSpecialAction gets the special action that should happen now.
func (game *Game) GetSpecialAction() Action {
	switch game.PlayerCount() {
	case 5:
		fallthrough
	case 6:
		switch game.Cards.TableFacist {
		case 3:
			return ActPolicyPeek
		case 4:
			return ActExecution
		case 5:
			return ActExecution
		}
	case 7:
		fallthrough
	case 8:
		switch game.Cards.TableFacist {
		case 2:
			return ActInvestigatePlayer
		case 3:
			return ActSelectPresident
		case 4:
			return ActExecution
		case 5:
			return ActExecution
		}
	case 9:
		fallthrough
	case 10:
		switch game.Cards.TableFacist {
		case 1:
			return ActInvestigatePlayer
		case 2:
			return ActInvestigatePlayer
		case 3:
			return ActSelectPresident
		case 4:
			return ActExecution
		case 5:
			return ActExecution
		}
	}
	return ActNothing
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

// Card gets the card corresponding to the given role
func (role Role) Card() Card {
	switch role {
	case RoleFacist:
		fallthrough
	case RoleHitler:
		return CardFacist
	default:
		return CardLiberal
	}
}

// The possible roles
const (
	RoleLiberal Role = "liberal"
	RoleFacist  Role = "facist"
	RoleHitler  Role = "hitler"
)
