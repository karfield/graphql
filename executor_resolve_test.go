package graphql_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/karfield/graphql"
	"github.com/karfield/graphql/testutil"
)

func testSchema(t *testing.T, testField *graphql.Field) graphql.Schema {
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"test": testField,
			},
		}),
	})
	if err != nil {
		t.Fatalf("Invalid schema: %v", err)
	}
	return schema
}

func TestExecutesResolveFunction_DefaultFunctionAccessesProperties(t *testing.T) {
	schema := testSchema(t, &graphql.Field{Type: graphql.String})

	source := map[string]interface{}{
		"test": "testValue",
	}

	expected := map[string]interface{}{
		"test": "testValue",
	}

	result, _ := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: `{ test }`,
		RootObject:    source,
	})
	if !reflect.DeepEqual(expected, result.Data) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result.Data))
	}
}

func TestExecutesResolveFunction_DefaultFunctionCallsMethods(t *testing.T) {
	schema := testSchema(t, &graphql.Field{Type: graphql.String})

	source := map[string]interface{}{
		"test": func() interface{} {
			return "testValue"
		},
	}

	expected := map[string]interface{}{
		"test": "testValue",
	}

	result, _ := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: `{ test }`,
		RootObject:    source,
	})
	if !reflect.DeepEqual(expected, result.Data) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result.Data))
	}
}

func TestExecutesResolveFunction_UsesProvidedResolveFunction(t *testing.T) {
	schema := testSchema(t, &graphql.Field{
		Type: graphql.String,
		Args: graphql.FieldConfigArgument{
			"aStr": &graphql.ArgumentConfig{Type: graphql.String},
			"aInt": &graphql.ArgumentConfig{Type: graphql.Int},
		},
		Resolve: graphql.ResolveField(func(p graphql.ResolveParams) (interface{}, error) {
			b, err := json.Marshal(p.Args)
			return string(b), err
		}),
	})

	expected := map[string]interface{}{
		"test": "{}",
	}
	result, _ := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: `{ test }`,
	})
	if !reflect.DeepEqual(expected, result.Data) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result.Data))
	}

	expected = map[string]interface{}{
		"test": `{"aStr":"String!"}`,
	}
	result, _ = graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: `{ test(aStr: "String!") }`,
	})
	if !reflect.DeepEqual(expected, result.Data) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result.Data))
	}

	expected = map[string]interface{}{
		"test": `{"aInt":-123,"aStr":"String!"}`,
	}
	result, _ = graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: `{ test(aInt: -123, aStr: "String!") }`,
	})
	if !reflect.DeepEqual(expected, result.Data) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result.Data))
	}
}

func TestExecutesResolveFunction_UsesProvidedResolveFunction_SourceIsStruct_WithoutJSONTags(t *testing.T) {

	// For structs without JSON tags, it will map to upper-cased exported field names
	type SubObjectWithoutJSONTags struct {
		Str string
		Int int
	}

	schema := testSchema(t, &graphql.Field{
		Type: graphql.NewObject(graphql.ObjectConfig{
			Name:        "SubObject",
			Description: "Maps GraphQL Object `SubObject` to Go struct `SubObjectWithoutJSONTags`",
			Fields: graphql.Fields{
				"Str": &graphql.Field{Type: graphql.String},
				"Int": &graphql.Field{Type: graphql.Int},
			},
		}),
		Args: graphql.FieldConfigArgument{
			"aStr": &graphql.ArgumentConfig{Type: graphql.String},
			"aInt": &graphql.ArgumentConfig{Type: graphql.Int},
		},
		Resolve: graphql.ResolveField(func(p graphql.ResolveParams) (interface{}, error) {
			aStr, _ := p.Args["aStr"].(string)
			aInt, _ := p.Args["aInt"].(int)
			return &SubObjectWithoutJSONTags{
				Str: aStr,
				Int: aInt,
			}, nil
		}),
	})

	expected := map[string]interface{}{
		"test": map[string]interface{}{
			"Str": "",
			"Int": 0,
		},
	}
	result, _ := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: `{ test { Str, Int } }`,
	})

	if !reflect.DeepEqual(expected, result.Data) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result.Data))
	}

	expected = map[string]interface{}{
		"test": map[string]interface{}{
			"Str": "String!",
			"Int": 0,
		},
	}
	result, _ = graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: `{ test(aStr: "String!") { Str, Int } }`,
	})
	if !reflect.DeepEqual(expected, result.Data) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result.Data))
	}

	expected = map[string]interface{}{
		"test": map[string]interface{}{
			"Str": "String!",
			"Int": -123,
		},
	}
	result, _ = graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: `{ test(aInt: -123, aStr: "String!") { Str, Int } }`,
	})
	if !reflect.DeepEqual(expected, result.Data) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result.Data))
	}
}

