package gmorm_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/miacio/gmorm"
)

type R9UserInfoModel struct {
	Id            int        `db:"id" json:"id"`
	Uid           string     `db:"uid" json:"uid"`
	CreateTime    *time.Time `db:"create_time" json:"create_time"`
	EffectiveTime *time.Time `db:"effective_time" json:"effective_time"`
	Mobile        string     `db:"mobile" json:"mobile"`
	Account       string     `db:"account" json:"account"`
	Password      string     `db:"password" json:"password"`
	Notes         *string    `db:"notes" json:"notes"`
}

func (R9UserInfoModel) TableName() string {
	return "r9_user_info"
}

var (
	gm *gmorm.GmOrm
)

func init() {
	DB, _ := sqlx.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local")
	DB.SetConnMaxLifetime(100)
	DB.SetMaxIdleConns(10)
	if err := DB.Ping(); err != nil {
		fmt.Println("open database fail")
		return
	}
	fmt.Println("open database success")

	gm = gmorm.New(DB)
}

func TestFind(t *testing.T) {
	r9 := gmorm.NewEngine[R9UserInfoModel](gm, R9UserInfoModel{})
	res := []R9UserInfoModel{}
	if err := r9.Find(&res, "id", "uid", "create_time", "account"); err != nil {
		t.Fatalf("find fail: %v", err)
	}

	for _, v := range res {
		m, _ := json.Marshal(v)
		fmt.Println(string(m))
	}
}

func TestInsert(t *testing.T) {
	r9 := gmorm.NewEngine[R9UserInfoModel](gm, R9UserInfoModel{})

	// n := time.Now()
	mb := "18888888888"

	r9a := make([]R9UserInfoModel, 0)
	for i := 0; i < 100; i++ {
		now := R9UserInfoModel{
			Uid:           strings.ReplaceAll(strings.ToUpper(uuid.NewString()), "-", ""),
			CreateTime:    nil,
			EffectiveTime: nil,
			Mobile:        mb,
			Account:       mb,
			Password:      mb,
			Notes:         &mb,
		}

		r9a = append(r9a, now)
	}
	r9 = r9.InsertNamed("db", r9a...)
	sql, _, _ := r9.Value()
	fmt.Println(sql)
	res, err := r9.Exec()
	if err != nil {
		t.Fatalf("insert fail: %v", err)
	}
	successLine, _ := res.RowsAffected()
	fmt.Printf("insert line is :%v", successLine)

}

func TestDelete(t *testing.T) {
	r9 := gmorm.NewEngine[R9UserInfoModel](gm, R9UserInfoModel{})
	if _, err := r9.Where("id > 2").Delete().Exec(); err != nil {
		t.Fatalf("insert fail: %v", err)
	}
}

func TestUpdate(t *testing.T) {
	r9 := gmorm.NewEngine[R9UserInfoModel](gm, R9UserInfoModel{})
	if _, err := r9.Where("id = 2").Set("effective_time = ?", time.Now()).Update().Exec(); err != nil {
		t.Fatalf("insert fail: %v", err)
	}
}
