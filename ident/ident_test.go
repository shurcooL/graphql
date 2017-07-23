package ident_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/shurcooL/graphql/ident"
)

func Example_lowerCamelCaseToMixedCaps() {
	fmt.Println(ident.ParseLowerCamelCase("clientMutationId").ToMixedCaps())

	// Output: ClientMutationID
}

func Example_screamingSnakeCaseToMixedCaps() {
	fmt.Println(ident.ParseScreamingSnakeCase("CLIENT_MUTATION_ID").ToMixedCaps())

	// Output: ClientMutationID
}

func Example_mixedCapsToLowerCamelCase() {
	fmt.Println(ident.ParseMixedCaps("ClientMutationID").ToLowerCamelCase())

	// Output: clientMutationId
}

func TestParseMixedCaps(t *testing.T) {
	tests := []struct {
		in   string
		want ident.Name
	}{
		{in: "ClientMutationID", want: ident.Name{"Client", "Mutation", "ID"}},
		{in: "StringURLAppend", want: ident.Name{"String", "URL", "Append"}},
		{in: "URLFrom", want: ident.Name{"URL", "From"}},
		{in: "SetURL", want: ident.Name{"Set", "URL"}},
		{in: "UIIP", want: ident.Name{"UI", "IP"}},
		{in: "URLHTMLFrom", want: ident.Name{"URL", "HTML", "From"}},
		{in: "SetURLHTML", want: ident.Name{"Set", "URL", "HTML"}},
		{in: "HTTPSQL", want: ident.Name{"HTTP", "SQL"}},
		{in: "HTTPSSQL", want: ident.Name{"HTTPS", "SQL"}},
	}
	for _, tc := range tests {
		got := ident.ParseMixedCaps(tc.in)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("got: %q, want: %q", got, tc.want)
		}
	}
}

func TestParseLowerCamelCase(t *testing.T) {
	tests := []struct {
		in   string
		want ident.Name
	}{
		{in: "clientMutationId", want: ident.Name{"client", "Mutation", "Id"}},
	}
	for _, tc := range tests {
		got := ident.ParseLowerCamelCase(tc.in)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("got: %q, want: %q", got, tc.want)
		}
	}
}

func TestParseScreamingSnakeCase(t *testing.T) {
	tests := []struct {
		in   string
		want ident.Name
	}{
		{in: "CLIENT_MUTATION_ID", want: ident.Name{"CLIENT", "MUTATION", "ID"}},
	}
	for _, tc := range tests {
		got := ident.ParseScreamingSnakeCase(tc.in)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("got: %q, want: %q", got, tc.want)
		}
	}
}

func TestWords_ToMixedCaps(t *testing.T) {
	tests := []struct {
		in   ident.Name
		want string
	}{
		{in: ident.Name{"client", "Mutation", "Id"}, want: "ClientMutationID"},
		{in: ident.Name{"CLIENT", "MUTATION", "ID"}, want: "ClientMutationID"},
	}
	for _, tc := range tests {
		got := tc.in.ToMixedCaps()
		if got != tc.want {
			t.Errorf("got: %q, want: %q", got, tc.want)
		}
	}
}

func TestWords_ToLowerCamelCase(t *testing.T) {
	tests := []struct {
		in   ident.Name
		want string
	}{
		{in: ident.Name{"client", "Mutation", "Id"}, want: "clientMutationId"},
		{in: ident.Name{"CLIENT", "MUTATION", "ID"}, want: "clientMutationId"},
	}
	for _, tc := range tests {
		got := tc.in.ToLowerCamelCase()
		if got != tc.want {
			t.Errorf("got: %q, want: %q", got, tc.want)
		}
	}
}

func TestMixedCapsToLowerCamelCase(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "DatabaseID", want: "databaseId"},
		{in: "URL", want: "url"},
		{in: "ID", want: "id"},
		{in: "CreatedAt", want: "createdAt"},
		{in: "Login", want: "login"},
		{in: "ResetAt", want: "resetAt"},
	}
	for _, tc := range tests {
		got := ident.ParseMixedCaps(tc.in).ToLowerCamelCase()
		if got != tc.want {
			t.Errorf("got: %q, want: %q", got, tc.want)
		}
	}
}
