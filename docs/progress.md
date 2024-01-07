# Progress

## Done so far

**Server**
- Websockets POC
- Local dev environment with `sqlc`
- SQL migrations with `golang-migrate`
- Start integrating Uber-FX for dependency injection

## Things to investigate

**Front-end**
- [ ] Understand how to start with frontend

**Server: Architecture**
- [ ] Figure out how to split code in packages
- [ ] Figure out how to determine whose turn it is in a game

**Server: Deserialization**
- [x] Byte-array to JSON conversion and vice-versa
- [ ] Message handler component with message structs

**Server: Logic**
- [ ] Game logic

**Server: Analytics**
- [ ] We could store logs of game events, by sending them somewhere

**Server: Auth**
- [ ] Figure out TLS for websockets
- [ ] Figure out where and how to store user credentials