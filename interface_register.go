package container

import (
	"errors"
	"reflect"
	"sync"
)

var interfaceRegisterSync sync.Once
var interfaceRegisterInstance *Register

type Register struct {
	interfacesMap map[string]string
}

func GetInterfaceRegisterInstance() *Register {
	interfaceRegisterSync.Do(func() {
		interfaceRegisterInstance = &Register{
			interfacesMap: make(map[string]string),
		}
	})
	return interfaceRegisterInstance
}

func (r *Register) Register(i interface{}, beanName string) {
	ref := reflect.ValueOf(i)
	if ref.Kind() != reflect.Ptr && !r.isInterface(ref.Elem().Interface()) {
		panic(errors.New("invalid interface type"))
	}

	r.interfacesMap[ref.Type().String()] = beanName
}

func (r *Register) GetInterfaceInjectBeanName(i interface{}) string {
	ref := reflect.ValueOf(i)
	if ref.Kind() != reflect.Ptr && !r.isInterface(ref.Elem().Interface()) {
		panic(errors.New("invalid interface type"))
	}
	if v, ok := r.interfacesMap[ref.Type().String()]; ok {
		return v
	}
	panic("interface not found.")
}

func (r *Register) isInterface(i interface{}) bool {
	if reflect.TypeOf(i).Kind() != reflect.Interface {
		return false
	}
	return true
}
