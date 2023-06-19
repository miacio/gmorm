package gmorm

import (
	"github.com/jmoiron/sqlx"
)

// GmOrm 核心ORM对象
type GmOrm struct {
	db *sqlx.DB
}

// New
func New(db *sqlx.DB) *GmOrm {
	return &GmOrm{
		db: db,
	}
}

// Register 注册表对象
func (g *GmOrm) GetEngine(t TbObj) *Engine[TbObj] {
	return NewEngine[TbObj](g, t)
}

func NewEngine[T TbObj](g *GmOrm, t T) *Engine[T] {
	return &Engine[T]{
		GmOrm:       g,
		t:           t,
		whereClause: make([]Clause, 0),
		setClause:   make([]Clause, 0),
		optionType:  0,
		err:         nil,
		sql:         "",
		params:      make([]interface{}, 0),
	}
}
