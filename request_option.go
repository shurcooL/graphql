package graphql

import "net/http"

// RequestOption allows you to modify the request before sending it to the
// server.
type RequestOption func(*http.Request) error

// WithHeader sets a header on the request.
func WithHeader(key, value string) RequestOption {
	return func(r *http.Request) error {
		r.Header.Set(key, value)
		return nil
	}
}
