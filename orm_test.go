package gmorm_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/miacio/gmorm"
)

type R9UserInfoModel struct {
	Id            int        `db:"id" json:"id"`
	Uid           string     `db:"uid" json:"uid"`
	CreateTime    *time.Time `db:"create_time" json:"create_time"`
	EffectiveTime *time.Time `db:"effective_time" json:"effective_time"`
	Mobile        *string    `db:"mobile" json:"mobile"`
	Account       *string    `db:"account" json:"account"`
	Password      *string    `db:"password" json:"password"`
	Notes         *string    `db:"notes" json:"notes"`
}

func (*R9UserInfoModel) TableName() string {
	return "r9_user_info"
}

func TestOpen(t *testing.T) {
	DB, _ := sqlx.Open("mysql", "root:123456&Mysql@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local")
	DB.SetConnMaxLifetime(100)
	DB.SetMaxIdleConns(10)
	if err := DB.Ping(); err != nil {
		fmt.Println("open database fail")
		return
	}
	fmt.Println("open database success")

	gm := gmorm.New(DB)
	r9 := gm.GetEngine(&R9UserInfoModel{})
	res := []R9UserInfoModel{}
	if err := r9.Find(&res); err != nil {
		t.Fatalf("find fail: %v", err)
	}

	for _, v := range res {
		m, _ := json.Marshal(v)
		fmt.Println(string(m))
	}

}
