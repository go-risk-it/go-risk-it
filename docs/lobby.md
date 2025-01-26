## Lobby

* clients connect to lobby service when they land on a page (websockets)
* 2 pages
  * landing page (matches overview, option to create a new match or join an existing one)
  * match lobby page (list of players, chat, option to start the match for the lobby owner)
    * when the match starts, all clients get a message with a game ID to connect to