package main

import (
	"./common"
	"./game"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/in_table", game.InTable)   	 //进桌
	mux.HandleFunc("/out_table", game.OutTable) 	 //离桌

	mux.HandleFunc("/lh_set_bet", game.LHSetBet)     //龙虎确定下注
	mux.HandleFunc("/lh_quit_bet", game.LHQuitBet)   //龙虎取消下注
	mux.HandleFunc("/lh_game_over", game.LHGameOver) //龙虎确定游戏结果

	mux.HandleFunc("/bjl_set_bet", game.BJLSetBet)     //百家乐确定下注
	mux.HandleFunc("/bjl_quit_bet", game.BjlQuitBet)   //百家乐取消下注
	mux.HandleFunc("/bjl_game_over", game.BjlGameOver) //百家乐游戏结果

	mux.HandleFunc("/test", Test)					 //测试
	http.ListenAndServe("0.0.0.0:8080", mux) 			//监听8080端口
}
//test
func Test(w http.ResponseWriter, r *http.Request)  {
	common.AjaxReturn(w,1,"success",nil)
	return
	code:=1
	msg:="success"
	backData:=common.ResData(code,msg,nil)
	io.WriteString(w, backData)
}