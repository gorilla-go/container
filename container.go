package container

import (
	"fmt"
	"reflect"
	"sync"
)

var (
	// 类型缓存
	typeCache      map[reflect.Type]string
	typeCacheMutex sync.RWMutex
)

func init() {
	typeCache = make(map[reflect.Type]string)
	typeCacheMutex = sync.RWMutex{}
}

type Container struct {
	// 单例注册
	singletons        map[string]any
	singletonsByAlias map[string]string

	// 懒汉式单例
	lazySingletons         map[string]any
	lazySingletonsByAlias  map[string]string
	lazySingletonsMutex    sync.RWMutex
	lazySingletonsMutexMap map[string]*sync.RWMutex

	// 对象提供者
	providers        map[string]any
	providersByAlias map[string]string

	// 实现绑定
	implements map[string]any
}

func NewContainer() *Container {
	return &Container{
		singletons:             make(map[string]any),
		singletonsByAlias:      make(map[string]string),
		providers:              make(map[string]any),
		providersByAlias:       make(map[string]string),
		lazySingletons:         make(map[string]any),
		lazySingletonsByAlias:  make(map[string]string),
		lazySingletonsMutex:    sync.RWMutex{},
		lazySingletonsMutexMap: make(map[string]*sync.RWMutex),
		implements:             make(map[string]any),
	}
}

// 获取泛型类型
func fetchGenericType[T any]() string {
	t := reflect.TypeOf((*T)(nil)).Elem()

	typeCacheMutex.RLock()
	if val, ok := typeCache[t]; ok {
		typeCacheMutex.RUnlock()
		return val
	}
	typeCacheMutex.RUnlock()

	typeCacheMutex.Lock()
	defer typeCacheMutex.Unlock()

	if _, ok := typeCache[t]; !ok {
		typeName := ""
		switch t.Kind() {
		case reflect.Ptr:
			elem := t.Elem()
			if elem.PkgPath() != "" {
				typeName = elem.PkgPath() + ".*" + elem.Name()
			} else {
				typeName = "builtin.*" + elem.Name()
			}
		default:
			if t.PkgPath() != "" {
				typeName = t.PkgPath() + "." + t.Name()
			} else {
				typeName = "builtin." + t.Name()
			}
		}
		typeCache[t] = typeName
	}
	return typeCache[t]
}

// 设置单例对象
func SetSingleton[T any](container *Container, class T) {
	container.singletons[fetchGenericType[T]()] = class
}

// 获取单例对象
func GetSingleton[T any](container *Container) (T, error) {
	genericName := fetchGenericType[T]()
	if t, ok := container.singletons[genericName]; ok {
		return t.(T), nil
	}
	var zero T
	return zero, fmt.Errorf("singleton %s not found", genericName)
}

// 获取单例对象，如果对象不存在则panic
func GetMustSingleton[T any](container *Container) T {
	t, err := GetSingleton[T](container)
	if err != nil {
		panic(err)
	}
	return t
}

// 设置单例对象并添加别名
func SetSingletonWithAlias[T any](container *Container, name string, class T) {
	SetSingleton(container, class)
	container.singletonsByAlias[name] = fetchGenericType[T]()
}

// 通过别名获取单例对象
func GetSingletonByAlias[T any](container *Container, name string) (T, error) {
	if t, ok := container.singletons[container.singletonsByAlias[name]]; ok {
		return t.(T), nil
	}
	var zero T
	return zero, fmt.Errorf("singleton alias %s not found", name)
}

// 通过别名获取单例对象，如果对象不存在则panic
func GetMustSingletonByAlias[T any](container *Container, name string) T {
	t, err := GetSingletonByAlias[T](container, name)
	if err != nil {
		panic(err)
	}
	return t
}

// 设置懒汉式单例
func SetLazySingleton[T any](container *Container, provideFunc func() T) {
	container.lazySingletons[fetchGenericType[T]()] = provideFunc
}

// 设置懒汉式单例并添加别名
func SetLazySingletonWithAlias[T any](container *Container, name string, provideFunc func() T) {
	genericName := fetchGenericType[T]()
	container.lazySingletons[genericName] = provideFunc
	container.lazySingletonsByAlias[name] = fetchGenericType[T]()
}

