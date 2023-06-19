package gmorm

import (
	"encoding/json"
	"reflect"
	"time"
)

// TbObj 表对象接口
type TbObj interface {
	TableName() string // 返回表名
}

type StringObj interface {
	String() any
}

// ColObj 字段值接口
type ColObj interface {
	Value() any
}

// KeyTo 字段key优化
func KeyTo(key string) string {
	if key != "" {
		if len(key) > 1 && key[0] != '`' {
			return "`" + key + "`"
		}
	}
	return key
}

// ObjTagMap 结构体对象依据tag获取数据转map
// 当结构体中存在tag字段但无值情况给nil值
// 当tag字段类型为slice或map时将会进行json处理
// 当tag字段类型为其他类型时,需要在对应类型上添加Value方法去重构它,该程序将会依据Value方法获取其字段值
// 此方法将作为sqlx的Named模式将结构体转换为map使用
func ObjTagMap(obj any, tag string) (map[string]interface{}, error) {
	params := make(map[string]interface{}, 0)
	if obj != nil {
		valueOf := reflect.ValueOf(obj)
		typeOf := reflect.TypeOf(obj)
		if reflect.TypeOf(obj).Kind() == reflect.Ptr {
			valueOf = reflect.ValueOf(obj).Elem()
			typeOf = reflect.TypeOf(obj).Elem()
		}
		numField := valueOf.NumField()
		for i := 0; i < numField; i++ {
			tag := typeOf.Field(i).Tag.Get(tag)
			if len(tag) > 0 && tag != "-" {
				params[tag] = nil

				switch valueOf.Field(i).Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16,
					reflect.Int32, reflect.Int64:
					params[tag] = valueOf.Field(i).Int()
				case reflect.Uint, reflect.Uint8, reflect.Uint16,
					reflect.Uint32, reflect.Uint64:
					params[tag] = valueOf.Field(i).Uint()
				case reflect.Float32, reflect.Float64:
					params[tag] = valueOf.Field(i).Float()
				case reflect.Bool:
					params[tag] = valueOf.Field(i).Bool()
				case reflect.String:
					if len(valueOf.Field(i).String()) > 0 {
						params[tag] = valueOf.Field(i).String()
					} else {
						params[tag] = ""
					}
				case reflect.Map:
					if !valueOf.Field(i).IsNil() {
						bytes, err := json.Marshal(valueOf.Field(i).Interface())
						if err != nil {
							return nil, err
						} else {
							params[tag] = string(bytes)
						}
					}
				case reflect.Slice:
					if ss, ok := valueOf.Field(i).Interface().([]string); ok {
						var pv string
						for _, sv := range ss {
							pv += sv + ","
						}
						if len(pv) >= len(",") && pv[len(pv)-len(","):] == "," {
							pv = pv[:len(pv)-1]
						}
						if len(pv) > 0 {
							params[tag] = pv
						}
					}
				default:
					if valueOf.Field(i).Kind() == reflect.Ptr {
						if valueOf.Field(i).IsZero() || valueOf.Field(i).IsNil() {
							continue
						}
					}
					if valueOf.Field(i).Type() == reflect.TypeOf(time.Time{}) || valueOf.Field(i).Type() == reflect.TypeOf(&time.Time{}) {
						params[tag] = valueOf.Field(i).Interface()
						continue
					}
					if co, ok := valueOf.Field(i).Interface().(ColObj); ok {
						if co != nil {
							params[tag] = co.Value()
						}
					}
				}
			}
		}
	}
	return params, nil
}
