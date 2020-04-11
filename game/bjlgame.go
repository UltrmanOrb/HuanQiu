package game

import (
	"../common"
	"../model"
	"database/sql"
	"net/http"
	"strconv"
)
//百家乐确定下注
func BJLSetBet(w http.ResponseWriter, r *http.Request)  {
	w.Header().Set("Access-Control-Allow-Origin", "*") //允许跨域
	var data common.BJLPostData
	data.UserId,_= strconv.Atoi(r.Header.Get("user_id"))
	data.DeskId,_=strconv.Atoi(r.PostFormValue("desk_id"))
	data.BootNum,_=strconv.Atoi(r.PostFormValue("boot_num"))
	data.PaveNum,_=strconv.Atoi(r.PostFormValue("pave_num"))

	data.Player,_=strconv.Atoi(r.PostFormValue("player"))
	data.PlayerPair,_=strconv.Atoi(r.PostFormValue("playerPair"))
	data.Tie,_=strconv.Atoi(r.PostFormValue("tie"))
	data.BankerPair,_=strconv.Atoi(r.PostFormValue("bankerPair"))
	data.Banker,_=strconv.Atoi(r.PostFormValue("banker"))

	betMoney :=data.Player+data.PlayerPair+data.Tie+data.BankerPair+data.Banker

	var db *sql.DB = common.ConMysql()
	defer db.Close()
	balance:=model.GetUserBalance(db,data.UserId)/100
	if balance< betMoney {
		common.AjaxReturn(w,0,"余额不足",nil)
		return
	}

	limit:=model.GetBjlDeskLimit(db,data.DeskId)//获取台桌限红

	if data.Player>0 {
		if data.Player<limit["minLimit"]||data.Player>limit["maxLimit"]{
			common.AjaxReturn(w,0,"下注失败",nil)
			return
		}else{
			data.Player=data.Player*100
		}
	}else{
		data.Player=0
	}

	if data.PlayerPair>0 {
		if data.PlayerPair<limit["pairMinLimit"]||data.PlayerPair>limit["pairMaxLimit"]{
			common.AjaxReturn(w,0,"下注失败",nil)
			return
		}else{
			data.PlayerPair=data.PlayerPair*100
		}
	}else{
		data.PlayerPair=0
	}

	if data.Tie>0 {
		if data.Tie<limit["tieMinLimit"]||data.Tie>limit["tieMaxLimit"]{
			common.AjaxReturn(w,0,"下注失败",nil)
			return
		}else{
			data.Tie=data.Tie*100
		}
	}else{
		data.Tie=0
	}

	if data.BankerPair>0 {
		if data.BankerPair<limit["pairMinLimit"]||data.BankerPair>limit["pairMaxLimit"]{
			common.AjaxReturn(w,0,"下注失败",nil)
			return
		}else{
			data.BankerPair=data.BankerPair*100
		}
	}else{
		data.BankerPair=0
	}

	if data.Banker>0 {
		if data.Banker<limit["minLimit"]||data.Banker>limit["maxLimit"]{
			common.AjaxReturn(w,0,"下注失败",nil)
			return
		}else{
			data.Banker=data.Banker*100
		}
	}else{
		data.Banker=0
	}
	code,msg:= model.BjlCertainBet(db,data)
	common.AjaxReturn(w,code,msg,nil)
}
//百家乐取消下注
func BjlQuitBet(w http.ResponseWriter, r *http.Request)  {
	w.Header().Set("Access-Control-Allow-Origin", "*") //允许跨域
	var data common.BJLPostData
	data.UserId,_= strconv.Atoi(r.Header.Get("user_id"))
	data.DeskId,_=strconv.Atoi(r.PostFormValue("desk_id"))
	data.BootNum,_=strconv.Atoi(r.PostFormValue("boot_num"))
	data.PaveNum,_=strconv.Atoi(r.PostFormValue("pave_num"))
	var db *sql.DB = common.ConMysql()
	defer db.Close()
	code,msg:= model.BjlCancelBet(db,data)
	common.AjaxReturn(w,code,msg,nil)
}
//百家乐确定游戏结果
func BjlGameOver(w http.ResponseWriter, r *http.Request)  {
	w.Header().Set("Access-Control-Allow-Origin", "*") //允许跨域
	var data common.BJLPostData
	data.DeskId,_=strconv.Atoi(r.PostFormValue("desk_id"))
	data.BootNum,_=strconv.Atoi(r.PostFormValue("boot_num"))
	data.PaveNum,_=strconv.Atoi(r.PostFormValue("pave_num"))
	gameNum:=r.PostFormValue("game_num")//
	go BjlGameDone(data,gameNum)//游戏结算
	go BjlGameRecord(data,gameNum)//游戏结果入库
}
//百家乐游戏结算
func BjlGameDone(data common.BJLPostData,gameNum string)  {
	var db *sql.DB = common.ConMysql()
	defer db.Close()
	model.GetBjlUserIds(db,data,gameNum)
	return
}
//百家乐游戏记录入库
func BjlGameRecord(data common.BJLPostData,gameNum string) {
	var db *sql.DB = common.ConMysql()
	defer db.Close()
	model.BJLgameRecord(db,data,gameNum)
	return
}
//获取百家乐台桌限红
func BjlDeskLimit(DeskId int) map[string]int {
	var db *sql.DB = common.ConMysql()
	defer db.Close()
	limit:=model.GetBjlDeskLimit(db,DeskId)
	return limit
}