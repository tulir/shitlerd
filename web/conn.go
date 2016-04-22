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
	"github.com/gorilla/websocket"
	"maunium.net/go/shitlerd/game"
	"net/http"
	"strings"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 30 * time.Second
	pingPeriod     = 25 * time.Second
	maxMessageSize = 1024
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
}

type connection struct {
	ws *websocket.Conn
	ch chan interface{}
	p  game.Player
}

func (c *connection) readPump() {
	defer func() {
		c.ws.Close()
	}()
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				fmt.Println("Unexpected close:", err)
			}
			break
		}

		var data = make(map[string]string)
		err = json.Unmarshal(message, data)
		if err != nil {
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
				return
			}

			err := c.writeJSON(new)
			if err != nil {
				fmt.Println("Disconnected:", err)
				return
			}
		case <-ticker.C:
			err := c.write(websocket.PingMessage, []byte{})
			if err != nil {
				return
			}
		}
	}
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	args := strings.Split(strings.Trim(r.RequestURI, "/"), "/")
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Failed to connect:", err)
		return
	}
	c := &connection{ws: ws, user: user}

	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	go c.writePump()

	c.user.GetNetworks().ForEach(func(net mauircdi.Network) {
		c.user.GetMessageChan() <- mauircdi.Message{Type: "netdata", Object: mauircdi.NetData{Name: net.GetName(), Connected: net.IsConnected()}}
		net.GetActiveChannels().ForEach(func(chd mauircdi.ChannelData) {
			c.user.GetMessageChan() <- mauircdi.Message{Type: "chandata", Object: chd}
		})
		c.user.GetMessageChan() <- mauircdi.Message{Type: "chanlist", Object: mauircdi.ChanList{Network: net.GetName(), List: net.GetAllChannels()}}
		c.user.GetMessageChan() <- mauircdi.Message{Type: "nickchange", Object: mauircdi.NickChange{Network: net.GetName(), Nick: net.GetNick()}}
	})

	c.readPump()
}