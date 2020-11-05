package spring

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

var repositories map[string]reflect.Value = make(map[string]reflect.Value)

type CRUDRepository struct {
	dataType interface{}
}

func AddRepository(repo interface{}, dataType interface{}) {
	var nilError = reflect.Zero(reflect.TypeOf((*error)(nil)).Elem())
	ptr := reflect.ValueOf(repo)
	if ptr.Kind() != reflect.Ptr {
		log.Fatal("Repository is not pointer")
	}
	ptr = ptr.Elem()
	repositories[ptr.Type().Name()] = ptr
	for i := 0; i < ptr.NumField(); i++ {
		if ptr.Type().Field(i).Name == "CRUDRepository" {
			field := ptr.Field(i)
			field.Set(reflect.ValueOf(CRUDRepository{dataType: dataType}))
		} else {
			// ...
			query := ptr.Type().Field(i).Tag.Get("query")
			if query != "" {
				// the implementation passed to MakeFunc.
				// It must work in terms of reflect.Values so that it is possible
				// to write code without knowing beforehand what the types
				// will be.
				implemention := func(in []reflect.Value) []reflect.Value {
					inVal := []interface{}{}
					for _, val := range in {
						inVal = append(inVal, val.Interface())
					}
					stmt, err := app.db.Prepare(query)
					if err != nil {
						return []reflect.Value{nilError, reflect.ValueOf(err)}
					}
					rows, err := stmt.Query(inVal...)
					if err != nil {
						return []reflect.Value{nilError, reflect.ValueOf(err)}
					}
					columns, err := rows.Columns()
					if err != nil {
						return []reflect.Value{nilError, reflect.ValueOf(err)}
					}
					colNum := len(columns)

					var values = make([]interface{}, colNum)
					for i := range values {
						var ii interface{}
						values[i] = &ii
					}
					for rows.Next() {
						err := rows.Scan(values...)
						if err != nil {
							return []reflect.Value{ptr, reflect.ValueOf(err)}
						}
						for i, colName := range columns {
							var raw_value = *(values[i].(*interface{}))
							var raw_type = reflect.TypeOf(raw_value)

							fmt.Println(colName, raw_type, raw_value)
						}
					}
					return []reflect.Value{reflect.ValueOf(rows), nilError}
				}
				// fptr is a pointer to a function.
				// Obtain the function value itself (likely nil) as a reflect.Value
				// so that we can query its type and then set the value.
				fn := ptr.Field(i)

				// Make a function of the right type.
				v := reflect.MakeFunc(fn.Type(), implemention)

				// Assign it to the value fn represents.
				fn.Set(v)
			}
		}
	}
}

func (c CRUDRepository) Save(model interface{}) error {
	log.Println("[Save]: starting database entry")
	valueOfDataType := reflect.Indirect(reflect.ValueOf(c.dataType))
	typeOf := valueOfDataType.Type()
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
	query, err := app.db.Prepare("INSERT INTO " + tableName + " (" + strings.Join(fieldNames, ",") + ") VALUES ( " + strings.Join(fieldMarkers, ",") + " );")
	log.Println("[Save]: saving ", fieldValues)
	_, err = query.Exec(fieldValues...)
	if err != nil {
		return err
	}
	log.Println("[Save]: ending database entry")
	return nil
}

func (c CRUDRepository) FindByID(id string) (map[string]interface{}, error) {
	log.Println("[FindByID]: starting database query")
	valueOfDataType := reflect.Indirect(reflect.ValueOf(c.dataType))
	tableName := strings.ToLower(valueOfDataType.Type().Name())
	log.Println(tableName, id)
	if valueOfDataType.NumMethod() != 0 {
		tableMethod := valueOfDataType.MethodByName("TableName")
		if tableMethod.IsValid() {
			tableNameValue := tableMethod.Call([]reflect.Value{})
			tableName = tableNameValue[0].Interface().(string)
		}
	}

	query, err := app.db.Prepare("SELECT * FROM " + tableName + " WHERE id = " + id + ";")
	if err != nil {
		return nil, err
	}
	rows, err := query.Query()
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
