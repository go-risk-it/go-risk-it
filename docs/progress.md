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

**Front-end**

- [ ] Understand how to start with frontend

**Server: Architecture**

- [x] Figure out how to split code in packages
- [ ] Figure out how to determine whose turn it is in a game
- [x] Add a db provider

**Server: Deserialization**

- [x] Byte-array to JSON conversion and vice-versa
- [ ] Message handler component with message structs

**Server: Logic**

- [ ] Game logic
- [ ] Map validation
- [ ] POST endpoint for creating a game
- [ ] GetFullState
- [ ] Frontend - Display full state

- [ ] Deploy move
- [ ] Attack move
- [ ] Reinforcement move

**Server: Analytics**

- [ ] We could store logs of game events, by sending them somewhere

**Server: Auth**

- [ ] Figure out TLS for websockets
- [ ] Figure out where and how to store user credentials
