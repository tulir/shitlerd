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
	"io/ioutil"
	"net/http"
	"strings"
)

var adjectives, animals []string

func init() {
	resp, err := http.Get("https://dl.maunium.net/adjectives")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	adjectives = strings.Split(string(data), "\n")

	resp, err = http.Get("https://dl.maunium.net/animals")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	animals = strings.Split(string(data), "\n")
}

// RandomName generates a random name for a game
func RandomName() string {
	adj1 := strings.Title(adjectives[r.Intn(len(adjectives))])
	adj2 := strings.Title(adjectives[r.Intn(len(adjectives))])
	animal := strings.Title(animals[r.Intn(len(animals))])
	return adj1 + adj2 + animal
}
