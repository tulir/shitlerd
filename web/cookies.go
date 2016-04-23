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

// Package web contains the HTTP server
package web

import (
	"github.com/gorilla/sessions"
	sgame "maunium.net/go/shitlerd/game"
	"net/http"
	"strings"
)

var store *sessions.CookieStore

func initStore(address string) {
	store = sessions.NewCookieStore([]byte("ThisDoesn'tNeedToBeSecret"))
	store.Options = &sessions.Options{
		Domain:   address,
		Path:     "/",
		MaxAge:   86400,
		Secure:   false,
		HttpOnly: true,
	}
}

func checkAuth(w http.ResponseWriter, r *http.Request) (*sgame.Player, string) {
	session, err := store.Get(r, "mauIRC")
	if err != nil {
		return nil, "invalidstore"
	}

	nameI := session.Values["name"]
	gameI := session.Values["game"]
	authTokenI := session.Values["authtoken"]
	if nameI == nil || gameI == nil || authTokenI == nil {
		return nil, "invalidsession"
	}
	name := nameI.(string)
	game := gameI.(string)
	authToken := authTokenI.(string)

	g, ok := sgame.Get(game)
	if !ok || g == nil || g.Ended {
		return nil, "invalidgame"
	}
	p := g.GetPlayer(name)
	if p.AuthToken != authToken {
		return nil, "invalidauthtoken"
	}
	return p, "success"
}

type notcon struct{}

func (notcon notcon) SendMessage(msg interface{}) {}

func join(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Add("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	args := strings.Split(strings.Trim(r.RequestURI, "/"), "/")
	gname := args[len(args)-1]

	name := r.Header.Get("Name")

	if len(name) == 0 || len(gname) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	game, ok := sgame.Get(gname)
	if !ok || game == nil {
		w.Write([]byte("gamenotfound"))
		w.WriteHeader(http.StatusNotFound)
		return
	} else if game.Started {
		w.Write([]byte("gamestarted"))
		w.WriteHeader(http.StatusUnauthorized)
	}

	status, player := game.Join(name, notcon{})

	switch status {
	case -1:
		w.Write([]byte("full"))
		w.WriteHeader(http.StatusUnauthorized)
	case -2:
		w.Write([]byte("nameused"))
		w.WriteHeader(http.StatusConflict)
	case -3:
		w.Write([]byte("gamestarted"))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	session, err := store.Get(r, "mauIRC")
	if err != nil {
		session, err = store.New(r, "mauIRC")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	session.Values["name"] = player.Name
	session.Values["game"] = game.Name
	session.Values["authtoken"] = player.AuthToken
	session.Save(r, w)
}
