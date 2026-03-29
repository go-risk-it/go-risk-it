package route

import (
	"net/http"
)

// PlainHandler handles requests with standard http types, returning errors.
type PlainHandler func(http.ResponseWriter, *http.Request) error
