package definition

import (
	"github.com/gorilla-go/container"
	"github.com/gorilla-go/container/definition/definition_type"
)

func Controller(name string, prototype interface{}, injects map[string]*ObjectDefinition) *ObjectDefinition {
	obj := &ObjectDefinition{
		name:               name,
		scope:              container.SINGLETON,
		propertyInjections: nil,
		prototype:          container.NewBean(prototype),
		Type:               definition_type.Controller,
		Value:              nil,
	}
	if injects != nil && len(injects) != 0 {
		obj.SetPropertyInjections(injects)
	}
	return obj
}
