package container

import (
	"reflect"
)

var BEAN_INIT_METHOD = "Init"
var BEAN_DESTROY_METHOD = "Destroy"

type Bean struct {
	value     interface{}
	ref       reflect.Value
	methodMap map[string]BeanMethod
}

type BeanMethod struct {
	ref  reflect.Value
	args []reflect.Type
}

func NewBean(i interface{}) *Bean {
	ref := reflect.ValueOf(i)
	methodMap := make(map[string]BeanMethod)
	f := func(r reflect.Type, isPtr bool) {
		for i := 0; i < r.NumMethod(); i++ {
			methodType := r.Method(i)
			name := methodType.Name

			var args []reflect.Type

			var v reflect.Value
			if ref.CanAddr() {
				m := ref.Addr().MethodByName(name)
				if m.IsValid() {
					v = m
					for i := 0; i < m.Type().NumIn(); i++ {
						args = append(args, m.Type().In(i))
					}
				}
			} else {
				m := ref.MethodByName(name)
				if m.IsValid() {
					v = m
					for i := 0; i < m.Type().NumIn(); i++ {
						args = append(args, m.Type().In(i))
					}
				}
			}

			methodMap[name] = BeanMethod{
				ref:  v,
				args: args,
			}
		}
	}

	if ref.Kind() == reflect.Struct || (ref.Kind() == reflect.Ptr && ref.Elem().Kind() == reflect.Struct) {
		f(ref.Type(), false)
		f(ref.Elem().Type(), true)
	}

	return &Bean{
		value:     i,
		ref:       ref,
		methodMap: methodMap,
	}
}

func (b *Bean) GetValue() interface{} {
	return b.value
}

func (b *Bean) GetReflection() reflect.Value {
	return b.ref
}

func (b *Bean) GetActualType() reflect.Kind {
	v := b.ref
	if v.Kind() == reflect.Ptr {
		return v.Elem().Kind()
	}
	return v.Kind()
}

func (b *Bean) HasInitMethod() bool {
	return b.HasMethod(BEAN_INIT_METHOD)
}

func (b *Bean) GetInitMethod() reflect.Value {
	return b.GetMethod(BEAN_INIT_METHOD)
}

func (b *Bean) HasDestroyMethod() bool {
	return b.HasMethod(BEAN_DESTROY_METHOD)
}

func (b *Bean) GetDestroyMethod() reflect.Value {
	return b.GetMethod(BEAN_DESTROY_METHOD)
}

func (b *Bean) HasMethod(name string) bool {
	if _, ok := b.methodMap[name]; ok {
		return true
	}
	return false
}

func (b *Bean) GetMethod(name string) reflect.Value {
	return b.methodMap[name].ref
}

func (b *Bean) GetMethodArgs(name string) []reflect.Type {
	return b.methodMap[name].args
}
