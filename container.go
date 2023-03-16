package container

import (
	"errors"
	"fmt"
	"github.com/gorilla-go/container/definition"
	"github.com/gorilla-go/container/definition/definition_type"
	"reflect"
	"strconv"
	"sync"
	"unsafe"
)

var scopeSort = map[BeanScope]int{
	SINGLETON: 0,
	REQUEST:   1,
}

var containerOnce sync.Once
var containerInstance *Container
var syncLock sync.Mutex

type Container struct {
	singletonPool map[string]*Bean
	requestPool   map[string]map[string]*Bean

	definitions        map[string]*definition.ObjectDefinition
	definitionsSort    map[string]int
	objectDefinitions  map[string]*definition.ObjectDefinition
	requestDefinitions map[string]*definition.ObjectDefinition
}

func ContainerInstance() *Container {
	containerOnce.Do(func() {
		containerInstance = &Container{
			singletonPool:      make(map[string]*Bean),
			requestPool:        make(map[string]map[string]*Bean),
			definitions:        make(map[string]*definition.ObjectDefinition),
			definitionsSort:    make(map[string]int),
			objectDefinitions:  make(map[string]*definition.ObjectDefinition),
			requestDefinitions: make(map[string]*definition.ObjectDefinition),
		}
	})

	return containerInstance
}

func (c *Container) Init() {
	c.initializeBeans(c.definitions)
}

func (c *Container) initializeBeans(definitions map[string]*definition.ObjectDefinition) {
	var f []string
	var sort []int
	for name, item := range definitions {
		if item.Type == definition_type.Value {
			continue
		}
		f = append(f, name)
		sort = append(sort, c.definitionsSort[name])
	}
	f = Array.Sort(f, sort, Array.DESC)

	for i := 0; i < len(f); i++ {
		beanName := f[i]
		definitionVar := definitions[beanName]
		if definitionVar.GetScope() == REQUEST {
			c.requestDefinitions[beanName] = definitionVar
			delete(c.definitions, beanName)
			continue
		} else {
			c.objectDefinitions[beanName] = definitionVar
			delete(c.definitions, beanName)
		}

		c.newBean(beanName, "")
	}
}

func (c *Container) GetStats() map[string]int {
	return map[string]int{
		"singleton":           len(c.singletonPool),
		"request":             len(c.requestPool),
		"definition":          len(c.definitions),
		"objectionDefinition": len(c.objectDefinitions),
		"requestDefinition":   len(c.requestDefinitions),
	}
}

func (c *Container) ToString() string {
	m := c.GetKeyStats()
	var str = ""
	var seq = len(m)
	var ii = 0
	for s, i := range m {
		str += s + ": " + strconv.Itoa(len(i))
		if ii+1 < seq {
			str += ", "
		}
		ii++
	}
	if str == "" {
		return "nil"
	}
	return str
}

func (c *Container) GetKeyStats() map[string][]string {
	return map[string][]string{
		"singleton": Map.Keys(c.singletonPool),
		"request":   Map.Keys(c.requestPool),
	}
}

func (c *Container) GetObjectDefinitions() map[string]*definition.ObjectDefinition {
	return c.objectDefinitions
}

func (c *Container) GetRequestDefinitions() map[string]*definition.ObjectDefinition {
	return c.requestDefinitions
}

func (c *Container) GetRequestPool() map[string]map[string]*Bean {
	return c.requestPool
}

func (c *Container) GetRequest(beanName string, requestUid string) *Bean {
	if v, ok := c.requestPool[requestUid][beanName]; ok {
		return v
	}

	if _, ok := c.requestDefinitions[beanName]; !ok {
		panic("Request bean(" + beanName + ") is not defined. please make sure scope is request.")
	}

	return c.newBean(beanName, requestUid)
}

func (c *Container) Get(beanName string) *Bean {
	if v, ok := c.singletonPool[beanName]; ok {
		return v
	}

	if _, ok := c.objectDefinitions[beanName]; !ok {
		panic(errors.New("The bean of " + beanName + " is not defined. please make sure scope is singleton."))
	}

	return c.newBean(c.objectDefinitions[beanName].GetName(), "")
}

func (c *Container) DestroyRequest(cid string) {
	beanGroup := c.requestPool[cid]
	for _, bean := range beanGroup {
		if bean.HasDestroyMethod() {
			bean.GetDestroyMethod().Call(nil)
		}
	}
	delete(c.requestPool, cid)
}

