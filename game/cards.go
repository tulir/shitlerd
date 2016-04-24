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

// Cards contains all the cards in the game.
type Cards struct {
	Deck         []Card
	Discarded    []Card
	TableLiberal int
	TableFacist  int
}

// Card is a single card (facist or liberal)
type Card string

// The possible card types
const (
	CardLiberal Card = "liberal"
	CardFacist  Card = "facist"
)

// CreateDeck creates a Cards object with 6 liberal and 11 facist cards in the deck
func CreateDeck() *Cards {
	var cards = &Cards{Deck: make([]Card, 17), Discarded: []Card{}, TableLiberal: 0, TableFacist: 0}
	liberal := 6
	facist := 11
	for i := 0; i < 17; i++ {
		if liberal == 0 && facist == 0 {
			break
		} else if liberal == 0 {
			cards.Deck[i] = CardFacist
			facist--
		} else if facist == 0 {
			cards.Deck[i] = CardLiberal
			liberal--
		} else {
			if r.Int()%2 == 0 {
				cards.Deck[i] = CardLiberal
				liberal--
			} else {
				cards.Deck[i] = CardFacist
				facist--
			}
		}
	}
	return cards
}

// PickCard picks one card from the deck
func (cards *Cards) PickCard() Card {
	card := cards.Deck[0]
	cards.Deck = cards.Deck[1:]
	return card
}

// PickCards picks `n` random cards from the deck
func (cards *Cards) PickCards() (picked []Card) {
	picked = cards.Deck[0:3]
	cards.Deck = cards.Deck[3:]
	return picked
}

// Peek peeks at the top three cards
func (cards *Cards) Peek() []Card {
	return cards.Deck[0:3]
}

// ResetDiscarded moves all discarded cards back to the deck
func (cards *Cards) ResetDiscarded() {
	for i := range cards.Discarded {
		j := r.Intn(i + 1)
		cards.Discarded[i], cards.Discarded[j] = cards.Discarded[j], cards.Discarded[i]
	}
	cards.Deck = append(cards.Deck, cards.Discarded...)
	cards.Discarded = []Card{}
}
