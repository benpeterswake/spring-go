package spring

import (
	"log"
	"reflect"
)

var services map[string]reflect.Value = make(map[string]reflect.Value)

func AddService(service interface{}) {
	serviceValue := reflect.ValueOf(service)
	if serviceValue.Kind() != reflect.Ptr {
		log.Fatal("Service is not pointer")
	}
	serviceValue = serviceValue.Elem()
	services[serviceValue.Type().Name()] = serviceValue
	for i := 0; i < serviceValue.NumField(); i++ {
		repo, ok := repositories[serviceValue.Field(i).Type().Name()]
		if ok {
			serviceValue.Field(i).Set(repo)
		}
	}
}