func TestExecutesResolveFunction_UsesProvidedResolveFunction_SourceIsStruct_WithJSONTags(t *testing.T) {

	// For structs without JSON tags, it will map to upper-cased exported field names
	type SubObjectWithJSONTags struct {
		OtherField string `json:""`
		Str        string `json:"str"`
		Int        int    `json:"int"`
	}

	schema := testSchema(t, &graphql.Field{
		Type: graphql.NewObject(graphql.ObjectConfig{
			Name:        "SubObject",
			Description: "Maps GraphQL Object `SubObject` to Go struct `SubObjectWithJSONTags`",
			Fields: graphql.Fields{
				"str": &graphql.Field{Type: graphql.String},
				"int": &graphql.Field{Type: graphql.Int},
			},
		}),
		Args: graphql.FieldConfigArgument{
			"aStr": &graphql.ArgumentConfig{Type: graphql.String},
			"aInt": &graphql.ArgumentConfig{Type: graphql.Int},
		},
		Resolve: graphql.ResolveField(func(p graphql.ResolveParams) (interface{}, error) {
			aStr, _ := p.Args["aStr"].(string)
			aInt, _ := p.Args["aInt"].(int)
			return &SubObjectWithJSONTags{
				Str: aStr,
				Int: aInt,
			}, nil
		}),
	})

	expected := map[string]interface{}{
		"test": map[string]interface{}{
			"str": "",
			"int": 0,
		},
	}
	result, _ := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: `{ test { str, int } }`,
	})

	if !reflect.DeepEqual(expected, result.Data) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result.Data))
	}

	expected = map[string]interface{}{
		"test": map[string]interface{}{
			"str": "String!",
			"int": 0,
		},
	}
	result, _ = graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: `{ test(aStr: "String!") { str, int } }`,
	})
	if !reflect.DeepEqual(expected, result.Data) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result.Data))
	}

	expected = map[string]interface{}{
		"test": map[string]interface{}{
			"str": "String!",
			"int": -123,
		},
	}
	result, _ = graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: `{ test(aInt: -123, aStr: "String!") { str, int } }`,
	})
	if !reflect.DeepEqual(expected, result.Data) {
		t.Fatalf("Unexpected result, Diff: %v", testutil.Diff(expected, result.Data))
	}
}

func testIsSelected(t *testing.T) graphql.Schema {
	return testSchema(t, &graphql.Field{
		Type: graphql.NewObject(graphql.ObjectConfig{
			Name: "SubObject",
			Fields: graphql.Fields{
				"str": &graphql.Field{
					Type: graphql.String,
					Resolve: graphql.ResolveField(func(p graphql.ResolveParams) (interface{}, error) {
						if !p.Info.IsFieldSelected("*") {
							t.Error("expect '/str' selected")
						}
						return 0, nil
					}),
				},
				"int": &graphql.Field{
					Type: graphql.Int,
				},
			},
		}),
		Resolve: graphql.ResolveField(func(p graphql.ResolveParams) (interface{}, error) {
			if !p.Info.IsFieldSelected("/str") {
				t.Error("expect '/str' selected")
			}
			if p.Info.IsFieldSelected("/int") {
				t.Error("expect '/int' not selected")
			}
			return struct{}{}, nil
		}),
	})
}

func TestExecutesResolveFunction_IsSelected_ByFieldAST(t *testing.T) {
	_, err := graphql.Do(graphql.Params{
		Schema:        testIsSelected(t),
		RequestString: `{ test { str } }`,
	})
	if err != nil {
		t.Error(err)
	}
}

func TestExecutesResolveFunction_IsSelected_ByFragment(t *testing.T) {
	_, err := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: `{ test { ... SubObjectFragment } } fragment SubObjectFragment on SubObject { str } `,
	})
	if err != nil {
		t.Error(err)
	}
}

func TestExecutesResolveFunction_IsSelected_ByInlineFragment(t *testing.T) {
	_, err := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: `{ test { ... on SubObject { str } } }`,
	})
	if err != nil {
		t.Error(err)
	}
}
