package httpx

import (
	"encoding/json"
	"net/http"
)

// Code is an custom HTTP code.
// It also implements error interface.
type Code int

func (c Code) Error() string {
	return Msgs[c]
}

// Common HTTP error code responsing to HTTP requests.
// It's in sync with Msgs.
const (
	Success Code = iota
	Failure

	ErrRequestDecodeJSON
	ErrInvalidEmail

	ErrInternalServer
)

// Msgs is an HTTP error code to flag map.
// It's in sync with Codes above.
var Msgs = map[Code]string{
	Success: "Success",
	Failure: "Failure",

	ErrRequestDecodeJSON: "Request JSON Decoding failed",
	ErrInvalidEmail:      "Invalid email",

	ErrInternalServer: "Internal server error",
}

// FinalResponse is a final uniform response to any request.
type FinalResponse struct {
	Code Code        `json:"code"`
	Msg  string      `json:"message"`
	Data interface{} `json:"data"`
}

// FinalizeResponse finalizes a response in an handler. It should panic when any error occurs
// so that the top recovery middleware could cache it.
func FinalizeResponse(w http.ResponseWriter, code Code, data interface{}) {
	if err := json.NewEncoder(w).Encode(&FinalResponse{
		Code: code,
		Msg:  Msgs[code],
		Data: data,
	}); err != nil {
		panic(err)
	}
}
