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

func AddRepository(repo interface{}, dataType interface{}) {
	repoPtr := reflect.ValueOf(repo)
	if repoPtr.Kind() != reflect.Ptr {
		log.Fatal("Repository is not pointer")
	}
	repoPtr = repoPtr.Elem()
	repoPtrType := repoPtr.Type()
	var nilError = reflect.Zero(reflect.TypeOf((*error)(nil)).Elem())
	repositories[repoPtrType.Name()] = repoPtr
	for i := 0; i < repoPtr.NumField(); i++ {
		field := repoPtrType.Field(i)
		if field.Name == "CRUDRepository" {
			field := repoPtr.Field(i)
			field.Set(reflect.ValueOf(CRUDRepository{dataType: dataType}))
		} else {
			// ...
			query := field.Tag.Get("query")
			if query != "" {
				// the implementation passed to MakeFunc.
				// It must work in terms of reflect.Values so that it is possible
				// to write code without knowing beforehand what the types
				// will be.
				implemention := func(in []reflect.Value) []reflect.Value {
					dataTypePtr := reflect.New(field.Type.Out(0).Elem())
					v := dataTypePtr.Elem()
					t := field.Type.Out(0).Elem()

					inVal := []interface{}{}
					for _, val := range in {
						inVal = append(inVal, val.Interface())
					}
					stmt, err := app.db.Prepare(query)
					if err != nil {
						return []reflect.Value{dataTypePtr, reflect.ValueOf(err)}
					}
					// TODO: check if exec or query
					rows, err := stmt.Query(inVal...)
					if err != nil {
						return []reflect.Value{dataTypePtr, reflect.ValueOf(err)}
					}

					cols, err := rows.Columns()
					if err != nil {
						return []reflect.Value{dataTypePtr, reflect.ValueOf(err)}
					}

					var m map[string]interface{}
					for rows.Next() {
						columns := make([]interface{}, len(cols))
						columnPointers := make([]interface{}, len(cols))
						for i := range columns {
							columnPointers[i] = &columns[i]
						}

						if err := rows.Scan(columnPointers...); err != nil {
							return []reflect.Value{dataTypePtr, reflect.ValueOf(err)}
						}

						m = make(map[string]interface{})
						for i, colName := range cols {
							val := columnPointers[i].(*interface{})
							m[colName] = *val
						}

					}

					for i := 0; i < v.NumField(); i++ {
						field := strings.Split(t.Field(i).Tag.Get("json"), ",")[0]

						if item, ok := m[field]; ok {
							if v.Field(i).CanSet() {
								if item != nil {
									switch v.Field(i).Kind() {
									case reflect.Int:
										v.Field(i).SetInt(item.(int64))
									case reflect.String:
										v.Field(i).SetString(item.(string))
									case reflect.Float32, reflect.Float64:
										v.Field(i).SetFloat(item.(float64))
									case reflect.Ptr:
										if reflect.ValueOf(item).Kind() == reflect.Bool {
											itemBool := item.(bool)
											v.Field(i).Set(reflect.ValueOf(&itemBool))
										}
									case reflect.Struct:
										v.Field(i).Set(reflect.ValueOf(item))
									default:
										fmt.Println(t.Field(i).Name, ": ", v.Field(i).Kind(), " - > - ", reflect.ValueOf(item).Kind()) // @todo remove after test out the Get methods
									}
								}
							}
						}
					}

					return []reflect.Value{dataTypePtr, nilError}
				}
				// frepoPtr is a pointer to a function.
				// Obtain the function value itself (likely nil) as a reflect.Value
				// so that we can query its type and then set the value.
				fn := repoPtr.Field(i)

				// Make a function of the right type.
				v := reflect.MakeFunc(fn.Type(), implemention)

				// Assign it to the value fn represents.
				fn.Set(v)
			}
		}
	}
}

type CRUDRepository struct {
	dataType interface{}
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

func (c CRUDRepository) FindByID(id string) (interface{}, error) {
	log.Println("[FindByID]: starting database query")
	value := reflect.ValueOf(c.dataType)
	v := value.Elem()
	t := v.Type()

	tableName := strings.ToLower(t.Name())
	log.Println(tableName, id)
	if v.NumMethod() != 0 {
		tableMethod := v.MethodByName("TableName")
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

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var m map[string]interface{}
	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		m = make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			m[colName] = *val
		}

	}

	for i := 0; i < v.NumField(); i++ {
		field := strings.Split(t.Field(i).Tag.Get("json"), ",")[0]

		if item, ok := m[field]; ok {
			if v.Field(i).CanSet() {
				if item != nil {
					switch v.Field(i).Kind() {
					case reflect.Int:
						v.Field(i).SetInt(item.(int64))
					case reflect.String:
						v.Field(i).SetString(item.(string))
					case reflect.Float32, reflect.Float64:
						v.Field(i).SetFloat(item.(float64))
					case reflect.Ptr:
						if reflect.ValueOf(item).Kind() == reflect.Bool {
							itemBool := item.(bool)
							v.Field(i).Set(reflect.ValueOf(&itemBool))
						}
					case reflect.Struct:
						v.Field(i).Set(reflect.ValueOf(item))
					default:
						fmt.Println(t.Field(i).Name, ": ", v.Field(i).Kind(), " - > - ", reflect.ValueOf(item).Kind()) // @todo remove after test out the Get methods
					}
				}
			}
		}
	}
	log.Println("[FindByID]: data: " + fmt.Sprintf("%v", m))
	log.Println("[FindByID]: ending database query")
	return value.Interface(), nil
}
