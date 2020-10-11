package spring

import (
	"reflect"
)

var services map[reflect.Type]reflect.Value = make(map[reflect.Type]reflect.Value)

func AddService(service interface{}) {
	serviceValue := reflect.New(reflect.TypeOf(service))
	services[serviceValue.Type()] = serviceValue
}
