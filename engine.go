package gmorm

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type optionType uint

const (
	_      optionType = iota
	SELECT            // 查询
	UPDATE            // 修改
	DELETE            // 删除
	INSERT            // 新增
)

// Engine 引擎
type Engine[T TbObj] struct {
	*GmOrm        // DB组件块
	t           T // 表结构体
	whereClause []Clause
	setClause   []Clause

	optionType optionType // 操作类型
	err        error      // 执行操作时产生的错误

	sql    string        // 生成的sql语句
	params []interface{} // 参数组
}

// Clear 重置
func (e *Engine[T]) Clear() *Engine[T] {
	e.whereClause = make([]Clause, 0)
	e.setClause = make([]Clause, 0)
	e.optionType = 0
	e.err = nil
	e.sql = ""
	e.params = make([]interface{}, 0)
	return e
}

// whereAppend
// end 是否结束当前子句
func (e *Engine[T]) whereAppend(end bool, condition string, vals ...interface{}) {
	var clause Clause
	if e.whereClause == nil || len(e.whereClause) == 0 {
		clause = NewClause()
		clause = NewClause()
		clause.Condition = append(clause.Condition, condition)
		clause.Params = append(clause.Params, vals...)
		clause.End = end
		e.whereClause = append(e.whereClause, clause)
		return
	}
	clause = e.whereClause[len(e.whereClause)-1]
	if !clause.End {
		clause.Condition = append(clause.Condition, condition)
		clause.Params = append(clause.Params, vals...)
		e.whereClause[len(e.whereClause)-1] = clause
	} else {
		newClause := NewClause()
		newClause.Condition = append(newClause.Condition, condition)
		newClause.Params = append(newClause.Params, vals...)
		e.whereClause = append(e.whereClause, newClause)
	}
}

// CloseClause 关闭当前子句
// 用于进入到下一条 or 子句 或 结束当前or子句操作
func (e *Engine[T]) CloseClause() *Engine[T] {
	if len(e.whereClause) > 0 {
		clause := e.whereClause[len(e.whereClause)-1]
		clause.End = true
		e.whereClause[len(e.whereClause)-1] = clause
	}
	return e
}

// Where
func (e *Engine[T]) Where(condition string, vals ...interface{}) *Engine[T] {
	e.CloseClause()
	e.whereAppend(false, condition, vals...)
	return e
}

// And
func (e *Engine[T]) And(condition string, vals ...interface{}) *Engine[T] {
	e.whereAppend(false, condition, vals...)
	return e
}

// Or
func (e *Engine[T]) Or(condition string, vals ...interface{}) *Engine[T] {
	e.CloseClause()
	e.whereAppend(false, condition, vals...)
	return e
}

// where where子句生成器
func (e *Engine[T]) where() (string, []interface{}) {
	sqlChain := make([]string, 0)
	params := make([]interface{}, 0)
	for _, c := range e.whereClause {
		sqlChain = append(sqlChain, strings.Join(c.Condition, " and "))
		params = append(params, c.Params...)
	}
	if len(sqlChain) > 1 {
		for i := range sqlChain {
			sqlChain[i] = "(" + sqlChain[i] + ")"
		}
	}
	return strings.Join(sqlChain, " or "), params
}

// Set
func (e *Engine[T]) Set(condition string, vals ...interface{}) *Engine[T] {
	var clause Clause
	if e.setClause == nil || len(e.setClause) == 0 {
		clause = NewClause()
	}
	clause.Condition = append(clause.Condition, condition)
	clause.Params = append(clause.Params, vals...)
	e.setClause = append(e.setClause, clause)
	return e
}

// set set子句生成器
func (e *Engine[T]) set() (string, []interface{}) {
	sqlChain := make([]string, 0)
	params := make([]interface{}, 0)
	for _, c := range e.setClause {
		sqlChain = append(sqlChain, strings.Join(c.Condition, ","))
		params = append(params, c.Params...)
	}
	return strings.Join(sqlChain, ","), params
}

func (e *Engine[T]) extractColumn(tag string, keyword bool) []string {
	result := make([]string, 0)

	valueOf := reflect.ValueOf(e.t)
	typeOf := reflect.TypeOf(e.t)

	if reflect.TypeOf(e.t).Kind() == reflect.Ptr {
		valueOf = reflect.ValueOf(e.t).Elem()
		typeOf = reflect.TypeOf(e.t).Elem()
	}
	numField := valueOf.NumField()
	for i := 0; i < numField; i++ {
		tag := typeOf.Field(i).Tag.Get(tag)
		if len(tag) > 0 && tag != "-" {
			if keyword {
				result = append(result, KeyTo(tag))
			} else {
				result = append(result, tag)
			}
		}
	}
	return result
}

