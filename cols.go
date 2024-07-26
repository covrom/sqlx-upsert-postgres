package upsert

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/jmoiron/sqlx"
)

type StructColumn[T any] struct {
	StructName string
	DBName     string
	IsPK       bool
}

type Columns[T any] []StructColumn[T]

func (c Columns[T]) PKs() []string {
	ret := make([]string, 0, len(c))
	for _, v := range c {
		if v.IsPK {
			ret = append(ret, v.DBName)
		}
	}
	return ret
}

func (c Columns[T]) DBs() []string {
	ret := make([]string, len(c))
	for i, v := range c {
		ret[i] = v.DBName
	}
	return ret
}

type Upserter interface {
	// returns database name of primary key columnts
	UpsertPrimaryKeyColumns() []string
	// returns database name of skipped columns
	UpsertSkipColumns() []string
}

// Tag "db" - name of the column in the database
// Tag "pk" - any non-empty value defines the column for conflict resolution
// Tag store:"-" skips the field in any case
func StructColumns[T any](mVal T) (Columns[T], error) {
	value := reflect.Indirect(reflect.ValueOf(mVal))
	if value.Kind() != reflect.Struct {
		return nil, fmt.Errorf("only structs are supported: %s is not a struct", value.Type())
	}
	typ := value.Type()
	ret := make(Columns[T], 0, typ.NumField())
	if err := fillColumns(typ, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func fillColumns[T any](typ reflect.Type, columns *Columns[T]) error {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("%s not a struct or a pointer to struct", typ.String())
	}

	var el T
	var ipks, isks []string

	ups, ok := any(el).(Upserter)
	if ok {
		ipks = ups.UpsertPrimaryKeyColumns()
		isks = ups.UpsertSkipColumns()
	}

	for i := 0; i < typ.NumField(); i++ {
		structField := typ.Field(i)
		if len(structField.PkgPath) > 0 {
			continue
		}

		if structField.Anonymous {
			if err := fillColumns(structField.Type, columns); err != nil {
				return err
			}
			continue
		}
		if structField.Tag.Get("store") == "-" {
			continue
		}

		dbTag := structField.Tag.Get("db")
		if dbTag == "" {
			dbTag = sqlx.NameMapper(structField.Name)
		}

		if slices.Contains(isks, dbTag) {
			continue
		}

		pkTag := structField.Tag.Get("pk")

		*columns = append(*columns, StructColumn[T]{
			StructName: structField.Name,
			DBName:     dbTag,
			IsPK:       (len(ipks) == 0 && pkTag != "") || slices.Contains(ipks, dbTag),
		})
	}

	return nil
}
