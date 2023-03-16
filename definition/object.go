package definition

import (
	"github.com/gorilla-go/container"
	"github.com/gorilla-go/container/definition/definition_type"
	"reflect"
)

type ObjectDefinition struct {
	name               string
	scope              container.BeanScope
	propertyInjections map[string]*ObjectDefinition
	prototype          *container.Bean
	Type               definition_type.Type
	Value              any
}

func Object(name string, prototype interface{}, injects map[string]*ObjectDefinition) *ObjectDefinition {
	obj := &ObjectDefinition{
		name:               name,
		scope:              container.REQUEST,
		propertyInjections: nil,
		prototype:          container.NewBean(prototype),
		Type:               definition_type.Object,
		Value:              nil,
	}
	if injects != nil && len(injects) != 0 {
		obj.SetPropertyInjections(injects)
	}
	return obj
}

func Singleton(name string, prototype interface{}, injects map[string]*ObjectDefinition) *ObjectDefinition {
	obj := &ObjectDefinition{
		name:               name,
		scope:              container.SINGLETON,
		propertyInjections: nil,
		prototype:          container.NewBean(prototype),
		Type:               definition_type.Object,
		Value:              nil,
	}
	if injects != nil && len(injects) != 0 {
		obj.SetPropertyInjections(injects)
	}
	return obj
}

func (o *ObjectDefinition) GetName() string {
	return o.name
}

func (o *ObjectDefinition) SetName(name string) {
	o.name = name
}

func (o *ObjectDefinition) GetPrototype() *container.Bean {
	return o.prototype
}

func (o *ObjectDefinition) SetPrototype(bean *container.Bean) {
	if bean.GetActualType() != reflect.Struct {
		panic("Invalid struct type.")
	}
	o.prototype = bean
}

func (o *ObjectDefinition) GetScope() container.BeanScope {
	return o.scope
}

func (o *ObjectDefinition) GetType() definition_type.Type {
	return o.Type
}

func (o *ObjectDefinition) GetConstValue() any {
	return o.Value
}

func (o *ObjectDefinition) SetScope(scope container.BeanScope) {
	o.scope = scope
}

func (o *ObjectDefinition) GetPropertyInjections() map[string]*ObjectDefinition {
	return o.propertyInjections
}

func (o *ObjectDefinition) SetPropertyInjections(s map[string]*ObjectDefinition) {
	if o.propertyInjections == nil {
		o.propertyInjections = make(map[string]*ObjectDefinition)
	}
	o.propertyInjections = s
}

func (o *ObjectDefinition) SetPropertyInjection(propertyName string, pi *ObjectDefinition) {
	if o.propertyInjections == nil {
		o.propertyInjections = make(map[string]*ObjectDefinition)
	}
	o.propertyInjections[propertyName] = pi
}
