package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/eskpil/ideal"
	"github.com/eskpil/ideal/_examples/base"
	"github.com/graphql-go/graphql"
)

func main() {
	runtime := ideal.NewRuntime()
	builder := ideal.NewBuilder(base.UserResolver)

	schema, err := builder.Build(runtime)
	if err != nil {
		panic(err)
	}

	runtimeParams := new(ideal.RuntimeParameters)

	runtimeParams.Headers = nil /* from our http framework */
	runtimeParams.Cookies = nil /* from our http framework*/

	// replace context.Background() with the context provided by the http framework
	ctx := context.WithValue(context.Background(), ideal.RuntimeParametersKey, runtimeParams)

	var input struct {
		Query     string                 `json:"query"`
		Operation string                 `json:"operationName"`
		Variables map[string]interface{} `json:"variables"`
	}

	// input is our json request body. Use the framework provided metehods of reading
	// the body. For the sake of the example we fill in the data manually.
	input.Query = `{ hello { name, email }}`

	params := graphql.Params{
		Context:        ctx,
		Schema:         schema,
		RequestString:  input.Query,
		VariableValues: input.Variables,
		OperationName:  input.Operation,
	}

	result := graphql.Do(params)
	output, _ := json.Marshal(result)
	fmt.Println(string(output))
}
