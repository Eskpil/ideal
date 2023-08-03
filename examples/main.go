package main

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/eskpil/ideal"
	"github.com/graphql-go/graphql"
)

type User struct {
	Name  string `gql:"name"`
	Email string `gql:"email"`
}

var hello = ideal.Query{
	Name: "hello",
	Type: reflect.TypeOf(User{}),

	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		return User{Name: "john doe", Email: "john@doe.com"}, nil
	},

	Description: "hello query returns user",
}

func main() {
	resolver := ideal.Resolver{}
	resolver.Query(hello)

	builder := ideal.NewBuilder(resolver)

	schema, err := builder.Build()
	if err != nil {
		panic(fmt.Sprintf("could not build schema: %v", err))
	}

	query := `{ hello { name, email }}`
	params := graphql.Params{Schema: schema, RequestString: query}

	result := graphql.Do(params)

	output, err := json.Marshal(result.Data)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(output))
	fmt.Println(result.Errors)
}
