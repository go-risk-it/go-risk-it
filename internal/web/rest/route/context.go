package route

import (
	"fmt"
	"net/http"
	"strconv"
)

// ExtractID parses the {id} path parameter from the request as an int64.
func ExtractID(r *http.Request) (int64, error) {
	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return -1, fmt.Errorf("invalid id: %w", err)
	}

	return int64(id), nil
}
