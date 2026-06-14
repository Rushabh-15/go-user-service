package models

// DateLayout is the canonical wire format for dob in requests and responses.
const DateLayout = "2006-01-02"

// CreateUserRequest is the body for POST /users (and reused by PUT on Day 3).
type CreateUserRequest struct {
	Name string `json:"name" validate:"required"`
	Dob  string `json:"dob" validate:"required,datetime=2006-01-02"`
}

// UserResponse is returned by create/update. It deliberately omits age,
// matching the task spec.
type UserResponse struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
	Dob  string `json:"dob"`
}

// UserWithAgeResponse is returned by get/list. Age is a separate type
// (not omitempty) so that a valid age of 0 is still serialized.
type UserWithAgeResponse struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
	Dob  string `json:"dob"`
	Age  int    `json:"age"`
}

// ErrorResponse is the uniform error envelope for every endpoint.
type ErrorResponse struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details,omitempty"`
}
