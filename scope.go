package container

import "errors"

type BeanScope int

const SINGLETON BeanScope = 0
const REQUEST BeanScope = 1

func (s BeanScope) ToString() string {
	switch s {
	case SINGLETON:
		return "Singleton"
	case REQUEST:
		return "Request"
	}
	panic(errors.New("invalid bean scope"))
}
