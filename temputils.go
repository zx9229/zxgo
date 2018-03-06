package zxgo

import (
	"reflect"
	"strconv"
	"strings"
)

// 通过map修改data(中的各个字段)的值.
func ModifyByMap(data interface{}, kvs map[string]string, upperKey bool) {
	var cacheKvs map[string]string = nil
	if upperKey {
		cacheKvs = make(map[string]string, 0)
		for k, v := range kvs { //TODO:怎么复制一个map.
			k2 := strings.ToUpper(k)
			cacheKvs[k2] = v
		}
	} else {
		cacheKvs = kvs
	}

	elem := reflect.ValueOf(data).Elem()
	elemType := elem.Type()

	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		if !field.CanSet() { //一般情况下,变量首字母是小写的,不可Set.
			continue
		}
		fieldKind := field.Kind()
		fieldName := elemType.Field(i).Name
		fieldNameFind := fieldName
		if upperKey {
			fieldNameFind = strings.ToUpper(fieldName)
		}
		if cacheV, isOk := cacheKvs[fieldNameFind]; isOk {
			switch fieldKind {
			case reflect.Bool:
				if b, err := strconv.ParseBool(cacheV); err == nil {
					field.SetBool(b)
				}
			case reflect.Int:
				fallthrough
			case reflect.Int8:
				fallthrough
			case reflect.Int16:
				fallthrough
			case reflect.Int32:
				fallthrough
			case reflect.Int64:
				fallthrough
			case reflect.Uint:
				fallthrough
			case reflect.Uint8:
				fallthrough
			case reflect.Uint16:
				fallthrough
			case reflect.Uint32:
				fallthrough
			case reflect.Uint64:
				if i, err := strconv.ParseInt(cacheV, 10, 64); err == nil {
					field.SetInt(i)
				}
			case reflect.Float32:
				fallthrough
			case reflect.Float64:
				if f, err := strconv.ParseFloat(cacheV, 64); err == nil {
					field.SetFloat(f)
				}
			case reflect.String:
				field.SetString(cacheV)
			default:
				panic("unknown fieldKind=" + strconv.Itoa(int(fieldKind)))
			}
		}
	}
}
