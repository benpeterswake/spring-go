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
		}
	}
}

func addFunction() {
	// swap is the implementation passed to MakeFunc.
	// It must work in terms of reflect.Values so that it is possible
	// to write code without knowing beforehand what the types
	// will be.
	swap := func(in []reflect.Value) []reflect.Value {
		return []reflect.Value{in[1], in[0]}
	}

	// makeSwap expects fptr to be a pointer to a nil function.
	// It sets that pointer to a new function created with MakeFunc.
	// When the function is invoked, reflect turns the arguments
	// into Values, calls swap, and then turns swap's result slice
	// into the values returned by the new function.
	makeSwap := func(fptr interface{}) {
		// fptr is a pointer to a function.
		// Obtain the function value itself (likely nil) as a reflect.Value
		// so that we can query its type and then set the value.
		fn := reflect.ValueOf(fptr).Elem()

		// Make a function of the right type.
		v := reflect.MakeFunc(fn.Type(), swap)

		// Assign it to the value fn represents.
		fn.Set(v)
	}

	// Make and call a swap function for ints.
	var intSwap func(int, int) (int, int)
	makeSwap(&intSwap)
	fmt.Println(intSwap(0, 1))

	// Make and call a swap function for float64s.
	var floatSwap func(float64, float64) (float64, float64)
	makeSwap(&floatSwap)
	fmt.Println(floatSwap(2.72, 3.14))
}

type CRUDRepository struct{}

func (c CRUDRepository) Save(model interface{}) error {
	log.Println("[Save]: starting database entry")
	typeOf := reflect.TypeOf(dataType)
	valueOfDataType := reflect.ValueOf(dataType)
	valueOf := reflect.ValueOf(model)
	tableName := strings.ToLower(typeOf.Name())
	if valueOfDataType.NumMethod() != 0 {
		tableMethod := valueOfDataType.MethodByName("TableName")
		if tableMethod.IsValid() {
			tableNameValue := tableMethod.Call([]reflect.Value{})
			tableName = tableNameValue[0].Interface().(string)
		}
	}
	sqlStatement := `INSERT INTO ` + tableName + ` (something) VALUES (`
	numOfFields := typeOf.NumField()
	fieldValues := []interface{}{}
	for i := 0; i < numOfFields; i++ {
		// check if field is ID serial
		field := typeOf.Field(i)
		id := field.Tag.Get("id")
		if id != "" {
			log.Println(id)
			continue
		}
		fieldValues = append(fieldValues, valueOf.Field(i).Interface())
		if (i + 1) == numOfFields {
			sqlStatement += "$" + strconv.FormatInt(int64(i), 10)
		} else {
			sqlStatement += "$" + strconv.FormatInt(int64(i), 10) + ","
		}
	}
	sqlStatement += ");"
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
		tableMethod := valueOfDataType.MethodByName("Table")
		if !tableMethod.IsNil() {
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
