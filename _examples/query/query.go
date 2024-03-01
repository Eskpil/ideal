package main

import (
	"encoding/json"
	"fmt"

	"github.com/eskpil/ideal"
	"github.com/graphql-go/graphql"
)

func main() {
	builder := ideal.NewBuilder(UserResolver)
	runtime := ideal.NewRuntime()
	schema, err := builder.Build(runtime)
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
