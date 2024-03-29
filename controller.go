package spring

import (
	"encoding/json"
	"log"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
)

func AddController(controller interface{}) {
	valueOf := reflect.ValueOf(controller)
	if valueOf.Kind() != reflect.Ptr {
		log.Fatal("Controller is not pointer")
	}
	typeOf := reflect.TypeOf(controller)
	valueOf = valueOf.Elem()
	for i := 0; i < valueOf.NumField(); i++ {
		repo, ok := services[valueOf.Field(i).Type().Name()]
		if ok {
			valueOf.Field(i).Set(repo)
		}
	}

	for i := 0; i < typeOf.NumMethod(); i++ {
		validMethod := false
		method := typeOf.Method(i)
		for i := 0; i < method.Type.NumOut(); i++ {
			methodType := method.Type.Out(i).Name()
			switch methodType {
			case "Handler":
				validMethod = true
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
		handlerInterface := handlerValue.Interface()
		if handlerInterface == nil {
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
		handler := generateHandler(handlerValue, httpVerb, consumes, produces)
		//
		app.muxRouter.HandleFunc(router, handler).Methods(httpVerb)
	}
}

func generateHandler(handlerValue reflect.Value, method, consumes, produces string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", produces)
		arg := reflect.TypeOf(handlerValue.Interface())
		ptrValue := reflect.New(arg.In(0).Elem())
		requestBody := ptrValue.Interface()
		if method != http.MethodGet {
			switch consumes {
			case "application/json":
				if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
					http.Error(w, "Error unmarsheling request body: "+err.Error(), 400)
					return
				}
				validateFunction := ptrValue.MethodByName("Validate")
				log.Println(validateFunction)
				if validateFunction.IsValid() {
					validation := validateFunction.Call([]reflect.Value{})
					isValid := validation[1].Interface().(bool)
					msg := validation[0].Interface().(string)
					if !isValid {
						http.Error(w, msg, 400)
						return
					}
				}
			}
		}
		var resp []reflect.Value
		vars := mux.Vars(r)
		params := r.URL.Query()
		resp = handlerValue.Elem().Call([]reflect.Value{reflect.ValueOf(requestBody), reflect.ValueOf(vars), reflect.ValueOf(params)})
		switch len(resp) {
		case 1:
			if resp[0].Interface() != nil {
				http.Error(w, resp[0].Interface().(error).Error(), 400)
				return
			}
			w.WriteHeader(200)
		case 2:
			if resp[1].Interface() != nil {
				http.Error(w, resp[1].Interface().(error).Error(), 400)
				return
			}
			w.WriteHeader(200)
			json.NewEncoder(w).Encode(resp[0].Interface())
		}
	}
}

func validateRequestBody() error {
	return nil
}
