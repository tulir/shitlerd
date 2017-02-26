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
package main

import (
	"flag"
	"fmt"

	_ "maunium.net/go/shitlerd/game"
	"maunium.net/go/shitlerd/web"
)

var address = flag.String("address", "localhost", "The address to bind the web server to.")
var port = flag.Int("port", 29305, "The port to bind the web server to.")

func main() {
	flag.Parse()
	web.Load(fmt.Sprintf("%s:%d", *address, *port))
}
