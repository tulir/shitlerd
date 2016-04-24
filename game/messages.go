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

// Type is the type of a message
type Type string

func (typ Type) String() string {
	return string(typ)
}

// The possible message types
const (
	TypeChat              Type = "chat"
	TypeJoin              Type = "join"
	TypeQuit              Type = "quit"
	TypeConnected         Type = "connected"
	TypeDisconnected      Type = "disconnected"
	TypeStart             Type = "start"
	TypeEnd               Type = "end"
	TypePresident         Type = "president"
	TypePickChancellor    Type = "pickchancellor"
	TypeStartVote         Type = "startvote"
	TypeVote              Type = "vote"
	TypePresidentDiscard  Type = "presidentdiscard"
	TypeChancellorDiscard Type = "chancellordiscard"
	TypeDiscard           Type = "discard"
	TypeCards             Type = "cards"
	TypeTable             Type = "table"
	TypeError             Type = "error"
	TypeEnact             Type = "enact"
	TypeVetoRequest       Type = "vetorequest"
	TypeVetoAccept        Type = "vetoaccept"
	TypeEnactForce        Type = "enactforce"
	TypePeek              Type = "peekcards"
	TypePeekBroadcast     Type = "peek"
	TypeInvestigate       Type = "investigate"
	TypeInvestigated      Type = "investigated"
	TypeInvestigateResult Type = "investigateresult"
	TypeExecute           Type = "execute"
	TypePresidentSelect   Type = "presidentselect"
	TypePresidentSelected Type = "presidentselected"
	TypeExecuted          Type = "executed"
)

// Chat contains the necessary fields for a chat message
type Chat struct {
	Type    Type   `json:"type"`
	Sender  string `json:"sender"`
	Message string `json:"message"`
}

// JoinQuit contains the necessary fields for join and quit messages
type JoinQuit struct {
	Type Type   `json:"type"`
	Name string `json:"name"`
}

// Start contains the necessary fields for a game start message
type Start struct {
	Type    Type            `json:"type"`
	Role    Role            `json:"role"`
	Players map[string]Role `json:"players"`
}

// End contains the necessary fields for a game end message
type End struct {
	Type   Type            `json:"type"`
	Winner Card            `json:"winner"`
	Roles  map[string]Role `json:"roles"`
}

// Error is sent to the client when something unexpected happens and the game ends
type Error struct {
	Type    Type   `json:"type"`
	Message string `json:"message"`
}

// President contains the necessary fields for a president selected announcement
type President struct {
	Type Type   `json:"type"`
	Name string `json:"name"`
}

// StartVote is sent to the clients when they should vote for president&chancellor
type StartVote struct {
	Type       Type   `json:"type"`
	President  string `json:"president"`
	Chancellor string `json:"chancellor"`
}

// VoteMessage contains the necessary field for a vote
type VoteMessage struct {
	Type Type `json:"type"`
	Vote Vote `json:"vote"`
}

// Discard is sent when the someone needs to discard one card
type Discard struct {
	Type Type   `json:"type"`
	Name string `json:"name"`
}

// CardsMessage is sent to the president and chancellor when they need to discard a card
type CardsMessage struct {
	Type  Type   `json:"type"`
	Cards []Card `json:"cards"`
}

// Table is sent to the clients to tell the status of the card table
type Table struct {
	Type         Type `json:"type"`
	Deck         int  `json:"deck"`
	Discarded    int  `json:"discarded"`
	TableLiberal int  `json:"tableLiberal"`
	TableFacist  int  `json:"tableFacist"`
}

// Enact is sent to the clients when the president and chancellor have enacted a policy
type Enact struct {
	Type       Type   `json:"type"`
	President  string `json:"president"`
	Chancellor string `json:"chancellor"`
	Policy     Card   `json:"policy"`
}

// EnactForce is sent to the client when a policy is enacted because of 3 failed governments
type EnactForce struct {
	Type   Type `json:"type"`
	Policy Card `json:"policy"`
}

// Veto is broadcasted when the chancellor wants to veto or the president accepts the veto of the current discard
type Veto struct {
	Type       Type   `json:"type"`
	President  string `json:"president"`
	Chancellor string `json:"chancellor"`
}

// PresidentAction is broadcasted when the president must perform a special action.
type PresidentAction struct {
	Type      Type   `json:"type"`
	President string `json:"president"`
}

// PresidentActionFinished is broadcasted when the president finishes an action
type PresidentActionFinished struct {
	Type      Type   `json:"type"`
	President string `json:"president"`
	Name      string `json:"name"`
}

// InvestigateResult is sent to the president when he/she has investigated someone
type InvestigateResult struct {
	Type   Type   `json:"type"`
	Name   string `json:"name"`
	Result Card   `json:"result"`
}
