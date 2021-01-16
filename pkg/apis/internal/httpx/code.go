package httpx

// Code is an custom HTTP code.
type Code int

// Msg returns the interpretation of code.
func (c Code) Msg() string {
	return Msgs[c]
}

// Common HTTP error code responsing to HTTP requests.
// It's in sync with Msgs.
const (
	Success Code = iota
	Failure

	ErrRequestDecodeJSON

	ErrAuthInvalidEmail
	ErrAuthInvalidVerificationCode
	ErrAuthVerificationCodeExpired
	ErrAuthInvalidOperation
	ErrAuthEmptyAlias
	ErrUnauthorized
	ErrAuthInvalidToken
	ErrAuthAlreadyDeregistered
	ErrAuthTokenExpired
	ErrAuthEmailAlreadyInUse

	ErrServiceUnavailable
)

// Msgs is an HTTP error code to flag map.
// It's in sync with Codes above.
var Msgs = map[Code]string{
	Success: "Success",
	Failure: "Failure",

	ErrRequestDecodeJSON: "Request JSON Decoding failed",

	ErrAuthInvalidEmail:            "Invalid email",
	ErrAuthInvalidVerificationCode: "Invalid verification code",
	ErrAuthVerificationCodeExpired: "Verification code expired",
	ErrAuthInvalidOperation:        "Invalid operation",
	ErrAuthEmptyAlias:              "Empty user alias",
	ErrUnauthorized:                "Unauthorized",
	ErrAuthInvalidToken:            "Invalid token",
	ErrAuthAlreadyDeregistered:     "Already deregistered",
	ErrAuthTokenExpired:            "Token expired",
	ErrAuthEmailAlreadyInUse:       "User email already in use",

	ErrServiceUnavailable: " Service unavailable",
}
