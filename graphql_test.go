package graphjson_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/Code-Hex/go-graphjson"
	"github.com/google/go-cmp/cmp"
	"github.com/shurcooL/graphql"
)

func TestUnmarshal(t *testing.T) {
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
	err := graphjson.Unmarshal([]byte(`{
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

func TestUnmarshal_graphqlTag(t *testing.T) {
	type query struct {
		Foo graphql.String `graphql:"baz"`
	}
	var got query
	err := graphjson.Unmarshal([]byte(`{
		"baz": "bar"
	}`), &got)
	if err != nil {
		t.Fatal(err)
	}
	want := query{
		Foo: "bar",
	}
	if !reflect.DeepEqual(got, want) {
		t.Error("not equal")
	}
}

func TestUnmarshal_jsonTag(t *testing.T) {
	type query struct {
		Foo graphql.String `json:"baz"`
	}
	var got query
	err := graphjson.Unmarshal([]byte(`{
		"foo": "bar"
	}`), &got)
	if err != nil {
		t.Fatal(err)
	}
	want := query{
		Foo: "bar",
	}
	if !reflect.DeepEqual(got, want) {
		t.Error("not equal")
	}
}

func TestUnmarshal_array(t *testing.T) {
	type query struct {
		Foo []graphql.String
		Bar []graphql.String
		Baz []graphql.String
	}
	var got query
	err := graphjson.Unmarshal([]byte(`{
		"foo": [
			"bar",
			"baz"
		],
		"bar": [],
		"baz": null
	}`), &got)
	if err != nil {
		t.Fatal(err)
	}
	want := query{
		Foo: []graphql.String{"bar", "baz"},
		Bar: []graphql.String{},
		Baz: []graphql.String(nil),
	}
	if !reflect.DeepEqual(got, want) {
		t.Error("not equal")
	}
}

// When unmarshaling into an array, its initial value should be overwritten
// (rather than appended to).
func TestUnmarshal_arrayReset(t *testing.T) {
	var got = []string{"initial"}
	err := graphjson.Unmarshal([]byte(`["bar", "baz"]`), &got)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"bar", "baz"}
	if !reflect.DeepEqual(got, want) {
		t.Error("not equal")
	}
}

func TestUnmarshal_objectArray(t *testing.T) {
	type query struct {
		Foo []struct {
			Name graphql.String
		}
	}
	var got query
	err := graphjson.Unmarshal([]byte(`{
		"foo": [
			{"name": "bar"},
			{"name": "baz"}
		]
	}`), &got)
	if err != nil {
		t.Fatal(err)
	}
	want := query{
		Foo: []struct{ Name graphql.String }{
			{"bar"},
			{"baz"},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Error("not equal")
	}
}

func TestUnmarshal_pointer(t *testing.T) {
	type query struct {
		Foo *graphql.String
		Bar *graphql.String
	}
	var got query
	got.Bar = new(graphql.String) // Test that got.Bar gets set to nil.
	err := graphjson.Unmarshal([]byte(`{
		"foo": "foo",
		"bar": null
	}`), &got)
	if err != nil {
		t.Fatal(err)
	}
	want := query{
		Foo: graphql.NewString("foo"),
		Bar: nil,
	}
	if !reflect.DeepEqual(got, want) {
		t.Error("not equal")
	}
}

func TestUnmarshal_objectPointerArray(t *testing.T) {
	type query struct {
		Foo []*struct {
			Name graphql.String
		}
	}
	var got query
	err := graphjson.Unmarshal([]byte(`{
		"foo": [
			{"name": "bar"},
			null,
			{"name": "baz"}
		]
	}`), &got)
	if err != nil {
		t.Fatal(err)
	}
	want := query{
		Foo: []*struct{ Name graphql.String }{
			{"bar"},
			nil,
			{"baz"},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Error("not equal")
	}
}

func TestUnmarshal_pointerWithInlineFragment(t *testing.T) {
	type actor struct {
		User struct {
			DatabaseID uint64
		} `graphql:"... on User"`
		Login string
	}
	type query struct {
		Author actor
		Editor *actor
	}
	var got query
	err := graphjson.Unmarshal([]byte(`{
		"author": {
			"databaseId": 1,
			"login": "test1"
		},
		"editor": {
			"databaseId": 2,
			"login": "test2"
		}
	}`), &got)
	if err != nil {
		t.Fatal(err)
	}
	var want query
	want.Author = actor{
		User:  struct{ DatabaseID uint64 }{1},
		Login: "test1",
	}
	want.Editor = &actor{
		User:  struct{ DatabaseID uint64 }{2},
		Login: "test2",
	}

	if !reflect.DeepEqual(got, want) {
		t.Error("not equal")
	}
}

func TestUnmarshal_unexportedField(t *testing.T) {
	type query struct {
		foo graphql.String
	}
	err := graphjson.Unmarshal([]byte(`{"foo": "bar"}`), new(query))
	if err == nil {
		t.Fatal("got error: nil, want: non-nil")
	}
	if got, want := err.Error(), "struct field for \"foo\" doesn't exist in any of 1 places to unmarshal"; got != want {
		t.Errorf("got error: %v, want: %v", got, want)
	}
}

func TestUnmarshal_multipleValues(t *testing.T) {
	type query struct {
		Foo graphql.String
	}
	err := graphjson.Unmarshal([]byte(`{"foo": "bar"}{"foo": "baz"}`), new(query))
	if err == nil {
		t.Fatal("got error: nil, want: non-nil")
	}
	if got, want := err.Error(), "invalid token '{' after top-level value"; got != want {
		t.Errorf("got error: %v, want: %v", got, want)
	}
}

func TestUnmarshal_union(t *testing.T) {
	/*
		{
			__typename
			... on ClosedEvent {
				createdAt
				actor {login}
			}
			... on ReopenedEvent {
				createdAt
				actor {login}
			}
		}
	*/
	type actor struct{ Login graphql.String }
	type closedEvent struct {
		Actor     actor
		CreatedAt time.Time
	}
	type reopenedEvent struct {
		Actor     actor
		CreatedAt time.Time
	}
	type issueTimelineItem struct {
		ClosedEvent   closedEvent   `graphql:"... on ClosedEvent"`
		ReopenedEvent reopenedEvent `graphql:"... on ReopenedEvent"`
	}
	var got issueTimelineItem
	err := graphjson.Unmarshal([]byte(`{
		"createdAt": "2017-06-29T04:12:01Z",
		"actor": {
			"login": "shurcooL-test"
		}
	}`), &got)
	if err != nil {
		t.Fatal(err)
	}
	want := issueTimelineItem{
		ClosedEvent: closedEvent{
			Actor: actor{
				Login: "shurcooL-test",
			},
			CreatedAt: time.Unix(1498709521, 0).UTC(),
		},
		ReopenedEvent: reopenedEvent{
			Actor: actor{
				Login: "shurcooL-test",
			},
			CreatedAt: time.Unix(1498709521, 0).UTC(),
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want, +got)\n%s", diff)
	}
}

func TestUnmarshal_union_typename(t *testing.T) {
	/*
		{
			__typename
			... on ClosedEvent {
				createdAt
				actor {login}
			}
			... on ReopenedEvent {
				createdAt
				actor {login}
			}
		}
	*/
	type actor struct{ Login graphql.String }
	type closedEvent struct {
		Actor     actor
		CreatedAt time.Time
	}
	type reopenedEvent struct {
		Actor     actor
		CreatedAt time.Time
	}
	type issueTimelineItem struct {
		ClosedEvent   closedEvent   `graphql:"... on ClosedEvent"`
		ReopenedEvent reopenedEvent `graphql:"... on ReopenedEvent"`
		Typename      string        `graphql:"__typename"`
	}
	var got issueTimelineItem
	err := graphjson.Unmarshal([]byte(`{
		"createdAt": "2017-06-29T04:12:01Z",
		"actor": {
			"login": "shurcooL-test"
		},
		"__typename": "ClosedEvent"
	}`), &got)
	if err != nil {
		t.Fatal(err)
	}
	want := issueTimelineItem{
		Typename: "ClosedEvent",
		ClosedEvent: closedEvent{
			Actor: actor{
				Login: "shurcooL-test",
			},
			CreatedAt: time.Unix(1498709521, 0).UTC(),
		},
		ReopenedEvent: reopenedEvent{},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want, +got)\n%s", diff)
	}
}

func TestUnmarshal_union_typename_in_array(t *testing.T) {
	type actor struct{ Login graphql.String }
	type closedEvent struct {
		Actor     actor
		CreatedAt time.Time
	}
	type reopenedEvent struct {
		Actor     actor
		CreatedAt time.Time
	}
	type issueTimelineItem struct {
		ClosedEvent   closedEvent   `graphql:"... on ClosedEvent"`
		ReopenedEvent reopenedEvent `graphql:"... on ReopenedEvent"`
		Typename      *string       `graphql:"__typename"`
	}
	type events struct {
		Foo []issueTimelineItem `graphql:"foo"`
	}
	var got events
	err := graphjson.Unmarshal([]byte(`{
		"foo": [
			{
				"createdAt": "2017-06-29T04:12:01Z",
				"actor": {
					"login": "shurcooL-test"
				},
				"__typename": "ClosedEvent"
			},
			{
				"createdAt": "2017-06-29T04:12:01Z",
				"actor": {
					"login": "shurcooL-test2"
				},
				"__typename": "ReopenedEvent"
			}
		]
	}`), &got)
	if err != nil {
		t.Fatal(err)
	}
	want := events{
		Foo: []issueTimelineItem{
			{
				Typename: stringP("ClosedEvent"),
				ClosedEvent: closedEvent{
					Actor: actor{
						Login: "shurcooL-test",
					},
					CreatedAt: time.Unix(1498709521, 0).UTC(),
				},
				ReopenedEvent: reopenedEvent{},
			},
			{
				Typename: stringP("ReopenedEvent"),
				ReopenedEvent: reopenedEvent{
					Actor: actor{
						Login: "shurcooL-test2",
					},
					CreatedAt: time.Unix(1498709521, 0).UTC(),
				},
				ClosedEvent: closedEvent{},
			},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want, +got)\n%s", diff)
	}
}

func stringP(s string) *string {
	return &s
}

// Issue https://github.com/shurcooL/githubv4/issues/18.
func TestUnmarshal_arrayInsideInlineFragment(t *testing.T) {
	/*
		query {
			search(type: ISSUE, first: 1, query: "type:pr repo:owner/name") {
				nodes {
					... on PullRequest {
						commits(last: 1) {
							nodes {
								url
							}
						}
					}
				}
			}
		}
	*/
	type query struct {
		Search struct {
			Nodes []struct {
				PullRequest struct {
					Commits struct {
						Nodes []struct {
							URL string `graphql:"url"`
						}
					} `graphql:"commits(last: 1)"`
				} `graphql:"... on PullRequest"`
			}
		} `graphql:"search(type: ISSUE, first: 1, query: \"type:pr repo:owner/name\")"`
	}
	var got query
	err := graphjson.Unmarshal([]byte(`{
		"search": {
			"nodes": [
				{
					"commits": {
						"nodes": [
							{
								"url": "https://example.org/commit/49e1"
							}
						]
					}
				}
			]
		}
	}`), &got)
	if err != nil {
		t.Fatal(err)
	}
	var want query
	want.Search.Nodes = make([]struct {
		PullRequest struct {
			Commits struct {
				Nodes []struct {
					URL string `graphql:"url"`
				}
			} `graphql:"commits(last: 1)"`
		} `graphql:"... on PullRequest"`
	}, 1)
	want.Search.Nodes[0].PullRequest.Commits.Nodes = make([]struct {
		URL string `graphql:"url"`
	}, 1)
	want.Search.Nodes[0].PullRequest.Commits.Nodes[0].URL = "https://example.org/commit/49e1"
	if !reflect.DeepEqual(got, want) {
		t.Error("not equal")
	}
}
