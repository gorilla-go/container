### Container Service
---
### Usage

##### Register Singleton
```go
type Register struct {}

func main() {
    container := NewContainer()
	SetSingleton(container, &Register{})
	_ = GetSingleton[*Register](container)
}
```

##### Register Singleton with Alias
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