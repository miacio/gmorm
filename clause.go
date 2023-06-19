package gmorm

// Clause 子句
// 单个块的子句为 and 当存在多子句时,使用 or 拼接
type Clause struct {
	Condition []string      // 语句(条件语句或设置语句) xxx = ? | xxx in (?) ...
	Params    []interface{} // 值
	End       bool          // 是否结束当前子句
}

// NewClause 创建子句
func NewClause() Clause {
	return Clause{
		Condition: make([]string, 0),
		Params:    make([]interface{}, 0),
	}
}