// 通过泛型名称获取懒汉式单例
func getLazySingletonByGenericName[T any](container *Container, genericName string) (T, error) {
	// 检查单例库中是否存在该对象
	if t, ok := container.singletons[genericName]; ok {
		return t.(T), nil
	}

	// 不存在则创建
	container.lazySingletonsMutex.Lock()
	if _, ok := container.lazySingletonsMutexMap[genericName]; !ok {
		container.lazySingletonsMutexMap[genericName] = &sync.RWMutex{}
	}
	container.lazySingletonsMutex.Unlock()
	container.lazySingletonsMutexMap[genericName].Lock()
	defer container.lazySingletonsMutexMap[genericName].Unlock()

	// 再次检查单例库中是否存在该对象
	if t, ok := container.singletons[genericName]; ok {
		return t.(T), nil
	}

	// 检查懒汉式单例库中是否存在该对象
	if _, ok := container.lazySingletons[genericName]; !ok {
		var zero T
		return zero, fmt.Errorf("lazy singleton %s not found", genericName)
	}
	// 设置单例
	container.singletons[genericName] = container.lazySingletons[genericName].(func() T)()
	return container.singletons[genericName].(T), nil
}

// 获取懒汉式单例
func GetLazySingleton[T any](container *Container) (T, error) {
	return getLazySingletonByGenericName[T](container, fetchGenericType[T]())
}

// 获取懒汉式单例，如果对象不存在则panic
func GetMustLazySingleton[T any](container *Container) T {
	t, err := GetLazySingleton[T](container)
	if err != nil {
		panic(err)
	}
	return t
}

// 通过别名获取懒汉式单例
func GetLazySingletonByAlias[T any](container *Container, name string) (T, error) {
	genericName, ok := container.lazySingletonsByAlias[name]
	if !ok {
		var zero T
		return zero, fmt.Errorf("lazy singleton alias %s not found", genericName)
	}
	return getLazySingletonByGenericName[T](container, genericName)
}

// 通过别名获取懒汉式单例，如果对象不存在则panic
func GetMustLazySingletonByAlias[T any](container *Container, name string) T {
	t, err := GetLazySingletonByAlias[T](container, name)
	if err != nil {
		panic(err)
	}
	return t
}

// 设置对象供应者
func SetProvider[T any](container *Container, provideFunc func() T) {
	container.providers[fetchGenericType[T]()] = provideFunc
}

// 设置提供者对象并添加别名
func SetProviderWithAlias[T any](container *Container, name string, provideFunc func() T) {
	SetProvider(container, provideFunc)
	container.providersByAlias[name] = fetchGenericType[T]()
}

// 通过泛型名称获取提供者对象
func getProviderByGenericName[T any](container *Container, genericName string) (T, error) {
	if t, ok := container.providers[genericName]; ok {
		return t.(func() T)(), nil
	}
	var zero T
	return zero, fmt.Errorf("provider %s not found", genericName)
}

// 获取提供者对象
func GetProvider[T any](container *Container) (T, error) {
	return getProviderByGenericName[T](container, fetchGenericType[T]())
}

// 获取提供者对象，如果对象不存在则panic
func GetMustProvider[T any](container *Container) T {
	t, err := GetProvider[T](container)
	if err != nil {
		panic(err)
	}
	return t
}

// 通过别名获取提供者对象
func GetProviderByAlias[T any](container *Container, name string) (T, error) {
	genericName, ok := container.providersByAlias[name]
	if !ok {
		var zero T
		return zero, fmt.Errorf("provider alias %s not found", name)
	}
	return getProviderByGenericName[T](container, genericName)
}

// 通过别名获取提供者对象，如果对象不存在则panic
func GetMustProviderByAlias[T any](container *Container, name string) T {
	t, err := GetProviderByAlias[T](container, name)
	if err != nil {
		panic(err)
	}
	return t
}

// 绑定实现
func BindImplement[T any](container *Container, implement any) {
	var _ T = implement.(T)
	container.implements[fetchGenericType[T]()] = implement
}

// 获取实现
func GetImplement[T any](container *Container) (T, error) {
	if t, ok := container.implements[fetchGenericType[T]()]; ok {
		return t.(T), nil
	}
	var zero T
	return zero, fmt.Errorf("implement %s not found", fetchGenericType[T]())
}

// 获取实现，如果实现不存在则panic
func GetMustImplement[T any](container *Container) T {
	t, err := GetImplement[T](container)
	if err != nil {
		panic(err)
	}
	return t
}
