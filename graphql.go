package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/shurcooL/graphql/internal/jsonutil"
	"golang.org/x/net/context/ctxhttp"
)

// Client is a GraphQL client.
type Client struct {
	url         string // GraphQL server URL.
	httpClient  *http.Client
	queryString queryType
}

// NewClient creates a GraphQL client targeting the specified GraphQL server URL.
// If httpClient is nil, then http.DefaultClient is used.
func NewClient(url string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		url:         url,
		httpClient:  httpClient,
		queryString: disabled,
	}
}

// EnableQueryString enables query string mode for graphql queries
func (c *Client) EnableQueryString() *Client {
	c.queryString = enabled
	return c
}

// DisableQueryString disables query string mode for graphql queries
func (c *Client) DisableQueryString() *Client {
	c.queryString = disabled
	return c
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
		return err
	}

	var resp *http.Response
	if op == queryOperation && c.queryString == enabled {
		resp, err = GetWithQueryString(ctx, c.httpClient, c.url, query, variables)
	} else {
		resp, err = ctxhttp.Post(ctx, c.httpClient, c.url, "application/json", &buf)
	}

	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("non-200 OK status code: %v body: %q", resp.Status, body)
	}
	var out struct {
		Data   *json.RawMessage
		Errors errors
		//Extensions interface{} // Unused.
	}
	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		// TODO: Consider including response body in returned error, if deemed helpful.
		return err
	}
	if out.Data != nil {
		err := jsonutil.UnmarshalGraphQL(*out.Data, v)
		if err != nil {
			// TODO: Consider including response body in returned error, if deemed helpful.
			return err
		}
	}
	if len(out.Errors) > 0 {
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

// GetWithQueryString sends an http get request with the query and variables as a query string
func GetWithQueryString(ctx context.Context, client *http.Client, graphqlURL string, query string, variables map[string]interface{}) (*http.Response, error) {
	queryString := url.QueryEscape(query)
	variableBytes, err := json.Marshal(variables)
	if err != nil {
		return &http.Response{}, err
	}
	variableString := url.QueryEscape(string(variableBytes))
	resp, err := ctxhttp.Get(ctx, client, graphqlURL+`?query=`+queryString+`&variables=`+variableString)
	return resp, err
}

type queryType uint8

const (
	disabled queryType = iota
	enabled
)
