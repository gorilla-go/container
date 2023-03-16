package container

import "errors"

type BeanScope int

const SINGLETON BeanScope = 0
const REQUEST BeanScope = 1

func (s BeanScope) ToString() string {
	switch s {
	case 0:
		return "Singleton"
	case 1:
		return "Request"
	}
	panic(errors.New("invalid bean scope"))
}
