// shitlerd - A manager for online Secret Hitler games
// Copyright (C) 2016-2017 Tulir Asokan

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

var registry map[string]*Game

func init() {
	registry = make(map[string]*Game)
}

// New creates a game and adds it to the registry
func New() string {
	name := RandomName()
	lcName := strings.ToLower(name)
	if game, ok := registry[lcName]; ok && game != nil && !game.Ended {
		name = RandomName()
	}
	game := CreateGame(name)
	registry[lcName] = game
	return name
}

// Get the game with the given name from the registry
func Get(name string) (*Game, bool) {
	game, ok := registry[strings.ToLower(name)]
	return game, ok
}

// Remove a game from the registry
func Remove(name string) bool {
	name = strings.ToLower(name)
	_, ok := registry[name]
	if !ok {
		return false
	}
	registry[name] = nil
	return true
}
