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
	DeckLiberal      int
	DeckFacist       int
	DiscardedLiberal int
	DiscardedFacist  int
	TableLiberal     int
	TableFacist      int
	PeekedCards      []Card
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

// PickCard picks one card from the deck
func (cards Cards) PickCard() Card {
	var picked Card
	if cards.TableFacist == 0 && cards.TableLiberal == 0 {
		cards.ResetDiscarded()
	}
	if cards.TableFacist == 0 {
		picked = CardLiberal
		cards.TableLiberal--
	} else if cards.TableLiberal == 0 {
		picked = CardFacist
		cards.TableFacist--
	} else {
		if r.Int()%2 == 0 {
			picked = CardLiberal
			cards.TableLiberal--
		} else {
			picked = CardFacist
			cards.TableFacist--
		}
	}

	if len(cards.PeekedCards) > 0 {
		newPicked := cards.PeekedCards[0]
		cards.PeekedCards[0] = picked
		return newPicked
	}
	return picked
}

// PickCards picks `n` random cards from the deck
func (cards Cards) PickCards() (picked []Card) {
	if len(cards.PeekedCards) > 0 {
		picked = make([]Card, 3)
		for i, card := range cards.PeekedCards {
			picked[i] = card
			switch card {
			case CardFacist:
				cards.TableFacist--
			case CardLiberal:
				cards.TableLiberal--
			}
		}
		cards.PeekedCards = []Card{}
		return
	}
	picked = make([]Card, 3)
	for i := 0; i < 3; i++ {
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
			if r.Int()%2 == 0 {
				picked[i] = CardLiberal
				cards.TableLiberal--
			} else {
				picked[i] = CardFacist
				cards.TableFacist--
			}
		}
	}
	return
}

// Peek peeks at the top three cards
func (cards Cards) Peek() []Card {
	cards.PeekedCards = make([]Card, 3)
	for i := 0; i < 3; i++ {
		if cards.TableFacist == 0 && cards.TableLiberal == 0 {
			cards.ResetDiscarded()
		}
		if cards.TableFacist == 0 {
			cards.PeekedCards[i] = CardLiberal
		} else if cards.TableLiberal == 0 {
			cards.PeekedCards[i] = CardFacist
		} else {
			if r.Int()%2 == 0 {
				cards.PeekedCards[i] = CardLiberal
			} else {
				cards.PeekedCards[i] = CardFacist
			}
		}
	}
	return cards.PeekedCards
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
