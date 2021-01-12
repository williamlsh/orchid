package httpx

import (
	"encoding/json"
	"net/http"
)

// FinalResponse is a final uniform response to any request.
type FinalResponse struct {
	Code Code        `json:"code"`
	Msg  string      `json:"message"`
	Data interface{} `json:"data,omitempty"`
}

// FinalizeResponse finalizes a response in an handler. It should panic when any error occurs
// so that the top recovery middleware could cache it.
func FinalizeResponse(w http.ResponseWriter, code Code, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(&FinalResponse{
		Code: code,
		Msg:  Msgs[code],
		Data: data,
	}); err != nil {
		panic(err)
	}
}
