### Container Service
---
### Usage

##### Register Singleton And Fetch
```go
type Register struct {}

func main() {
    container := NewContainer()
	SetSingleton(container, &Register{})
	_ = GetSingleton[*Register](container)
}
```

##### Register Singleton With Alias And Fetch By Alias Or Generic Type
```go
type Register struct {}

func main() {
    container := NewContainer()
	SetSingletonWithAlias(container, "register", &Register{})
    // get singleton by generic type
	_ = GetSingleton[*Register](container)

    // get by alias
    _ = GetSingletonByAlias[*Register](container, "register")
}
```

##### Register Lazy Singleton And Fetch
> Lazy singleton will be built when first fetch it. and keep singleton on next.
```go
type Register struct {}

func main() {
    container := NewContainer()
	SetLazySingleton(container, func () *Register {
        return &Register{}
    })
    
    // will panic when this action called.
	_ = GetSingleton[*Register](container)

    // will return Register Struct Successfully.
    _ = GetLazySingleton[*Register](container)

    // this also return register struct, because register had been built at before.
    _ = GetSingleton[*Register](container)
}
```