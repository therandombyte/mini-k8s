// before writing handlers, need supporting functions
// this one will help to return a json http response which handlers can reuse

package util

import (
	"encoding/json"
	"net/http"
)

// encode obj as JSON into the response body
func WriteJSON(w http.ResponseWriter, code int, obj any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(obj)
}
