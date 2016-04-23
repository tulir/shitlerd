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
	"encoding/json"
	"fmt"
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

type jsonCookies struct {
	Name      string `json:"name"`
	Game      string `json:"game"`
	AuthToken string `json:"authtoken"`
}

func checkAuth(w http.ResponseWriter, r *http.Request) (*sgame.Player, string) {
	var cookies = jsonCookies{}

	session, err := store.Get(r, "mauIRC")
	if err == nil {
		nameI := session.Values["name"]
		gameI := session.Values["game"]
		authTokenI := session.Values["authtoken"]
		if nameI == nil || gameI == nil || authTokenI == nil {
			return nil, "invalidsession"
		}
		cookies.Name = nameI.(string)
		cookies.Game = gameI.(string)
		cookies.AuthToken = authTokenI.(string)
	} else {
		dec := json.NewDecoder(r.Body)
		dec.Decode(cookies)
	}

	if len(cookies.Game) == 0 || len(cookies.Name) == 0 || len(cookies.AuthToken) == 0 {
		return nil, "invaliddata"
	}

	g, ok := sgame.Get(cookies.Game)
	if !ok || g == nil || g.Ended {
		return nil, "invalidgame"
	}
	p := g.GetPlayer(cookies.Name)
	if p.AuthToken != cookies.AuthToken {
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
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "{\"success\": false, \"message\": \"%s\"}", "gamenotfound")
		return
	} else if game.Started {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "{\"success\": false, \"message\": \"%s\"}", "gamestarted")
	}

	status, player := game.Join(name, notcon{})

	switch status {
	case -1:
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "{\"success\": false, \"message\": \"%s\"}", "full")
	case -2:
		w.WriteHeader(http.StatusConflict)
		fmt.Fprintf(w, "{\"success\": false, \"message\": \"%s\"}", "nameused")
	case -3:
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "{\"success\": false, \"message\": \"%s\"}", "gamestarted")
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

	fmt.Fprintf(w, "{\"success\": true, \"name\": \"%s\", \"game\": \"%s\", \"authtoken\": \"%s\"}", player.Name, game.Name, player.AuthToken)
}
