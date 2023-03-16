package definition

import "github.com/gorilla-go/container/definition/definition_type"

func Value(v any) *ObjectDefinition {
	return &ObjectDefinition{
		Type:  definition_type.Const,
		Value: v,
	}
}
