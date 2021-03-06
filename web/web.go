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

// Package web contains the HTTP server
package web

import (
	"flag"
	"net/http"

	"github.com/gorilla/context"
	"maunium.net/go/shitlerd/game"
)

var trustOrigin = flag.Bool("trustOrigin", false, "Trust Origin headers for WebSockets")

// Load the web server
func Load(addr string) {
	if *trustOrigin {
		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}
	}
	http.HandleFunc("/create", create)
	http.HandleFunc("/socket", serveWs)
	err := http.ListenAndServe(addr, context.ClearHandler(http.DefaultServeMux))
	if err != nil {
		panic(err)
	}
}

func create(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(game.New()))
}