func (c *Container) AddDefinition(od *definition.ObjectDefinition) {
	c.definitions[od.GetName()] = od
	if _, ok := c.definitionsSort[od.GetName()]; !ok {
		c.definitionsSort[od.GetName()] = 0
	}

	propertyInjections := od.GetPropertyInjections()
	if propertyInjections != nil && len(propertyInjections) > 0 {
		for _, definition := range propertyInjections {
			if scopeSort[od.GetScope()] < scopeSort[definition.GetScope()] {
				panic(
					errors.New(
						fmt.Sprintf(
							"%s bean can not include %s property at %s",
							od.GetScope().ToString(),
							definition.GetScope().ToString(),
							od.GetName(),
						),
					),
				)
			}
			c.definitionsSort[definition.GetName()] = c.definitionsSort[od.GetName()] + 1
		}
		c.AddDefinitions(Map.Values(propertyInjections))
	}
}

func (c *Container) AddDefinitions(ods []*definition.ObjectDefinition) {
	for _, od := range ods {
		c.AddDefinition(od)
	}
}

func (c *Container) newBean(beanName string, cid string) *Bean {
	if bean, ok := c.singletonPool[beanName]; ok {
		return bean
	}

	definitionVar := c.getNewObjectDefinition(beanName)
	scope := definitionVar.GetScope()

	// hook...

	prototype := definitionVar.GetPrototype()

	reflectObject := c.newInstance(prototype)
	propertyInjections := definitionVar.GetPropertyInjections()

	if len(propertyInjections) != 0 {
		reflectObject = c.newProperty(reflectObject, propertyInjections, cid)
	}

	bean := NewBean(reflectObject.Interface())

	// bean init.
	if bean.HasInitMethod() {
		bean.GetInitMethod().Call(nil)
	}

	return c.setNewBean(beanName, scope, bean, cid)
}

func (c *Container) newInstance(bean *Bean) reflect.Value {
	refValue := reflect.ValueOf(bean.GetValue())
	if refValue.Kind() == reflect.Ptr {
		return reflect.New(refValue.Elem().Type())
	}
	return reflect.New(refValue.Type()).Elem()
}

func (c *Container) newProperty(v reflect.Value, m map[string]*definition.ObjectDefinition, cid string) reflect.Value {
	cp := v
	vk := v.Kind()
	if vk == reflect.Ptr {
		v = v.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		currentPropertyName := v.Type().Field(i).Name

		var injectionObj reflect.Value
		if propertyValue, ok := m[currentPropertyName]; ok {
			needField := v.Field(i)
			needKind := needField.Kind()

			if propertyValue.Type == definition_type.Const {
				injectionObj = reflect.ValueOf(propertyValue.GetConstValue())
			} else {

				annotationName := propertyValue.GetName()
				scope := propertyValue.GetScope()

				var bean *Bean
				switch scope {
				case REQUEST:
					bean = c.GetRequest(annotationName, cid)
					break
				default:
					bean = c.Get(annotationName)
					break
				}

				injectionObj = reflect.ValueOf(bean.GetValue())
			}

			givenKind := injectionObj.Kind()
			if needKind != givenKind && needKind != reflect.Interface {
				panic(
					errors.New(
						"property " + currentPropertyName + " is " + needKind.String() + " but " +
							givenKind.String() + " given at " + v.Type().String(),
					),
				)
			}

			if v.CanAddr() {
				method := v.Addr().MethodByName("Set" + Strings.UcFirst(currentPropertyName))
				if method.IsValid() {
					method.Call([]reflect.Value{injectionObj})
					continue
				}
			}

			currentPropertyUnsafe := reflect.NewAt(needField.Type(), unsafe.Pointer(needField.UnsafeAddr())).Elem()
			currentPropertyUnsafe.Set(injectionObj)
		}
	}

	if vk == reflect.Ptr {
		return cp
	}
	return v
}

func (c *Container) doBeanInit(v reflect.Value) {

}

func (c *Container) getNewObjectDefinition(name string) *definition.ObjectDefinition {
	if objectDefinition, ok := c.objectDefinitions[name]; ok {
		return objectDefinition
	}

	if objectDefinition, ok := c.requestDefinitions[name]; ok {
		return objectDefinition
	}

	panic(errors.New("Bean name of " + name + " is not defined."))
}

func (c *Container) setNewBean(beanName string, scope BeanScope, object *Bean, cid string) *Bean {
	switch scope {
	case SINGLETON:
		c.singletonPool[beanName] = object
		break
	case REQUEST:
		syncLock.Lock()
		defer syncLock.Unlock()
		if _, ok := c.requestPool[cid]; !ok {
			c.requestPool[cid] = make(map[string]*Bean)
		}
		c.requestPool[cid][beanName] = object
		break
	}
	return object
}
