package zxxorm

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
)

func EngineInsertOne(engine *xorm.Engine, bean interface{}) error {
	affected, err := engine.InsertOne(bean)
	if (affected <= 0 && err == nil) || (affected > 0 && err != nil) {
		panic(fmt.Sprintf("xorm的逻辑异常,InsertOne,affected=%v,err=%v", affected, err))
	}
	return err
}

func SessionInsertOne(session *xorm.Session, bean interface{}) error {
	affected, err := session.InsertOne(bean)
	if (affected <= 0 && err == nil) || (affected > 0 && err != nil) {
		panic(fmt.Sprintf("xorm的逻辑异常,InsertOne,affected=%v,err=%v", affected, err))
	}
	return err
}

func Update(engine *xorm.Engine, bean interface{}, condiBeans ...interface{}) error {
	affected, err := engine.Update(bean, condiBeans...)
	if (affected <= 0 && err == nil) || (affected > 0 && err != nil) {
		panic(fmt.Sprintf("xorm的逻辑异常,Update,affected=%v,err=%v", affected, err))
	}
	return err
}

// 其行为如下所示(我感觉着不用"事务"也没问题):
// 1. 无主键=>执行insert语句=>结束函数.
// 2. 有主键=>where(主键)&update=>报错=>结束函数.
//                             =>无错=>修改了数据=>结束函数.
//                                   =>未修改数据=>执行insert语句=>结束函数.
func Upsert(engine *xorm.Engine, bean interface{}) error {
	var err error

	var tbInfo *xorm.Table = engine.TableInfo(bean)
	if tbInfo == nil {
		err = fmt.Errorf("找不到对应的TableInfo")
		return err
	}

	if tbInfo.PrimaryKeys == nil || len(tbInfo.PrimaryKeys) == 0 {
		var affected int64
		affected, err = engine.InsertOne(bean)
		if (affected <= 0 && err == nil) || (affected > 0 && err != nil) {
			panic(fmt.Sprintf("xorm的逻辑异常,InsertOne,affected=%v,err=%v", affected, err))
		}
		return err
	}

	var query string = strings.Join(tbInfo.PrimaryKeys, " = ? AND ") + " = ?"
	var args []interface{} = make([]interface{}, 0, len(tbInfo.PrimaryKeys))

	for _, col := range tbInfo.Columns() { //我们在这里假设[tbInfo.PrimaryKeys]和[tbInfo.Columns()]的顺序是一致的.
		var isPkField bool = false
		var __PkIndex int = -1
		for idx, pkFieldName := range tbInfo.PrimaryKeys {
			if pkFieldName == col.Name {
				isPkField = true
				__PkIndex = idx
				break
			}
		}
		if !isPkField {
			continue
		}
		if len(args) != __PkIndex {
			err = errors.New("假设不成立,请修改代码,让pkName和pkValue对应起来")
			return err
		}

		var fieldValuePtr *reflect.Value = nil
		if fieldValuePtr, err = col.ValueOf(bean); err != nil {
			return err
		}

		var arg interface{}
		if arg, err = value2Interface(col, *fieldValuePtr); err != nil {
			return err
		}

		args = append(args, arg)
	}

	var affected int64
	if affected, err = engine.Where(query, args...).Update(bean); err != nil {
		if affected > 0 {
			panic(fmt.Sprintf("xorm的逻辑异常,Where+Update,affected=%v,err=%v", affected, err))
		}
		return err
	}
	if affected <= 0 {
		affected, err = engine.InsertOne(bean)
	}
	if (affected <= 0 && err == nil) || (affected > 0 && err != nil) {
		panic(fmt.Sprintf("xorm的逻辑异常,Where+Update/InsertOne,affected=%v,err=%v", affected, err))
	}

	return err
}

// 函数[value2Interface]是从[github.com\go-xorm\xorm\session_convert.go]的
// [func (session *Session) value2Interface(col *core.Column, fieldValue reflect.Value) (interface{}, error)]里面摘取出来的,同时做了删减.
// convert a field value of a struct to interface for put into db
func value2Interface(col *core.Column, fieldValue reflect.Value) (interface{}, error) {
	if fieldValue.CanAddr() {
		if fieldConvert, ok := fieldValue.Addr().Interface().(core.Conversion); ok {
			data, err := fieldConvert.ToDB()
			if err != nil {
				return 0, err
			}
			if col.SQLType.IsBlob() {
				return data, nil
			}
			return string(data), nil
		}
	}

	if fieldConvert, ok := fieldValue.Interface().(core.Conversion); ok {
		data, err := fieldConvert.ToDB()
		if err != nil {
			return 0, err
		}
		if col.SQLType.IsBlob() {
			return data, nil
		}
		return string(data), nil
	}

	fieldType := fieldValue.Type()
	k := fieldType.Kind()
	if k == reflect.Ptr {
		if fieldValue.IsNil() {
			return nil, nil
		} else if !fieldValue.IsValid() {
			fmt.Println("the field[", col.FieldName, "] is invalid")
			return nil, nil
		} else {
			// !nashtsai! deference pointer type to instance type
			fieldValue = fieldValue.Elem()
			fieldType = fieldValue.Type()
			k = fieldType.Kind()
		}
	}

	switch k {
	case reflect.Bool:
		return fieldValue.Bool(), nil
	case reflect.String:
		return fieldValue.String(), nil
	case reflect.Struct:
		// xorm 里面是支持的, 因为函数拿不出来, 这里不予支持.
		return nil, fmt.Errorf("Unsupported type %v", fieldValue.Type())
	case reflect.Complex64, reflect.Complex128:
		bytes, err := json.Marshal(fieldValue.Interface())
		if err != nil {
			fmt.Println(err)
			return 0, err
		}
		return string(bytes), nil
	case reflect.Array, reflect.Slice, reflect.Map:
		if !fieldValue.IsValid() {
			return fieldValue.Interface(), nil
		}

		if col.SQLType.IsText() {
			bytes, err := json.Marshal(fieldValue.Interface())
			if err != nil {
				fmt.Println(err)
				return 0, err
			}
			return string(bytes), nil
		} else if col.SQLType.IsBlob() {
			var bytes []byte
			var err error
			if (k == reflect.Array || k == reflect.Slice) &&
				(fieldValue.Type().Elem().Kind() == reflect.Uint8) {
				bytes = fieldValue.Bytes()
			} else {
				bytes, err = json.Marshal(fieldValue.Interface())
				if err != nil {
					fmt.Println(err)
					return 0, err
				}
			}
			return bytes, nil
		}
		return nil, xorm.ErrUnSupportedType
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		return int64(fieldValue.Uint()), nil
	default:
		return fieldValue.Interface(), nil
	}
}
