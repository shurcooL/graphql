package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/shurcooL/graphql/internal/jsonutil"
	"golang.org/x/net/context/ctxhttp"
)

// Client is a GraphQL client.
type Client struct {
	url            string // GraphQL server URL.
	httpClient     *http.Client
	requestOptions []RequestOption
}

// NewClient creates a GraphQL client targeting the specified GraphQL server URL.
// If httpClient is nil, then http.DefaultClient is used.
func NewClient(url string, httpClient *http.Client, opts ...RequestOption) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		url:            url,
		httpClient:     httpClient,
		requestOptions: opts,
	}
}

// Query executes a single GraphQL query request,
// with a query derived from q, populating the response into it.
// q should be a pointer to struct that corresponds to the GraphQL schema.
func (c *Client) Query(ctx context.Context, q interface{}, variables map[string]interface{}, opts ...RequestOption) error {
	return c.do(ctx, queryOperation, q, variables, opts)
}

// Mutate executes a single GraphQL mutation request,
// with a mutation derived from m, populating the response into it.
// m should be a pointer to struct that corresponds to the GraphQL schema.
func (c *Client) Mutate(ctx context.Context, m interface{}, variables map[string]interface{}, opts ...RequestOption) error {
	return c.do(ctx, mutationOperation, m, variables, opts)
}

// do executes a single GraphQL operation.
func (c *Client) do(ctx context.Context, op operationType, v interface{}, variables map[string]interface{}, opts []RequestOption) error {
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
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.url, &buf)
	if err != nil {
		return &RequestError{Err: err}
	}
	req.Header.Set("Content-Type", "application/json")

	var allOpts []RequestOption
	allOpts = append(allOpts, c.requestOptions...)
	allOpts = append(allOpts, opts...)

	for _, opt := range allOpts {
		if err := opt(req); err != nil {
			return &OptionError{Err: err}
		}
	}

	resp, err := ctxhttp.Do(ctx, c.httpClient, req)
	if err != nil {
		return &ResponseError{Err: err}
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &ResponseError{Err: err}
	}

	if resp.StatusCode != http.StatusOK {
		return &ServerError{
			Body:   body,
			Status: resp.Status,
		}
	}

	var out struct {
		Data       *json.RawMessage
		Errors     errors
		Extensions interface{}
	}
	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		return &BodyError{Err: err, Body: body}
	}

	if out.Data != nil {
		err := jsonutil.UnmarshalGraphQL(*out.Data, v)
		if err != nil {
			return &BodyError{Err: err, Body: body}
		}
	}

	if len(out.Errors) > 0 {
		if out.Extensions != nil {
			return newErrorsWithExtensions(out.Errors, out.Extensions)
		}

		return out.Errors
	}

	return nil
}

// errors represents the "errors" array in a response from a GraphQL server.
// If returned via error interface, the slice is expected to contain at least 1 element.
//
// Specification: https://facebook.github.io/graphql/#sec-Errors.
type errors []struct {
	Message   string
	Locations []struct {
		Line   int
		Column int
	}
}

// Error implements error interface.
func (e errors) Error() string {
	return e[0].Message
}

type operationType uint8

const (
	queryOperation operationType = iota
	mutationOperation
	//subscriptionOperation // Unused.
)

type ErrorsWithExtensions struct {
	errors     errors
	extensions interface{}
}

func newErrorsWithExtensions(err errors, extensions interface{}) ErrorsWithExtensions {
	return ErrorsWithExtensions{errors: err, extensions: extensions}
}

func (e ErrorsWithExtensions) Error() string {
	return e.errors[0].Message
}

func (e ErrorsWithExtensions) Extensions() interface{} {
	return e.extensions
}
