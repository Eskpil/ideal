package ideal

import (
	"fmt"
	"github.com/graphql-go/graphql"
	"reflect"
	"strings"
)

type Field struct {
	Name string

	Middleware []MiddlewareFunc

	Arguments reflect.Type
	Type      reflect.Type
	Resolve   ResolveFunc

	Description       string
	DeprecationReason string
}

type Query struct {
	Name string

	Middleware []MiddlewareFunc

	Arguments reflect.Type
	Type      reflect.Type
	Resolve   ResolveFunc

	Description       string
	DeprecationReason string
}

type Mutation struct {
	Name string

	Middleware []MiddlewareFunc

	Arguments reflect.Type
	Type      reflect.Type
	Resolve   ResolveFunc

	Description       string
	DeprecationReason string
}

type Resolver struct {
	Type reflect.Type

	Queries   []Query
	Mutations []Mutation
	Fields    []Field
}

type Builder struct {
	Resolvers []Resolver

	objectCache map[reflect.Type]*graphql.Object
	inputCache  map[reflect.Type]graphql.FieldConfigArgument

	rootQuery        *graphql.Object
	rootMutation     *graphql.Object
	rootSubscription *graphql.Object
}

func (r *Resolver) Query(query Query) {
	r.Queries = append(r.Queries, query)
}

func NewBuilder(resolvers ...Resolver) *Builder {
	builder := new(Builder)

	builder.Resolvers = resolvers

	builder.objectCache = make(map[reflect.Type]*graphql.Object)
	builder.inputCache = make(map[reflect.Type]graphql.FieldConfigArgument)

	return builder
}

func (b *Builder) AddResolver(resolver Resolver) {
	b.Resolvers = append(b.Resolvers, resolver)
}

func (b *Builder) introspectField(r reflect.Type) graphql.Type {
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
		return b.introspectField(r.Elem())
	case reflect.Array:
	case reflect.Slice:
		return graphql.NewList(b.introspectField(r.Elem()))

	case reflect.Struct:
		if r.String() == "time.Time" {
			return graphql.DateTime
		}
		return b.introspect(r).(*graphql.Object)
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

func (b *Builder) introspect(of reflect.Type) interface{} {
	if of.Kind() == reflect.Struct {
		if object, ok := b.objectCache[of]; ok {
			return object
		}

		fields := graphql.Fields{}

		for i := 0; i < of.NumField(); i++ {
			r := of.Field(i)

			tags := strings.Split(r.Tag.Get("gql"), ",")

			if tags[0] == "-" {
				continue
			}

			name := r.Name
			if tags[0] != "" {
				name = tags[0]
			}

			field := graphql.Field{
				Name: name,
				Type: b.introspectField(r.Type),
			}

			fields[name] = &field
		}

		object := graphql.NewObject(graphql.ObjectConfig{
			Name:   of.Name(),
			Fields: fields,
		})

		b.objectCache[of] = object

		return object
	} else if of.Kind() == reflect.Array {
		return graphql.NewList(b.introspect(of.Elem()).(*graphql.Object))
	} else if of.Kind() == reflect.Slice {
		return graphql.NewList(b.introspect(of.Elem()).(*graphql.Object))
	} else {
		panic("input must be a struct, array or slice")
	}

	return nil
}

func (b *Builder) introspectInput(t reflect.Type) graphql.FieldConfigArgument {
	config := graphql.FieldConfigArgument{}
	obj := b.introspect(t).(*graphql.Object)

	for key, value := range obj.Fields() {
		config[key] = &graphql.ArgumentConfig{Type: value.Type}
	}

	return config
}

func (b *Builder) lookupArguments(t reflect.Type) graphql.FieldConfigArgument {
	input, ok := b.inputCache[t]
	if !ok {
		input = b.introspectInput(t)
		b.inputCache[t] = input
	}

	return input
}

func (b *Builder) Build(runtime *Runtime) (graphql.Schema, error) {
	rootMutationFields := graphql.Fields{}
	rootQueryFields := graphql.Fields{}

	for _, resolver := range b.Resolvers {
		if resolver.Type != nil {
			object := b.introspect(resolver.Type).(*graphql.Object)

			for _, fieldResolver := range resolver.Fields {
				args := graphql.FieldConfigArgument{}
				if fieldResolver.Arguments != nil {
					args = b.lookupArguments(fieldResolver.Arguments)
				}

				runtime.AddHandler(fieldResolver.Name, fieldResolver.Resolve, fieldResolver.Middleware...)

				field := graphql.Field{
					Name: fieldResolver.Name,

					Type: b.introspect(fieldResolver.Type).(graphql.Output),
					Args: args,

					Resolve: runtime.Resolve,

					Description:       fieldResolver.Name,
					DeprecationReason: fieldResolver.DeprecationReason,
				}

				object.AddFieldConfig(fieldResolver.Name, &field)
			}
		}

		for _, query := range resolver.Queries {
			if query.Type == nil {
				panic("query type must not be nil")
			}

			args := graphql.FieldConfigArgument{}
			if query.Arguments != nil {
				args = b.lookupArguments(query.Arguments)
			}

			runtime.AddHandler(query.Name, query.Resolve, query.Middleware...)

			field := graphql.Field{
				Name: query.Name,

				Type: b.introspect(query.Type).(graphql.Output),
				Args: args,

				Resolve: runtime.Resolve,

				Description:       query.Name,
				DeprecationReason: query.DeprecationReason,
			}

			rootQueryFields[query.Name] = &field
		}

		for _, mutation := range resolver.Mutations {
			if mutation.Type == nil {
				panic("mutation type must not be nil")
			}

			args := graphql.FieldConfigArgument{}
			if mutation.Arguments != nil {
				args = b.lookupArguments(mutation.Arguments)
			}

			runtime.AddHandler(mutation.Name, mutation.Resolve, mutation.Middleware...)

			field := graphql.Field{
				Name: mutation.Name,

				Type: b.introspect(mutation.Type).(*graphql.Object),
				Args: args,

				Resolve: runtime.Resolve,

				Description:       mutation.Name,
				DeprecationReason: mutation.DeprecationReason,
			}

			rootMutationFields[mutation.Name] = &field
		}
	}

	if 0 >= len(rootMutationFields) {
		schema, err := graphql.NewSchema(graphql.SchemaConfig{
			Query: graphql.NewObject(graphql.ObjectConfig{
				Name:   "RootQuery",
				Fields: rootQueryFields,
			}),
		})

		return schema, err
	}

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Mutation: graphql.NewObject(graphql.ObjectConfig{
			Name:   "RootMutation",
			Fields: rootMutationFields,
		}),
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name:   "RootQuery",
			Fields: rootQueryFields,
		}),
	})

	return schema, err

}
