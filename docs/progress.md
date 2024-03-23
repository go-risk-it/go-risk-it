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
- [ ] Improve region-to-player assignment
- [ ] Improve troops-region assignment
- [x] Define deploy-move API
- [ ] Implement deploy-move API
  - [ ] Polish up deploy phase
  - [ ] Test deploy phase
  - [ ] Expose POST endpoint for deploy phase
- [ ] Store playerID and gameID in HTTP/WS headers

**Front-end**

- [x] Understand how to start with frontend
- [x] Implement deploy-move API

**Server: Architecture**

- [x] Figure out how to split code in packages
- [x] Figure out how to determine whose turn it is in a game
- [x] Add a db provider

**Server: Deserialization**

- [x] Byte-array to JSON conversion and vice-versa
- [x] Message handler component with message structs

**Server: Logic**

- [x] Game logic
- [ ] Map validation
- [ ] POST endpoint for creating a game
- [x] Frontend - Display full state

- [ ] Deploy move
- [ ] Attack move
- [ ] Reinforcement move

**Server: Analytics**

- [ ] We could store logs of game events, by sending them somewhere

**Server: Auth**

- [ ] Figure out TLS for websockets
- [ ] Figure out where and how to store user credentials
