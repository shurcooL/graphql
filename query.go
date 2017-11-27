package graphql

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
	"sort"

	"github.com/shurcooL/graphql/ident"
)

func constructQuery(v interface{}, variables map[string]interface{}) string {
	query := query(v)
	if variables != nil {
		return "query(" + queryArguments(variables) + ")" + query
	}
	return query
}

func constructMutation(v interface{}, variables map[string]interface{}) string {
	query := query(v)
	if variables != nil {
		return "mutation(" + queryArguments(variables) + ")" + query
	}
	return "mutation" + query
}

// queryArguments constructs a minified arguments string for variables.
//
// E.g., map[string]interface{}{"a": Int(123), "b": NewBoolean(true)} -> "$a:Int!$b:Boolean".
func queryArguments(variables map[string]interface{}) string {
	sorted := make([]string, 0, len(variables))
	for k := range variables {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)
	var s string
	for _, k := range sorted {
		v := variables[k]
		s += "$" + k + ":"
		t := reflect.TypeOf(v)
		switch t.Kind() {
		case reflect.Slice, reflect.Array:
			// TODO: Support t.Elem() being a pointer, if needed. Probably want to do this recursively.
			s += "[" + t.Elem().Name() + "!]" // E.g., "[IssueState!]".
		case reflect.Ptr:
			// Pointer is an optional type, so no "!" at the end.
			s += t.Elem().Name() // E.g., "Int".
		default:
			name := t.Name()
			if name == "string" { // HACK: Workaround for https://github.com/shurcooL/githubql/issues/12.
				name = "ID"
			}
			// Value is a required type, so add "!" to the end.
			s += name + "!" // E.g., "Int!".
		}
	}
	return s
}

// query uses writeQuery to recursively construct
// a minified query string from the provided struct v.
//
// E.g., struct{Foo Int, BarBaz *Boolean} -> "{foo,barBaz}".
func query(v interface{}) string {
	var buf bytes.Buffer
	writeQuery(&buf, reflect.TypeOf(v), false)
	return buf.String()
}

// writeQuery writes a minified query for t to w. If inline is true,
// the struct fields of t are inlined into parent struct.
func writeQuery(w io.Writer, t reflect.Type, inline bool) {
	switch t.Kind() {
	case reflect.Ptr, reflect.Slice:
		writeQuery(w, t.Elem(), false)
	case reflect.Struct:
		// If the type implements json.Unmarshaler, it's a scalar. Don't expand it.
		if reflect.PtrTo(t).Implements(jsonUnmarshaler) {
			return
		}
		if !inline {
			io.WriteString(w, "{")
		}
		for i := 0; i < t.NumField(); i++ {
			if i != 0 {
				io.WriteString(w, ",")
			}
			f := t.Field(i)
			value, ok := f.Tag.Lookup("graphql")
			inlineField := f.Anonymous && !ok
			if !inlineField {
				if ok {
					io.WriteString(w, value)
				} else {
					io.WriteString(w, ident.ParseMixedCaps(f.Name).ToLowerCamelCase())
				}
			}
			writeQuery(w, f.Type, inlineField)
		}
		if !inline {
			io.WriteString(w, "}")
		}
	}
}

var jsonUnmarshaler = reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()
