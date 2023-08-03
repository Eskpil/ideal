# Ideal

Ideal is a graphql code first builder for golang. Ideal inspects your
provided types with the reflect package and produces graphql types for
you under the hood.

This package is greatly inspired by
[type-graphql](https://github.com/MichalLytek/type-graphql) over in the
node.js world.

## Documentation

Similar to the internal json package you can tag your struct fields to
produce a different name. Example

```go
package example

import "github.com/eskpil/ideal"

type User struct {
    Name  string `gql:"name"`
    Email string `gql:"email"`
}
```

You define your queries and mutations like this

```go
package example

import "github.com/eskpil/ideal"

var Hello = ideal.Query{
    Name: "hello",
    Type: reflect.TypeOf(User{})

    Resolve: func(params graphql.ResolveParams) (interface{}, error) {
        return User{Name: "john", Email: "john@doe.com"}
    }

    Description: "hello"
}
```

To produce our builder we need a builder

```go
package example

import "github.com/eskpil/ideal"

func main() {
	resolver := ideal.Resolver{}
	resolver.Query(hello)

	builder := ideal.NewBuilder(resolver)
	
	schema, err := builder.Build()
}
```

From this point you have a graphql schema like the one defined in 
[graphql-go](https://github.com/graphql-go/graphql) on which you can execute graphql requests.


```go
package example

import "github.com/eskpil/ideal"

func main() {
    params := graphql.Params{Schema: schema, RequestString: `{ hello {
    name, email } }`} 

    result := graphql.Do(params)
}
```

For a look at the complete example, [read](./examples/main.go)
