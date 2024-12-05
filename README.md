### Container Service
---
### Usage

##### Register Singleton And Fetch
```go
type Register struct {}

func main() {
    container := NewContainer()

    // set singleton.
	SetSingleton(container, &Register{})
    // get singleton by generic
	_ = GetSingleton[*Register](container)

    // set singleton by alias
    SetSingletonWithAlias(container, "action", &Action{})
    // get singleton by generic type
	_ = GetSingleton[*Action](container)
    // get by alias, and the execution efficiency is higher than the former.
    _ = GetSingletonByAlias[*Action](container, "action")
}
```

##### Register Lazy Singleton And Fetch
> Lazy singleton will be built when fetch it first. and keep singleton on next.
```go
type Register struct {}

func main() {
    container := NewContainer()

    // set lazy singleton. this func will only save, and not exec.
	SetLazySingleton(container, func () *Register {
        return &Register{}
    })
    
    // will panic when this action called.
	_ = GetSingleton[*Register](container)

    // will return register struct successfully. on this time, register action func done once.
    _ = GetLazySingleton[*Register](container)

    // this also return register struct, because register had been built at before and keep singleton.
    _ = GetSingleton[*Register](container)

    
    // set lazy singleton by alias
    SetLazySingletonWithAlias(container, "action", func () *Action {
        return &Action{}
    })

    // fetch by alias
    _ = GetLazySingletonByAlias[*Action](container, "action")
}
```