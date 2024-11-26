package container

import (
	"fmt"
	"testing"
)

type Student struct {
	Name string
	Age  int
}

func (s *Student) GetAge() int {
	return s.Age
}

type IStudent interface {
	GetAge() int
}

type IStudent2 interface {
	GetName() string
}

func Test_getObjectName(t *testing.T) {
	container := NewContainer()
	BindImplement[IStudent](container, &Student{
		Name: "name",
		Age:  18,
	})

	student := GetMustImplement[IStudent](container)
	fmt.Println(student.GetAge())
}
