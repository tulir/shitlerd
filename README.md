# Secret Hitler Server
A server to handle online Secret Hitler games.

## API
### Creating a game
You can create a game by making a GET request to `/create`. This will simply return the name of the newly created game.

### Connecting
The connection is made using WebSockets. The primary (currently the only) socket is at `/socket`.

Once connected, the client must send a join message in JSON format. The message must contain at least the fields `type` with the value `join`, `game` with the name of the game (case-insensitive) and `name` with the username of the client. The join message may also contain the field `authtoken` which should contain the token to retake a username (after a disconnection, for example).

The response for the join message will contain at least the field `success`. If true, the response should also contain `authtoken` which can be used to rejoin with the same name. If unsuccessful, the response will contain the field `message` with a simple error code.

Possible errors:
* `gamenotfound` - The given game does not exist (see the section Creating a game)
* `gamestarted` - The game has already started and no valid auth token was given
* `full` - The game is full and no valid auth token was given
* `nameused` - The name is already in used and no valid auth token was given
* `invalidname` - The name is invalid (names must be [a-zA-Z0-9_-]{3,16})

### Game protocol
Every message must contain the field `type` to identify what the message should contain.
Messages that the server receives at the wrong time or from the wrong user are ignored.
**All** fields in client -> server messages must be JSON strings!

#### Client -> server messages
* Type `chat` - A chat message.
  * Field `message` - The message to send.
* Type `start` - Tell the server to start the game. Ignored if the game is already started or has less than 5 players.
* Type `vote` - Vote for a president+chancellor combination. Ignored if the game isn't in a voting state.
  * Field `vote` - The vote value, `ja` or `nein`.
* Type `pickchancellor` - Pick a chancellor.
  * Field `name` - The name of the chancellor to pick.
* Type `discard` - Discard a card.
  * Field `index` - The index of the card (from the cards the server sent the client).
* Type `vetorequest` - Request veto. There must be 5 fascist cards on the table.
* Type `vetoaccept` - Accept veto request. The chancellor must have requested a veto first.
* Type `investigate`, `execute`, `presidentselect` - Sent by the president when he/she performs a special action. The special action `peek` requires no answer.
  * Field `name` - The person the action is performed on.

#### Server -> client messages
* Type `chat` - A chat message.
  * Field `message` - The message.
  * Field `sender` - The name of the user who sent the message.
* Type `join`, `quit` - A player joined or left the game.
  * Field `name` - The name of the player who joined or left the game.
* Type `connected`, `disconnected` - A player connected or disconnected
  * Field `name` - The player who connected/disconnected.
* Type `start` - The game has started.
  * Field `role` - The secret role of the user.
  * Field `players` - A map of players and their roles. All roles will be. `unknown` if the client is liberal or the client is hitler and there are over 6 players.
* Type `startvote` - The president has picked his/her chancellor and players msut vote.
  * Field `president` - The name of the president.
  * Field `chancellor` - The name of the chancellor.
* Type `governmentfailed` - The vote has failed.
  * Field `times` - The amount of times the government has failed by now.
  * Field `veto` - True if the fail was caused by the president and chancellor vetoing the card pick.
* Type `presidentdiscard` - The vote has ended successfully and the president has received three cards, one of which he/she must discard. The government fail counter is reset when this event occurs.
  * Field `name` - The name of the president.
* Type `chancellordiscard` - The president has discarded one card and the chancellor has received the remaining two.
  * Field `name` - The name of the chancellor.
* Type `cards` - Two or three cards, one of which must be discarded.
  * Field `cards` - An array of strings. A string will either be `liberal` or `fascist`. Fuck you if you can't figure out which string means which card.
* Type `table` - The current status of the table.
  * Field `deck` - The number of cards in the deck.
  * Field `discarded` - The number of discarded cards.
  * Field `tableLiberal` - The number of liberal cards on the table.
  * Field `tableFascist` - The number of fascist cards on the table.
* Type `enact` - The president and chancellor have both discarded a card and the remaining card is enacted.
  * Field `president` - The name of the president.
  * Field `chancellor` - The name of the chancellor.
  * Field `policy` - The policy of the card they enacted (`liberal` or `fascist`)
* Type `enactforce` - Three failed elections or vetos have occured and the first card in the deck is forcefully enacted.
  * Field `policy` - The policy of the card enacted (`liberal` or `fascist`)
* Types `vetorequest`, `vetoaccept` - The chancellor has requested or the president has accepted to veto the current pick.
  * Field `president` - The name of the president.
  * Field `chancellor` - The name of the chancellor.
* Type `peek`, `investigate`, `presidentselect`, `execute` - The president must perform a special action.
  * Field `president` - The name of the president.
* Type `peekcards` - Sent to the president when the special action `peek` is invoked.
  * Field `cards` - The top three cards from the deck.
* Type `investigateresult` - The result of the investigation.
  * Field `name` - The name of the player who was investigated
  * Field `result` - The party of the player (`liberal` or `fascist`). Hitler is a fascist too.
* Type `investigated`, `presidentselected`, `executed` - Broadcasted once the president has completed a special action.
  * There is no broadcast for the action `peek`, since the game goes on instantly after the president receives the peek cards.
  * Field `president` - The name of the president.
  * Field `name` - The name of the player the action was performed on.
* Type `error` - The server has encountered an internal error and the game has been terminated.
  * Field `message` - A human-readable error message.
* Type `end` - The game has naturally ended.
  * Field `winner` - The side that won (`liberal` or `fascist`).
  * Field `roles` - A map of the roles of all players.

# Attribution
["Secret Hitler"](http://secrethitler.com/) is a game designed by Max Temkin, Mike Boxleiter, Tommy Maranges, and Mackenzie Schubert. This adaptation is neither affiliated with, nor endorsed by the copyright holders.
