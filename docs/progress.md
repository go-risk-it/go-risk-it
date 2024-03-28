# Progress

## Done so far

**Server**

- Websockets POC
- Local dev environment with `sqlc`
- SQL migrations with `golang-migrate`
- Start integrating Uber-FX for dependency injection

## Things to investigate

**General**

- [x] Figure out what the fuck is a context?!?
- [x] Figure out how to do testing
- [x] Figure out how to format code automatically
- [x] Figure out fx modules
- [x] Create realistic input
- [x] Define deploy-move API
- [x] Implement deploy-move API
  - [x] Polish up deploy phase
  - [x] Test deploy phase
  - [x] Expose POST endpoint for deploy phase 
  - [x] Emit `GameStateChanged` in `SetGamePhaseQ`
  - [x] Handle `GameStateChanged`
- [ ] Improve region-to-player assignment
- [ ] Improve troops-region assignment
- [ ] Check if game is ended
- [ ] Figure out authentication
- [ ] When fetching state and the game is not there, err should be returned but instead we get empty array for player state and board state. Decouple state fetching from message broadcasting for better handling

**Front-end**

- [x] Understand how to start with frontend
- [x] Display full state
- [ ] Implement deploy-move API
- [ ] Figure out how to specify GameID when connecting via websockets (after OnOpen, send a "subscribe" message with GameID?)

**Server: Architecture**

- [x] Figure out how to split code in packages
- [x] Figure out how to determine whose turn it is in a game
- [x] Add a db provider
- [ ] Figure out error handling and propagation

**Server: Deserialization**

- [x] Byte-array to JSON conversion and vice-versa
- [x] Message handler component with message structs

**Server: Logic**

- [x] Game logic
- [ ] Map validation
- [x] POST endpoint for creating a game
- [x] Deploy move
- [ ] Attack move
- [ ] Reinforcement move
- [ ] Cards move

**Server: Analytics**

- [ ] We could store logs of game events, by sending them somewhere

**Server: Auth**

- [ ] Figure out TLS for websockets
- [ ] Store playerID and authToken in HTTP request headers
- [ ] Figure out where and how to store user credentials
