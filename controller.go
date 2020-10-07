package spring

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
)

type Handler struct {
	Route       string
	Handler     interface{}
	Params      []string
	Method      string
	Produces    string
	Consumes    string
	RequestBody interface{}
}

func AddController(controller interface{}) {
	typeOf := reflect.TypeOf(controller)
	valueOf := reflect.ValueOf(controller)
	for i := 0; i < typeOf.NumMethod(); i++ {
		//
		method := typeOf.Method(i).Name
		value := valueOf.MethodByName(method).Call([]reflect.Value{})
		//
		routeValue := value[0].FieldByName("Route")
		httpVerbValue := value[0].FieldByName("Method")
		// handlerValue := value[0].FieldByName("Handler")
		consumesValue := value[0].FieldByName("Consumes")
		producesValue := value[0].FieldByName("Produces")
		requestBodyValue := value[0].FieldByName("RequestBody")
		//
		router := routeValue.Interface().(string)
		httpVerb := httpVerbValue.Interface().(string)
		produces := producesValue.Interface().(string)
		consumes := consumesValue.Interface().(string)
		handler := generateHandler(requestBodyValue, consumes, produces)
		//
		var r *mux.Router
		Container.Make(&r)
		r.HandleFunc(router, handler).Methods(httpVerb)
	}
}

func generateHandler(requestBodyValue reflect.Value, consumes, produces string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", produces)
		var body map[string]interface{}
		switch consumes {
		case "application/json":
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
		}

		if err := validateRequestBody(requestBodyValue, body); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		resp := getReservationImpl(body)
		json.NewEncoder(w).Encode(resp)
	}
}

func getReservationImpl(test map[string]interface{}) string {
	return "Worked"
}

func validateRequestBody(requestBodyValue reflect.Value, body map[string]interface{}) error {
	// what i want is to access the fields in the struct and then valid that against the incoming request map
	v := requestBodyValue.Elem().Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := field.Tag.Get("json")
		fieldType := field.Type
		value, ok := body[fieldName]
		if !ok {
			return errors.New("Missing field " + fieldName)
		}

		if fmt.Sprintf("%T", value) != fieldType.String() {
			return errors.New("Invalid type " + fieldType.String() + " for field " + fieldName)
		}
	}
	return nil
}
