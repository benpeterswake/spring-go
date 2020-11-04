package spring

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

var dataType interface{}

func AddRepository(repo interface{}, typ interface{}) {
	dataType = typ
	// loop over the methods on a struct
	// obj := reflect.ValueOf(repo)
	typeOf := reflect.TypeOf(repo)
	for i := 0; i < typeOf.NumField(); i++ {
		if typeOf.Field(i).Name != "CRUDRepository" {
			// ...
			query := typeOf.Field(i).Tag.Get("query")
			if query != "" {
				// // the implementation passed to MakeFunc.
				// // It must work in terms of reflect.Values so that it is possible
				// // to write code without knowing beforehand what the types
				// // will be.
				// implemention := func(in []reflect.Value) []reflect.Value {
				// 	inVal := []interface{}{}
				// 	for _, val := range in {
				// 		inVal = append(inVal, val.Interface())
				// 	}
				// 	result, err := app.db.Exec(query, inVal...)
				// 	if err != nil {
				// 		return []reflect.Value{reflect.ValueOf(nil), reflect.ValueOf(err)}
				// 	}
				// 	return []reflect.Value{reflect.ValueOf(result), reflect.ValueOf(nil)}
				// }
				// // fptr is a pointer to a function.
				// // Obtain the function value itself (likely nil) as a reflect.Value
				// // so that we can query its type and then set the value.
				// fn := obj.Field(i)

				// // Make a function of the right type.
				// v := reflect.MakeFunc(fn.Type(), implemention)

				// // Assign it to the value fn represents.
				// fn.Set(v)
			}
		}
	}
}

type CRUDRepository struct{}

func (c CRUDRepository) Save(model interface{}) error {
	log.Println("[Save]: starting database entry")
	typeOf := reflect.TypeOf(dataType)
	valueOfDataType := reflect.ValueOf(dataType)
	valueOf := reflect.Indirect(reflect.ValueOf(model))
	tableName := strings.ToLower(typeOf.Name())
	if valueOfDataType.NumMethod() != 0 {
		tableMethod := valueOfDataType.MethodByName("TableName")
		if tableMethod.IsValid() {
			tableNameValue := tableMethod.Call([]reflect.Value{})
			tableName = tableNameValue[0].Interface().(string)
		}
	}

	fieldValues := []interface{}{}
	fieldNames := []string{}
	fieldMarkers := []string{}
	numOfFields := typeOf.NumField()
	for i := 0; i < numOfFields; i++ {
		// check if field is ID serial
		field := typeOf.Field(i)
		id := field.Tag.Get("id")
		if id != "" {
			continue
		}
		fieldValues = append(fieldValues, valueOf.Field(i).Interface())
		jsonTag := field.Tag.Get("json")
		fieldNames = append(fieldNames, jsonTag)
		fieldMarkers = append(fieldMarkers, "$"+strconv.FormatInt(int64(i), 10))
	}
	sqlStatement := "INSERT INTO " + tableName + " (" + strings.Join(fieldNames, ",") + ") VALUES ( " + strings.Join(fieldMarkers, ",") + " );"
	log.Println("[Save]: saving ", fieldValues)
	_, err := app.db.Exec(sqlStatement, fieldValues...)
	if err != nil {
		return err
	}
	log.Println("[Save]: ending database entry")
	return nil
}

func (c CRUDRepository) FindByID(id string) (map[string]interface{}, error) {
	log.Println("[FindByID]: starting database query")
	typeOf := reflect.TypeOf(dataType)
	valueOfDataType := reflect.ValueOf(dataType)
	tableName := strings.ToLower(typeOf.Name())
	if valueOfDataType.NumMethod() != 0 {
		tableMethod := valueOfDataType.MethodByName("TableName")
		if tableMethod.IsValid() {
			tableNameValue := tableMethod.Call([]reflect.Value{})
			tableName = tableNameValue[0].Interface().(string)
		}
	}
	query := "SELECT * FROM " + tableName + " WHERE id = " + id + ";"
	rows, err := app.db.Query(query)
	if err != nil {
		return nil, err
	}
	cols, _ := rows.Columns()
	m := make(map[string]interface{})
	for rows.Next() {
		// Create a slice of interface{}'s to represent each column,
		// and a second slice to contain pointers to each item in the columns slice.
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		// Scan the result into the column pointers...
		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.

		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			m[colName] = *val
		}
		// Outputs: map[columnName:value columnName2:value2 columnName3:value3 ...]
	}
	log.Println("[FindByID]: data: " + fmt.Sprintf("%v", m))
	log.Println("[FindByID]: ending database query")
	return m, nil
}
