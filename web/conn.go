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
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"maunium.net/go/shitlerd/game"
)

var debug = flag.Bool("wsDebug", false, "Print WebSocket connection debug/log messages")

const (
	writeWait      = 5 * time.Second
	pongWait       = 10 * time.Second
	pingPeriod     = 5 * time.Second
	maxMessageSize = 1024
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
}

type connection struct {
	ws *websocket.Conn
	ch chan interface{}
	p  *game.Player
}

func (c *connection) SendMessage(msg interface{}) {
	c.ch <- msg
}

func (c *connection) Close() {
	c.write(websocket.CloseMessage, []byte{})
	c.p = nil
}

func (c *connection) readPump() {
	defer func() {
		c.ws.Close()
	}()
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				fmt.Println("Unexpected close:", err)
				if c.p != nil && c.p.Conn == c {
					c.p.Disconnect()
					c.p = nil
				}
			}
			break
		}

		var data = make(map[string]interface{})
		err = json.Unmarshal(message, &data)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if c.p == nil {
			if data["type"] == "join" {
				c.ch <- c.join(data)
			}
			continue
		}

		c.p.ReceiveMessage(data)
	}
}

func (c *connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

func (c *connection) writeJSON(payload interface{}) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteJSON(payload)
}

func (c *connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case new, ok := <-c.ch:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				if c.p != nil && c.p.Conn == c {
					c.p.Disconnect()
					c.p = nil
				}
				return
			}
			err := c.writeJSON(new)
			if err != nil {
				fmt.Println("Disconnected:", err)
				if c.p != nil && c.p.Conn == c {
					c.p.Disconnect()
					c.p = nil
				}
				return
			}
		case <-ticker.C:
			err := c.write(websocket.PingMessage, []byte{})
			if err != nil {
				if c.p != nil && c.p.Conn == c {
					c.p.Disconnect()
					c.p = nil
				}
				return
			}
		}
	}
}

func (c *connection) join(data map[string]interface{}) (response map[string]interface{}) {
	response = make(map[string]interface{})
	g, ok := game.Get(data["game"].(string))
	if !ok || g == nil {
		response["success"] = false
		response["message"] = "gamenotfound"
		response["game"] = data["game"]
		response["name"] = data["name"]
		return
	}

	response["game"] = g.Name
	authtoken, _ := data["authtoken"].(string)

	state, p := g.Join(data["name"].(string), authtoken, c)
	if p != nil {
		response["name"] = p.Name
	} else {
		response["name"] = data["name"]
	}

	if _, isInt := state.(int); isInt {
		c.p = p
		response["success"] = true
		response["authtoken"] = p.AuthToken
		players := make(map[string]bool)
		for _, p := range g.Players {
			if p != nil {
				players[p.Name] = p.Connected
			}
		}
		response["players"] = players
		response["started"] = g.Started
		if g.Started {
			response["table"] = g.GetTable()
			response["role"] = p.Role
			toLiberals, toFascists := g.MapRoles()
			pc := g.PlayerCount()
			if p.Role == game.RoleLiberal || (pc > 6 && p.Role == game.RoleHitler) {
				response["players"] = toLiberals
			} else if p.Role == game.RoleFascist || (pc < 7 && p.Role == game.RoleHitler) {
				response["players"] = toFascists
			}
		}
	} else {
		response["success"] = false
		response["message"] = state
	}
	return
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Failed to connect:", err)
		return
	}

	c := &connection{ws: ws, ch: make(chan interface{})}
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	go c.writePump()
	c.readPump()
}
