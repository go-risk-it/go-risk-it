// Package message defines the WebSocket message envelope format.
//
// Every message sent over a game or lobby WebSocket is a [Message] with
// a discriminating [Type] field and a JSON payload. [BuildMessage]
// serializes any payload into this envelope, producing a json.RawMessage
// ready for dispatch.
//
// # Layer
//
// Web — WebSocket message serialization.
package message
