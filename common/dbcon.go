package common

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

//常量数据库连接定义
const mysql = "huanqiu:huanqiu@tcp(114.67.97.70:3306)/huanqiu"
//mysql连接
func ConMysql() *sql.DB {
	db, err := sql.Open("mysql", mysql)
	if err != nil {
		fmt.Println("数据库连接错误！")
	}
	return db
}
