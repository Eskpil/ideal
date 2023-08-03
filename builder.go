package ideal

import (
	"fmt"
	"github.com/graphql-go/graphql"
	"reflect"
	"strings"
)

type Query struct {
	Name string

	Arguments reflect.Type
	Type      reflect.Type
	Resolve   graphql.FieldResolveFn

	Description       string
	DeprecationReason string
}

type Mutation struct {
	Name string

	Arguments reflect.Type
	Type      reflect.Type
	Resolve   graphql.FieldResolveFn

	Description       string
	DeprecationReason string
}

type Resolver struct {
	Queries   []Query
	Mutations []Mutation
}

type Builder struct {
	Resolvers []*Resolver

	objectCache map[reflect.Type]*graphql.Object
	inputCache  map[reflect.Type]graphql.FieldConfigArgument

	rootQuery        *graphql.Object
	rootMutation     *graphql.Object
	rootSubscription *graphql.Object
}

func (r *Resolver) Query(query Query) {
	r.Queries = append(r.Queries, query)
}

func NewBuilder(resolvers []*Resolver) *Builder {
	builder := new(Builder)

	builder.Resolvers = resolvers

	builder.objectCache = make(map[reflect.Type]*graphql.Object)
	builder.inputCache = make(map[reflect.Type]graphql.FieldConfigArgument)

	return builder
}

func (s *Builder) introspectField(r reflect.Type) graphql.Type {
	switch r.Kind() {
	case reflect.Bool:
		return graphql.Boolean
	case reflect.Int:
	case reflect.Int8:
	case reflect.Int16:
	case reflect.Int32:
	case reflect.Int64:
	case reflect.Uint:
	case reflect.Uint8:
	case reflect.Uint16:
	case reflect.Uint32:
	case reflect.Uint64:
	case reflect.Uintptr:
		return graphql.Int
	case reflect.Float32:
	case reflect.Float64:
		return graphql.Float
	case reflect.String:
		return graphql.String

	case reflect.Pointer:
		// TODO: Optional of some kind? idk
		return s.introspectField(r.Elem())
	case reflect.Array:
	case reflect.Slice:
		return graphql.NewList(s.introspectField(r.Elem()))

	case reflect.Struct:
		fields := s.introspect(r)
		return graphql.NewObject(graphql.ObjectConfig{
			Name:   r.Name(),
			Fields: fields,
		})

	case reflect.Complex64:
	case reflect.Complex128:
	case reflect.Chan:
	case reflect.Func:
	case reflect.Interface:
	case reflect.Map:
	case reflect.UnsafePointer:
		panic(fmt.Sprintf("unsupported type \"%s\"", r.String()))
	}

	panic("unreachable")
}

func (s *Builder) introspect(of reflect.Type) graphql.Fields {
	fields := graphql.Fields{}

	if of.Kind() == reflect.Struct {
		for i := 0; i < of.NumField(); i++ {
			r := of.Field(i)

			tags := strings.Split(r.Tag.Get("gql"), ",")

			name := r.Name
			if tags[0] != "" {
				name = tags[0]
			}

			field := graphql.Field{
				Name: name,
				Type: s.introspectField(r.Type),
			}

			fields[name] = &field
		}
	} else {
		panic("input must be a struct")
	}

	return fields
}

func (s *Builder) introspectObject(t reflect.Type) *graphql.Object {
	object := graphql.NewObject(graphql.ObjectConfig{
		Name:   t.Name(),
		Fields: s.introspect(t),
	})

	return object
}

func (s *Builder) introspectInput(t reflect.Type) graphql.FieldConfigArgument {
	config := graphql.FieldConfigArgument{}
	fields := s.introspect(t)

	for key, value := range fields {
		config[key] = &graphql.ArgumentConfig{Type: value.Type}
	}

	return config
}

func (s *Builder) lookupObject(t reflect.Type) *graphql.Object {
	object, ok := s.objectCache[t]
	if !ok {
		object = s.introspectObject(t)
		s.objectCache[t] = object
	}

	return object
}

func (s *Builder) lookupArguments(t reflect.Type) graphql.FieldConfigArgument {
	input, ok := s.inputCache[t]
	if !ok {
		input = s.introspectInput(t)
		s.inputCache[t] = input
	}

	return input
}

func (s *Builder) Build() (graphql.Schema, error) {
	rootMutationFields := graphql.Fields{}
	rootQueryFields := graphql.Fields{}

	for _, resolver := range s.Resolvers {
		for _, query := range resolver.Queries {
			field := graphql.Field{
				Name: query.Name,

				Type: s.lookupObject(query.Type),
				Args: s.lookupArguments(query.Arguments),

				Resolve: query.Resolve,

				Description:       query.Name,
				DeprecationReason: query.DeprecationReason,
			}

			rootQueryFields[query.Name] = &field
		}
	}

	if 0 >= len(rootMutationFields) {
		return graphql.NewSchema(graphql.SchemaConfig{
			Query: graphql.NewObject(graphql.ObjectConfig{
				Name:   "RootQuery",
				Fields: rootQueryFields,
			}),
		})
	}

	return graphql.NewSchema(graphql.SchemaConfig{
		Mutation: graphql.NewObject(graphql.ObjectConfig{
			Name:   "RootMutation",
			Fields: rootMutationFields,
		}),
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name:   "RootQuery",
			Fields: rootQueryFields,
		}),
	})

}
