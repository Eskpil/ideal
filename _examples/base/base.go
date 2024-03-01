package base

import (
	"reflect"

	"github.com/eskpil/ideal"
)

type User struct {
	Name  string `gql:"name"`
	Email string `gql:"email"`
}

var hello = ideal.Query{
	Name: "hello",
	Type: reflect.TypeOf(User{}),

	Resolve: func(c *ideal.Context) (interface{}, error) {
		return User{Name: "john doe", Email: "john@doe.com"}, nil
	},

	Description: "hello query returns user",
}

var UserResolver ideal.Resolver

func init() {
	UserResolver.Query(hello)
}
