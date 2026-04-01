// Package restutils provides shared HTTP response and request helpers.
//
// [DecodeRequest] decodes and validates JSON request bodies with strict
// parsing (unknown fields rejected, size limited to 1 MB). [WriteJSON]
// marshals a typed payload to JSON and writes it with the given status code.
// [WriteError] and [WriteErrorWithTrace] map domain error categories to HTTP
// status codes and produce a standard [ErrorResponse] JSON envelope.
//
// # Layer
//
// Web — HTTP request decoding and error response formatting.
package restutils
