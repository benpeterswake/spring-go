package spring

import (
	"encoding/json"
	"log"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
	"gitlab.com/BPeters58/go-spring-app/model"
)

func AddController(controller interface{}) {
	typeOf := reflect.TypeOf(controller)
	valueOf := reflect.ValueOf(controller)
	for i := 0; i < typeOf.NumMethod(); i++ {
		//
		validMethod := true
		method := typeOf.Method(i)
		for i := 0; i < method.Type.NumOut(); i++ {
			methodType := reflect.New(method.Type.Out(i)).Interface()
			switch methodType.(type) {
			case Handler:
				validMethod = false
				break
			}
		}
		if !validMethod {
			continue
		}
		value := valueOf.MethodByName(method.Name).Call([]reflect.Value{})
		//
		routeValue := value[0].FieldByName("Route")
		httpVerbValue := value[0].FieldByName("Method")
		handlerValue := value[0].FieldByName("Handler")
		consumesValue := value[0].FieldByName("Consumes")
		producesValue := value[0].FieldByName("Produces")
		//
		router := routeValue.Interface().(string)
		httpVerb := httpVerbValue.Interface().(string)
		produces := producesValue.Interface().(string)
		consumes := consumesValue.Interface().(string)
		if handlerValue.Interface() == nil {
			log.Fatal("Handler is required")
		}
		if router == "" {
			log.Fatal("Route is required")
		}
		if httpVerb == "" {
			log.Fatal("Method is required")
		}
		if produces == "" {
			log.Fatal("Produces is required")
		}
		if consumes == "" {
			log.Fatal("Consumes is required")
		}
		handler := generateHandler(handlerValue, consumes, produces)
		//
		var r *mux.Router
		Container.Make(&r)
		r.HandleFunc(router, handler).Methods(httpVerb)
	}
}

func generateHandler(handlerValue reflect.Value, consumes, produces string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", produces)
		handlerType := handlerValue.Elem().Type()
		requestPointer := reflect.New(handlerType.In(0))
		log.Println(requestPointer.Type())
		requestBody := reflect.Indirect(reflect.ValueOf(requestPointer.Interface())).Interface().(model.Test)
		switch consumes {
		case "application/json":
			if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
		}
		resp := handlerValue.Elem().Call([]reflect.Value{reflect.ValueOf(requestBody)})
		json.NewEncoder(w).Encode(resp[0].Interface())
	}
}

func validateRequestBody() error {
	return nil
}