// Select select语句生成器
func (e *Engine[T]) Select(columns ...string) *Engine[T] {
	if e.optionType == 0 && e.err == nil {
		sqlTemp := "SELECT %s FROM %s"
		var sql string
		if columns == nil {
			columns = e.extractColumn("db", true)
		}
		if columns != nil && len(columns) > 0 {
			sql = fmt.Sprintf(sqlTemp, strings.Join(columns, ","), e.t.TableName())
		} else {
			sql = fmt.Sprintf(sqlTemp, "*", e.t.TableName())
		}

		where, params := e.where()
		if where != "" {
			sql = sql + " WHERE " + where
		}
		e.sql = sql
		e.params = params
		e.optionType = SELECT
	} else {
		e.err = errors.New("sql engine only option error")
	}
	return e
}

func (se *Engine[T]) Count() *Engine[T] {
	return se.Select("count(1)")
}

// Update update语句生成器
// 此方法执行时如果update语句没有where条件将会抛出错误
func (e *Engine[T]) Update() *Engine[T] {
	if e.optionType == 0 && e.err == nil {
		sqlTemp := "UPDATE %s SET %s WHERE %s"
		params := make([]interface{}, 0)
		setSql, setParams := e.set()
		if setSql == "" {
			e.err = errors.New("no update column")
			return e
		}
		whereSql, whereParams := e.where()
		if whereSql == "" {
			e.err = errors.New("where clause is empty")
			return e
		}
		params = append(params, setParams...)
		params = append(params, whereParams...)
		sql := fmt.Sprintf(sqlTemp, e.t.TableName(), setSql, whereSql)
		e.sql = sql
		e.params = params
		e.optionType = UPDATE
	} else {
		e.err = errors.New("sql engine only option error")
	}
	return e
}

// Delete delete语句生成器
// 此方法执行时如果delete语句没有where条件将会抛出错误
func (e *Engine[T]) Delete() *Engine[T] {
	if e.optionType == 0 && e.err == nil {
		sqlTemp := "DELETE FROM %s WHERE %s"
		params := make([]interface{}, 0)
		whereSql, whereParams := e.where()
		if whereSql == "" {
			e.err = errors.New("where clause is empty")
			return e
		}
		params = append(params, whereParams...)
		e.sql = fmt.Sprintf(sqlTemp, e.t.TableName(), whereSql)
		e.params = params
		e.optionType = DELETE
	} else {
		e.err = errors.New("sql engine only option error")
	}
	return e
}

// Insert insert named语句生成器(允许生成批量插入)
// 此方法依据tag获取字段名称,并将依据此tag的值设定为列名进行插入语句生成
func (e *Engine[T]) InsertNamed(tag string, objs ...T) *Engine[T] {
	if e.optionType == 0 && e.err == nil {
		sqlTemp := "INSERT INTO %s (%s) VALUES (%s)"

		columns := e.extractColumn(tag, false)
		if columns == nil {
			e.err = errors.New("columns is not found")
			return e
		}
		valColumns := make([]string, 0)
		for i := range columns {
			valColumns = append(valColumns, ":"+columns[i])
			columns[i] = KeyTo(columns[i])
		}
		e.sql = fmt.Sprintf(sqlTemp, e.t.TableName(), strings.Join(columns, ","), strings.Join(valColumns, ","))
		e.params = make([]interface{}, 0)
		for i := range objs {
			param, err := ObjTagMap(objs[i], tag)
			if err != nil {
				e.err = err
				return e
			}
			e.params = append(e.params, param)
		}
		e.optionType = INSERT
	} else {
		e.err = errors.New("sql engine only option error")
	}
	return e
}

func (e *Engine[T]) Get(obj any, columns ...string) error {
	if e.err != nil {
		return e.err
	}
	if e.optionType == 0 {
		e.Select(columns...)
	}
	if e.optionType == SELECT {
		err := e.db.Get(obj, e.sql, e.params...)
		e.Clear()
		return err
	}
	return errors.New("unknown option type")
}

func (e *Engine[T]) Find(obj any, columns ...string) error {
	if e.err != nil {
		return e.err
	}
	if e.optionType == 0 {
		e.Select(columns...)
	}
	if e.optionType == SELECT {
		err := e.db.Select(obj, e.sql, e.params...)
		e.Clear()
		return err
	}
	return errors.New("unknown option type")
}

func (e *Engine[T]) Exec() (sql.Result, error) {
	if e.err != nil {
		return nil, e.err
	}
	var result sql.Result
	var err error
	switch e.optionType {
	case UPDATE, DELETE:
		result, err = e.db.Exec(e.sql, e.params...)
		e.Clear()
		return result, err
	case INSERT:
		result, err = e.db.NamedExec(e.sql, e.params)
		e.Clear()
		return result, err
	}
	return nil, errors.New("unknow option type")
}

func (e *Engine[T]) Value() (string, []interface{}, error) {
	return e.sql, e.params, e.err
}
