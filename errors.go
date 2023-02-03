package graphql

import "fmt"

// RequestError represents an error building a request.
type RequestError struct{ Err error }

func (e *RequestError) Error() string { return fmt.Sprintf("request error: %v", e.Err) }
func (e *RequestError) Unwrap() error { return e.Err }

// OptionError represents an error modifiying a request.
type OptionError struct{ Err error }

func (e *OptionError) Error() string { return fmt.Sprintf("request option error: %v", e.Err) }
func (e *OptionError) Unwrap() error { return e.Err }

// ResponseError represents a response error, either with getting a response
// from the server (eg. network error), or reading its body.
type ResponseError struct{ Err error }

func (e *ResponseError) Error() string { return fmt.Sprintf("request option error: %v", e.Err) }
func (e *ResponseError) Unwrap() error { return e.Err }

// ServerError indicates that the server returned a response but it was not what
// we consider a successful one.
type ServerError struct {
	Body   []byte
	Status string
}

func (e *ServerError) Error() string {
	return fmt.Sprintf("non-200 OK status code: %v body: %q", e.Status, e.Body)
}

// BodyError indicates that the server responded with the right status code but
// the body was unexpected and it did not parse as a valid GraphQL response.
type BodyError struct {
	Body []byte
	Err  error
}

func (e *BodyError) Error() string {
	return fmt.Sprintf("could not parse the body: %v, body: %q", e.Err, e.Body)
}

func (e *BodyError) Unwrap() error { return e.Err }
