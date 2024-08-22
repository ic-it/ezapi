package ezapi

import "encoding/json"

// Unmarshal Error (400)
type DefaultUnmarshalError struct {
	Err error
}

func (e DefaultUnmarshalError) Error() string {
	return e.Err.Error()
}

func (e DefaultUnmarshalError) Render(ctx BaseContext) error {
	w := ctx.GetW()
	w.WriteHeader(400)
	errorBody := EzAPIError{Message: "Error unmarshalling request: " + e.Error()}
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(errorBody)
}

// Internal Server Error (500)
type DefaultInternalError struct {
	Err error
}

func (e DefaultInternalError) Error() string {
	return e.Err.Error()
}

func (e DefaultInternalError) Render(ctx BaseContext) error {
	w := ctx.GetW()
	w.WriteHeader(500)
	errorBody := EzAPIError{Message: "Internal server error: " + e.Error()}
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(errorBody)
}

type EzAPIError struct {
	Message string `json:"message"`
}
