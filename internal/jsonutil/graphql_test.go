package jsonutil_test

import (
	"reflect"
	"testing"

	"github.com/shurcooL/graphql"
	"github.com/shurcooL/graphql/internal/jsonutil"
)

func TestUnmarshalGraphQL(t *testing.T) {
	/*
		query {
			me {
				name
				height
			}
		}
	*/
	type query struct {
		Me struct {
			Name   graphql.String
			Height graphql.Float
		}
	}
	var got query
	err := jsonutil.UnmarshalGraphQL([]byte(`{
		"me": {
			"name": "Luke Skywalker",
			"height": 1.72
		}
	}`), &got)
	if err != nil {
		t.Fatal(err)
	}
	var want query
	want.Me.Name = "Luke Skywalker"
	want.Me.Height = 1.72
	if !reflect.DeepEqual(got, want) {
		t.Error("not equal")
	}
}

// TODO: Port the rest of the jsonutil tests from githubql.
