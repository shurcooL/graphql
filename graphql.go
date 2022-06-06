package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/shurcooL/graphql/internal/jsonutil"
	"golang.org/x/net/context/ctxhttp"
)

// Client is a GraphQL client.
type Client struct {
	url        string // GraphQL server URL.
	httpClient *http.Client
}

// NewClient creates a GraphQL client targeting the specified GraphQL server URL.
// If httpClient is nil, then http.DefaultClient is used.
func NewClient(url string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		url:        url,
		httpClient: httpClient,
	}
}

// Query executes a single GraphQL query request,
// with a query derived from q, populating the response into it.
// q should be a pointer to struct that corresponds to the GraphQL schema.
func (c *Client) Query(ctx context.Context, q interface{}, variables map[string]interface{}) error {
	return c.do(ctx, queryOperation, q, variables)
}

// Mutate executes a single GraphQL mutation request,
// with a mutation derived from m, populating the response into it.
// m should be a pointer to struct that corresponds to the GraphQL schema.
func (c *Client) Mutate(ctx context.Context, m interface{}, variables map[string]interface{}) error {
	return c.do(ctx, mutationOperation, m, variables)
}

// do executes a single GraphQL operation.
func (c *Client) do(ctx context.Context, op operationType, v interface{}, variables map[string]interface{}) error {
	var query string
	switch op {
	case queryOperation:
		query = constructQuery(v, variables)
	case mutationOperation:
		query = constructMutation(v, variables)
	}
	in := struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables,omitempty"`
	}{
		Query:     query,
		Variables: variables,
	}
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(in)
	if err != nil {
		return JSONMarshalError(err)
	}
	resp, err := ctxhttp.Post(ctx, c.httpClient, c.url, "application/json", &buf)
	if err != nil {
		return HTTPRequestError(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err := HTTPResponseError{
			Status: resp.StatusCode,
		}
		body, e := ioutil.ReadAll(resp.Body)
		if e != nil {
			err.Err = e
		} else {
			err.Body = body
		}
		return err
	}
	var out struct {
		Data   *json.RawMessage
		Errors Errors
		//Extensions interface{} // Unused.
	}
	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		// TODO: Consider including response body in returned error, if deemed helpful.
		return JSONUnmarshalError(err)
	}
	if out.Data != nil {
		err := jsonutil.UnmarshalGraphQL(*out.Data, v)
		if err != nil {
			// TODO: Consider including response body in returned error, if deemed helpful.
			return UnmarshalError(err)
		}
	}
	if len(out.Errors) > 0 {
		return out.Errors
	}
	return nil
}

// JSONMarshalError represents the JSON Marshal error
// when encoding the request body.
type JSONMarshalError error

// JSONUnmarshalError represents the JSON Unmarshal error
// when decoding the response body.
type JSONUnmarshalError error

// HTTPRequestError represents the http client error
// when sending the request.
type HTTPRequestError error

// HTTPResponseError represents the http client error
// when the response has a non-200 status code.
type HTTPResponseError struct {
	Status int
	Body   []byte
	Err    error
}

// UnmarshalError represents the JSON Unmarshal error
// when decoding the GraphQL response data.
type UnmarshalError error

// Errors represents the "errors" array in a response from a GraphQL server.
// If returned via error interface, the slice is expected to contain at least 1 element.
//
// Specification: https://facebook.github.io/graphql/#sec-Errors.
type Errors []struct {
	Message   string
	Locations []struct {
		Line   int
		Column int
	}
	Path []interface{}
	Type string
}

// Error implements error interface.
func (e Errors) Error() string {
	return e[0].Message
}

// Error implements error interface.
func (e HTTPResponseError) Error() string {
	return fmt.Sprintf("non-200 OK status code: %d body: %q", e.Status, e.Body)
}

type operationType uint8

const (
	queryOperation operationType = iota
	mutationOperation
	//subscriptionOperation // Unused.
)
