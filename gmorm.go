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
	return &Engine[TbObj]{
		t:           t,
		GmOrm:       g,
		whereClause: make([]Clause, 0),
		setClause:   make([]Clause, 0),
		optionType:  0,
		err:         nil,
		sql:         "",
		params:      make([]interface{}, 0),
	}
}
