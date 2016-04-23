# Secret Hitler Server

A server to handle online Secret Hitler games.

## API

### Joining games
Joining games is done with a HTTP POST request to `/join/<game>`. The users nickname must be included in the header field `Name`.
The response body will always be in JSON format and contain the `successful` field.

If successful, the response JSON will contain the username, the game name and the auth token. These details will also be included as encrypted cookies, so web clients don't have to store the response data manually.

If unsuccessful, the response JSON will contain a simple error message string in the field `message`. Explanations for error messages:
* `gamenotfound` - The given game does not exist.
* `gamestarted` - The game has already started.
* `full` - The game is full (10 players)
* `nameused` - The given name is already in use by another player.

### Connection
Once joined, the actual socket connection can be made. Currently this uses WebSockets, but adding support for TCP would not be difficult.
The WebSocket connection is located at `/socket`.

When connecting, you must supply either the cookies or the JSON provided by `/join`. Web browsers will most likely handle cookies on their own and thus web-based clients don't need 
